package board

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

var boardListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(boardListItem).ID }},
		{Header: "Name", Value: func(v any) string { return v.(boardListItem).Name }},
		{Header: "Description", Value: func(v any) string { return truncate(v.(boardListItem).Description, 40) }},
		{Header: "Column Layout", Value: func(v any) string { return v.(boardListItem).ColumnLayout }},
		{Header: "URL", Value: func(v any) string { return v.(boardListItem).URL }},
	},
}

func NewListCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List boards",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runBoardList(cmd.Context(), opts)
		},
	}
}

func runBoardList(ctx context.Context, opts *options.RootOptions) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.ListBoardsWithResponse(ctx, keyEditor(key))
	if err != nil {
		return fmt.Errorf("listing boards: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	items := make([]boardListItem, len(*resp.JSON200))
	for i, b := range *resp.JSON200 {
		item := boardListItem{
			Name: b.Name,
		}
		if b.Id != nil {
			item.ID = *b.Id
		}
		if b.Description != nil {
			item.Description = *b.Description
		}
		if b.LayoutGeneration != nil {
			item.ColumnLayout = string(*b.LayoutGeneration)
		}
		if b.Links != nil && b.Links.BoardUrl != nil {
			item.URL = *b.Links.BoardUrl
		}
		items[i] = item
	}

	return opts.OutputWriter().Write(items, boardListTable)
}
