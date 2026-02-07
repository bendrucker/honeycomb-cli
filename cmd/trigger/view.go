package trigger

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var triggerViewTable = output.TableDef{
	Columns: []output.Column{
		{Header: "Field", Value: func(v any) string { return v.([2]string)[0] }},
		{Header: "Value", Value: func(v any) string { return v.([2]string)[1] }},
	},
}

func NewViewCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "view <trigger-id>",
		Short: "View a trigger",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(cmd.Context(), opts, *dataset, args[0])
		},
	}
}

func runView(ctx context.Context, opts *options.RootOptions, dataset, triggerID string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetTriggerWithResponse(ctx, dataset, triggerID, keyEditor(key))
	if err != nil {
		return fmt.Errorf("getting trigger: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	detail := toDetail(*resp.JSON200)
	return opts.OutputWriter().Write(detail, triggerViewTable)
}
