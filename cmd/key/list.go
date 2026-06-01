package key

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

var keyTypes = []string{
	string(api.Ingest),
	string(api.Configuration),
}

func NewListCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var (
		keyType    string
		legacyType string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List API keys",
		Example: `  # List API keys
  honeycomb key list

  # List only ingest keys
  honeycomb key list --key-type ingest`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			filterType := keyType
			if filterType == "" {
				filterType = legacyType
			}
			client, err := opts.ClientFor(team, options.AuthManagement)
			if err != nil {
				return err
			}
			return runKeyList(cmd.Context(), opts, client, *team, filterType)
		},
	}

	cmd.Flags().StringVar(&keyType, "key-type", "", "Filter by key type (ingest, configuration)")
	cmd.Flags().StringVar(&legacyType, "type", "", "Filter by key type (ingest, configuration)")
	_ = cmd.Flags().MarkHidden("type")
	_ = cmd.Flags().MarkDeprecated("type", "use --key-type")

	return cmd
}

func runKeyList(ctx context.Context, opts *options.RootOptions, client *api.ClientWithResponses, team, filterType string) error {
	if err := command.ValidateEnum("key-type", filterType, keyTypes); err != nil {
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

	return opts.OutputWriterList().WriteList(items, keyListTable, "No keys found.")
}
