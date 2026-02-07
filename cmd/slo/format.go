package slo

import (
	"fmt"
	"time"

	"github.com/bendrucker/honeycomb-cli/internal/api"
)

func formatTarget(targetPerMillion int) string {
	pct := float64(targetPerMillion) / 10000.0
	return fmt.Sprintf("%g%%", pct)
}

func formatTimePeriod(days int) string {
	return fmt.Sprintf("%dd", days)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

type sloItem struct {
	ID               string `json:"id"                        yaml:"id"`
	Name             string `json:"name"                      yaml:"name"`
	TargetPerMillion int    `json:"target_per_million"         yaml:"target_per_million"`
	TimePeriodDays   int    `json:"time_period_days"           yaml:"time_period_days"`
	SLIAlias         string `json:"sli_alias"                  yaml:"sli_alias"`
	Description      string `json:"description,omitempty"      yaml:"description,omitempty"`
}

type sloDetail struct {
	ID               string   `json:"id"                         yaml:"id"`
	Name             string   `json:"name"                       yaml:"name"`
	Description      string   `json:"description,omitempty"      yaml:"description,omitempty"`
	TargetPerMillion int      `json:"target_per_million"         yaml:"target_per_million"`
	TimePeriodDays   int      `json:"time_period_days"           yaml:"time_period_days"`
	SLIAlias         string   `json:"sli_alias"                  yaml:"sli_alias"`
	DatasetSlugs     []string `json:"dataset_slugs,omitempty"    yaml:"dataset_slugs,omitempty"`
	CreatedAt        string   `json:"created_at,omitempty"       yaml:"created_at,omitempty"`
	UpdatedAt        string   `json:"updated_at,omitempty"       yaml:"updated_at,omitempty"`
	ResetAt          string   `json:"reset_at,omitempty"         yaml:"reset_at,omitempty"`

	// Detailed fields (only populated with --detailed)
	Compliance      *float64 `json:"compliance,omitempty"       yaml:"compliance,omitempty"`
	BudgetRemaining *float64 `json:"budget_remaining,omitempty" yaml:"budget_remaining,omitempty"`
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
		Name:             s.Name,
		TargetPerMillion: s.TargetPerMillion,
		TimePeriodDays:   s.TimePeriodDays,
		SLIAlias:         s.Sli.Alias,
	}
	if s.Id != nil {
		d.ID = *s.Id
	}
	if s.Description != nil {
		d.Description = *s.Description
	}
	if s.CreatedAt != nil {
		d.CreatedAt = s.CreatedAt.Format(time.RFC3339)
	}
	if s.UpdatedAt != nil {
		d.UpdatedAt = s.UpdatedAt.Format(time.RFC3339)
	}
	if s.ResetAt.IsSpecified() && !s.ResetAt.IsNull() {
		d.ResetAt = s.ResetAt.MustGet().Format(time.RFC3339)
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
