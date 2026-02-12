package slo

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewHistoryCmd(opts *options.RootOptions, _ *string) *cobra.Command {
	var (
		sloIDs    []string
		startTime int
		endTime   int
	)

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Get SLO historical compliance and budget data",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runHistory(cmd.Context(), opts, sloIDs, startTime, endTime)
		},
	}

	cmd.Flags().StringSliceVar(&sloIDs, "slo-id", nil, "SLO IDs to retrieve history for (required, repeatable)")
	cmd.Flags().IntVar(&startTime, "start-time", 0, "Start time as Unix timestamp (required)")
	cmd.Flags().IntVar(&endTime, "end-time", 0, "End time as Unix timestamp (required)")

	_ = cmd.MarkFlagRequired("slo-id")
	_ = cmd.MarkFlagRequired("start-time")
	_ = cmd.MarkFlagRequired("end-time")

	return cmd
}

func runHistory(ctx context.Context, opts *options.RootOptions, sloIDs []string, startTime, endTime int) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	ids := make([]interface{}, len(sloIDs))
	for i, id := range sloIDs {
		ids[i] = id
	}

	body := api.SLOHistoryRequest{
		Ids:       ids,
		StartTime: startTime,
		EndTime:   endTime,
	}

	resp, err := client.GetSloHistoryWithResponse(ctx, body, keyEditor(key))
	if err != nil {
		return fmt.Errorf("getting SLO history: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	data := *resp.JSON200

	return opts.OutputWriter().WriteDynamic(data, historyTable(data))
}

func historyTable(data api.SLOHistoryResponse) output.DynamicTableDef {
	headers := []string{"SLO ID", "Timestamp", "Compliance", "Budget Remaining"}
	var rows [][]string

	for sloID, entries := range data {
		for _, entry := range entries {
			ts := ""
			if entry.Timestamp != nil {
				ts = strconv.Itoa(*entry.Timestamp)
			}
			compliance := ""
			if entry.Compliance != nil {
				compliance = fmt.Sprintf("%g%%", *entry.Compliance)
			}
			budget := ""
			if entry.BudgetRemaining != nil {
				budget = fmt.Sprintf("%g", *entry.BudgetRemaining)
			}
			rows = append(rows, []string{sloID, ts, compliance, budget})
		}
	}

	return output.DynamicTableDef{
		Headers: headers,
		Rows:    rows,
	}
}
