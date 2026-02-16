package environment

import (
	"encoding/json"
	"strconv"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

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
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

type environmentDetail struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	Description     string `json:"description,omitempty"`
	Color           string `json:"color,omitempty"`
	DeleteProtected bool   `json:"delete_protected"`
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
	return opts.OutputWriter().WriteFields(detail, []output.Field{
		{Label: "ID", Value: detail.ID},
		{Label: "Name", Value: detail.Name},
		{Label: "Slug", Value: detail.Slug},
		{Label: "Description", Value: detail.Description},
		{Label: "Color", Value: detail.Color},
		{Label: "Delete Protected", Value: strconv.FormatBool(detail.DeleteProtected)},
	})
}
