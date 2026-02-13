package query

import (
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
)

type annotationItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	QueryID string `json:"query_id"`
	Source  string `json:"source,omitempty"`
}

type annotationDetail struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	QueryID     string `json:"query_id"`
	Source      string `json:"source,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

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
	fields := []output.Field{
		{Label: "ID", Value: detail.ID},
		{Label: "Name", Value: detail.Name},
		{Label: "Description", Value: detail.Description},
		{Label: "Query ID", Value: detail.QueryID},
	}
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
