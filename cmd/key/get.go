package key

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewGetCmd(opts *options.RootOptions, team *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an API key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.RequireTeam(team); err != nil {
				return err
			}
			return runKeyGet(cmd.Context(), opts, *team, args[0])
		},
	}
}

func runKeyGet(ctx context.Context, opts *options.RootOptions, team, id string) error {
	auth, err := opts.KeyEditor(config.KeyManagement)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetApiKeyWithResponse(ctx, api.TeamSlug(team), api.ID(id), auth)
	if err != nil {
		return fmt.Errorf("getting API key: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.ApplicationvndApiJSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeKeyDetail(opts, objectToDetail(resp.ApplicationvndApiJSON200.Data))
}
