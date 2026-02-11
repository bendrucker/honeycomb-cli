package marker

import (
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewSettingUpdateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		settingType string
		color       string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a marker setting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSettingUpdate(cmd, opts, *dataset, args[0], settingType, color)
		},
	}

	cmd.Flags().StringVar(&settingType, "type", "", "Marker setting type")
	cmd.Flags().StringVar(&color, "color", "", "Marker setting color (hex)")

	return cmd
}

func runSettingUpdate(cmd *cobra.Command, opts *options.RootOptions, dataset, settingID, settingType, color string) error {
	ctx := cmd.Context()

	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
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

	existing, err := findSetting(*listResp.JSON200, settingID)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("type") {
		existing.Type = settingType
	}
	if cmd.Flags().Changed("color") {
		existing.Color = color
	}

	resp, err := client.UpdateMarkerSettingsWithResponse(ctx, api.DatasetSlugOrAll(dataset), settingID, api.UpdateMarkerSettingsJSONRequestBody(existing), keyEditor(key))
	if err != nil {
		return fmt.Errorf("updating marker setting: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeSettingDetail(opts, toSettingItem(*resp.JSON200))
}
