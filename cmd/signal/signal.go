package signal

import (
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/oapi-codegen/nullable"
	"github.com/spf13/cobra"
)

var measuredSignals = []string{
	string(api.ErrorRate),
	string(api.Presence),
}

var sensitivities = []string{
	string(api.Low),
	string(api.Medium),
	string(api.High),
}

var statuses = []string{
	string(api.AnomalySignalStatusOnboarding),
	string(api.AnomalySignalStatusTraining),
	string(api.AnomalySignalStatusActive),
	string(api.AnomalySignalStatusOff),
	string(api.AnomalySignalStatusIneligible),
}

func NewCmd(opts *options.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "signal",
		Short:   "Manage anomaly detection signals",
		Aliases: []string{"signals"},
		Example: `  # List signals that are currently anomalous
  honeycomb signal list --anomalous

  # Get a signal by ID
  honeycomb signal get sig-abc123

  # Lower a signal's sensitivity
  honeycomb signal update sig-abc123 --sensitivity low`,
	}

	cmd.AddCommand(NewListCmd(opts))
	cmd.AddCommand(NewGetCmd(opts))
	cmd.AddCommand(NewUpdateCmd(opts))
	cmd.AddCommand(NewAnomaliesCmd(opts))

	return command.Group(cmd)
}

type signalItem struct {
	ID             string `json:"id" col:"ID"`
	ServiceName    string `json:"service_name,omitempty" col:"Service"`
	DatasetSlug    string `json:"dataset_slug,omitempty" col:"Dataset"`
	MeasuredSignal string `json:"measured_signal,omitempty" col:"Measured"`
	Status         string `json:"status,omitempty" col:"Status"`
	Enabled        bool   `json:"enabled" col:"Enabled"`
	Anomalous      bool   `json:"currently_anomalous" col:"Anomalous"`
}

type signalDetail struct {
	ID                   string          `json:"id" detail:"ID"`
	ServiceName          string          `json:"service_name,omitempty" detail:"Service"`
	DatasetSlug          string          `json:"dataset_slug,omitempty" detail:"Dataset"`
	EnvironmentSlug      string          `json:"environment_slug,omitempty" detail:"Environment"`
	MeasuredSignal       string          `json:"measured_signal,omitempty" detail:"Measured"`
	Status               string          `json:"status,omitempty" detail:"Status"`
	Enabled              bool            `json:"enabled" detail:"Enabled"`
	Sensitivity          string          `json:"sensitivity,omitempty" detail:"Sensitivity"`
	Anomalous            bool            `json:"currently_anomalous" detail:"Anomalous"`
	AutoInvestigate      bool            `json:"auto_investigate" detail:"Auto Investigate"`
	LastAnomalyStartedAt *int            `json:"last_anomaly_started_at,omitempty"`
	LastAnomalyEndedAt   *int            `json:"last_anomaly_ended_at,omitempty"`
	Recipients           []recipientItem `json:"recipients,omitempty"`
	CreatedAt            string          `json:"created_at,omitempty" detail:"Created At"`
	UpdatedAt            string          `json:"updated_at,omitempty" detail:"Updated At"`
}

type recipientItem struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type,omitempty"`
	Target string `json:"target,omitempty"`
	Muted  bool   `json:"muted,omitempty"`
}

func toItem(s api.Signal) signalItem {
	return signalItem{
		ID:             deref.String(s.Id),
		ServiceName:    deref.String(s.ServiceName),
		DatasetSlug:    deref.String(s.DatasetSlug),
		MeasuredSignal: deref.Enum(s.MeasuredSignal),
		Status:         deref.Enum(s.Status),
		Enabled:        s.Enabled,
		Anomalous:      deref.Bool(s.CurrentlyAnomalous),
	}
}

func toDetail(s api.SignalDetailResponse) signalDetail {
	d := signalDetail{
		ID:                   deref.String(s.Id),
		ServiceName:          deref.String(s.ServiceName),
		DatasetSlug:          deref.String(s.DatasetSlug),
		EnvironmentSlug:      deref.String(s.EnvironmentSlug),
		MeasuredSignal:       deref.Enum(s.MeasuredSignal),
		Status:               deref.Enum(s.Status),
		Enabled:              s.Enabled,
		Sensitivity:          string(s.Sensitivity),
		Anomalous:            deref.Bool(s.CurrentlyAnomalous),
		AutoInvestigate:      deref.Bool(s.AutoInvestigate),
		LastAnomalyStartedAt: nullableInt(s.LastAnomalyStartedAt),
		LastAnomalyEndedAt:   nullableInt(s.LastAnomalyEndedAt),
		CreatedAt:            deref.Time(s.CreatedAt),
		UpdatedAt:            deref.Time(s.UpdatedAt),
	}

	for _, r := range s.Recipients {
		item := recipientItem{
			ID:     deref.String(r.Id),
			Type:   deref.Enum(r.Type),
			Target: deref.String(r.Target),
		}
		if r.Details != nil {
			item.Muted = deref.Bool(r.Details.Muted)
		}
		d.Recipients = append(d.Recipients, item)
	}

	return d
}

func nullableInt(n nullable.Nullable[int]) *int {
	if !n.IsSpecified() || n.IsNull() {
		return nil
	}
	v, err := n.Get()
	if err != nil {
		return nil
	}
	return &v
}

func writeSignalDetail(opts *options.RootOptions, detail signalDetail) error {
	fields := output.FieldsFromTags(detail)

	if detail.LastAnomalyStartedAt != nil {
		fields = append(fields, output.Field{Label: "Last Anomaly Started", Value: formatEpoch(*detail.LastAnomalyStartedAt)})
	}

	if detail.LastAnomalyEndedAt != nil {
		fields = append(fields, output.Field{Label: "Last Anomaly Ended", Value: formatEpoch(*detail.LastAnomalyEndedAt)})
	}

	if len(detail.Recipients) > 0 {
		labels := make([]string, len(detail.Recipients))
		for i, r := range detail.Recipients {
			labels[i] = recipientLabel(r)
		}
		fields = append(fields, output.Field{Label: "Recipients", Value: strings.Join(labels, ", ")})
	}

	return opts.OutputWriter().WriteFields(detail, fields)
}

func recipientLabel(r recipientItem) string {
	label := r.Target
	if label == "" {
		label = r.ID
	}
	if r.Muted {
		label += " (muted)"
	}
	return label
}
