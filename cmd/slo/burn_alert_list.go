package slo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewBurnAlertListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var sloID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List burn alerts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runBurnAlertList(cmd.Context(), opts, *dataset, sloID)
		},
	}

	cmd.Flags().StringVar(&sloID, "slo-id", "", "Filter by SLO ID (required)")
	_ = cmd.MarkFlagRequired("slo-id")

	return cmd
}

func runBurnAlertList(ctx context.Context, opts *options.RootOptions, dataset, sloID string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	params := &api.ListBurnAlertsBySloParams{SloId: sloID}
	resp, err := client.ListBurnAlertsBySloWithResponse(ctx, dataset, params, keyEditor(key))
	if err != nil {
		return fmt.Errorf("listing burn alerts: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	var raw []json.RawMessage
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing burn alerts response: %w", err)
	}

	items := make([]burnAlertItem, len(raw))
	for i, r := range raw {
		if err := json.Unmarshal(r, &items[i]); err != nil {
			return fmt.Errorf("parsing burn alert: %w", err)
		}
	}

	return opts.OutputWriter().Write(items, burnAlertListTable)
}
