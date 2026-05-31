package recipient

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(opts *options.RootOptions) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <recipient-id>",
		Short: "Delete a recipient",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd.Context(), opts, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(ctx context.Context, opts *options.RootOptions, recipientID string, yes bool) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	proceed, err := command.ConfirmDelete(opts.IOStreams, yes, "recipient", recipientID, nil)
	if err != nil {
		return err
	}
	if !proceed {
		return fmt.Errorf("aborted")
	}

	resp, err := client.DeleteRecipientWithResponse(ctx, recipientID)
	if err != nil {
		return fmt.Errorf("deleting recipient: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(recipientID, fmt.Sprintf("Recipient %s deleted", recipientID))
}
