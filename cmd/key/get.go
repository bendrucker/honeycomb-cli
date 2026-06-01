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
		Example: `  # Get an API key by ID
  honeycomb key get abc123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.RequireTeam(team); err != nil {
				return err
			}
			return runKeyGet(cmd.Context(), opts, *team, args[0])
		},
	}
}

func runKeyGet(ctx context.Context, opts *options.RootOptions, team, id string) error {
	client, err := opts.Client(config.KeyManagement)
	if err != nil {
		return err
	}

	resp, err := client.GetApiKeyWithResponse(ctx, api.TeamSlug(team), api.ID(id))
	if err != nil {
		return fmt.Errorf("getting API key: %w", err)
	}

	key, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON200)
	if err != nil {
		return err
	}

	return writeKeyDetail(opts, objectToDetail(key.Data))
}
