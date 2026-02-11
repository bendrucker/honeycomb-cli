package recipient

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewListCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List recipients",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRecipientList(cmd.Context(), opts)
		},
	}
}

func runRecipientList(ctx context.Context, opts *options.RootOptions) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.ListRecipientsWithResponse(ctx, keyEditor(key))
	if err != nil {
		return fmt.Errorf("listing recipients: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	items, err := parseRecipientListBody(resp.Body)
	if err != nil {
		return err
	}

	return opts.OutputWriter().Write(items, recipientListTable)
}
