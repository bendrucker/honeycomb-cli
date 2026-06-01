package slo

import (
	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type burnAlertItem struct {
	ID        string `json:"id" col:"ID"`
	AlertType string `json:"alert_type" col:"Alert Type"`
	SloID     string `json:"slo_id,omitempty" col:"SLO ID"`
	CreatedAt string `json:"created_at,omitempty" col:"Created At"`
}

type burnAlertDetail struct {
	ID                                    string `json:"id" detail:"ID"`
	AlertType                             string `json:"alert_type" detail:"Alert Type"`
	Description                           string `json:"description,omitempty"`
	SloID                                 string `json:"slo_id,omitempty"`
	CreatedAt                             string `json:"created_at,omitempty"`
	UpdatedAt                             string `json:"updated_at,omitempty"`
	ExhaustionMinutes                     *int   `json:"exhaustion_minutes,omitempty"`
	BudgetRateDecreaseThresholdPerMillion *int   `json:"budget_rate_decrease_threshold_per_million,omitempty"`
	BudgetRateWindowMinutes               *int   `json:"budget_rate_window_minutes,omitempty"`
}

var burnAlertListTable = output.TableFromTags[burnAlertItem]()

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

	return command.Group(cmd)
}
