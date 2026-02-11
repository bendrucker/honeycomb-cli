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

func NewSettingDeleteCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a marker setting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSettingDelete(cmd.Context(), opts, *dataset, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runSettingDelete(ctx context.Context, opts *options.RootOptions, dataset, settingID string, yes bool) error {
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

		listResp, err := client.ListMarkerSettingsWithResponse(ctx, api.DatasetSlugOrAll(dataset), keyEditor(key))
		if err != nil {
			return fmt.Errorf("listing marker settings: %w", err)
		}

		if err := api.CheckResponse(listResp.StatusCode(), listResp.Body); err != nil {
			return err
		}

		if listResp.JSON200 == nil {
			return fmt.Errorf("unexpected response: %s", listResp.Status())
		}

		s, err := findSetting(*listResp.JSON200, settingID)
		if err != nil {
			return err
		}

		answer, err := prompt.Choice(opts.IOStreams.Err, opts.IOStreams.In,
			fmt.Sprintf("Delete marker setting %q (type: %s)? (y/N): ", settingID, s.Type),
			[]string{"y", "N"},
		)
		if err != nil {
			return err
		}
		if answer != "y" {
			return nil
		}
	}

	resp, err := client.DeleteMarkerSettingsWithResponse(ctx, api.DatasetSlugOrAll(dataset), settingID, keyEditor(key))
	if err != nil {
		return fmt.Errorf("deleting marker setting: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(opts.IOStreams.Err, "Deleted marker setting %s\n", settingID)
	return nil
}
