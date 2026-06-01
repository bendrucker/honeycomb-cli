package board

import (
	"encoding/json"
	"fmt"
	"strings"

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
	ID      string          `json:"id" detail:"ID"`
	Name    string          `json:"name" detail:"Name"`
	Filters json.RawMessage `json:"filters,omitempty"`
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
		d.Filters = raw
	}
	return d
}

func writeViewDetail(opts *options.RootOptions, detail viewDetail) error {
	fields := output.FieldsFromTags(detail)
	fields = append(fields, output.Field{Label: "Filters", Value: formatFilters(detail.Filters)})
	return opts.OutputWriter().WriteFields(detail, fields)
}

// formatFilters renders a view's filters as one "column operation value" line
// per filter. An empty raw message (no filters) renders as an em-dash.
func formatFilters(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "—"
	}

	var filters []struct {
		Column    string `json:"column"`
		Operation string `json:"operation"`
		Value     any    `json:"value,omitempty"`
	}
	if err := json.Unmarshal(raw, &filters); err != nil {
		return string(raw)
	}
	if len(filters) == 0 {
		return "—"
	}

	lines := make([]string, len(filters))
	for i, f := range filters {
		if f.Value != nil {
			lines[i] = fmt.Sprintf("%s %s %v", f.Column, f.Operation, f.Value)
		} else {
			lines[i] = fmt.Sprintf("%s %s", f.Column, f.Operation)
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
	}

	cmd.PersistentFlags().StringVar(&board, "board", "", "Board ID (required)")
	_ = cmd.MarkPersistentFlagRequired("board")

	cmd.AddCommand(NewViewListCmd(opts, &board))
	cmd.AddCommand(NewViewGetCmd(opts, &board))
	cmd.AddCommand(NewViewCreateCmd(opts, &board))
	cmd.AddCommand(NewViewUpdateCmd(opts, &board))
	cmd.AddCommand(NewViewDeleteCmd(opts, &board))

	return cmd
}
