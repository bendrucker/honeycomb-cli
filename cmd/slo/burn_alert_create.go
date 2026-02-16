package slo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewBurnAlertCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file                string
		sloID               string
		alertType           string
		exhaustionMinutes   int
		budgetRateWindowMin int
		budgetRateThreshold int
		recipients          []string
		description         string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a burn alert",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if file != "" {
				return runBurnAlertCreateFromFile(cmd.Context(), opts, *dataset, file)
			}
			return runBurnAlertCreateFromFlags(cmd.Context(), opts, *dataset, sloID, alertType, exhaustionMinutes, budgetRateWindowMin, budgetRateThreshold, recipients, description)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&sloID, "slo-id", "", "SLO ID")
	cmd.Flags().StringVar(&alertType, "alert-type", "", "Alert type (exhaustion_time or budget_rate)")
	cmd.Flags().IntVar(&exhaustionMinutes, "exhaustion-minutes", 0, "Minutes until budget exhaustion (required for exhaustion_time)")
	cmd.Flags().IntVar(&budgetRateWindowMin, "budget-rate-window-minutes", 0, "Budget rate window in minutes (required for budget_rate)")
	cmd.Flags().IntVar(&budgetRateThreshold, "budget-rate-threshold", 0, "Budget rate decrease threshold per million (required for budget_rate)")
	cmd.Flags().StringSliceVar(&recipients, "recipient", nil, "Recipient ID (repeatable)")
	cmd.Flags().StringVar(&description, "description", "", "Description")

	cmd.MarkFlagsMutuallyExclusive("file", "slo-id")
	cmd.MarkFlagsMutuallyExclusive("file", "alert-type")
	cmd.MarkFlagsMutuallyExclusive("file", "exhaustion-minutes")
	cmd.MarkFlagsMutuallyExclusive("file", "budget-rate-window-minutes")
	cmd.MarkFlagsMutuallyExclusive("file", "budget-rate-threshold")
	cmd.MarkFlagsMutuallyExclusive("file", "recipient")
	cmd.MarkFlagsMutuallyExclusive("file", "description")

	return cmd
}

func runBurnAlertCreateFromFile(ctx context.Context, opts *options.RootOptions, dataset, file string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	data, err := readFile(opts, file)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.CreateBurnAlertWithBodyWithResponse(ctx, dataset, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("creating burn alert: %w", err)
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

func runBurnAlertCreateFromFlags(ctx context.Context, opts *options.RootOptions, dataset, sloID, alertType string, exhaustionMinutes, budgetRateWindowMin, budgetRateThreshold int, recipients []string, description string) error {
	if sloID == "" {
		return fmt.Errorf("--slo-id is required")
	}
	if alertType == "" {
		return fmt.Errorf("--alert-type is required")
	}
	if len(recipients) == 0 {
		return fmt.Errorf("at least one --recipient is required")
	}

	body := map[string]any{
		"alert_type": alertType,
		"slo":        map[string]string{"id": sloID},
	}

	rcpts := make([]map[string]string, len(recipients))
	for i, id := range recipients {
		rcpts[i] = map[string]string{"id": id}
	}
	body["recipients"] = rcpts

	switch alertType {
	case "exhaustion_time":
		if exhaustionMinutes == 0 {
			return fmt.Errorf("--exhaustion-minutes is required when --alert-type=exhaustion_time")
		}
		body["exhaustion_minutes"] = exhaustionMinutes
	case "budget_rate":
		if budgetRateWindowMin == 0 {
			return fmt.Errorf("--budget-rate-window-minutes is required when --alert-type=budget_rate")
		}
		if budgetRateThreshold == 0 {
			return fmt.Errorf("--budget-rate-threshold is required when --alert-type=budget_rate")
		}
		body["budget_rate_window_minutes"] = budgetRateWindowMin
		body["budget_rate_decrease_threshold_per_million"] = budgetRateThreshold
	default:
		return fmt.Errorf("--alert-type must be exhaustion_time or budget_rate")
	}

	if description != "" {
		body["description"] = description
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding burn alert: %w", err)
	}

	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.CreateBurnAlertWithBodyWithResponse(ctx, dataset, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("creating burn alert: %w", err)
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
