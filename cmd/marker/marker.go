package marker

import (
	"context"
	"fmt"
	"net/http"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type markerItem struct {
	ID        string `json:"id"                     yaml:"id"`
	Type      string `json:"type,omitempty"          yaml:"type,omitempty"`
	Message   string `json:"message,omitempty"       yaml:"message,omitempty"`
	URL       string `json:"url,omitempty"           yaml:"url,omitempty"`
	StartTime *int   `json:"start_time,omitempty"    yaml:"start_time,omitempty"`
	EndTime   *int   `json:"end_time,omitempty"      yaml:"end_time,omitempty"`
	Color     string `json:"color,omitempty"         yaml:"color,omitempty"`
	CreatedAt string `json:"created_at,omitempty"    yaml:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"    yaml:"updated_at,omitempty"`
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
	item := markerItem{
		StartTime: m.StartTime,
		EndTime:   m.EndTime,
	}
	if m.Id != nil {
		item.ID = *m.Id
	}
	if m.Type != nil {
		item.Type = *m.Type
	}
	if m.Message != nil {
		item.Message = *m.Message
	}
	if m.Url != nil {
		item.URL = *m.Url
	}
	if m.Color != nil {
		item.Color = *m.Color
	}
	if m.CreatedAt != nil {
		item.CreatedAt = *m.CreatedAt
	}
	if m.UpdatedAt != nil {
		item.UpdatedAt = *m.UpdatedAt
	}
	return item
}

func writeDetail(opts *options.RootOptions, item markerItem) error {
	format := opts.ResolveFormat()
	if format != "table" {
		return opts.OutputWriter().Write(item, output.TableDef{})
	}

	tw := tabwriter.NewWriter(opts.IOStreams.Out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "ID:\t%s\n", item.ID)
	_, _ = fmt.Fprintf(tw, "Type:\t%s\n", item.Type)
	_, _ = fmt.Fprintf(tw, "Message:\t%s\n", item.Message)
	_, _ = fmt.Fprintf(tw, "URL:\t%s\n", item.URL)
	if item.StartTime != nil {
		_, _ = fmt.Fprintf(tw, "Start Time:\t%d\n", *item.StartTime)
	}
	if item.EndTime != nil {
		_, _ = fmt.Fprintf(tw, "End Time:\t%d\n", *item.EndTime)
	}
	_, _ = fmt.Fprintf(tw, "Color:\t%s\n", item.Color)
	_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", item.CreatedAt)
	_, _ = fmt.Fprintf(tw, "Updated At:\t%s\n", item.UpdatedAt)
	return tw.Flush()
}

func findMarker(markers []api.Marker, id string) (api.Marker, error) {
	for _, m := range markers {
		if m.Id != nil && *m.Id == id {
			return m, nil
		}
	}
	return api.Marker{}, fmt.Errorf("marker %q not found", id)
}

func keyEditor(key string) api.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		config.ApplyAuth(req, config.KeyConfig, key)
		return nil
	}
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
	cmd.AddCommand(NewViewCmd(opts, &dataset))
	cmd.AddCommand(NewCreateCmd(opts, &dataset))
	cmd.AddCommand(NewUpdateCmd(opts, &dataset))
	cmd.AddCommand(NewDeleteCmd(opts, &dataset))

	return cmd
}
