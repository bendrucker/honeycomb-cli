package column

import (
	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type calculatedItem struct {
	ID          string `json:"id" col:"ID"`
	Alias       string `json:"alias" col:"Alias"`
	Expression  string `json:"expression" col:"Expression"`
	Description string `json:"description,omitempty" col:"Description"`
}

type calculatedDetail struct {
	ID          string `json:"id" detail:"ID"`
	Alias       string `json:"alias" detail:"Alias"`
	Expression  string `json:"expression" detail:"Expression"`
	Description string `json:"description,omitempty" detail:"Description"`
	CreatedAt   string `json:"created_at,omitempty" detail:"Created At"`
	UpdatedAt   string `json:"updated_at,omitempty" detail:"Updated At"`
}

var calculatedListTable = output.TableFromTags[calculatedItem]()

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
	return opts.OutputWriter().WriteFields(d, output.FieldsFromTags(d))
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

	return command.Group(cmd)
}
