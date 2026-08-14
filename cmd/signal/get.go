package signal

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewGetCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <signal-id>",
		Short: "Get a signal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd.Context(), opts, args[0])
		},
	}
}

func runGet(ctx context.Context, opts *options.RootOptions, id string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	resp, err := client.GetSignalWithResponse(ctx, id)
	if err != nil {
		return fmt.Errorf("getting signal: %w", err)
	}

	signal, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeSignalDetail(opts, toDetail(*signal))
}
