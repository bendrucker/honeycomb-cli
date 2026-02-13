package recipient

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type triggerItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Disabled    bool   `json:"disabled"`
	Triggered   bool   `json:"triggered"`
	AlertType   string `json:"alert_type,omitempty"`
	Threshold   string `json:"threshold,omitempty"`
}

var triggerListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(triggerItem).ID }},
		{Header: "Name", Value: func(v any) string { return v.(triggerItem).Name }},
		{Header: "Description", Value: func(v any) string { return v.(triggerItem).Description }},
		{Header: "Disabled", Value: func(v any) string { return strconv.FormatBool(v.(triggerItem).Disabled) }},
		{Header: "Triggered", Value: func(v any) string { return strconv.FormatBool(v.(triggerItem).Triggered) }},
		{Header: "Alert Type", Value: func(v any) string { return v.(triggerItem).AlertType }},
		{Header: "Threshold", Value: func(v any) string { return v.(triggerItem).Threshold }},
	},
}

func NewTriggersCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "triggers <recipient-id>",
		Short: "List triggers associated with a recipient",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTriggers(cmd.Context(), opts, args[0])
		},
	}
}

func runTriggers(ctx context.Context, opts *options.RootOptions, recipientID string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.ListTriggersWithRecipientWithResponse(ctx, recipientID, auth)
	if err != nil {
		return fmt.Errorf("listing triggers for recipient: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	items := make([]triggerItem, len(*resp.JSON200))
	for i, t := range *resp.JSON200 {
		items[i] = toTriggerItem(t)
	}

	return opts.OutputWriter().Write(items, triggerListTable)
}

func toTriggerItem(t api.TriggerResponse) triggerItem {
	item := triggerItem{}
	if t.Id != nil {
		item.ID = *t.Id
	}
	if t.Name != nil {
		item.Name = *t.Name
	}
	if t.Description != nil {
		item.Description = *t.Description
	}
	if t.Disabled != nil {
		item.Disabled = *t.Disabled
	}
	if t.Triggered != nil {
		item.Triggered = *t.Triggered
	}
	if t.AlertType != nil {
		item.AlertType = string(*t.AlertType)
	}
	if t.Threshold != nil {
		item.Threshold = fmt.Sprintf("%s %g", t.Threshold.Op, t.Threshold.Value)
	}
	return item
}
