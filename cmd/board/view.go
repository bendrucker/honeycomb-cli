package board

import (
	"encoding/json"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type viewItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type viewDetail struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Filters json.RawMessage `json:"filters,omitempty"`
}

var viewListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(viewItem).ID }},
		{Header: "Name", Value: func(v any) string { return v.(viewItem).Name }},
	},
}

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
	return opts.OutputWriter().WriteFields(detail, []output.Field{
		{Label: "ID", Value: detail.ID},
		{Label: "Name", Value: detail.Name},
	})
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
