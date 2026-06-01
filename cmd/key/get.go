package key

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
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
			client, err := opts.ClientFor(team, options.AuthManagement)
			if err != nil {
				return err
			}
			return runKeyGet(cmd.Context(), opts, client, *team, args[0])
		},
	}
}

func runKeyGet(ctx context.Context, opts *options.RootOptions, client *api.ClientWithResponses, team, id string) error {
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
