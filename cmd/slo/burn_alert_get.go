package slo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
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
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetBurnAlertWithResponse(ctx, dataset, burnAlertID, keyEditor(key))
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
	return opts.OutputWriter().WriteValue(detail, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID:\t%s\n", detail.ID)
		_, _ = fmt.Fprintf(tw, "Alert Type:\t%s\n", detail.AlertType)
		if detail.Description != "" {
			_, _ = fmt.Fprintf(tw, "Description:\t%s\n", detail.Description)
		}
		if detail.SloID != "" {
			_, _ = fmt.Fprintf(tw, "SLO ID:\t%s\n", detail.SloID)
		}
		if detail.ExhaustionMinutes != nil {
			_, _ = fmt.Fprintf(tw, "Exhaustion Minutes:\t%d\n", *detail.ExhaustionMinutes)
		}
		if detail.BudgetRateDecreaseThresholdPerMillion != nil {
			_, _ = fmt.Fprintf(tw, "Budget Rate Threshold:\t%d\n", *detail.BudgetRateDecreaseThresholdPerMillion)
		}
		if detail.BudgetRateWindowMinutes != nil {
			_, _ = fmt.Fprintf(tw, "Budget Rate Window:\t%d min\n", *detail.BudgetRateWindowMinutes)
		}
		if detail.CreatedAt != "" {
			_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", detail.CreatedAt)
		}
		if detail.UpdatedAt != "" {
			_, _ = fmt.Fprintf(tw, "Updated At:\t%s\n", detail.UpdatedAt)
		}
		return tw.Flush()
	})
}
