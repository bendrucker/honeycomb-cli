package query

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
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
	d := annotationDetail{
		Name:    a.Name,
		QueryID: a.QueryId,
	}
	if a.Id != nil {
		d.ID = *a.Id
	}
	if a.Description != nil {
		d.Description = *a.Description
	}
	if a.Source != nil {
		d.Source = string(*a.Source)
	}
	if a.CreatedAt != nil {
		d.CreatedAt = a.CreatedAt.Format(time.RFC3339)
	}
	if a.UpdatedAt != nil {
		d.UpdatedAt = a.UpdatedAt.Format(time.RFC3339)
	}
	return d
}

func writeAnnotationDetail(opts *options.RootOptions, detail annotationDetail) error {
	return opts.OutputWriter().WriteValue(detail, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID:\t%s\n", detail.ID)
		_, _ = fmt.Fprintf(tw, "Name:\t%s\n", detail.Name)
		_, _ = fmt.Fprintf(tw, "Description:\t%s\n", detail.Description)
		_, _ = fmt.Fprintf(tw, "Query ID:\t%s\n", detail.QueryID)
		if detail.Source != "" {
			_, _ = fmt.Fprintf(tw, "Source:\t%s\n", detail.Source)
		}
		if detail.CreatedAt != "" {
			_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", detail.CreatedAt)
		}
		if detail.UpdatedAt != "" {
			_, _ = fmt.Fprintf(tw, "Updated At:\t%s\n", detail.UpdatedAt)
		}
		return tw.Flush()
	})
}
