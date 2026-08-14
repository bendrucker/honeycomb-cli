package environment

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var environmentListTable = output.TableFromTags[environmentItem]()

func NewListCmd(opts *options.RootOptions, team *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List environments",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := opts.ClientFor(team, options.AuthManagement)
			if err != nil {
				return err
			}
			return runEnvironmentList(cmd.Context(), opts, client, *team)
		},
	}
}

func runEnvironmentList(ctx context.Context, opts *options.RootOptions, client *api.ClientWithResponses, team string) error {
	var items []environmentItem
	params := &api.ListEnvironmentsParams{}
	var cursor string
	for {
		resp, err := client.ListEnvironmentsWithResponse(ctx, team, params)
		if err != nil {
			return fmt.Errorf("listing environments: %w", err)
		}

		list, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON200)
		if err != nil {
			return err
		}

		for _, e := range list.Data {
			items = append(items, envToItem(e))
		}

		cursor, err = api.NextPageCursor(list.Links, cursor)
		if err != nil {
			return err
		}
		if cursor == "" {
			break
		}
		params.PageAfter = &cursor
	}

	return opts.OutputWriterList().WriteList(items, environmentListTable, "No environments found.")
}
