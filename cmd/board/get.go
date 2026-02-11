package board

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewGetCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "get <board-id>",
		Aliases: []string{"view"},
		Short:   "Get a board",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBoardGet(cmd.Context(), opts, args[0])
		},
	}
}

func runBoardGet(ctx context.Context, opts *options.RootOptions, boardID string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetBoardWithResponse(ctx, boardID, keyEditor(key))
	if err != nil {
		return fmt.Errorf("getting board: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeBoardDetail(opts, boardToDetail(*resp.JSON200))
}
