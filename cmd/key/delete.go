package key

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an API key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeyDelete(cmd.Context(), opts, *team, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runKeyDelete(ctx context.Context, opts *options.RootOptions, team, id string, yes bool) error {
	auth, err := opts.KeyEditor(config.KeyManagement)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	if !yes {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--yes is required in non-interactive mode")
		}

		answer, err := prompt.Choice(opts.IOStreams.Err, opts.IOStreams.In,
			fmt.Sprintf("Delete API key %s? (y/N): ", id),
			[]string{"y", "N"},
		)
		if err != nil {
			return err
		}
		if answer != "y" {
			return nil
		}
	}

	resp, err := client.DeleteApiKeyWithResponse(ctx, api.TeamSlug(team), api.ID(id), auth)
	if err != nil {
		return fmt.Errorf("deleting API key: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(id, fmt.Sprintf("Deleted API key %s", id))
}
