package slo

import (
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type burnAlertItem struct {
	ID        string `json:"id"                   yaml:"id"`
	AlertType string `json:"alert_type"           yaml:"alert_type"`
	SloID     string `json:"slo_id,omitempty"     yaml:"slo_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty" yaml:"created_at,omitempty"`
}

type burnAlertDetail struct {
	ID                                    string `json:"id"                                                  yaml:"id"`
	AlertType                             string `json:"alert_type"                                          yaml:"alert_type"`
	Description                           string `json:"description,omitempty"                               yaml:"description,omitempty"`
	SloID                                 string `json:"slo_id,omitempty"                                    yaml:"slo_id,omitempty"`
	CreatedAt                             string `json:"created_at,omitempty"                                yaml:"created_at,omitempty"`
	UpdatedAt                             string `json:"updated_at,omitempty"                                yaml:"updated_at,omitempty"`
	ExhaustionMinutes                     *int   `json:"exhaustion_minutes,omitempty"                        yaml:"exhaustion_minutes,omitempty"`
	BudgetRateDecreaseThresholdPerMillion *int   `json:"budget_rate_decrease_threshold_per_million,omitempty" yaml:"budget_rate_decrease_threshold_per_million,omitempty"`
	BudgetRateWindowMinutes               *int   `json:"budget_rate_window_minutes,omitempty"                yaml:"budget_rate_window_minutes,omitempty"`
}

var burnAlertListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(burnAlertItem).ID }},
		{Header: "Alert Type", Value: func(v any) string { return v.(burnAlertItem).AlertType }},
		{Header: "SLO ID", Value: func(v any) string { return v.(burnAlertItem).SloID }},
		{Header: "Created At", Value: func(v any) string { return v.(burnAlertItem).CreatedAt }},
	},
}

func NewBurnAlertCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "burn-alert",
		Short:   "Manage SLO burn alerts",
		Aliases: []string{"burn-alerts"},
	}

	cmd.AddCommand(NewBurnAlertListCmd(opts, dataset))
	cmd.AddCommand(NewBurnAlertGetCmd(opts, dataset))
	cmd.AddCommand(NewBurnAlertCreateCmd(opts, dataset))
	cmd.AddCommand(NewBurnAlertUpdateCmd(opts, dataset))
	cmd.AddCommand(NewBurnAlertDeleteCmd(opts, dataset))

	return cmd
}
