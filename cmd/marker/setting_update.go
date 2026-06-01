package marker

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
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

	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	if settingType == "" || color == "" {
		existing, err := findMarkerSetting(ctx, client, dataset, settingID)
		if err != nil {
			return err
		}
		if settingType == "" {
			settingType = existing.Type
		}
		if color == "" {
			color = existing.Color
		}
	}

	body := api.UpdateMarkerSettingsJSONRequestBody{
		Type:  settingType,
		Color: color,
	}

	resp, err := client.UpdateMarkerSettingsWithResponse(ctx, api.DatasetSlugOrAll(dataset), settingID, body)
	if err != nil {
		return fmt.Errorf("updating marker setting: %w", err)
	}

	setting, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeSettingDetail(opts, toSettingItem(*setting))
}

func findMarkerSetting(ctx context.Context, client *api.ClientWithResponses, dataset, settingID string) (*api.MarkerSetting, error) {
	resp, err := client.ListMarkerSettingsWithResponse(ctx, api.DatasetSlugOrAll(dataset))
	if err != nil {
		return nil, fmt.Errorf("listing marker settings: %w", err)
	}
	settings, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return nil, err
	}
	for _, s := range *settings {
		if s.Id != nil && *s.Id == settingID {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("marker setting %s not found", settingID)
}
