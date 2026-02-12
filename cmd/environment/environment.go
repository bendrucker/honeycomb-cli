package environment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	var team string

	cmd := &cobra.Command{
		Use:     "environment",
		Short:   "Manage environments",
		Aliases: []string{"environments", "env"},
	}

	cmd.PersistentFlags().StringVar(&team, "team", "", "Team slug (required)")
	_ = cmd.MarkPersistentFlagRequired("team")

	cmd.AddCommand(NewListCmd(opts, &team))
	cmd.AddCommand(NewGetCmd(opts, &team))
	cmd.AddCommand(NewCreateCmd(opts, &team))
	cmd.AddCommand(NewUpdateCmd(opts, &team))
	cmd.AddCommand(NewDeleteCmd(opts, &team))

	return cmd
}

func keyEditor(key string) api.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		config.ApplyAuth(req, config.KeyManagement, key)
		return nil
	}
}

type environmentItem struct {
	ID          string `json:"id"                     yaml:"id"`
	Name        string `json:"name"                   yaml:"name"`
	Slug        string `json:"slug"                   yaml:"slug"`
	Description string `json:"description,omitempty"   yaml:"description,omitempty"`
	Color       string `json:"color,omitempty"         yaml:"color,omitempty"`
}

type environmentDetail struct {
	ID              string `json:"id"                      yaml:"id"`
	Name            string `json:"name"                    yaml:"name"`
	Slug            string `json:"slug"                    yaml:"slug"`
	Description     string `json:"description,omitempty"   yaml:"description,omitempty"`
	Color           string `json:"color,omitempty"         yaml:"color,omitempty"`
	DeleteProtected bool   `json:"delete_protected"        yaml:"delete_protected"`
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
	return opts.OutputWriter().WriteValue(detail, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID:\t%s\n", detail.ID)
		_, _ = fmt.Fprintf(tw, "Name:\t%s\n", detail.Name)
		_, _ = fmt.Fprintf(tw, "Slug:\t%s\n", detail.Slug)
		_, _ = fmt.Fprintf(tw, "Description:\t%s\n", detail.Description)
		_, _ = fmt.Fprintf(tw, "Color:\t%s\n", detail.Color)
		_, _ = fmt.Fprintf(tw, "Delete Protected:\t%t\n", detail.DeleteProtected)
		return tw.Flush()
	})
}
