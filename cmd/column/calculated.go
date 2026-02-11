package column

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type calculatedItem struct {
	ID          string `json:"id"                    yaml:"id"`
	Alias       string `json:"alias"                 yaml:"alias"`
	Expression  string `json:"expression"            yaml:"expression"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type calculatedDetail struct {
	ID          string `json:"id"                    yaml:"id"`
	Alias       string `json:"alias"                 yaml:"alias"`
	Expression  string `json:"expression"            yaml:"expression"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"  yaml:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"  yaml:"updated_at,omitempty"`
}

var calculatedListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(calculatedItem).ID }},
		{Header: "Alias", Value: func(v any) string { return v.(calculatedItem).Alias }},
		{Header: "Expression", Value: func(v any) string { return v.(calculatedItem).Expression }},
		{Header: "Description", Value: func(v any) string { return v.(calculatedItem).Description }},
	},
}

func toCalculatedItem(f api.CalculatedField) calculatedItem {
	item := calculatedItem{
		Alias:      f.Alias,
		Expression: f.Expression,
	}
	if f.Id != nil {
		item.ID = *f.Id
	}
	if f.Description != nil {
		item.Description = *f.Description
	}
	return item
}

func toCalculatedDetail(f api.CalculatedField) calculatedDetail {
	d := calculatedDetail{
		Alias:      f.Alias,
		Expression: f.Expression,
	}
	if f.Id != nil {
		d.ID = *f.Id
	}
	if f.Description != nil {
		d.Description = *f.Description
	}
	if f.CreatedAt != nil {
		d.CreatedAt = *f.CreatedAt
	}
	if f.UpdatedAt != nil {
		d.UpdatedAt = *f.UpdatedAt
	}
	return d
}

func writeCalculatedDetail(opts *options.RootOptions, f api.CalculatedField) error {
	d := toCalculatedDetail(f)
	return opts.OutputWriter().WriteValue(d, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID:\t%s\n", d.ID)
		_, _ = fmt.Fprintf(tw, "Alias:\t%s\n", d.Alias)
		_, _ = fmt.Fprintf(tw, "Expression:\t%s\n", d.Expression)
		_, _ = fmt.Fprintf(tw, "Description:\t%s\n", d.Description)
		_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", d.CreatedAt)
		_, _ = fmt.Fprintf(tw, "Updated At:\t%s\n", d.UpdatedAt)
		return tw.Flush()
	})
}

func NewCalculatedCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "calculated",
		Short:   "Manage calculated fields",
		Aliases: []string{"calc"},
	}

	cmd.AddCommand(NewCalculatedListCmd(opts, dataset))
	cmd.AddCommand(NewCalculatedGetCmd(opts, dataset))
	cmd.AddCommand(NewCalculatedCreateCmd(opts, dataset))
	cmd.AddCommand(NewCalculatedUpdateCmd(opts, dataset))
	cmd.AddCommand(NewCalculatedDeleteCmd(opts, dataset))

	return cmd
}
