package marker

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
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
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	if !yes {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--yes is required in non-interactive mode")
		}

		// Fetch marker for confirmation display
		listResp, err := client.GetMarkerWithResponse(ctx, dataset, keyEditor(key))
		if err != nil {
			return fmt.Errorf("listing markers: %w", err)
		}

		if err := api.CheckResponse(listResp.StatusCode(), listResp.Body); err != nil {
			return err
		}

		if listResp.JSON200 == nil {
			return fmt.Errorf("unexpected response: %s", listResp.Status())
		}

		m, err := findMarker(*listResp.JSON200, markerID)
		if err != nil {
			return err
		}

		markerType := ""
		if m.Type != nil {
			markerType = *m.Type
		}

		answer, err := prompt.Choice(opts.IOStreams.Err, opts.IOStreams.In,
			fmt.Sprintf("Delete marker %q (type: %s)? (y/N): ", markerID, markerType),
			[]string{"y", "N"},
		)
		if err != nil {
			return err
		}
		if answer != "y" {
			return nil
		}
	}

	resp, err := client.DeleteMarkerWithResponse(ctx, dataset, markerID, keyEditor(key))
	if err != nil {
		return fmt.Errorf("deleting marker: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(markerID, fmt.Sprintf("Deleted marker %s", markerID))
}
