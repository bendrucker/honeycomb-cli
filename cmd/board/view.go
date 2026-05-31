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
	return opts.OutputWriter().WriteFields(detail, output.FieldsFromTags(detail))
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
