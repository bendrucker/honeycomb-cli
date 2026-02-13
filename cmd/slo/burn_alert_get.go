package slo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewBurnAlertGetCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <burn-alert-id>",
		Short: "Get a burn alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBurnAlertGet(cmd.Context(), opts, *dataset, args[0])
		},
	}
}

func runBurnAlertGet(ctx context.Context, opts *options.RootOptions, dataset, burnAlertID string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetBurnAlertWithResponse(ctx, dataset, burnAlertID, auth)
	if err != nil {
		return fmt.Errorf("getting burn alert: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	var detail burnAlertDetail
	if err := json.Unmarshal(resp.Body, &detail); err != nil {
		return fmt.Errorf("parsing burn alert response: %w", err)
	}

	return writeBurnAlertDetail(opts, detail)
}

func writeBurnAlertDetail(opts *options.RootOptions, detail burnAlertDetail) error {
	fields := []output.Field{
		{Label: "ID", Value: detail.ID},
		{Label: "Alert Type", Value: detail.AlertType},
	}
	if detail.Description != "" {
		fields = append(fields, output.Field{Label: "Description", Value: detail.Description})
	}
	if detail.SloID != "" {
		fields = append(fields, output.Field{Label: "SLO ID", Value: detail.SloID})
	}
	if detail.ExhaustionMinutes != nil {
		fields = append(fields, output.Field{Label: "Exhaustion Minutes", Value: fmt.Sprintf("%d", *detail.ExhaustionMinutes)})
	}
	if detail.BudgetRateDecreaseThresholdPerMillion != nil {
		fields = append(fields, output.Field{Label: "Budget Rate Threshold", Value: fmt.Sprintf("%d", *detail.BudgetRateDecreaseThresholdPerMillion)})
	}
	if detail.BudgetRateWindowMinutes != nil {
		fields = append(fields, output.Field{Label: "Budget Rate Window", Value: fmt.Sprintf("%d min", *detail.BudgetRateWindowMinutes)})
	}
	if detail.CreatedAt != "" {
		fields = append(fields, output.Field{Label: "Created At", Value: detail.CreatedAt})
	}
	if detail.UpdatedAt != "" {
		fields = append(fields, output.Field{Label: "Updated At", Value: detail.UpdatedAt})
	}
	return opts.OutputWriter().WriteFields(detail, fields)
}
