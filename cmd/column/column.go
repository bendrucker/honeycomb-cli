package column

import (
	"strconv"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type columnItem struct {
	ID          string `json:"id"`
	KeyName     string `json:"key_name"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Hidden      bool   `json:"hidden"`
	LastWritten string `json:"last_written,omitempty"`
}

type columnDetail struct {
	ID          string `json:"id"`
	KeyName     string `json:"key_name"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Hidden      bool   `json:"hidden"`
	LastWritten string `json:"last_written,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
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
	return columnDetail{
		ID:          deref.String(c.Id),
		KeyName:     c.KeyName,
		Type:        deref.Enum(c.Type),
		Description: deref.String(c.Description),
		Hidden:      deref.Bool(c.Hidden),
		LastWritten: deref.String(c.LastWritten),
		CreatedAt:   deref.String(c.CreatedAt),
		UpdatedAt:   deref.String(c.UpdatedAt),
	}
}

func writeColumnDetail(opts *options.RootOptions, c api.Column) error {
	d := columnToDetail(c)
	return opts.OutputWriter().WriteFields(d, []output.Field{
		{Label: "ID", Value: d.ID},
		{Label: "Key Name", Value: d.KeyName},
		{Label: "Type", Value: d.Type},
		{Label: "Description", Value: d.Description},
		{Label: "Hidden", Value: strconv.FormatBool(d.Hidden)},
		{Label: "Last Written", Value: d.LastWritten},
		{Label: "Created At", Value: d.CreatedAt},
		{Label: "Updated At", Value: d.UpdatedAt},
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
	cmd.AddCommand(NewGetCmd(opts, &dataset))
	cmd.AddCommand(NewCreateCmd(opts, &dataset))
	cmd.AddCommand(NewUpdateCmd(opts, &dataset))
	cmd.AddCommand(NewDeleteCmd(opts, &dataset))
	cmd.AddCommand(NewCalculatedCmd(opts, &dataset))

	return cmd
}
