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
	profile string
}

func newKeyringTokenStore(profile string) *keyringTokenStore {
	return &keyringTokenStore{profile: profile}
}

func (s *keyringTokenStore) GetToken(ctx context.Context) (*transport.Token, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var token transport.Token
	err := config.GetMCPToken(s.profile, &token)
	if errors.Is(err, keyring.ErrNotFound) {
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
	if err := config.SetMCPToken(s.profile, token); err != nil {
		return fmt.Errorf("saving MCP token: %w", err)
	}
	return nil
}

// oauthConfig builds the library OAuth configuration for a profile, wiring the
// keyring-backed token store, the requested scopes, a loopback redirect URI per
// RFC 8252, and PKCE (S256). The config key is never referenced here.
func oauthConfig(profile, redirectURI string) transport.OAuthConfig {
	return transport.OAuthConfig{
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
func authorizeInteractive(ctx context.Context, ios *iostreams.IOStreams, authErr error, listener net.Listener) error {
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

func callbackHandler(ch chan<- callbackParams) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		params := callbackParams{
			code:    q.Get("code"),
			state:   q.Get("state"),
			errCode: q.Get("error"),
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!doctype html><html><body>` +
			`<h1>Authorization complete</h1>` +
			`<p>You can close this window and return to the terminal.</p>` +
			`</body></html>`))

		ch <- params
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
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", addr.Port)
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
