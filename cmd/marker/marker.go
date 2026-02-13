package marker

import (
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type markerItem struct {
	ID        string `json:"id"`
	Type      string `json:"type,omitempty"`
	Message   string `json:"message,omitempty"`
	URL       string `json:"url,omitempty"`
	StartTime *int   `json:"start_time,omitempty"`
	EndTime   *int   `json:"end_time,omitempty"`
	Color     string `json:"color,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

var markerListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(markerItem).ID }},
		{Header: "Type", Value: func(v any) string { return v.(markerItem).Type }},
		{Header: "Message", Value: func(v any) string { return v.(markerItem).Message }},
		{Header: "Start Time", Value: func(v any) string {
			if st := v.(markerItem).StartTime; st != nil {
				return fmt.Sprintf("%d", *st)
			}
			return ""
		}},
		{Header: "End Time", Value: func(v any) string {
			if et := v.(markerItem).EndTime; et != nil {
				return fmt.Sprintf("%d", *et)
			}
			return ""
		}},
		{Header: "Color", Value: func(v any) string { return v.(markerItem).Color }},
	},
}

func markerToItem(m api.Marker) markerItem {
	return markerItem{
		ID:        deref.String(m.Id),
		Type:      deref.String(m.Type),
		Message:   deref.String(m.Message),
		URL:       deref.String(m.Url),
		StartTime: m.StartTime,
		EndTime:   m.EndTime,
		Color:     deref.String(m.Color),
		CreatedAt: deref.String(m.CreatedAt),
		UpdatedAt: deref.String(m.UpdatedAt),
	}
}

func writeDetail(opts *options.RootOptions, item markerItem) error {
	fields := []output.Field{
		{Label: "ID", Value: item.ID},
		{Label: "Type", Value: item.Type},
		{Label: "Message", Value: item.Message},
		{Label: "URL", Value: item.URL},
	}
	if item.StartTime != nil {
		fields = append(fields, output.Field{Label: "Start Time", Value: fmt.Sprintf("%d", *item.StartTime)})
	}
	if item.EndTime != nil {
		fields = append(fields, output.Field{Label: "End Time", Value: fmt.Sprintf("%d", *item.EndTime)})
	}
	fields = append(fields,
		output.Field{Label: "Color", Value: item.Color},
		output.Field{Label: "Created At", Value: item.CreatedAt},
		output.Field{Label: "Updated At", Value: item.UpdatedAt},
	)
	return opts.OutputWriter().WriteFields(item, fields)
}

func findMarker(markers []api.Marker, id string) (api.Marker, error) {
	for _, m := range markers {
		if m.Id != nil && *m.Id == id {
			return m, nil
		}
	}
	return api.Marker{}, fmt.Errorf("marker %q not found", id)
}

func NewCmd(opts *options.RootOptions) *cobra.Command {
	var dataset string

	cmd := &cobra.Command{
		Use:     "marker",
		Short:   "Manage markers",
		Aliases: []string{"markers"},
	}

	cmd.PersistentFlags().StringVar(&dataset, "dataset", "", "Dataset slug (required)")
	_ = cmd.MarkPersistentFlagRequired("dataset")

	cmd.AddCommand(NewListCmd(opts, &dataset))
	cmd.AddCommand(NewCreateCmd(opts, &dataset))
	cmd.AddCommand(NewUpdateCmd(opts, &dataset))
	cmd.AddCommand(NewDeleteCmd(opts, &dataset))
	cmd.AddCommand(NewSettingCmd(opts, &dataset))

	return cmd
}
