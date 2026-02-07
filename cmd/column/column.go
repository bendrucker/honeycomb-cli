package column

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type columnItem struct {
	ID          string `json:"id"                     yaml:"id"`
	KeyName     string `json:"key_name"               yaml:"key_name"`
	Type        string `json:"type,omitempty"          yaml:"type,omitempty"`
	Description string `json:"description,omitempty"   yaml:"description,omitempty"`
	Hidden      bool   `json:"hidden"                  yaml:"hidden"`
	LastWritten string `json:"last_written,omitempty"  yaml:"last_written,omitempty"`
}

type columnDetail struct {
	ID          string `json:"id"                     yaml:"id"`
	KeyName     string `json:"key_name"               yaml:"key_name"`
	Type        string `json:"type,omitempty"          yaml:"type,omitempty"`
	Description string `json:"description,omitempty"   yaml:"description,omitempty"`
	Hidden      bool   `json:"hidden"                  yaml:"hidden"`
	LastWritten string `json:"last_written,omitempty"  yaml:"last_written,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"    yaml:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"    yaml:"updated_at,omitempty"`
}

var columnListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(columnItem).ID }},
		{Header: "Key Name", Value: func(v any) string { return v.(columnItem).KeyName }},
		{Header: "Type", Value: func(v any) string { return v.(columnItem).Type }},
		{Header: "Description", Value: func(v any) string { return v.(columnItem).Description }},
		{Header: "Hidden", Value: func(v any) string {
			if v.(columnItem).Hidden {
				return "yes"
			}
			return "no"
		}},
		{Header: "Last Written", Value: func(v any) string { return v.(columnItem).LastWritten }},
	},
}

func columnToDetail(c api.Column) columnDetail {
	d := columnDetail{
		KeyName: c.KeyName,
	}
	if c.Id != nil {
		d.ID = *c.Id
	}
	if c.Type != nil {
		d.Type = string(*c.Type)
	}
	if c.Description != nil {
		d.Description = *c.Description
	}
	if c.Hidden != nil {
		d.Hidden = *c.Hidden
	}
	if c.LastWritten != nil {
		d.LastWritten = *c.LastWritten
	}
	if c.CreatedAt != nil {
		d.CreatedAt = *c.CreatedAt
	}
	if c.UpdatedAt != nil {
		d.UpdatedAt = *c.UpdatedAt
	}
	return d
}

func writeColumnDetail(opts *options.RootOptions, c api.Column) error {
	d := columnToDetail(c)
	return opts.OutputWriter().WriteValue(d, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID:\t%s\n", d.ID)
		_, _ = fmt.Fprintf(tw, "Key Name:\t%s\n", d.KeyName)
		_, _ = fmt.Fprintf(tw, "Type:\t%s\n", d.Type)
		_, _ = fmt.Fprintf(tw, "Description:\t%s\n", d.Description)
		_, _ = fmt.Fprintf(tw, "Hidden:\t%v\n", d.Hidden)
		_, _ = fmt.Fprintf(tw, "Last Written:\t%s\n", d.LastWritten)
		_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", d.CreatedAt)
		_, _ = fmt.Fprintf(tw, "Updated At:\t%s\n", d.UpdatedAt)
		return tw.Flush()
	})
}

func NewCmd(opts *options.RootOptions) *cobra.Command {
	var dataset string

	cmd := &cobra.Command{
		Use:     "column",
		Short:   "Manage columns",
		Aliases: []string{"columns"},
	}

	cmd.PersistentFlags().StringVar(&dataset, "dataset", "", "Dataset slug (required)")
	_ = cmd.MarkPersistentFlagRequired("dataset")

	cmd.AddCommand(NewListCmd(opts, &dataset))
	cmd.AddCommand(NewViewCmd(opts, &dataset))
	cmd.AddCommand(NewCreateCmd(opts, &dataset))
	cmd.AddCommand(NewUpdateCmd(opts, &dataset))
	cmd.AddCommand(NewDeleteCmd(opts, &dataset))

	return cmd
}

func keyEditor(key string) api.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		config.ApplyAuth(req, config.KeyConfig, key)
		return nil
	}
}
