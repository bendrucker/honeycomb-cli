package marker

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List markers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMarkerList(cmd.Context(), opts, *dataset)
		},
	}
}

func runMarkerList(ctx context.Context, opts *options.RootOptions, dataset string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetMarkerWithResponse(ctx, dataset, keyEditor(key))
	if err != nil {
		return fmt.Errorf("listing markers: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	items := make([]markerItem, len(*resp.JSON200))
	for i, m := range *resp.JSON200 {
		items[i] = markerToItem(m)
	}

	return opts.OutputWriter().Write(items, markerListTable)
}
