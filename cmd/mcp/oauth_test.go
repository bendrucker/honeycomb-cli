package mcp

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
)

func TestKeyringTokenStore(t *testing.T) {
	const profile = "tokenstore-test"
	t.Cleanup(func() { _ = config.NewMCPStore(profile).DeleteToken() })

	store := newKeyringTokenStore(profile)
	ctx := context.Background()

	if _, err := store.GetToken(ctx); !errors.Is(err, transport.ErrNoToken) {
		t.Fatalf("GetToken on empty store = %v, want ErrNoToken", err)
	}

	want := &transport.Token{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(time.Hour).UTC().Truncate(time.Second),
	}
	if err := store.SaveToken(ctx, want); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	got, err := store.GetToken(ctx)
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if got.AccessToken != want.AccessToken {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, want.AccessToken)
	}
	if got.RefreshToken != want.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, want.RefreshToken)
	}
	if !got.ExpiresAt.Equal(want.ExpiresAt) {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, want.ExpiresAt)
	}
}

func TestKeyringTokenStore_ContextCancelled(t *testing.T) {
	store := newKeyringTokenStore("ctx-test")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := store.GetToken(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("GetToken = %v, want context.Canceled", err)
	}
	if err := store.SaveToken(ctx, &transport.Token{}); !errors.Is(err, context.Canceled) {
		t.Errorf("SaveToken = %v, want context.Canceled", err)
	}
}

func TestResolveToken(t *testing.T) {
	for _, tc := range []struct {
		name string
		flag string
		env  string
		want string
	}{
		{name: "flag wins", flag: "flag-tok", env: "env-tok", want: "flag-tok"},
		{name: "env fallback", flag: "", env: "env-tok", want: "env-tok"},
		{name: "neither", flag: "", env: "", want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(tokenEnvVar, tc.env)
			if got := resolveToken(tc.flag); got != tc.want {
				t.Errorf("resolveToken(%q) = %q, want %q", tc.flag, got, tc.want)
			}
		})
	}
}

func TestRemediateAuthError(t *testing.T) {
	for _, tc := range []struct {
		name        string
		err         error
		wantRemedy  bool
		wantSameErr bool
	}{
		{name: "nil", err: nil, wantRemedy: false},
		{name: "status 401", err: errors.New("request failed with status 401"), wantRemedy: true},
		{name: "status 403", err: errors.New("request failed with status 403"), wantRemedy: true},
		{name: "unauthorized text", err: errors.New("Unauthorized"), wantRemedy: true},
		{name: "forbidden text", err: errors.New("Forbidden"), wantRemedy: true},
		{name: "benign 401 digits", err: errors.New("read 401 bytes from response"), wantRemedy: false, wantSameErr: true},
		{name: "benign 403 digits", err: errors.New("payload was 403 bytes"), wantRemedy: false, wantSameErr: true},
		{name: "unrelated", err: errors.New("connection refused"), wantRemedy: false, wantSameErr: true},
		{name: "non-interactive passthrough", err: errNonInteractiveAuth, wantRemedy: false, wantSameErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := remediateAuthError(tc.err)
			if tc.err == nil {
				if got != nil {
					t.Fatalf("remediateAuthError(nil) = %v, want nil", got)
				}
				return
			}
			if !errors.Is(got, tc.err) {
				t.Errorf("remediated error does not wrap original: %v", got)
			}
			hasRemedy := strings.Contains(got.Error(), tokenEnvVar) &&
				strings.Contains(got.Error(), "mcp --help")
			if hasRemedy != tc.wantRemedy {
				t.Errorf("remedy present = %v, want %v (%q)", hasRemedy, tc.wantRemedy, got.Error())
			}
			if tc.wantSameErr && got.Error() != tc.err.Error() {
				t.Errorf("error was modified: %q", got.Error())
			}
		})
	}
}

func TestNewLoopbackListener(t *testing.T) {
	l, redirectURI, err := newLoopbackListener()
	if err != nil {
		t.Fatalf("newLoopbackListener: %v", err)
	}
	defer func() { _ = l.Close() }()

	if host := l.Addr().(*net.TCPAddr).IP.String(); host != "127.0.0.1" {
		t.Errorf("listener bound to %q, want 127.0.0.1", host)
	}
	if !strings.HasPrefix(redirectURI, "http://127.0.0.1:") {
		t.Errorf("redirect URI = %q, want loopback", redirectURI)
	}
	if !strings.HasSuffix(redirectURI, "/callback") {
		t.Errorf("redirect URI = %q, want /callback path", redirectURI)
	}
}

