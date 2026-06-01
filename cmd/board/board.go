package board

import (
	"encoding/json"

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
	fields = append(fields, output.Field{Label: "Preset Filters", Value: string(detail.PresetFilters)})
	return opts.OutputWriter().WriteFields(detail, fields)
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
