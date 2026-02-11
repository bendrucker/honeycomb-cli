package recipient

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

func NewDeleteCmd(opts *options.RootOptions) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a recipient",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRecipientDelete(cmd.Context(), opts, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runRecipientDelete(ctx context.Context, opts *options.RootOptions, recipientID string, yes bool) error {
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

		resp, err := client.GetRecipientWithResponse(ctx, recipientID, keyEditor(key))
		if err != nil {
			return fmt.Errorf("getting recipient: %w", err)
		}

		if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		detail, err := parseRecipientBody(resp.Body)
		if err != nil {
			return err
		}

		label := detail.Type
		if detail.Target != "" {
			label += " (" + detail.Target + ")"
		}

		answer, err := prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, fmt.Sprintf("Delete recipient %s? (y/N): ", label))
		if err != nil {
			return err
		}
		if !strings.EqualFold(answer, "y") {
			return fmt.Errorf("aborted")
		}
	}

	resp, err := client.DeleteRecipientWithResponse(ctx, recipientID, keyEditor(key))
	if err != nil {
		return fmt.Errorf("deleting recipient: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(opts.IOStreams.Err, "Recipient %s deleted\n", recipientID)
	return nil
}
