package slo

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewBurnAlertUpdateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file                string
		exhaustionMinutes   int
		budgetRateWindow    int
		budgetRateThreshold int
		recipients          []string
		description         string
	)

	cmd := &cobra.Command{
		Use:   "update <burn-alert-id>",
		Short: "Update a burn alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBurnAlertUpdate(cmd, opts, *dataset, args[0], file, exhaustionMinutes, budgetRateWindow, budgetRateThreshold, recipients, description)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().IntVar(&exhaustionMinutes, "exhaustion-minutes", 0, "Minutes until budget exhaustion (for exhaustion_time alerts)")
	cmd.Flags().IntVar(&budgetRateWindow, "budget-rate-window-minutes", 0, "Time window in minutes for budget rate calculation")
	cmd.Flags().IntVar(&budgetRateThreshold, "budget-rate-threshold", 0, "Budget decrease threshold per million")
	cmd.Flags().StringSliceVar(&recipients, "recipient", nil, "Recipient ID (repeatable)")
	cmd.Flags().StringVar(&description, "description", "", "Burn alert description")

	for _, flag := range []string{"exhaustion-minutes", "budget-rate-window-minutes", "budget-rate-threshold", "recipient", "description"} {
		cmd.MarkFlagsMutuallyExclusive("file", flag)
	}

	return cmd
}

var burnAlertUpdateFlags = []string{"exhaustion-minutes", "budget-rate-window-minutes", "budget-rate-threshold", "recipient", "description"}

func runBurnAlertUpdate(cmd *cobra.Command, opts *options.RootOptions, dataset, burnAlertID, file string, exhaustionMinutes, budgetRateWindow, budgetRateThreshold int, recipients []string, description string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	ctx := cmd.Context()

	var data []byte
	if file != "" {
		data, err = readFile(opts, file)
		if err != nil {
			return err
		}
	} else if hasAnyFlag(cmd, burnAlertUpdateFlags...) {
		resp, err := client.GetBurnAlertWithResponse(ctx, dataset, burnAlertID, auth)
		if err != nil {
			return fmt.Errorf("getting burn alert: %w", err)
		}
		if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		var current map[string]any
		if err := json.Unmarshal(resp.Body, &current); err != nil {
			return fmt.Errorf("parsing burn alert: %w", err)
		}

		applyBurnAlertFlags(cmd, current, exhaustionMinutes, budgetRateWindow, budgetRateThreshold, recipients, description)

		data, err = json.Marshal(current)
		if err != nil {
			return fmt.Errorf("encoding burn alert: %w", err)
		}
	} else {
		return fmt.Errorf("--file, --exhaustion-minutes, --budget-rate-window-minutes, --budget-rate-threshold, --recipient, or --description is required")
	}

	resp, err := client.UpdateBurnAlertWithBodyWithResponse(ctx, dataset, burnAlertID, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("updating burn alert: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	var detail burnAlertDetail
	if err := json.Unmarshal(resp.Body, &detail); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	return writeBurnAlertDetail(opts, detail)
}

func hasAnyFlag(cmd *cobra.Command, names ...string) bool {
	for _, n := range names {
		if cmd.Flags().Changed(n) {
			return true
		}
	}
	return false
}

func applyBurnAlertFlags(cmd *cobra.Command, current map[string]any, exhaustionMinutes, budgetRateWindow, budgetRateThreshold int, recipients []string, description string) {
	if cmd.Flags().Changed("exhaustion-minutes") {
		current["exhaustion_minutes"] = exhaustionMinutes
	}
	if cmd.Flags().Changed("budget-rate-window-minutes") {
		current["budget_rate_window_minutes"] = budgetRateWindow
	}
	if cmd.Flags().Changed("budget-rate-threshold") {
		current["budget_rate_decrease_threshold_per_million"] = budgetRateThreshold
	}
	if cmd.Flags().Changed("description") {
		current["description"] = description
	}
	if cmd.Flags().Changed("recipient") {
		r := make([]map[string]any, len(recipients))
		for i, id := range recipients {
			r[i] = map[string]any{"id": id}
		}
		current["recipients"] = r
	}

	// Remove read-only fields
	delete(current, "id")
	delete(current, "created_at")
	delete(current, "updated_at")
	delete(current, "triggered")
	delete(current, "slo_id")
}
