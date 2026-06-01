package board

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewGetCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <board-id>",
		Short: "Get a board",
		Example: `  # Get a board by ID
  honeycomb board get abc123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBoardGet(cmd.Context(), opts, args[0])
		},
	}
}

func runBoardGet(ctx context.Context, opts *options.RootOptions, boardID string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	resp, err := client.GetBoardWithResponse(ctx, boardID)
	if err != nil {
		return fmt.Errorf("getting board: %w", err)
	}

	board, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeBoardDetail(opts, boardToDetail(*board))
}
