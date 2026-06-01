package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var toolsTable = output.TableDef{
	Columns: []output.Column{
		{Header: "Name", Value: func(v any) string { return v.(mcp.Tool).Name }},
		{Header: "Description", Value: func(v any) string { return v.(mcp.Tool).Description }},
	},
}

func newToolsCmd(opts *options.RootOptions, token *string, factory clientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "tools",
		Short: "List available MCP tools",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTools(cmd.Context(), opts, derefToken(token), factory)
		},
	}
}

func runTools(ctx context.Context, opts *options.RootOptions, token string, factory clientFactory) error {
	c, err := connect(ctx, opts, token, factory)
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
