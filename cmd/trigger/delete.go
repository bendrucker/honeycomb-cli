package trigger

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <trigger-id>",
		Short: "Delete a trigger",
		Example: `  # Delete a trigger, prompting for confirmation
  honeycomb trigger delete abc123 --dataset my-dataset

  # Delete without confirmation
  honeycomb trigger delete abc123 --dataset my-dataset --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd.Context(), opts, *dataset, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(ctx context.Context, opts *options.RootOptions, dataset, triggerID string, yes bool) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	proceed, err := command.ConfirmDelete(opts.IOStreams, yes, "trigger", triggerID, func() (string, error) {
		resp, err := client.GetTriggerWithResponse(ctx, dataset, triggerID)
		if err != nil {
			return "", fmt.Errorf("getting trigger: %w", err)
		}
		if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
			return "", err
		}
		if resp.JSON200 != nil && resp.JSON200.Name != nil {
			return *resp.JSON200.Name, nil
		}
		return "", nil
	})
	if err != nil {
		return err
	}
	if !proceed {
		return nil
	}

	resp, err := client.DeleteTriggerWithResponse(ctx, dataset, triggerID)
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
