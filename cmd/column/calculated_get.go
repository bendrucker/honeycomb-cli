package column

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewCalculatedGetCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a calculated column",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCalculatedGet(cmd.Context(), opts, *dataset, args[0])
		},
	}
}

func runCalculatedGet(ctx context.Context, opts *options.RootOptions, dataset, id string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	resp, err := client.GetCalculatedFieldWithResponse(ctx, dataset, id)
	if err != nil {
		return fmt.Errorf("getting calculated column: %w", err)
	}

	field, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeCalculatedDetail(opts, *field)
}
