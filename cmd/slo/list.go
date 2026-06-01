package slo

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
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
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	resp, err := client.ListSlosWithResponse(ctx, dataset)
	if err != nil {
		return fmt.Errorf("listing SLOs: %w", err)
	}

	slos, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	items := make([]sloItem, len(*slos))
	for i, s := range *slos {
		items[i] = sloItem{
			ID:               deref.String(s.Id),
			Name:             s.Name,
			Description:      deref.String(s.Description),
			TargetPerMillion: s.TargetPerMillion,
			TimePeriodDays:   s.TimePeriodDays,
			SLIAlias:         s.Sli.Alias,
		}
	}

	return opts.OutputWriterList().Write(items, sloListTable)
}
