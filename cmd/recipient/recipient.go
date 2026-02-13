package recipient

import (
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "recipient",
		Short:   "Manage recipients",
		Aliases: []string{"recipients"},
	}

	cmd.AddCommand(NewListCmd(opts))
	cmd.AddCommand(NewGetCmd(opts))
	cmd.AddCommand(NewCreateCmd(opts))
	cmd.AddCommand(NewUpdateCmd(opts))
	cmd.AddCommand(NewDeleteCmd(opts))
	cmd.AddCommand(NewTriggersCmd(opts))

	return cmd
}

type recipientItem struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Target string `json:"target,omitempty"`
}

type recipientDetail struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Details   map[string]any `json:"details,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
	UpdatedAt string         `json:"updated_at,omitempty"`
}

func unmarshalRecipients(body []byte) ([]recipientDetail, error) {
	var raw []map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing recipients: %w", err)
	}

	items := make([]recipientDetail, len(raw))
	for i, r := range raw {
		items[i] = mapToDetail(r)
	}
	return items, nil
}

func unmarshalRecipient(body []byte) (recipientDetail, error) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return recipientDetail{}, fmt.Errorf("parsing recipient: %w", err)
	}
	return mapToDetail(raw), nil
}

func mapToDetail(raw map[string]any) recipientDetail {
	d := recipientDetail{}
	if id, ok := raw["id"].(string); ok {
		d.ID = id
	}
	if t, ok := raw["type"].(string); ok {
		d.Type = t
	}
	if ca, ok := raw["created_at"].(string); ok {
		d.CreatedAt = ca
	}
	if ua, ok := raw["updated_at"].(string); ok {
		d.UpdatedAt = ua
	}

	if nested, ok := raw["details"].(map[string]any); ok {
		d.Details = nested
	} else {
		details := make(map[string]any)
		for k, v := range raw {
			switch k {
			case "id", "type", "created_at", "updated_at":
				continue
			default:
				details[k] = v
			}
		}
		if len(details) > 0 {
			d.Details = details
		}
	}

	return d
}

func detailToItem(d recipientDetail) recipientItem {
	item := recipientItem{
		ID:   d.ID,
		Type: d.Type,
	}
	item.Target = extractTarget(d)
	return item
}

func extractTarget(d recipientDetail) string {
	if d.Details == nil {
		return ""
	}
	for _, field := range []string{"name", "url", "channel", "address", "email_address", "integration_key", "webhook_url"} {
		if v, ok := d.Details[field].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

func writeRecipientDetail(opts *options.RootOptions, detail recipientDetail) error {
	fields := []output.Field{
		{Label: "ID", Value: detail.ID},
		{Label: "Type", Value: detail.Type},
	}
	if target := extractTarget(detail); target != "" {
		fields = append(fields, output.Field{Label: "Target", Value: target})
	}
	if detail.CreatedAt != "" {
		fields = append(fields, output.Field{Label: "Created At", Value: detail.CreatedAt})
	}
	if detail.UpdatedAt != "" {
		fields = append(fields, output.Field{Label: "Updated At", Value: detail.UpdatedAt})
	}
	return opts.OutputWriter().WriteFields(detail, fields)
}
