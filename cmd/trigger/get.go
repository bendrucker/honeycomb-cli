package trigger

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewGetCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <trigger-id>",
		Short: "Get a trigger",
		Example: `  # Get a trigger by ID
  honeycomb trigger get abc123 --dataset my-dataset`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd.Context(), opts, *dataset, args[0])
		},
	}
}

func runGet(ctx context.Context, opts *options.RootOptions, dataset, triggerID string) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	resp, err := client.GetTriggerWithResponse(ctx, dataset, triggerID)
	if err != nil {
		return fmt.Errorf("getting trigger: %w", err)
	}

	trigger, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeTriggerDetail(opts, toDetail(*trigger))
}
