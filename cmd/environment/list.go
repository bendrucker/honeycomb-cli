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

var environmentListTable = output.TableFromTags[environmentItem]()

func NewListCmd(opts *options.RootOptions, team *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List environments",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := opts.RequireTeam(team); err != nil {
				return err
			}
			return runEnvironmentList(cmd.Context(), opts, *team)
		},
	}
}

func runEnvironmentList(ctx context.Context, opts *options.RootOptions, team string) error {
	client, err := opts.Client(config.KeyManagement)
	if err != nil {
		return err
	}

	resp, err := client.ListEnvironmentsWithResponse(ctx, team, nil)
	if err != nil {
		return fmt.Errorf("listing environments: %w", err)
	}

	list, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON200)
	if err != nil {
		return err
	}

	items := make([]environmentItem, len(list.Data))
	for i, e := range list.Data {
		items[i] = envToItem(e)
	}

	return opts.OutputWriterList().Write(items, environmentListTable)
}
