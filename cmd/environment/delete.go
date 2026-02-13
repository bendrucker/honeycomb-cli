package environment

import (
	"context"
	"fmt"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <environment-id>",
		Short: "Delete an environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnvironmentDelete(cmd.Context(), opts, *team, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runEnvironmentDelete(ctx context.Context, opts *options.RootOptions, team, envID string, yes bool) error {
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

		getResp, err := client.GetEnvironmentWithResponse(ctx, team, envID, auth)
		if err != nil {
			return fmt.Errorf("getting environment: %w", err)
		}
		if err := api.CheckResponse(getResp.StatusCode(), getResp.Body); err != nil {
			return err
		}

		name := envID
		if getResp.ApplicationvndApiJSON200 != nil {
			name = getResp.ApplicationvndApiJSON200.Data.Attributes.Name
		}

		answer, err := prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, fmt.Sprintf("Delete environment %q? (y/N): ", name))
		if err != nil {
			return err
		}
		if !strings.EqualFold(answer, "y") {
			return fmt.Errorf("aborted")
		}
	}

	resp, err := client.DeleteEnvironmentWithResponse(ctx, team, envID, auth)
	if err != nil {
		return fmt.Errorf("deleting environment: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(envID, fmt.Sprintf("Environment %s deleted", envID))
}
