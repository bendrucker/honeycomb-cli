package key

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an API key",
		Example: `  # Delete an API key, prompting for confirmation
  honeycomb key delete abc123

  # Delete without confirmation
  honeycomb key delete abc123 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.ClientFor(team, options.AuthManagement)
			if err != nil {
				return err
			}
			return runKeyDelete(cmd.Context(), opts, client, *team, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runKeyDelete(ctx context.Context, opts *options.RootOptions, client *api.ClientWithResponses, team, id string, yes bool) error {
	proceed, err := command.ConfirmDelete(opts.IOStreams, yes, "API key", id, nil)
	if err != nil {
		return err
	}
	if !proceed {
		return nil
	}

	resp, err := client.DeleteApiKeyWithResponse(ctx, api.TeamSlug(team), api.ID(id))
	if err != nil {
		return fmt.Errorf("deleting API key: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(id, fmt.Sprintf("Deleted API key %s", id))
}
