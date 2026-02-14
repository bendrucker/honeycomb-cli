package marker

import (
	"bytes"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		markerType string
		message    string
		url        string
		startTime  int
		endTime    int
		color      string
	)

	cmd := &cobra.Command{
		Use:   "update <marker-id>",
		Short: "Update a marker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMarkerUpdate(cmd, opts, *dataset, args[0], markerType, message, url, startTime, endTime, color)
		},
	}

	cmd.Flags().StringVar(&markerType, "type", "", "Marker type")
	cmd.Flags().StringVar(&message, "message", "", "Marker message")
	cmd.Flags().StringVar(&url, "url", "", "URL associated with the marker")
	cmd.Flags().IntVar(&startTime, "start-time", 0, "Start time as Unix timestamp")
	cmd.Flags().IntVar(&endTime, "end-time", 0, "End time as Unix timestamp")
	cmd.Flags().StringVar(&color, "color", "", "Marker color")

	return cmd
}

func runMarkerUpdate(cmd *cobra.Command, opts *options.RootOptions, dataset, markerID, markerType, message, url string, startTime, endTime int, color string) error {
	ctx := cmd.Context()

	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	// Fetch existing marker (no individual GET — list and filter)
	listResp, err := client.GetMarkerWithResponse(ctx, dataset, auth)
	if err != nil {
		return fmt.Errorf("listing markers: %w", err)
	}

	if err := api.CheckResponse(listResp.StatusCode(), listResp.Body); err != nil {
		return err
	}

	if listResp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", listResp.Status())
	}

	existing, err := findMarker(*listResp.JSON200, markerID)
	if err != nil {
		return err
	}

	// Merge flag overrides into existing marker (API is full PUT)
	if cmd.Flags().Changed("type") {
		existing.Type = &markerType
	}
	if cmd.Flags().Changed("message") {
		existing.Message = &message
	}
	if cmd.Flags().Changed("url") {
		existing.Url = &url
	}
	if cmd.Flags().Changed("start-time") {
		existing.StartTime = &startTime
	}
	if cmd.Flags().Changed("end-time") {
		existing.EndTime = &endTime
	}
	if cmd.Flags().Changed("color") {
		existing.Color = &color
	}

	data, err := api.MarshalStrippingReadOnly(existing, "Marker")
	if err != nil {
		return fmt.Errorf("encoding marker: %w", err)
	}

	resp, err := client.UpdateMarkerWithBodyWithResponse(ctx, dataset, markerID, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("updating marker: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeDetail(opts, markerToItem(*resp.JSON200))
}
