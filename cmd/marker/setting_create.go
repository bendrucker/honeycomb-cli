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

func NewSettingCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		settingType string
		color       string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a marker setting",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if settingType == "" && opts.IOStreams.CanPrompt() {
				v, err := prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Type: ")
				if err != nil {
					return err
				}
				settingType = v
			}

			if color == "" && opts.IOStreams.CanPrompt() {
				v, err := prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Color: ")
				if err != nil {
					return err
				}
				color = v
			}

			body := api.MarkerSetting{
				Type:  settingType,
				Color: color,
			}

			return runSettingCreate(cmd.Context(), opts, *dataset, body)
		},
	}

	cmd.Flags().StringVar(&settingType, "type", "", "Marker setting type (e.g., deploys)")
	cmd.Flags().StringVar(&color, "color", "", "Marker setting color (hex, e.g., #F96E11)")

	return cmd
}

func runSettingCreate(ctx context.Context, opts *options.RootOptions, dataset string, body api.CreateMarkerSettingJSONRequestBody) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.CreateMarkerSettingWithResponse(ctx, api.DatasetSlugOrAll(dataset), body, auth)
	if err != nil {
		return fmt.Errorf("creating marker setting: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeSettingDetail(opts, toSettingItem(*resp.JSON201))
}
