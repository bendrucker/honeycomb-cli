package column

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCalculatedDeleteCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a calculated column",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCalculatedDelete(cmd.Context(), opts, *dataset, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runCalculatedDelete(ctx context.Context, opts *options.RootOptions, dataset, id string, yes bool) error {
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

		getResp, err := client.GetCalculatedFieldWithResponse(ctx, dataset, id, keyEditor(key))
		if err != nil {
			return fmt.Errorf("getting calculated column: %w", err)
		}

		if err := api.CheckResponse(getResp.StatusCode(), getResp.Body); err != nil {
			return err
		}

		if getResp.JSON200 == nil {
			return fmt.Errorf("unexpected response: %s", getResp.Status())
		}

		answer, err := prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, fmt.Sprintf("Delete calculated column %q? (y/N): ", getResp.JSON200.Alias))
		if err != nil {
			return err
		}
		if answer != "y" && answer != "Y" {
			return fmt.Errorf("aborted")
		}
	}

	resp, err := client.DeleteCalculatedFieldWithResponse(ctx, dataset, id, keyEditor(key))
	if err != nil {
		return fmt.Errorf("deleting calculated column: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(id, fmt.Sprintf("Calculated column %s deleted", id))
}
