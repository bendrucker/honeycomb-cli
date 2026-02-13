package board

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewViewListCmd(opts *options.RootOptions, board *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List board views",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runViewList(cmd.Context(), opts, *board)
		},
	}
}

func runViewList(ctx context.Context, opts *options.RootOptions, boardID string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.ListBoardViewsWithResponse(ctx, boardID, auth)
	if err != nil {
		return fmt.Errorf("listing board views: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	items := make([]viewItem, len(*resp.JSON200))
	for i, v := range *resp.JSON200 {
		items[i] = viewResponseToItem(v)
	}

	return opts.OutputWriter().Write(items, viewListTable)
}
