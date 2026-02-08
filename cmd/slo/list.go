package slo

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var sloListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(sloItem).ID }},
		{Header: "Name", Value: func(v any) string { return v.(sloItem).Name }},
		{Header: "Target", Value: func(v any) string { return formatTarget(v.(sloItem).TargetPerMillion) }},
		{Header: "Time Period", Value: func(v any) string { return formatTimePeriod(v.(sloItem).TimePeriodDays) }},
		{Header: "SLI Alias", Value: func(v any) string { return v.(sloItem).SLIAlias }},
		{Header: "Description", Value: func(v any) string { return truncate(v.(sloItem).Description, 40) }},
	},
}

func NewListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List SLOs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSLOList(cmd.Context(), opts, *dataset)
		},
	}
}

func runSLOList(ctx context.Context, opts *options.RootOptions, dataset string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.ListSlosWithResponse(ctx, dataset, keyEditor(key))
	if err != nil {
		return fmt.Errorf("listing SLOs: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	items := make([]sloItem, len(*resp.JSON200))
	for i, s := range *resp.JSON200 {
		item := sloItem{
			Name:             s.Name,
			TargetPerMillion: s.TargetPerMillion,
			TimePeriodDays:   s.TimePeriodDays,
			SLIAlias:         s.Sli.Alias,
		}
		if s.Id != nil {
			item.ID = *s.Id
		}
		if s.Description != nil {
			item.Description = *s.Description
		}
		items[i] = item
	}

	return opts.OutputWriter().Write(items, sloListTable)
}
