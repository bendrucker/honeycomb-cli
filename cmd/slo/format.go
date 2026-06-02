package slo

import (
	"fmt"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
)

// targetPerMillion is an SLO target expressed in parts per million. It marshals
// as its underlying integer while rendering as a percentage in detail tables.
type targetPerMillion int

func (t targetPerMillion) FormatField() string {
	pct := float64(t) / 10000.0
	return fmt.Sprintf("%g%%", pct)
}

// timePeriodDays is an SLO time period in days. It marshals as its underlying
// integer while rendering with a trailing "d" in detail tables.
type timePeriodDays int

func (d timePeriodDays) FormatField() string {
	return fmt.Sprintf("%dd", d)
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
	ID               string           `json:"id" detail:"ID"`
	Name             string           `json:"name" detail:"Name"`
	Description      string           `json:"description,omitempty" detail:"Description"`
	TargetPerMillion targetPerMillion `json:"target_per_million" detail:"Target"`
	TimePeriodDays   timePeriodDays   `json:"time_period_days" detail:"Time Period"`
	SLIAlias         string           `json:"sli_alias" detail:"SLI Alias"`
	DatasetSlugs     []string         `json:"dataset_slugs,omitempty"`
	CreatedAt        string           `json:"created_at,omitempty"`
	UpdatedAt        string           `json:"updated_at,omitempty"`
	ResetAt          string           `json:"reset_at,omitempty"`

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
		TargetPerMillion: targetPerMillion(s.TargetPerMillion),
		TimePeriodDays:   timePeriodDays(s.TimePeriodDays),
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
	fields := output.FieldsFromTags(detail)
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
