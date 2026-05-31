package recipient

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var recipientListTable = output.TableFromTags[recipientItem]()

func NewListCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List recipients",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runList(cmd.Context(), opts)
		},
	}
}

func runList(ctx context.Context, opts *options.RootOptions) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	resp, err := client.ListRecipientsWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("listing recipients: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	details, err := unmarshalRecipients(resp.Body)
	if err != nil {
		return err
	}

	items := make([]recipientItem, len(details))
	for i, d := range details {
		items[i] = detailToItem(d)
	}

	return opts.OutputWriterList().Write(items, recipientListTable)
}
