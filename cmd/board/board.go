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
	ID            string        `json:"id" detail:"ID"`
	Name          string        `json:"name" detail:"Name"`
	Description   string        `json:"description,omitempty" detail:"Description"`
	Type          string        `json:"type" detail:"Type"`
	URL           string        `json:"url,omitempty" detail:"URL"`
	PresetFilters presetFilters `json:"preset_filters,omitempty" detail:"Preset Filters"`
	Panels        panels        `json:"panels,omitempty" detail:"Panels"`
}

func writeBoardDetail(opts *options.RootOptions, detail boardDetail) error {
	return opts.OutputWriter().WriteFields(detail, output.FieldsFromTags(detail))
}

// panels wraps a board's raw panels JSON. It preserves RawMessage's JSON
// passthrough via MarshalJSON while carrying its own table rendering.
type panels json.RawMessage

func (p panels) MarshalJSON() ([]byte, error) {
	return json.RawMessage(p).MarshalJSON()
}

func (p *panels) UnmarshalJSON(data []byte) error {
	return (*json.RawMessage)(p).UnmarshalJSON(data)
}

// FormatField summarizes a board's panels as one line per panel, pairing each
// panel's type with the query, SLO, or text it references. An empty raw message
// (no panels) renders as an em-dash.
func (p panels) FormatField() string {
	raw := json.RawMessage(p)
	if len(raw) == 0 {
		return "—"
	}

	var parsed []struct {
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
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return string(raw)
	}
	if len(parsed) == 0 {
		return "—"
	}

	lines := make([]string, len(parsed))
	for i, p := range parsed {
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

// presetFilters wraps a board's raw preset_filters JSON. It preserves
// RawMessage's JSON passthrough while rendering as its raw string in tables,
// matching the previous string(rawMessage) behavior.
type presetFilters json.RawMessage

func (p presetFilters) MarshalJSON() ([]byte, error) {
	return json.RawMessage(p).MarshalJSON()
}

func (p *presetFilters) UnmarshalJSON(data []byte) error {
	return (*json.RawMessage)(p).UnmarshalJSON(data)
}

func (p presetFilters) FormatField() string {
	return string(p)
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
		d.PresetFilters = presetFilters(raw)
	}
	if b.Panels != nil {
		raw, _ := json.Marshal(b.Panels)
		d.Panels = panels(raw)
	}
	return d
}
