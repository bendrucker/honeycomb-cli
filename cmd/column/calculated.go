package column

import (
	"fmt"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type calculatedItem struct {
	ID          string `json:"id"`
	Alias       string `json:"alias"`
	Expression  string `json:"expression"`
	Description string `json:"description,omitempty"`
}

type calculatedDetail struct {
	ID          string `json:"id"`
	Alias       string `json:"alias"`
	Expression  string `json:"expression"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

var calculatedListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(calculatedItem).ID }},
		{Header: "Alias", Value: func(v any) string { return v.(calculatedItem).Alias }},
		{Header: "Expression", Value: func(v any) string { return v.(calculatedItem).Expression }},
		{Header: "Description", Value: func(v any) string { return v.(calculatedItem).Description }},
	},
}

func toCalculatedItem(c api.CalculatedField) calculatedItem {
	item := calculatedItem{
		Alias:      c.Alias,
		Expression: c.Expression,
	}
	if c.Id != nil {
		item.ID = *c.Id
	}
	if c.Description != nil {
		item.Description = *c.Description
	}
	return item
}

func toCalculatedDetail(c api.CalculatedField) calculatedDetail {
	d := calculatedDetail{
		Alias:      c.Alias,
		Expression: c.Expression,
	}
	if c.Id != nil {
		d.ID = *c.Id
	}
	if c.Description != nil {
		d.Description = *c.Description
	}
	if c.CreatedAt != nil {
		d.CreatedAt = *c.CreatedAt
	}
	if c.UpdatedAt != nil {
		d.UpdatedAt = *c.UpdatedAt
	}
	return d
}

func writeCalculatedDetail(opts *options.RootOptions, c api.CalculatedField) error {
	d := toCalculatedDetail(c)
	format := opts.ResolveFormat()
	if format != "table" {
		return opts.OutputWriter().Write(d, output.TableDef{})
	}

	tw := tabwriter.NewWriter(opts.IOStreams.Out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "ID:\t%s\n", d.ID)
	_, _ = fmt.Fprintf(tw, "Alias:\t%s\n", d.Alias)
	_, _ = fmt.Fprintf(tw, "Expression:\t%s\n", d.Expression)
	_, _ = fmt.Fprintf(tw, "Description:\t%s\n", d.Description)
	_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", d.CreatedAt)
	_, _ = fmt.Fprintf(tw, "Updated At:\t%s\n", d.UpdatedAt)
	return tw.Flush()
}

func NewCalculatedCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "calculated",
		Short:   "Manage calculated columns",
		Aliases: []string{"calc"},
	}

	cmd.AddCommand(NewCalculatedListCmd(opts, dataset))
	cmd.AddCommand(NewCalculatedGetCmd(opts, dataset))
	cmd.AddCommand(NewCalculatedCreateCmd(opts, dataset))
	cmd.AddCommand(NewCalculatedUpdateCmd(opts, dataset))
	cmd.AddCommand(NewCalculatedDeleteCmd(opts, dataset))

	return cmd
}
