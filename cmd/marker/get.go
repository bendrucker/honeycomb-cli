package marker

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewGetCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <marker-id>",
		Short: "Get a marker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMarkerGet(cmd.Context(), opts, *dataset, args[0])
		},
	}
}

func runMarkerGet(ctx context.Context, opts *options.RootOptions, dataset, markerID string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	// No individual GET: list and filter.
	resp, err := client.GetMarkerWithResponse(ctx, dataset)
	if err != nil {
		return fmt.Errorf("listing markers: %w", err)
	}

	markers, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	marker, err := findMarker(*markers, markerID)
	if err != nil {
		return err
	}

	return writeDetail(opts, markerToItem(marker))
}
