package marker

import (
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type settingItem struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
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
		ID:        deref.String(s.Id),
		Type:      s.Type,
		Color:     s.Color,
		CreatedAt: deref.String(s.CreatedAt),
	}
	if s.UpdatedAt.IsSpecified() && !s.UpdatedAt.IsNull() {
		item.UpdatedAt = s.UpdatedAt.MustGet()
	}
	return item
}

func writeSettingDetail(opts *options.RootOptions, item settingItem) error {
	return opts.OutputWriter().WriteFields(item, []output.Field{
		{Label: "ID", Value: item.ID},
		{Label: "Type", Value: item.Type},
		{Label: "Color", Value: item.Color},
		{Label: "Created At", Value: item.CreatedAt},
		{Label: "Updated At", Value: item.UpdatedAt},
	})
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
