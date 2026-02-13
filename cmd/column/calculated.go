package column

import (
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
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
	return calculatedItem{
		ID:          deref.String(c.Id),
		Alias:       c.Alias,
		Expression:  c.Expression,
		Description: deref.String(c.Description),
	}
}

func toCalculatedDetail(c api.CalculatedField) calculatedDetail {
	return calculatedDetail{
		ID:          deref.String(c.Id),
		Alias:       c.Alias,
		Expression:  c.Expression,
		Description: deref.String(c.Description),
		CreatedAt:   deref.String(c.CreatedAt),
		UpdatedAt:   deref.String(c.UpdatedAt),
	}
}

func writeCalculatedDetail(opts *options.RootOptions, c api.CalculatedField) error {
	d := toCalculatedDetail(c)
	return opts.OutputWriter().WriteFields(d, []output.Field{
		{Label: "ID", Value: d.ID},
		{Label: "Alias", Value: d.Alias},
		{Label: "Expression", Value: d.Expression},
		{Label: "Description", Value: d.Description},
		{Label: "Created At", Value: d.CreatedAt},
		{Label: "Updated At", Value: d.UpdatedAt},
	})
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
