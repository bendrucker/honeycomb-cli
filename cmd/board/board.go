package board

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
		Use:     "board",
		Short:   "Manage boards",
		Aliases: []string{"boards"},
	}

	cmd.AddCommand(NewListCmd(opts))
	cmd.AddCommand(NewGetCmd(opts))
	cmd.AddCommand(NewCreateCmd(opts))
	cmd.AddCommand(NewUpdateCmd(opts))
	cmd.AddCommand(NewDeleteCmd(opts))
	cmd.AddCommand(NewViewCmd(opts))

	return cmd
}

func keyEditor(key string) api.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		config.ApplyAuth(req, config.KeyConfig, key)
		return nil
	}
}

type boardListItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	ColumnLayout string `json:"column_layout,omitempty"`
	URL          string `json:"url,omitempty"`
}

type boardDetail struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	Type         string          `json:"type"`
	ColumnLayout string          `json:"column_layout,omitempty"`
	URL          string          `json:"url,omitempty"`
	Panels       json.RawMessage `json:"panels,omitempty"`
}

func writeBoardDetail(opts *options.RootOptions, detail boardDetail) error {
	return opts.OutputWriter().WriteValue(detail, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID:\t%s\n", detail.ID)
		_, _ = fmt.Fprintf(tw, "Name:\t%s\n", detail.Name)
		_, _ = fmt.Fprintf(tw, "Description:\t%s\n", detail.Description)
		_, _ = fmt.Fprintf(tw, "Type:\t%s\n", detail.Type)
		_, _ = fmt.Fprintf(tw, "Column Layout:\t%s\n", detail.ColumnLayout)
		_, _ = fmt.Fprintf(tw, "URL:\t%s\n", detail.URL)
		return tw.Flush()
	})
}

func readBoardJSON(r io.Reader) (api.Board, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return api.Board{}, fmt.Errorf("reading input: %w", err)
	}

	var board api.Board
	if err := json.Unmarshal(data, &board); err != nil {
		return api.Board{}, fmt.Errorf("parsing board JSON: %w", err)
	}
	return board, nil
}

func boardToDetail(b api.Board) boardDetail {
	d := boardDetail{
		Name: b.Name,
		Type: string(b.Type),
	}
	if b.Id != nil {
		d.ID = *b.Id
	}
	if b.Description != nil {
		d.Description = *b.Description
	}
	if b.LayoutGeneration != nil {
		d.ColumnLayout = string(*b.LayoutGeneration)
	}
	if b.Links != nil && b.Links.BoardUrl != nil {
		d.URL = *b.Links.BoardUrl
	}
	if b.Panels != nil {
		raw, _ := json.Marshal(b.Panels)
		d.Panels = raw
	}
	return d
}
