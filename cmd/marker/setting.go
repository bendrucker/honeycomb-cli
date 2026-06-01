package marker

import (
	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type settingItem struct {
	ID        string `json:"id" col:"ID" detail:"ID"`
	Type      string `json:"type" col:"Type" detail:"Type"`
	Color     string `json:"color" col:"Color" detail:"Color"`
	CreatedAt string `json:"created_at,omitempty" detail:"Created At"`
	UpdatedAt string `json:"updated_at,omitempty" detail:"Updated At"`
}

var settingListTable = output.TableFromTags[settingItem]()

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
	return opts.OutputWriter().WriteFields(item, output.FieldsFromTags(item))
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

	return command.Group(cmd)
}
