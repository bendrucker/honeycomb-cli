package column

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewCalculatedGetCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a calculated field",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCalculatedGet(cmd.Context(), opts, *dataset, args[0])
		},
	}
}

func runCalculatedGet(ctx context.Context, opts *options.RootOptions, dataset, fieldID string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetCalculatedFieldWithResponse(ctx, api.DatasetSlugOrAll(dataset), fieldID, keyEditor(key))
	if err != nil {
		return fmt.Errorf("getting calculated field: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeCalculatedDetail(opts, *resp.JSON200)
}
