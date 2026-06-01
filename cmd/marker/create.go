package marker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file       string
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
			if file != "" {
				return runMarkerCreateFromFile(cmd.Context(), opts, *dataset, file)
			}

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

			if !opts.IOStreams.CanPrompt() {
				if markerType == "" {
					return fmt.Errorf("--type is required in non-interactive mode")
				}
				if message == "" {
					return fmt.Errorf("--message is required in non-interactive mode")
				}
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

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&markerType, "type", "", "Marker type (e.g., deploy)")
	cmd.Flags().StringVar(&message, "message", "", "Marker message")
	cmd.Flags().StringVar(&url, "url", "", "URL associated with the marker")
	cmd.Flags().IntVar(&startTime, "start-time", 0, "Start time as Unix timestamp (defaults to now)")
	cmd.Flags().IntVar(&endTime, "end-time", 0, "End time as Unix timestamp")
	cmd.Flags().StringVar(&color, "color", "", "Marker color")

	for _, name := range []string{"type", "message", "url", "start-time", "end-time", "color"} {
		cmd.MarkFlagsMutuallyExclusive("file", name)
	}

	return cmd
}

func runMarkerCreateFromFile(ctx context.Context, opts *options.RootOptions, dataset, file string) error {
	data, err := command.ReadDefinitionFile(opts.IOStreams, file)
	if err != nil {
		return err
	}

	var body api.CreateMarkerJSONRequestBody
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("parsing marker JSON: %w", err)
	}

	return runMarkerCreate(ctx, opts, dataset, body)
}

func runMarkerCreate(ctx context.Context, opts *options.RootOptions, dataset string, body api.CreateMarkerJSONRequestBody) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	resp, err := client.CreateMarkerWithResponse(ctx, dataset, body)
	if err != nil {
		return fmt.Errorf("creating marker: %w", err)
	}

	marker, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON201)
	if err != nil {
		return err
	}

	return writeDetail(opts, markerToItem(*marker))
}
