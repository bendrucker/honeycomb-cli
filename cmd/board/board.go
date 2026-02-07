package board

import (
	"context"
	"encoding/json"
	"net/http"

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
	cmd.AddCommand(NewViewCmd(opts))
	cmd.AddCommand(NewCreateCmd(opts))
	cmd.AddCommand(NewUpdateCmd(opts))
	cmd.AddCommand(NewDeleteCmd(opts))

	return cmd
}

func keyEditor(key string) api.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		config.ApplyAuth(req, config.KeyConfig, key)
		return nil
	}
}

type boardListItem struct {
	ID           string `json:"id"                      yaml:"id"`
	Name         string `json:"name"                    yaml:"name"`
	Description  string `json:"description,omitempty"    yaml:"description,omitempty"`
	ColumnLayout string `json:"column_layout,omitempty"  yaml:"column_layout,omitempty"`
	URL          string `json:"url,omitempty"            yaml:"url,omitempty"`
}

type boardDetail struct {
	ID           string          `json:"id"                        yaml:"id"`
	Name         string          `json:"name"                      yaml:"name"`
	Description  string          `json:"description,omitempty"     yaml:"description,omitempty"`
	Type         string          `json:"type"                      yaml:"type"`
	ColumnLayout string          `json:"column_layout,omitempty"   yaml:"column_layout,omitempty"`
	URL          string          `json:"url,omitempty"             yaml:"url,omitempty"`
	Panels       json.RawMessage `json:"panels,omitempty"          yaml:"-"`
	PanelsAny    any             `json:"-"                         yaml:"panels,omitempty"`
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
		var panels any
		_ = json.Unmarshal(raw, &panels)
		d.PanelsAny = panels
	}
	return d
}
