package recipient

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewGetCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <recipient-id>",
		Short: "Get a recipient",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd.Context(), opts, args[0])
		},
	}
}

func runGet(ctx context.Context, opts *options.RootOptions, recipientID string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	resp, err := client.GetRecipientWithResponse(ctx, recipientID)
	if err != nil {
		return fmt.Errorf("getting recipient: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	detail, err := unmarshalRecipient(resp.Body)
	if err != nil {
		return err
	}

	return writeRecipientDetail(opts, detail)
}
