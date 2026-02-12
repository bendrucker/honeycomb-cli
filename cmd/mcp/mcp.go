package mcp

import (
	"context"
	"fmt"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/spf13/cobra"
)

type clientFactory func(ctx context.Context, mcpURL, key string) (*mcpclient.Client, error)

func defaultClientFactory(ctx context.Context, mcpURL, key string) (*mcpclient.Client, error) {
	c, err := mcpclient.NewStreamableHttpClient(mcpURL,
		transport.WithHTTPHeaders(map[string]string{
			"Authorization": "Bearer " + key,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("creating MCP client: %w", err)
	}

	if err := c.Start(ctx); err != nil {
		return nil, fmt.Errorf("starting MCP client: %w", err)
	}

	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "honeycomb-cli",
				Version: "0.1.0",
			},
		},
	})
	if err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("initializing MCP session: %w", err)
	}

	return c, nil
}

func NewCmd(opts *options.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Interact with the Honeycomb MCP server",
	}

	cmd.PersistentFlags().StringVar(&opts.MCPUrl, "mcp-url", "", "MCP server URL")

	cmd.AddCommand(newToolsCmd(opts, nil))
	cmd.AddCommand(newCallCmd(opts, nil))

	return cmd
}
