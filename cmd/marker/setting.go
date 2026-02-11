package marker

import (
	"fmt"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type settingItem struct {
	ID        string `json:"id"                     yaml:"id"`
	Type      string `json:"type"                   yaml:"type"`
	Color     string `json:"color"                  yaml:"color"`
	CreatedAt string `json:"created_at,omitempty"    yaml:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"    yaml:"updated_at,omitempty"`
}

var settingListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(settingItem).ID }},
		{Header: "Type", Value: func(v any) string { return v.(settingItem).Type }},
		{Header: "Color", Value: func(v any) string { return v.(settingItem).Color }},
	},
}

func toSettingItem(s api.MarkerSetting) settingItem {
	item := settingItem{
		Type:  s.Type,
		Color: s.Color,
	}
	if s.Id != nil {
		item.ID = *s.Id
	}
	if s.CreatedAt != nil {
		item.CreatedAt = *s.CreatedAt
	}
	if s.UpdatedAt.IsSpecified() && !s.UpdatedAt.IsNull() {
		item.UpdatedAt = s.UpdatedAt.MustGet()
	}
	return item
}

func writeSettingDetail(opts *options.RootOptions, item settingItem) error {
	format := opts.ResolveFormat()
	if format != "table" {
		return opts.OutputWriter().Write(item, output.TableDef{})
	}

	tw := tabwriter.NewWriter(opts.IOStreams.Out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "ID:\t%s\n", item.ID)
	_, _ = fmt.Fprintf(tw, "Type:\t%s\n", item.Type)
	_, _ = fmt.Fprintf(tw, "Color:\t%s\n", item.Color)
	_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", item.CreatedAt)
	_, _ = fmt.Fprintf(tw, "Updated At:\t%s\n", item.UpdatedAt)
	return tw.Flush()
}

func NewSettingCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "setting",
		Short:   "Manage marker settings",
		Aliases: []string{"settings"},
	}

	cmd.AddCommand(NewSettingListCmd(opts, dataset))
	cmd.AddCommand(NewSettingCreateCmd(opts, dataset))
	cmd.AddCommand(NewSettingUpdateCmd(opts, dataset))
	cmd.AddCommand(NewSettingDeleteCmd(opts, dataset))

	return cmd
}
