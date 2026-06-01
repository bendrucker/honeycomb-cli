package slo

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var sloListTable = output.TableDef{
	Columns: []output.Column{
		output.Col("ID", func(s sloItem) string { return s.ID }),
		output.Col("Name", func(s sloItem) string { return s.Name }),
		output.Col("Target", func(s sloItem) string { return formatTarget(s.TargetPerMillion) }),
		output.Col("Time Period", func(s sloItem) string { return formatTimePeriod(s.TimePeriodDays) }),
		output.Col("SLI Alias", func(s sloItem) string { return s.SLIAlias }),
		output.Col("Description", func(s sloItem) string { return output.Truncate(s.Description, 40) }),
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
	client, err := opts.ClientFor(nil, options.AuthConfig)
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

	return opts.OutputWriterList().WriteList(items, sloListTable, "No SLOs found.")
}
