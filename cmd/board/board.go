package board

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "board",
		Short:   "Manage boards",
		Aliases: []string{"boards"},
		Example: `  # List boards
  honeycomb board list

  # Get a board by ID
  honeycomb board get abc123

  # Create a board from a file
  honeycomb board create --file board.json`,
	}

	cmd.AddCommand(NewListCmd(opts))
	cmd.AddCommand(NewGetCmd(opts))
	cmd.AddCommand(NewCreateCmd(opts))
	cmd.AddCommand(NewUpdateCmd(opts))
	cmd.AddCommand(NewDeleteCmd(opts))
	cmd.AddCommand(NewViewCmd(opts))

	return command.Group(cmd)
}

type boardListItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

type boardDetail struct {
	ID            string          `json:"id" detail:"ID"`
	Name          string          `json:"name" detail:"Name"`
	Description   string          `json:"description,omitempty" detail:"Description"`
	Type          string          `json:"type" detail:"Type"`
	URL           string          `json:"url,omitempty" detail:"URL"`
	PresetFilters json.RawMessage `json:"preset_filters,omitempty"`
	Panels        json.RawMessage `json:"panels,omitempty"`
}

func writeBoardDetail(opts *options.RootOptions, detail boardDetail) error {
	fields := output.FieldsFromTags(detail)
	fields = append(fields,
		output.Field{Label: "Panels", Value: formatPanels(detail.Panels)},
		output.Field{Label: "Preset Filters", Value: string(detail.PresetFilters)},
	)
	return opts.OutputWriter().WriteFields(detail, fields)
}

// formatPanels summarizes a board's panels as one line per panel, pairing each
// panel's type with the query, SLO, or text it references. An empty raw message
// (no panels) renders as an em-dash.
func formatPanels(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "—"
	}

	var panels []struct {
		Type       string `json:"type"`
		QueryPanel *struct {
			QueryId string `json:"query_id"`
		} `json:"query_panel"`
		SLOPanel *struct {
			SloId string `json:"slo_id"`
		} `json:"slo_panel"`
		TextPanel *struct {
			Content string `json:"content"`
		} `json:"text_panel"`
	}
	if err := json.Unmarshal(raw, &panels); err != nil {
		return string(raw)
	}
	if len(panels) == 0 {
		return "—"
	}

	lines := make([]string, len(panels))
	for i, p := range panels {
		switch {
		case p.QueryPanel != nil:
			lines[i] = fmt.Sprintf("%s (query %s)", p.Type, p.QueryPanel.QueryId)
		case p.SLOPanel != nil:
			lines[i] = fmt.Sprintf("%s (slo %s)", p.Type, p.SLOPanel.SloId)
		default:
			lines[i] = p.Type
		}
	}
	return strings.Join(lines, "\n")
}

func boardToDetail(b api.Board) boardDetail {
	d := boardDetail{
		ID:          deref.String(b.Id),
		Name:        b.Name,
		Type:        string(b.Type),
		Description: deref.String(b.Description),
	}
	if b.Links != nil {
		d.URL = deref.String(b.Links.BoardUrl)
	}
	if b.PresetFilters != nil {
		raw, _ := json.Marshal(b.PresetFilters)
		d.PresetFilters = raw
	}
	if b.Panels != nil {
		raw, _ := json.Marshal(b.Panels)
		d.Panels = raw
	}
	return d
}
