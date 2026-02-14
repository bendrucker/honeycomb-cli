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
			return runKeyList(cmd.Context(), opts, *team, filterType)
		},
	}

	cmd.Flags().StringVar(&filterType, "type", "", "Filter by key type (ingest, configuration)")

	return cmd
}

func runKeyList(ctx context.Context, opts *options.RootOptions, team, filterType string) error {
	auth, err := opts.KeyEditor(config.KeyManagement)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	var params *api.ListApiKeysParams
	if filterType != "" {
		ft := api.ListApiKeysParamsFilterType(filterType)
		params = &api.ListApiKeysParams{FilterType: &ft}
	}

	resp, err := client.ListApiKeysWithResponse(ctx, api.TeamSlug(team), params, auth)
	if err != nil {
		return fmt.Errorf("listing API keys: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.ApplicationvndApiJSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	items := make([]keyItem, len(resp.ApplicationvndApiJSON200.Data))
	for i, obj := range resp.ApplicationvndApiJSON200.Data {
		items[i] = objectToItem(obj)
	}

	return opts.OutputWriterList().Write(items, keyListTable)
}
