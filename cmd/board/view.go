package board

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
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
	item := viewItem{}
	if v.Id != nil {
		item.ID = *v.Id
	}
	if v.Name != nil {
		item.Name = *v.Name
	}
	return item
}

func viewResponseToDetail(v api.BoardViewResponse) viewDetail {
	d := viewDetail{}
	if v.Id != nil {
		d.ID = *v.Id
	}
	if v.Name != nil {
		d.Name = *v.Name
	}
	if v.Filters != nil {
		raw, _ := json.Marshal(v.Filters)
		d.Filters = raw
	}
	return d
}

func writeViewDetail(opts *options.RootOptions, detail viewDetail) error {
	return opts.OutputWriter().WriteValue(detail, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID:\t%s\n", detail.ID)
		_, _ = fmt.Fprintf(tw, "Name:\t%s\n", detail.Name)
		return tw.Flush()
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
