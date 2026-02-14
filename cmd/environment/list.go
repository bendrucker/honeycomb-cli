package environment

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var environmentListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(environmentItem).ID }},
		{Header: "Name", Value: func(v any) string { return v.(environmentItem).Name }},
		{Header: "Slug", Value: func(v any) string { return v.(environmentItem).Slug }},
		{Header: "Description", Value: func(v any) string { return v.(environmentItem).Description }},
		{Header: "Color", Value: func(v any) string { return v.(environmentItem).Color }},
	},
}

func NewListCmd(opts *options.RootOptions, team *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List environments",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEnvironmentList(cmd.Context(), opts, *team)
		},
	}
}

func runEnvironmentList(ctx context.Context, opts *options.RootOptions, team string) error {
	auth, err := opts.KeyEditor(config.KeyManagement)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.ListEnvironmentsWithResponse(ctx, team, nil, auth)
	if err != nil {
		return fmt.Errorf("listing environments: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.ApplicationvndApiJSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	items := make([]environmentItem, len(resp.ApplicationvndApiJSON200.Data))
	for i, e := range resp.ApplicationvndApiJSON200.Data {
		items[i] = envToItem(e)
	}

	return opts.OutputWriterList().Write(items, environmentListTable)
}
