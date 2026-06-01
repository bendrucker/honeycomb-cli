package column

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
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
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	proceed, err := command.ConfirmDelete(opts.IOStreams, yes, "calculated column", id, func() (string, error) {
		getResp, err := client.GetCalculatedFieldWithResponse(ctx, dataset, id)
		if err != nil {
			return "", fmt.Errorf("getting calculated column: %w", err)
		}

		if err := api.CheckResponse(getResp.StatusCode(), getResp.Body); err != nil {
			return "", err
		}

		if getResp.JSON200 == nil {
			return "", fmt.Errorf("unexpected response: %s", getResp.Status())
		}

		return getResp.JSON200.Alias, nil
	})
	if err != nil {
		return err
	}
	if !proceed {
		return fmt.Errorf("aborted")
	}

	resp, err := client.DeleteCalculatedFieldWithResponse(ctx, dataset, id)
	if err != nil {
		return fmt.Errorf("deleting calculated column: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(id, fmt.Sprintf("Calculated column %s deleted", id))
}
