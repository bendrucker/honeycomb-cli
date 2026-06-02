package mcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/zalando/go-keyring"

	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
)

// clientName is the name registered with the MCP server during Dynamic Client
// Registration and reported in the MCP handshake.
const clientName = "honeycomb-cli"

// oauthScopes are the scopes the CLI requests from the Honeycomb MCP server.
var oauthScopes = []string{"mcp:read", "mcp:write"}

// keyringTokenStore implements transport.TokenStore backed by the OS keyring,
// persisting the OAuth token set under the {profile}:mcp entry. Tokens never
// touch disk in plaintext and are never logged.
type keyringTokenStore struct {
	store config.MCPStore
}

func newKeyringTokenStore(profile string) *keyringTokenStore {
	return &keyringTokenStore{store: config.NewMCPStore(profile)}
}

func (s *keyringTokenStore) GetToken(ctx context.Context) (*transport.Token, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var token transport.Token
	err := s.store.Token(&token)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, transport.ErrNoToken
	}
	// A stored-but-unparseable entry is unusable. Treat it as no token so the
	// library re-runs the authorization flow rather than hard-failing both
	// refresh and re-auth.
	if errors.Is(err, config.ErrMCPTokenCorrupt) {
		return nil, transport.ErrNoToken
	}
	if err != nil {
		return nil, fmt.Errorf("reading MCP token: %w", err)
	}
	return &token, nil
}

func (s *keyringTokenStore) SaveToken(ctx context.Context, token *transport.Token) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.store.SetToken(token); err != nil {
		return fmt.Errorf("saving MCP token: %w", err)
	}
	return nil
}

// oauthConfig builds the library OAuth configuration for a profile, wiring the
// keyring-backed token store, the requested scopes, a loopback redirect URI per
// RFC 8252, and PKCE (S256). The config key is never referenced here.
//
// A previously persisted DCR client ID is seeded so the refresh-token grant can
// send a stable client_id across CLI invocations. A missing or unreadable entry
// leaves the client ID empty, which triggers a fresh registration.
func oauthConfig(profile, redirectURI string) transport.OAuthConfig {
	clientID, _ := config.NewMCPStore(profile).ClientID()
	return transport.OAuthConfig{
		ClientID:    clientID,
		ClientURI:   "https://github.com/bendrucker/honeycomb-cli",
		RedirectURI: redirectURI,
		Scopes:      oauthScopes,
		TokenStore:  newKeyringTokenStore(profile),
		PKCEEnabled: true,
	}
}

// newOAuthClient constructs an OAuth-capable streamable HTTP MCP client for the
// given URL and profile.
func newOAuthClient(mcpURL, profile, redirectURI string) (*mcpclient.Client, error) {
	c, err := mcpclient.NewOAuthStreamableHttpClient(mcpURL, oauthConfig(profile, redirectURI))
	if err != nil {
		return nil, fmt.Errorf("creating MCP client: %w", err)
	}
	return c, nil
}

