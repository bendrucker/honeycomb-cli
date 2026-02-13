package slo

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

func NewDeleteCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <slo-id>",
		Short: "Delete an SLO",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSLODelete(cmd.Context(), opts, *dataset, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runSLODelete(ctx context.Context, opts *options.RootOptions, dataset, sloID string, yes bool) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
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

		s, err := getSLO(ctx, client, auth, dataset, sloID)
		if err != nil {
			return err
		}

		answer, err := prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, fmt.Sprintf("Delete SLO %q? (y/N): ", s.Name))
		if err != nil {
			return err
		}
		if !strings.EqualFold(answer, "y") {
			return fmt.Errorf("aborted")
		}
	}

	resp, err := client.DeleteSloWithResponse(ctx, dataset, sloID, auth)
	if err != nil {
		return fmt.Errorf("deleting SLO: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(sloID, fmt.Sprintf("SLO %s deleted", sloID))
}
