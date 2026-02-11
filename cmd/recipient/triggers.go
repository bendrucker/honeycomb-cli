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

type triggerItem struct {
	ID      string `json:"id"                      yaml:"id"`
	Name    string `json:"name"                    yaml:"name"`
	Dataset string `json:"dataset_slug,omitempty"  yaml:"dataset_slug,omitempty"`
}

var triggerListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(triggerItem).ID }},
		{Header: "Name", Value: func(v any) string { return v.(triggerItem).Name }},
		{Header: "Dataset", Value: func(v any) string { return v.(triggerItem).Dataset }},
	},
}

func NewTriggersCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "triggers <recipient-id>",
		Short: "List triggers for a recipient",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRecipientTriggers(cmd.Context(), opts, args[0])
		},
	}
}

func runRecipientTriggers(ctx context.Context, opts *options.RootOptions, recipientID string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.ListTriggersWithRecipientWithResponse(ctx, recipientID, keyEditor(key))
	if err != nil {
		return fmt.Errorf("listing triggers: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	items := make([]triggerItem, len(*resp.JSON200))
	for i, t := range *resp.JSON200 {
		item := triggerItem{}
		if t.Id != nil {
			item.ID = *t.Id
		}
		if t.Name != nil {
			item.Name = *t.Name
		}
		if t.DatasetSlug != nil {
			item.Dataset = *t.DatasetSlug
		}
		items[i] = item
	}

	return opts.OutputWriter().Write(items, triggerListTable)
}
