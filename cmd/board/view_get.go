package board

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewViewGetCmd(opts *options.RootOptions, board *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <view-id>",
		Short: "Get a board view",
		Example: `  # Get a board view by ID
  honeycomb board view get view123 --board abc123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runViewGet(cmd.Context(), opts, *board, args[0])
		},
	}
}

func runViewGet(ctx context.Context, opts *options.RootOptions, boardID, viewID string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	resp, err := client.GetBoardViewWithResponse(ctx, boardID, viewID)
	if err != nil {
		return fmt.Errorf("getting board view: %w", err)
	}

	view, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeViewDetail(opts, viewResponseToDetail(*view))
}
