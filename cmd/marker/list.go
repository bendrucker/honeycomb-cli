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
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	resp, err := client.GetMarkerWithResponse(ctx, dataset)
	if err != nil {
		return fmt.Errorf("listing markers: %w", err)
	}

	markers, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	items := make([]markerItem, len(*markers))
	for i, m := range *markers {
		items[i] = markerToItem(m)
	}

	return opts.OutputWriterList().WriteList(items, markerListTable, "No markers found.")
}
