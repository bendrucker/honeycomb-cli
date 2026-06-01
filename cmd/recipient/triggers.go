package recipient

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type triggerItem struct {
	ID          string `json:"id" col:"ID"`
	Name        string `json:"name" col:"Name"`
	Description string `json:"description,omitempty" col:"Description"`
	Disabled    bool   `json:"disabled" col:"Disabled"`
	Triggered   bool   `json:"triggered" col:"Triggered"`
	AlertType   string `json:"alert_type,omitempty" col:"Alert Type"`
	Threshold   string `json:"threshold,omitempty" col:"Threshold"`
}

var triggerListTable = output.TableFromTags[triggerItem]()

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
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	resp, err := client.ListTriggersWithRecipientWithResponse(ctx, recipientID)
	if err != nil {
		return fmt.Errorf("listing triggers for recipient: %w", err)
	}

	triggers, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	items := make([]triggerItem, len(*triggers))
	for i, t := range *triggers {
		items[i] = toTriggerItem(t)
	}

	return opts.OutputWriterList().WriteList(items, triggerListTable, "No triggers found.")
}

func toTriggerItem(t api.TriggerResponse) triggerItem {
	item := triggerItem{
		ID:          deref.String(t.Id),
		Name:        deref.String(t.Name),
		Description: deref.String(t.Description),
		Disabled:    deref.Bool(t.Disabled),
		Triggered:   deref.Bool(t.Triggered),
		AlertType:   deref.Enum(t.AlertType),
	}
	if t.Threshold != nil {
		item.Threshold = fmt.Sprintf("%s %g", t.Threshold.Op, t.Threshold.Value)
	}
	return item
}
