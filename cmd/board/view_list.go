package board

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewViewListCmd(opts *options.RootOptions, board *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List board views",
		Example: `  # List views on a board
  honeycomb board view list --board abc123`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runViewList(cmd.Context(), opts, *board)
		},
	}
}

func runViewList(ctx context.Context, opts *options.RootOptions, boardID string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	resp, err := client.ListBoardViewsWithResponse(ctx, boardID)
	if err != nil {
		return fmt.Errorf("listing board views: %w", err)
	}

	views, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	items := make([]viewItem, len(*views))
	for i, v := range *views {
		items[i] = viewResponseToItem(v)
	}

	return opts.OutputWriterList().WriteList(items, viewListTable, "No board views found.")
}
