package environment

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <environment-id>",
		Short: "Delete an environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.RequireTeam(team); err != nil {
				return err
			}
			return runEnvironmentDelete(cmd.Context(), opts, *team, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runEnvironmentDelete(ctx context.Context, opts *options.RootOptions, team, envID string, yes bool) error {
	client, err := opts.Client(config.KeyManagement)
	if err != nil {
		return err
	}

	proceed, err := command.ConfirmDelete(opts.IOStreams, yes, "environment", envID, func() (string, error) {
		getResp, err := client.GetEnvironmentWithResponse(ctx, team, envID)
		if err != nil {
			return "", fmt.Errorf("getting environment: %w", err)
		}
		if err := api.CheckResponse(getResp.StatusCode(), getResp.Body); err != nil {
			return "", err
		}
		if getResp.ApplicationvndApiJSON200 != nil {
			return getResp.ApplicationvndApiJSON200.Data.Attributes.Name, nil
		}
		return "", nil
	})
	if err != nil {
		return err
	}
	if !proceed {
		return fmt.Errorf("aborted")
	}

	resp, err := client.DeleteEnvironmentWithResponse(ctx, team, envID)
	if err != nil {
		return fmt.Errorf("deleting environment: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(envID, fmt.Sprintf("Environment %s deleted", envID))
}
