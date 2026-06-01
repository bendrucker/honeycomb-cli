package query

import (
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type annotationItem struct {
	ID      string `json:"id" col:"ID"`
	Name    string `json:"name" col:"Name"`
	QueryID string `json:"query_id" col:"Query ID"`
	Source  string `json:"source,omitempty" col:"Source"`
}

type annotationDetail struct {
	ID          string `json:"id" detail:"ID"`
	Name        string `json:"name" detail:"Name"`
	Description string `json:"description,omitempty" detail:"Description"`
	QueryID     string `json:"query_id" detail:"Query ID"`
	Source      string `json:"source,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

var annotationListTable = output.TableFromTags[annotationItem]()

func annotationToDetail(a api.QueryAnnotation) annotationDetail {
	return annotationDetail{
		ID:          deref.String(a.Id),
		Name:        a.Name,
		Description: deref.String(a.Description),
		QueryID:     a.QueryId,
		Source:      deref.Enum(a.Source),
		CreatedAt:   deref.Time(a.CreatedAt),
		UpdatedAt:   deref.Time(a.UpdatedAt),
	}
}

func writeAnnotationDetail(opts *options.RootOptions, detail annotationDetail) error {
	fields := output.FieldsFromTags(detail)
	if detail.Source != "" {
		fields = append(fields, output.Field{Label: "Source", Value: detail.Source})
	}
	if detail.CreatedAt != "" {
		fields = append(fields, output.Field{Label: "Created At", Value: detail.CreatedAt})
	}
	if detail.UpdatedAt != "" {
		fields = append(fields, output.Field{Label: "Updated At", Value: detail.UpdatedAt})
	}
	return opts.OutputWriter().WriteFields(detail, fields)
}

func NewAnnotationCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "annotation",
		Short:   "Manage query annotations (saved queries)",
		Aliases: []string{"annotations"},
	}

	cmd.AddCommand(NewListCmd(opts, dataset))
	cmd.AddCommand(NewViewCmd(opts, dataset))
	cmd.AddCommand(NewCreateCmd(opts, dataset))
	cmd.AddCommand(NewUpdateCmd(opts, dataset))
	cmd.AddCommand(NewDeleteCmd(opts, dataset))

	return cmd
}
