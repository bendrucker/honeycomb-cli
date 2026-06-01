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

var recipientListTable = output.TableDef{Columns: []output.Column{
	{Header: "ID", Value: func(v any) string { return v.(recipientDetail).ID }},
	{Header: "Type", Value: func(v any) string { return v.(recipientDetail).Type }},
	{Header: "Target", Value: func(v any) string { return extractTarget(v.(recipientDetail)) }},
}}

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

	return opts.OutputWriterList().Write(details, recipientListTable)
}
