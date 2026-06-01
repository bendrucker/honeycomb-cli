package column

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <column-id>",
		Short: "Delete a column",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runColumnDelete(cmd.Context(), opts, *dataset, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runColumnDelete(ctx context.Context, opts *options.RootOptions, dataset, columnID string, yes bool) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	proceed, err := command.ConfirmDelete(opts.IOStreams, yes, "column", columnID, func() (string, error) {
		getResp, err := client.GetColumnWithResponse(ctx, dataset, columnID)
		if err != nil {
			return "", fmt.Errorf("getting column: %w", err)
		}

		if err := api.CheckResponse(getResp.StatusCode(), getResp.Body); err != nil {
			return "", err
		}

		if getResp.JSON200 == nil {
			return "", fmt.Errorf("unexpected response: %s", getResp.Status())
		}

		return getResp.JSON200.KeyName, nil
	})
	if err != nil {
		return err
	}
	if !proceed {
		return fmt.Errorf("aborted")
	}

	resp, err := client.DeleteColumnWithResponse(ctx, dataset, columnID)
	if err != nil {
		return fmt.Errorf("deleting column: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(columnID, fmt.Sprintf("Column %s deleted", columnID))
}
