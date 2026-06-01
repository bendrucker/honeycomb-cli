package environment

import (
	"encoding/json"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var environmentColors = []string{
	string(api.Blue),
	string(api.Gold),
	string(api.Green),
	string(api.LightBlue),
	string(api.LightGold),
	string(api.LightGreen),
	string(api.LightPurple),
	string(api.LightRed),
	string(api.Purple),
	string(api.Red),
}

func colorFlagUsage() string {
	return "Environment color: " + strings.Join(environmentColors, ", ")
}

func NewCmd(opts *options.RootOptions) *cobra.Command {
	var team string

	cmd := &cobra.Command{
		Use:     "environment",
		Short:   "Manage environments",
		Aliases: []string{"environments", "env"},
	}

	cmd.PersistentFlags().StringVar(&team, "team", "", "Team slug")

	cmd.AddCommand(NewListCmd(opts, &team))
	cmd.AddCommand(NewGetCmd(opts, &team))
	cmd.AddCommand(NewCreateCmd(opts, &team))
	cmd.AddCommand(NewUpdateCmd(opts, &team))
	cmd.AddCommand(NewDeleteCmd(opts, &team))

	return cmd
}

type environmentItem struct {
	ID          string `json:"id" col:"ID"`
	Name        string `json:"name" col:"Name"`
	Slug        string `json:"slug" col:"Slug"`
	Description string `json:"description,omitempty" col:"Description"`
	Color       string `json:"color,omitempty" col:"Color"`
}

type environmentDetail struct {
	ID              string `json:"id" detail:"ID"`
	Name            string `json:"name" detail:"Name"`
	Slug            string `json:"slug" detail:"Slug"`
	Description     string `json:"description,omitempty" detail:"Description"`
	Color           string `json:"color,omitempty" detail:"Color"`
	DeleteProtected bool   `json:"delete_protected" detail:"Delete Protected"`
}

func colorString(c api.Environment_Attributes_Color) string {
	b, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	var s string
	if json.Unmarshal(b, &s) == nil {
		return s
	}
	return ""
}

func envToItem(e api.Environment) environmentItem {
	return environmentItem{
		ID:          e.Id,
		Name:        e.Attributes.Name,
		Slug:        e.Attributes.Slug,
		Description: e.Attributes.Description,
		Color:       colorString(e.Attributes.Color),
	}
}

func envToDetail(e api.Environment) environmentDetail {
	return environmentDetail{
		ID:              e.Id,
		Name:            e.Attributes.Name,
		Slug:            e.Attributes.Slug,
		Description:     e.Attributes.Description,
		Color:           colorString(e.Attributes.Color),
		DeleteProtected: e.Attributes.Settings.DeleteProtected,
	}
}

func writeEnvironmentDetail(opts *options.RootOptions, detail environmentDetail) error {
	return opts.OutputWriter().WriteFields(detail, output.FieldsFromTags(detail))
}
