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
		Use:   "history [slo-id...]",
		Short: "Get SLO historical compliance and budget data",
		Long: "Get SLO historical compliance and budget data.\n\n" +
			"SLO IDs may be passed as positional arguments, via --slo-id (repeatable), or both. " +
			"At least one SLO ID is required; the two sources are merged and de-duplicated.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ids := mergeSLOIDs(args, sloIDs)
			if len(ids) == 0 {
				return fmt.Errorf("at least one SLO ID is required (positional argument or --slo-id)")
			}
			return runHistory(cmd.Context(), opts, ids, startTime, endTime)
		},
	}

	cmd.Flags().StringSliceVar(&sloIDs, "slo-id", nil, "SLO IDs to retrieve history for (repeatable; may also be passed as positional arguments)")
	cmd.Flags().IntVar(&startTime, "start-time", 0, "Start time as Unix timestamp (required)")
	cmd.Flags().IntVar(&endTime, "end-time", 0, "End time as Unix timestamp (required)")

	_ = cmd.MarkFlagRequired("start-time")
	_ = cmd.MarkFlagRequired("end-time")

	return cmd
}

func mergeSLOIDs(args, flagIDs []string) []string {
	seen := make(map[string]bool, len(args)+len(flagIDs))
	var ids []string
	for _, id := range append(append([]string{}, args...), flagIDs...) {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids
}

func runHistory(ctx context.Context, opts *options.RootOptions, sloIDs []string, startTime, endTime int) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
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

	resp, err := client.GetSloHistoryWithResponse(ctx, body)
	if err != nil {
		return fmt.Errorf("getting SLO history: %w", err)
	}

	history, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	data := *history

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
