package trigger

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

var triggerListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(triggerItem).ID }},
		{Header: "Name", Value: func(v any) string { return v.(triggerItem).Name }},
		{Header: "Description", Value: func(v any) string { return v.(triggerItem).Description }},
		{Header: "Disabled", Value: func(v any) string { return strconv.FormatBool(v.(triggerItem).Disabled) }},
		{Header: "Triggered", Value: func(v any) string { return strconv.FormatBool(v.(triggerItem).Triggered) }},
		{Header: "Alert Type", Value: func(v any) string { return v.(triggerItem).AlertType }},
		{Header: "Threshold", Value: func(v any) string { return v.(triggerItem).Threshold }},
	},
}

func NewListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List triggers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runList(cmd.Context(), opts, *dataset)
		},
	}
}

func runList(ctx context.Context, opts *options.RootOptions, dataset string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.ListTriggersWithResponse(ctx, dataset, auth)
	if err != nil {
		return fmt.Errorf("listing triggers: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	items := make([]triggerItem, len(*resp.JSON200))
	for i, t := range *resp.JSON200 {
		items[i] = toItem(t)
	}

	return opts.OutputWriter().Write(items, triggerListTable)
}
