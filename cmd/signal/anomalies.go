package signal

import (
	"context"
	"fmt"
	"time"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type anomalyItem struct {
	ID          string  `json:"id,omitempty"`
	StartedAt   int     `json:"started_at,omitempty"`
	EndedAt     int     `json:"ended_at,omitempty"`
	Measurement float32 `json:"measurement,omitempty"`
	LowerBound  float32 `json:"lower_bound"`
	UpperBound  float32 `json:"upper_bound"`
}

var anomalyListTable = output.TableDef{Columns: []output.Column{
	output.Col("ID", func(a anomalyItem) string { return a.ID }),
	output.Col("Started", func(a anomalyItem) string { return formatEpoch(a.StartedAt) }),
	output.Col("Ended", func(a anomalyItem) string { return formatEpoch(a.EndedAt) }),
	output.Col("Measurement", func(a anomalyItem) string { return fmt.Sprintf("%g", a.Measurement) }),
	output.Col("Normal Range", func(a anomalyItem) string {
		return fmt.Sprintf("%g - %g", a.LowerBound, a.UpperBound)
	}),
}}

func NewAnomaliesCmd(opts *options.RootOptions) *cobra.Command {
	var (
		startTime int
		endTime   int
	)

	cmd := &cobra.Command{
		Use:   "anomalies <signal-id>",
		Short: "List resolved anomalies for a signal",
		Long: "List resolved anomalies for a signal.\n\n" +
			"The time range is required and may span no more than 30 days.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnomalies(cmd.Context(), opts, args[0], startTime, endTime)
		},
	}

	cmd.Flags().IntVar(&startTime, "start-time", 0, "Start time as Unix timestamp (required)")
	cmd.Flags().IntVar(&endTime, "end-time", 0, "End time as Unix timestamp (required)")

	_ = cmd.MarkFlagRequired("start-time")
	_ = cmd.MarkFlagRequired("end-time")

	return cmd
}

func runAnomalies(ctx context.Context, opts *options.RootOptions, id string, startTime, endTime int) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	params := &api.ListSignalHistoricalAnomaliesParams{
		StartTime: startTime,
		EndTime:   endTime,
	}

	var items []anomalyItem
	for {
		resp, err := client.ListSignalHistoricalAnomaliesWithResponse(ctx, id, params)
		if err != nil {
			return fmt.Errorf("listing anomalies: %w", err)
		}

		page, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
		if err != nil {
			return err
		}

		for _, a := range page.HistoricalAnomalies {
			items = append(items, anomalyItem{
				ID:          deref.String(a.Id),
				StartedAt:   deref.Int(a.StartedAt),
				EndedAt:     deref.Int(a.EndedAt),
				Measurement: deref.Val(a.Measurement),
				LowerBound:  a.NormalRange.Lower,
				UpperBound:  a.NormalRange.Upper,
			})
		}

		cursor, err := nextPageCursor(page.Links)
		if err != nil {
			return err
		}
		if cursor == "" {
			break
		}
		params.PageAfter = &cursor
	}

	return opts.OutputWriterList().WriteList(items, anomalyListTable, "No anomalies found.")
}

func formatEpoch(seconds int) string {
	if seconds == 0 {
		return ""
	}
	return time.Unix(int64(seconds), 0).UTC().Format(time.RFC3339)
}
