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

type viewItem struct {
	ID   string `json:"id" col:"ID"`
	Name string `json:"name" col:"Name"`
}

type viewDetail struct {
	ID      string  `json:"id" detail:"ID"`
	Name    string  `json:"name" detail:"Name"`
	Filters filters `json:"filters,omitempty" detail:"Filters"`
}

var viewListTable = output.TableFromTags[viewItem]()

func viewResponseToItem(v api.BoardViewResponse) viewItem {
	return viewItem{
		ID:   deref.String(v.Id),
		Name: deref.String(v.Name),
	}
}

func viewResponseToDetail(v api.BoardViewResponse) viewDetail {
	d := viewDetail{
		ID:   deref.String(v.Id),
		Name: deref.String(v.Name),
	}
	if v.Filters != nil {
		raw, _ := json.Marshal(v.Filters)
		d.Filters = filters(raw)
	}
	return d
}

func writeViewDetail(opts *options.RootOptions, detail viewDetail) error {
	return opts.OutputWriter().WriteFields(detail, output.FieldsFromTags(detail))
}

// filters wraps a view's raw filters JSON. It preserves RawMessage's JSON
// passthrough via MarshalJSON while carrying its own table rendering.
type filters json.RawMessage

func (f filters) MarshalJSON() ([]byte, error) {
	return json.RawMessage(f).MarshalJSON()
}

func (f *filters) UnmarshalJSON(data []byte) error {
	return (*json.RawMessage)(f).UnmarshalJSON(data)
}

// FormatField renders a view's filters as one "column operation value" line
// per filter. An empty raw message (no filters) renders as an em-dash.
func (f filters) FormatField() string {
	raw := json.RawMessage(f)
	if len(raw) == 0 {
		return "—"
	}

	var parsed []struct {
		Column    string `json:"column"`
		Operation string `json:"operation"`
		Value     any    `json:"value,omitempty"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return string(raw)
	}
	if len(parsed) == 0 {
		return "—"
	}

	lines := make([]string, len(parsed))
	for i, p := range parsed {
		if p.Value != nil {
			lines[i] = fmt.Sprintf("%s %s %v", p.Column, p.Operation, p.Value)
		} else {
			lines[i] = fmt.Sprintf("%s %s", p.Column, p.Operation)
		}
	}
	return strings.Join(lines, "\n")
}

func NewViewCmd(opts *options.RootOptions) *cobra.Command {
	var board string

	cmd := &cobra.Command{
		Use:     "view",
		Short:   "Manage board views",
		Aliases: []string{"views"},
		Example: `  # List views on a board
  honeycomb board view list --board abc123

  # Get a view by ID
  honeycomb board view get view123 --board abc123`,
	}

	cmd.PersistentFlags().StringVar(&board, "board", "", "Board ID (required)")
	_ = cmd.MarkPersistentFlagRequired("board")

	cmd.AddCommand(NewViewListCmd(opts, &board))
	cmd.AddCommand(NewViewGetCmd(opts, &board))
	cmd.AddCommand(NewViewCreateCmd(opts, &board))
	cmd.AddCommand(NewViewUpdateCmd(opts, &board))
	cmd.AddCommand(NewViewDeleteCmd(opts, &board))

	return command.Group(cmd)
}
