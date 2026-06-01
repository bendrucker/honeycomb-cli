package board

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
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
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	resp, err := client.ListBoardsWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("listing boards: %w", err)
	}

	boards, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	items := make([]boardListItem, len(*boards))
	for i, b := range *boards {
		item := boardListItem{
			ID:           deref.String(b.Id),
			Name:         b.Name,
			Description:  deref.String(b.Description),
			ColumnLayout: deref.Enum(b.LayoutGeneration),
		}
		if b.Links != nil {
			item.URL = deref.String(b.Links.BoardUrl)
		}
		items[i] = item
	}

	return opts.OutputWriterList().Write(items, boardListTable)
}
