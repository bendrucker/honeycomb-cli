package marker

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewGetCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:     "get <marker-id>",
		Aliases: []string{"view"},
		Short:   "Get a marker",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMarkerGet(cmd.Context(), opts, *dataset, args[0])
		},
	}
}

func runMarkerGet(ctx context.Context, opts *options.RootOptions, dataset, markerID string) error {
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

	m, err := findMarker(*resp.JSON200, markerID)
	if err != nil {
		return err
	}

	return writeDetail(opts, markerToItem(m))
}