// authorizeInteractive runs the authorization-code flow (Dynamic Client
// Registration + PKCE S256) when the server reports that authorization is
// required. It starts a loopback callback server, opens the browser to the
// authorization URL, validates the returned state, and exchanges the code for
// tokens, which the handler persists through the keyring token store.
func authorizeInteractive(ctx context.Context, ios *iostreams.IOStreams, profile string, authErr error, listener net.Listener) error {
	handler := mcpclient.GetOAuthHandler(authErr)
	if handler == nil {
		return fmt.Errorf("MCP server requires authorization but no OAuth handler was provided")
	}

	codeVerifier, err := mcpclient.GenerateCodeVerifier()
	if err != nil {
		return fmt.Errorf("generating PKCE verifier: %w", err)
	}
	codeChallenge := mcpclient.GenerateCodeChallenge(codeVerifier)

	state, err := mcpclient.GenerateState()
	if err != nil {
		return fmt.Errorf("generating OAuth state: %w", err)
	}

	if handler.GetClientID() == "" {
		if err := handler.RegisterClient(ctx, clientName); err != nil {
			return fmt.Errorf("registering OAuth client: %w", err)
		}
		// Persist the registered client ID so a later invocation's refresh grant
		// sends a stable client_id instead of forcing a fresh browser flow. A
		// failure here is non-fatal: the in-memory handler still works for this
		// session.
		if id := handler.GetClientID(); id != "" {
			_ = config.NewMCPStore(profile).SetClientID(id)
		}
	}

	authURL, err := handler.GetAuthorizationURL(ctx, state, codeChallenge)
	if err != nil {
		return fmt.Errorf("building authorization URL: %w", err)
	}

	callbackChan := make(chan callbackParams, 1)
	server := &http.Server{
		Handler:           callbackHandler(callbackChan),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			callbackChan <- callbackParams{err: serveErr}
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	_, _ = fmt.Fprintf(ios.Err, "Opening browser to authorize Honeycomb MCP access.\n")
	_, _ = fmt.Fprintf(ios.Err, "If the browser does not open, visit:\n  %s\n", authURL)
	openBrowser(ios, authURL)

	var params callbackParams
	select {
	case params = <-callbackChan:
	case <-ctx.Done():
		return ctx.Err()
	}
	if params.err != nil {
		return fmt.Errorf("waiting for authorization callback: %w", params.err)
	}
	if params.errCode != "" {
		return fmt.Errorf("authorization denied: %s", params.errCode)
	}
	if params.state != state {
		return fmt.Errorf("authorization state mismatch (possible CSRF); aborting")
	}
	if params.code == "" {
		return fmt.Errorf("authorization callback returned no code")
	}

	if err := handler.ProcessAuthorizationResponse(ctx, params.code, params.state, codeVerifier); err != nil {
		return fmt.Errorf("exchanging authorization code: %w", err)
	}

	_, _ = fmt.Fprintln(ios.Err, "Authorization successful.")
	return nil
}

type callbackParams struct {
	code    string
	state   string
	errCode string
	err     error
}

const callbackPath = "/callback"

const successPage = `<!doctype html><html><body>` +
	`<h1>Authorization complete</h1>` +
	`<p>You can close this window and return to the terminal.</p>` +
	`</body></html>`

const failurePage = `<!doctype html><html><body>` +
	`<h1>Authorization failed</h1>` +
	`<p>Return to the terminal.</p>` +
	`</body></html>`

// callbackHandler serves the loopback OAuth redirect. It handles only GET
// requests to the redirect path that carry an authorization result (a code or
// an error). Browser probes such as /favicon.ico, or bare hits to the redirect
// path with neither parameter, get a 404/204 and are never delivered to the
// channel, so exactly one meaningful result reaches the waiting flow without
// leaking a blocked-send goroutine. The Go flow remains authoritative for state
// validation; the rendered page only reflects whether the callback looks valid.
func callbackHandler(ch chan<- callbackParams) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		q := r.URL.Query()
		code := q.Get("code")
		errCode := q.Get("error")
		state := q.Get("state")

		// A request carrying neither a code nor an error is not an
		// authorization response (e.g. a browser probe). Do not consume the
		// channel slot for it.
		if code == "" && errCode == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if errCode == "" && code != "" {
			_, _ = w.Write([]byte(successPage))
		} else {
			_, _ = w.Write([]byte(failurePage))
		}

		ch <- callbackParams{code: code, state: state, errCode: errCode}
	})
	return mux
}

// newLoopbackListener binds a callback listener to a loopback address per
// RFC 8252, letting the OS choose the port. The returned redirect URI uses
// 127.0.0.1 with that port.
func newLoopbackListener() (net.Listener, string, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", fmt.Errorf("starting loopback callback server: %w", err)
	}
	addr := l.Addr().(*net.TCPAddr)
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d%s", addr.Port, callbackPath)
	return l, redirectURI, nil
}

func openBrowser(ios *iostreams.IOStreams, url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		_, _ = fmt.Fprintf(ios.Err, "Could not open browser automatically: %v\n", err)
	}
}
