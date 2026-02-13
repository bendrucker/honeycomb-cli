package marker

import (
	"context"
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

	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	if settingType == "" || color == "" {
		existing, err := findMarkerSetting(ctx, client, dataset, settingID, auth)
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

	resp, err := client.UpdateMarkerSettingsWithResponse(ctx, api.DatasetSlugOrAll(dataset), settingID, body, auth)
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

func findMarkerSetting(ctx context.Context, client *api.ClientWithResponses, dataset, settingID string, auth api.RequestEditorFn) (*api.MarkerSetting, error) {
	resp, err := client.ListMarkerSettingsWithResponse(ctx, api.DatasetSlugOrAll(dataset), auth)
	if err != nil {
		return nil, fmt.Errorf("listing marker settings: %w", err)
	}
	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response: %s", resp.Status())
	}
	for _, s := range *resp.JSON200 {
		if s.Id != nil && *s.Id == settingID {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("marker setting %s not found", settingID)
}
