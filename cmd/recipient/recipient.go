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
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "recipient",
		Short:   "Manage notification recipients",
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
	ID     string `json:"id"                yaml:"id"`
	Type   string `json:"type"              yaml:"type"`
	Target string `json:"target,omitempty"  yaml:"target,omitempty"`
}

type recipientDetail struct {
	ID         string          `json:"id"                    yaml:"id"`
	Type       string          `json:"type"                  yaml:"type"`
	Target     string          `json:"target,omitempty"      yaml:"target,omitempty"`
	CreatedAt  string          `json:"created_at,omitempty"  yaml:"created_at,omitempty"`
	UpdatedAt  string          `json:"updated_at,omitempty"  yaml:"updated_at,omitempty"`
	Details    json.RawMessage `json:"details,omitempty"     yaml:"-"`
	DetailsAny any             `json:"-"                     yaml:"details,omitempty"`
}

type rawRecipient struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	Details   json.RawMessage `json:"details"`
}

func extractTarget(details json.RawMessage) string {
	var d map[string]any
	if json.Unmarshal(details, &d) != nil {
		return ""
	}
	for _, key := range []string{"email_address", "slack_channel", "webhook_name", "url"} {
		if v, ok := d[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func parseRecipientBody(body []byte) (recipientDetail, error) {
	var raw rawRecipient
	if err := json.Unmarshal(body, &raw); err != nil {
		return recipientDetail{}, fmt.Errorf("parsing recipient: %w", err)
	}

	detail := recipientDetail{
		ID:        raw.ID,
		Type:      raw.Type,
		Target:    extractTarget(raw.Details),
		CreatedAt: raw.CreatedAt,
		UpdatedAt: raw.UpdatedAt,
		Details:   raw.Details,
	}

	if len(raw.Details) > 0 {
		var detailsAny any
		_ = json.Unmarshal(raw.Details, &detailsAny)
		detail.DetailsAny = detailsAny
	}

	return detail, nil
}

func parseRecipientListBody(body []byte) ([]recipientItem, error) {
	var raws []rawRecipient
	if err := json.Unmarshal(body, &raws); err != nil {
		return nil, fmt.Errorf("parsing recipients: %w", err)
	}

	items := make([]recipientItem, len(raws))
	for i, raw := range raws {
		items[i] = recipientItem{
			ID:     raw.ID,
			Type:   raw.Type,
			Target: extractTarget(raw.Details),
		}
	}
	return items, nil
}

var recipientListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(recipientItem).ID }},
		{Header: "Type", Value: func(v any) string { return v.(recipientItem).Type }},
		{Header: "Target", Value: func(v any) string { return v.(recipientItem).Target }},
	},
}

func writeRecipientDetail(opts *options.RootOptions, detail recipientDetail) error {
	return opts.OutputWriter().WriteValue(detail, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID:\t%s\n", detail.ID)
		_, _ = fmt.Fprintf(tw, "Type:\t%s\n", detail.Type)
		_, _ = fmt.Fprintf(tw, "Target:\t%s\n", detail.Target)
		_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", detail.CreatedAt)
		_, _ = fmt.Fprintf(tw, "Updated At:\t%s\n", detail.UpdatedAt)
		return tw.Flush()
	})
}
