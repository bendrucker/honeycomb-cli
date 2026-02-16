package board

import (
	"encoding/json"
	"fmt"
	"io"

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
	}

	cmd.AddCommand(NewListCmd(opts))
	cmd.AddCommand(NewGetCmd(opts))
	cmd.AddCommand(NewCreateCmd(opts))
	cmd.AddCommand(NewUpdateCmd(opts))
	cmd.AddCommand(NewDeleteCmd(opts))
	cmd.AddCommand(NewViewCmd(opts))

	return cmd
}

type boardListItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	ColumnLayout string `json:"column_layout,omitempty"`
	URL          string `json:"url,omitempty"`
}

type boardDetail struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description,omitempty"`
	Type          string          `json:"type"`
	ColumnLayout  string          `json:"column_layout,omitempty"`
	URL           string          `json:"url,omitempty"`
	PresetFilters json.RawMessage `json:"preset_filters,omitempty"`
	Panels        json.RawMessage `json:"panels,omitempty"`
}

func writeBoardDetail(opts *options.RootOptions, detail boardDetail) error {
	return opts.OutputWriter().WriteFields(detail, []output.Field{
		{Label: "ID", Value: detail.ID},
		{Label: "Name", Value: detail.Name},
		{Label: "Description", Value: detail.Description},
		{Label: "Type", Value: detail.Type},
		{Label: "Column Layout", Value: detail.ColumnLayout},
		{Label: "URL", Value: detail.URL},
		{Label: "Preset Filters", Value: string(detail.PresetFilters)},
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
		ID:           deref.String(b.Id),
		Name:         b.Name,
		Type:         string(b.Type),
		Description:  deref.String(b.Description),
		ColumnLayout: deref.Enum(b.LayoutGeneration),
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