func TestCallbackHandler(t *testing.T) {
	for _, tc := range []struct {
		name       string
		path       string
		wantStatus int
		wantSend   bool
		wantParams callbackParams
		wantBody   string
	}{
		{
			name:       "valid code renders success and delivers",
			path:       "/callback?code=abc&state=xyz",
			wantStatus: http.StatusOK,
			wantSend:   true,
			wantParams: callbackParams{code: "abc", state: "xyz"},
			wantBody:   "Authorization complete",
		},
		{
			name:       "error param renders failure and delivers",
			path:       "/callback?error=access_denied&state=xyz",
			wantStatus: http.StatusOK,
			wantSend:   true,
			wantParams: callbackParams{errCode: "access_denied", state: "xyz"},
			wantBody:   "Authorization failed",
		},
		{
			name:       "favicon probe is ignored",
			path:       "/favicon.ico",
			wantStatus: http.StatusNotFound,
			wantSend:   false,
		},
		{
			name:       "bare callback without params is ignored",
			path:       "/callback",
			wantStatus: http.StatusNoContent,
			wantSend:   false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ch := make(chan callbackParams, 1)
			srv := httptestServer(t, callbackHandler(ch))

			resp, err := http.Get(srv + tc.path)
			if err != nil {
				t.Fatalf("GET %s: %v", tc.path, err)
			}
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}
			if tc.wantBody != "" && !strings.Contains(string(body), tc.wantBody) {
				t.Errorf("body = %q, want it to contain %q", string(body), tc.wantBody)
			}

			select {
			case params := <-ch:
				if !tc.wantSend {
					t.Fatalf("unexpected channel delivery: %+v", params)
				}
				if params != tc.wantParams {
					t.Errorf("params = %+v, want %+v", params, tc.wantParams)
				}
			case <-time.After(200 * time.Millisecond):
				if tc.wantSend {
					t.Fatal("callback params not received")
				}
			}
		})
	}
}

func TestKeyringTokenStore_CorruptJSON(t *testing.T) {
	const profile = "corrupt-token-test"
	t.Cleanup(func() { _ = config.NewMCPStore(profile).DeleteToken() })

	// Store an entry whose JSON is valid but cannot decode into a Token: a
	// non-RFC3339 string for the time-typed expires_at field.
	if err := config.NewMCPStore(profile).SetToken(map[string]any{"expires_at": "not-a-time"}); err != nil {
		t.Fatalf("SetToken: %v", err)
	}

	store := newKeyringTokenStore(profile)
	if _, err := store.GetToken(context.Background()); !errors.Is(err, transport.ErrNoToken) {
		t.Fatalf("GetToken on corrupt entry = %v, want ErrNoToken", err)
	}
}

// httptestServer starts an HTTP server on a loopback listener serving handler
// and returns its base URL, shutting it down at test end.
func httptestServer(t *testing.T, handler http.Handler) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: time.Second}
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	})
	return "http://" + l.Addr().String()
}

func TestNonInteractiveAuthError(t *testing.T) {
	if !strings.Contains(errNonInteractiveAuth.Error(), tokenEnvVar) {
		t.Errorf("non-interactive error should mention %s: %q", tokenEnvVar, errNonInteractiveAuth.Error())
	}
}

// TestConnectNonInteractiveOAuthRequired drives the real defaultClientFactory
// against a stub MCP endpoint that always returns 401 with an OAuth
// WWW-Authenticate challenge. With no TTY and no token, the connect path must
// surface the actionable non-interactive auth error rather than retrying a
// browser flow.
func TestConnectNonInteractiveOAuthRequired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	ts := iostreams.Test(t) // non-TTY: CanPrompt() == false
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		MCPUrl:    srv.URL,
	}

	_, err := connect(context.Background(), opts, "", nil)
	if err == nil {
		t.Fatal("expected an auth error")
	}
	if !errors.Is(err, errNonInteractiveAuth) {
		t.Fatalf("error = %v, want errNonInteractiveAuth", err)
	}
}

func TestConnectWithToken(t *testing.T) {
	// connectWithToken builds a real streamable client that fails fast against a
	// dead address; this asserts the headless path constructs and starts without
	// touching the keyring or config key.
	conn := connectOptions{mcpURL: "http://127.0.0.1:1/mcp", token: "headless-token"}
	_, err := connectWithToken(context.Background(), conn)
	if err == nil {
		t.Fatal("expected connection error against dead address")
	}
	// The error must come from the transport, not from a missing credential.
	if mcpclient.IsOAuthAuthorizationRequiredError(err) {
		t.Errorf("headless token path should not require OAuth: %v", err)
	}
}
