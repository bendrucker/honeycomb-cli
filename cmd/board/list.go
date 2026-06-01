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

var boardListTable = output.TableDef{
	Columns: []output.Column{
		output.Col("ID", func(b boardListItem) string { return b.ID }),
		output.Col("Name", func(b boardListItem) string { return b.Name }),
		output.Col("Description", func(b boardListItem) string { return output.Truncate(b.Description, 40) }),
		output.Col("URL", func(b boardListItem) string { return b.URL }),
	},
}

func NewListCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List boards",
		Example: `  # List boards
  honeycomb board list

  # List boards as JSON
  honeycomb board list --format json`,
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
			ID:          deref.String(b.Id),
			Name:        b.Name,
			Description: deref.String(b.Description),
		}
		if b.Links != nil {
			item.URL = deref.String(b.Links.BoardUrl)
		}
		items[i] = item
	}

	return opts.OutputWriterList().WriteList(items, boardListTable, "No boards found.")
}
