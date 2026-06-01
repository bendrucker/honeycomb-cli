package key

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewListCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var filterType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List API keys",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := opts.RequireTeam(team); err != nil {
				return err
			}
			return runKeyList(cmd.Context(), opts, *team, filterType)
		},
	}

	cmd.Flags().StringVar(&filterType, "type", "", "Filter by key type (ingest, configuration)")

	return cmd
}

func runKeyList(ctx context.Context, opts *options.RootOptions, team, filterType string) error {
	client, err := opts.Client(config.KeyManagement)
	if err != nil {
		return err
	}

	var params *api.ListApiKeysParams
	if filterType != "" {
		ft := api.ListApiKeysParamsFilterType(filterType)
		params = &api.ListApiKeysParams{FilterType: &ft}
	}

	resp, err := client.ListApiKeysWithResponse(ctx, api.TeamSlug(team), params)
	if err != nil {
		return fmt.Errorf("listing API keys: %w", err)
	}

	list, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON200)
	if err != nil {
		return err
	}

	items := make([]keyItem, len(list.Data))
	for i, obj := range list.Data {
		items[i] = objectToItem(obj)
	}

	return opts.OutputWriterList().Write(items, keyListTable)
}
