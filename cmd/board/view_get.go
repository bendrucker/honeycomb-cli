package board

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewViewGetCmd(opts *options.RootOptions, board *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <view-id>",
		Short: "Get a board view",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runViewGet(cmd.Context(), opts, *board, args[0])
		},
	}
}

func runViewGet(ctx context.Context, opts *options.RootOptions, boardID, viewID string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetBoardViewWithResponse(ctx, boardID, viewID, auth)
	if err != nil {
		return fmt.Errorf("getting board view: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeViewDetail(opts, viewResponseToDetail(*resp.JSON200))
}
