package mcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/spf13/cobra"
)

// tokenEnvVar holds a pre-issued MCP access token for headless (CI) use,
// bypassing the interactive browser flow.
const tokenEnvVar = "HONEYCOMB_MCP_TOKEN"

// connectOptions carries everything the client factory needs to authenticate
// and open an MCP session. This is the OAuth path that options.AuthMCPOAuth
// names: it deliberately excludes the Honeycomb config API key, which the
// OAuth-protected MCP server does not accept, and reaches the server through the
// OAuth transport below rather than opts.Client / opts.ClientFor.
type connectOptions struct {
	root   *options.RootOptions
	mcpURL string
	// token, when set, is a pre-issued bearer token used for headless auth.
	token string
}

// clientFactory opens an initialized MCP client session.
type clientFactory func(ctx context.Context, conn connectOptions) (*mcpclient.Client, error)

var experimentalOnce sync.Once

// warnExperimental prints a one-time stderr notice that the command is
// experimental. It fires once per process so repeated calls in tests or
// scripts do not spam the notice.
func warnExperimental(opts *options.RootOptions) {
	experimentalOnce.Do(func() {
		_, _ = fmt.Fprintln(opts.IOStreams.Err, "Warning: 'honeycomb mcp' is experimental and may change.")
	})
}

func newInitializeRequest() mcp.InitializeRequest {
	return mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    clientName,
				Version: "0.1.0",
			},
		},
	}
}

// defaultClientFactory connects to the MCP server using OAuth. With a headless
// token it sends a static bearer; otherwise it relies on the keyring token
// store and, when the server reports authorization is required, runs the
// interactive authorization-code flow once and retries.
func defaultClientFactory(ctx context.Context, conn connectOptions) (*mcpclient.Client, error) {
	if conn.token != "" {
		return connectWithToken(ctx, conn)
	}
	return connectWithOAuth(ctx, conn)
}

func connectWithToken(ctx context.Context, conn connectOptions) (*mcpclient.Client, error) {
	c, err := mcpclient.NewStreamableHttpClient(conn.mcpURL,
		transport.WithHTTPHeaders(map[string]string{
			"Authorization": "Bearer " + conn.token,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("creating MCP client: %w", err)
	}
	return startAndInitialize(ctx, c)
}

func connectWithOAuth(ctx context.Context, conn connectOptions) (*mcpclient.Client, error) {
	listener, redirectURI, err := newLoopbackListener()
	if err != nil {
		return nil, err
	}
	defer func() { _ = listener.Close() }()

	c, err := newOAuthClient(conn.mcpURL, conn.root.ActiveProfile(), redirectURI)
	if err != nil {
		return nil, err
	}

	session, err := startAndInitialize(ctx, c)
	if err == nil {
		return session, nil
	}

	if !mcpclient.IsOAuthAuthorizationRequiredError(err) {
		return nil, err
	}

	if !conn.root.IOStreams.CanPrompt() {
		_ = c.Close()
		return nil, errNonInteractiveAuth
	}

	if authErr := authorizeInteractive(ctx, conn.root.IOStreams, conn.root.ActiveProfile(), err, listener); authErr != nil {
		_ = c.Close()
		return nil, authErr
	}

	return startAndInitialize(ctx, c)
}

func startAndInitialize(ctx context.Context, c *mcpclient.Client) (*mcpclient.Client, error) {
	if err := c.Start(ctx); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("starting MCP client: %w", err)
	}

	if _, err := c.Initialize(ctx, newInitializeRequest()); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("initializing MCP session: %w", err)
	}

	return c, nil
}

var errNonInteractiveAuth = errors.New(
	"MCP authorization required but session is non-interactive; " +
		"set " + tokenEnvVar + " (or pass --token) with a pre-issued MCP access token to authenticate without a browser",
)

const mcpLongDescription = `Interact with the Honeycomb MCP server.

EXPERIMENTAL: this command and its authentication flow may change.

Authentication:
  The Honeycomb MCP server is an OAuth 2.1 protected resource. The CLI does NOT
  use your Honeycomb config API key here. Instead it runs an OAuth
  authorization-code flow (Dynamic Client Registration + PKCE) the first time
  you connect, opening a browser to authorize access. Access and refresh tokens
  are stored in your OS keyring under the active profile and refreshed
  automatically.

  For non-interactive use (CI), set HONEYCOMB_MCP_TOKEN (or pass --token) with a
  pre-issued MCP access token to skip the browser flow.

  Run 'honeycomb auth logout' to clear stored MCP tokens along with API keys.`

// derefToken returns the value behind a token flag pointer, or empty when nil.
func derefToken(token *string) string {
	if token == nil {
		return ""
	}
	return *token
}

// resolveToken returns the headless token from the --token flag, falling back
// to the HONEYCOMB_MCP_TOKEN environment variable.
func resolveToken(flagToken string) string {
	if flagToken != "" {
		return flagToken
	}
	return os.Getenv(tokenEnvVar)
}

// connect opens an MCP session for a subcommand, applying the experimental
// warning, the token fallback, and the configured factory.
func connect(ctx context.Context, opts *options.RootOptions, flagToken string, factory clientFactory) (*mcpclient.Client, error) {
	warnExperimental(opts)

	if factory == nil {
		factory = defaultClientFactory
	}

	conn := connectOptions{
		root:   opts,
		mcpURL: opts.ResolveMCPUrl(),
		token:  resolveToken(flagToken),
	}

	c, err := factory(ctx, conn)
	if err != nil {
		return nil, remediateAuthError(err)
	}
	return c, nil
}

// remediateAuthError appends actionable guidance when an error indicates the
// MCP server rejected the request for authentication or authorization reasons.
func remediateAuthError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, errNonInteractiveAuth) {
		return err
	}

	if isAuthRejection(err) {
		return fmt.Errorf("%w (the MCP server rejected the request as unauthenticated; "+
			"the CLI authenticates to MCP via OAuth, not your API key: re-run interactively to "+
			"authorize in a browser, set %s for CI, or run 'honeycomb auth logout' to clear stale "+
			"tokens; see 'honeycomb mcp --help')", err, tokenEnvVar)
	}
	return err
}

// isAuthRejection reports whether err represents the MCP server rejecting the
// request for authentication or authorization reasons. The typed library
// errors are the primary signal; the string fallback is tightened to specific
// phrases so an unrelated error whose text merely contains "401"/"403" digits
// (e.g. a byte count) does not misfire.
func isAuthRejection(err error) bool {
	if mcpclient.IsOAuthAuthorizationRequiredError(err) ||
		mcpclient.IsAuthorizationRequiredError(err) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unauthorized") ||
		strings.Contains(msg, "forbidden") ||
		strings.Contains(msg, "status 401") ||
		strings.Contains(msg, "status 403")
}

func NewCmd(opts *options.RootOptions) *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Interact with the Honeycomb MCP server (experimental)",
		Long:  mcpLongDescription,
	}

	cmd.PersistentFlags().StringVar(&opts.MCPUrl, "mcp-url", "", "MCP server URL")
	cmd.PersistentFlags().StringVar(&token, "token", "", "Pre-issued MCP access token (or set "+tokenEnvVar+")")

	cmd.AddCommand(newToolsCmd(opts, &token, nil))
	cmd.AddCommand(newCallCmd(opts, &token, nil))

	return command.Group(cmd)
}
