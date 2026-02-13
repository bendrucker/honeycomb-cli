package trigger

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
		Use:   "delete <trigger-id>",
		Short: "Delete a trigger",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd.Context(), opts, *dataset, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(ctx context.Context, opts *options.RootOptions, dataset, triggerID string, yes bool) error {
	key, err := opts.RequireKey(config.KeyConfig)
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

		resp, err := client.GetTriggerWithResponse(ctx, dataset, triggerID, keyEditor(key))
		if err != nil {
			return fmt.Errorf("getting trigger: %w", err)
		}
		if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		name := triggerID
		if resp.JSON200 != nil && resp.JSON200.Name != nil {
			name = *resp.JSON200.Name
		}

		answer, err := prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, fmt.Sprintf("Delete trigger %q? (y/N): ", name))
		if err != nil {
			return err
		}
		if !strings.EqualFold(answer, "y") {
			return nil
		}
	}

	resp, err := client.DeleteTriggerWithResponse(ctx, dataset, triggerID, keyEditor(key))
	if err != nil {
		return fmt.Errorf("deleting trigger: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.StatusCode() != 204 {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return opts.OutputWriter().WriteDeleted(triggerID, fmt.Sprintf("Deleted trigger %s", triggerID))
}
