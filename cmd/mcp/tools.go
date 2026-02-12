package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var toolsTable = output.TableDef{
	Columns: []output.Column{
		{Header: "Name", Value: func(v any) string { return v.(mcp.Tool).Name }},
		{Header: "Description", Value: func(v any) string { return v.(mcp.Tool).Description }},
	},
}

func newToolsCmd(opts *options.RootOptions, factory clientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "tools",
		Short: "List available MCP tools",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTools(cmd.Context(), opts, factory)
		},
	}
}

func runTools(ctx context.Context, opts *options.RootOptions, factory clientFactory) error {
	if factory == nil {
		factory = defaultClientFactory
	}

	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	c, err := factory(ctx, opts.ResolveMCPUrl(), key)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()

	result, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("listing tools: %w", err)
	}

	return opts.OutputWriter().Write(result.Tools, toolsTable)
}
