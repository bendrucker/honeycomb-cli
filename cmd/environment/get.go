package environment

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewGetCmd(opts *options.RootOptions, team *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <environment-id>",
		Short: "Get an environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.ClientFor(team, options.AuthManagement)
			if err != nil {
				return err
			}
			return runEnvironmentGet(cmd.Context(), opts, client, *team, args[0])
		},
	}
}

func runEnvironmentGet(ctx context.Context, opts *options.RootOptions, client *api.ClientWithResponses, team, envID string) error {
	resp, err := client.GetEnvironmentWithResponse(ctx, team, envID)
	if err != nil {
		return fmt.Errorf("getting environment: %w", err)
	}

	env, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON200)
	if err != nil {
		return err
	}

	return writeEnvironmentDetail(opts, envToDetail(env.Data))
}
