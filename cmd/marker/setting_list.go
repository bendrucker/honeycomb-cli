package marker

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewSettingListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List marker settings",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSettingList(cmd.Context(), opts, *dataset)
		},
	}
}

func runSettingList(ctx context.Context, opts *options.RootOptions, dataset string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	resp, err := client.ListMarkerSettingsWithResponse(ctx, api.DatasetSlugOrAll(dataset))
	if err != nil {
		return fmt.Errorf("listing marker settings: %w", err)
	}

	settings, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	items := make([]settingItem, len(*settings))
	for i, s := range *settings {
		items[i] = toSettingItem(s)
	}

	return opts.OutputWriterList().WriteList(items, settingListTable, "No marker settings found.")
}
