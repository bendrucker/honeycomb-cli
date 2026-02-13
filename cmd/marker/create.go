package marker

import (
	"context"
	"fmt"
	"time"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		markerType string
		message    string
		url        string
		startTime  int
		endTime    int
		color      string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a marker",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if markerType == "" && opts.IOStreams.CanPrompt() {
				v, err := prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Type: ")
				if err != nil {
					return err
				}
				markerType = v
			}

			if message == "" && opts.IOStreams.CanPrompt() {
				v, err := prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Message: ")
				if err != nil {
					return err
				}
				message = v
			}

			if !cmd.Flags().Changed("start-time") {
				startTime = int(time.Now().Unix())
			}

			body := api.CreateMarkerJSONRequestBody{
				Type:      &markerType,
				Message:   &message,
				StartTime: &startTime,
			}
			if url != "" {
				body.Url = &url
			}
			if cmd.Flags().Changed("end-time") {
				body.EndTime = &endTime
			}
			if color != "" {
				body.Color = &color
			}

			return runMarkerCreate(cmd.Context(), opts, *dataset, body)
		},
	}

	cmd.Flags().StringVar(&markerType, "type", "", "Marker type (e.g., deploy)")
	cmd.Flags().StringVar(&message, "message", "", "Marker message")
	cmd.Flags().StringVar(&url, "url", "", "URL associated with the marker")
	cmd.Flags().IntVar(&startTime, "start-time", 0, "Start time as Unix timestamp (defaults to now)")
	cmd.Flags().IntVar(&endTime, "end-time", 0, "End time as Unix timestamp")
	cmd.Flags().StringVar(&color, "color", "", "Marker color")

	return cmd
}

func runMarkerCreate(ctx context.Context, opts *options.RootOptions, dataset string, body api.CreateMarkerJSONRequestBody) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.CreateMarkerWithResponse(ctx, dataset, body, auth)
	if err != nil {
		return fmt.Errorf("creating marker: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeDetail(opts, markerToItem(*resp.JSON201))
}
