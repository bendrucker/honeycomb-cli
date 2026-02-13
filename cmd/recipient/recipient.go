package recipient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
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

func keyEditor(key string) api.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		config.ApplyAuth(req, config.KeyConfig, key)
		return nil
	}
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
	return opts.OutputWriter().WriteValue(detail, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID:\t%s\n", detail.ID)
		_, _ = fmt.Fprintf(tw, "Type:\t%s\n", detail.Type)
		target := extractTarget(detail)
		if target != "" {
			_, _ = fmt.Fprintf(tw, "Target:\t%s\n", target)
		}
		if detail.CreatedAt != "" {
			_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", detail.CreatedAt)
		}
		if detail.UpdatedAt != "" {
			_, _ = fmt.Fprintf(tw, "Updated At:\t%s\n", detail.UpdatedAt)
		}
		return tw.Flush()
	})
}
