package slo

import (
	"fmt"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
)

func formatTarget(targetPerMillion int) string {
	pct := float64(targetPerMillion) / 10000.0
	return fmt.Sprintf("%g%%", pct)
}

func formatTimePeriod(days int) string {
	return fmt.Sprintf("%dd", days)
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-3]) + "..."
}

type sloItem struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	TargetPerMillion int    `json:"target_per_million"`
	TimePeriodDays   int    `json:"time_period_days"`
	SLIAlias         string `json:"sli_alias"`
	Description      string `json:"description,omitempty"`
}

type sloDetail struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description,omitempty"`
	TargetPerMillion int      `json:"target_per_million"`
	TimePeriodDays   int      `json:"time_period_days"`
	SLIAlias         string   `json:"sli_alias"`
	DatasetSlugs     []string `json:"dataset_slugs,omitempty"`
	CreatedAt        string   `json:"created_at,omitempty"`
	UpdatedAt        string   `json:"updated_at,omitempty"`
	ResetAt          string   `json:"reset_at,omitempty"`

	// Detailed fields (only populated with --detailed)
	Compliance      *float64 `json:"compliance,omitempty"`
	BudgetRemaining *float64 `json:"budget_remaining,omitempty"`
}

// sloDetailedResponse extends api.SLO with the detailed fields that
// are not generated (SLODetailedResponse = SLO is a type alias).
type sloDetailedResponse struct {
	api.SLO
	Compliance      *float64 `json:"compliance,omitempty"`
	BudgetRemaining *float64 `json:"budget_remaining,omitempty"`
}

func sloToDetail(s api.SLO) sloDetail {
	d := sloDetail{
		ID:               deref.String(s.Id),
		Name:             s.Name,
		Description:      deref.String(s.Description),
		TargetPerMillion: s.TargetPerMillion,
		TimePeriodDays:   s.TimePeriodDays,
		SLIAlias:         s.Sli.Alias,
		CreatedAt:        deref.Time(s.CreatedAt),
		UpdatedAt:        deref.Time(s.UpdatedAt),
	}
	if s.ResetAt.IsSpecified() && !s.ResetAt.IsNull() {
		d.ResetAt = s.ResetAt.MustGet().Format("2006-01-02T15:04:05Z07:00")
	}
	if s.DatasetSlugs != nil {
		for _, v := range *s.DatasetSlugs {
			if slug, ok := v.(string); ok {
				d.DatasetSlugs = append(d.DatasetSlugs, slug)
			}
		}
	}
	return d
}

func detailedToDetail(s sloDetailedResponse) sloDetail {
	d := sloToDetail(s.SLO)
	d.Compliance = s.Compliance
	d.BudgetRemaining = s.BudgetRemaining
	return d
}

func writeSloDetail(opts *options.RootOptions, detail sloDetail) error {
	fields := []output.Field{
		{Label: "ID", Value: detail.ID},
		{Label: "Name", Value: detail.Name},
		{Label: "Description", Value: detail.Description},
		{Label: "SLI Alias", Value: detail.SLIAlias},
		{Label: "Target", Value: formatTarget(detail.TargetPerMillion)},
		{Label: "Time Period", Value: formatTimePeriod(detail.TimePeriodDays)},
	}
	if len(detail.DatasetSlugs) > 0 {
		fields = append(fields, output.Field{Label: "Datasets", Value: strings.Join(detail.DatasetSlugs, ", ")})
	}
	if detail.CreatedAt != "" {
		fields = append(fields, output.Field{Label: "Created At", Value: detail.CreatedAt})
	}
	if detail.UpdatedAt != "" {
		fields = append(fields, output.Field{Label: "Updated At", Value: detail.UpdatedAt})
	}
	if detail.ResetAt != "" {
		fields = append(fields, output.Field{Label: "Reset At", Value: detail.ResetAt})
	}
	if detail.Compliance != nil {
		fields = append(fields, output.Field{Label: "Compliance", Value: fmt.Sprintf("%g%%", *detail.Compliance)})
	}
	if detail.BudgetRemaining != nil {
		fields = append(fields, output.Field{Label: "Budget Remaining", Value: fmt.Sprintf("%g", *detail.BudgetRemaining)})
	}
	return opts.OutputWriter().WriteFields(detail, fields)
}
