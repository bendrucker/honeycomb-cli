package marker

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <marker-id>",
		Short: "Delete a marker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMarkerDelete(cmd.Context(), opts, *dataset, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runMarkerDelete(ctx context.Context, opts *options.RootOptions, dataset, markerID string, yes bool) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	proceed, err := command.ConfirmDelete(opts.IOStreams, yes, "marker", markerID, func() (string, error) {
		listResp, err := client.GetMarkerWithResponse(ctx, dataset)
		if err != nil {
			return "", fmt.Errorf("listing markers: %w", err)
		}

		if err := api.CheckResponse(listResp.StatusCode(), listResp.Body); err != nil {
			return "", err
		}

		if listResp.JSON200 == nil {
			return "", fmt.Errorf("unexpected response: %s", listResp.Status())
		}

		if _, err := findMarker(*listResp.JSON200, markerID); err != nil {
			return "", err
		}

		return "", nil
	})
	if err != nil {
		return err
	}
	if !proceed {
		return nil
	}

	resp, err := client.DeleteMarkerWithResponse(ctx, dataset, markerID)
	if err != nil {
		return fmt.Errorf("deleting marker: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(markerID, fmt.Sprintf("Deleted marker %s", markerID))
}
