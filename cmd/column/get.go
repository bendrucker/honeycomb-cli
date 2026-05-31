package column

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
		Use:   "get <column-id>",
		Short: "Get a column",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runColumnGet(cmd.Context(), opts, *dataset, args[0])
		},
	}
}

func runColumnGet(ctx context.Context, opts *options.RootOptions, dataset, columnID string) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	resp, err := client.GetColumnWithResponse(ctx, dataset, columnID)
	if err != nil {
		return fmt.Errorf("getting column: %w", err)
	}

	col, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeColumnDetail(opts, *col)
}
