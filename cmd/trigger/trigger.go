package trigger

import (
	"fmt"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	var dataset string

	cmd := &cobra.Command{
		Use:     "trigger",
		Short:   "Manage triggers",
		Aliases: []string{"triggers"},
		Example: `  # List triggers for a dataset
  honeycomb trigger list --dataset my-dataset

  # Get a trigger by ID
  honeycomb trigger get abc123 --dataset my-dataset`,
	}

	cmd.PersistentFlags().StringVar(&dataset, "dataset", "", "Dataset slug (required)")
	_ = cmd.MarkPersistentFlagRequired("dataset")

	cmd.AddCommand(NewListCmd(opts, &dataset))
	cmd.AddCommand(NewGetCmd(opts, &dataset))
	cmd.AddCommand(NewCreateCmd(opts, &dataset))
	cmd.AddCommand(NewUpdateCmd(opts, &dataset))
	cmd.AddCommand(NewDeleteCmd(opts, &dataset))

	return command.Group(cmd)
}

type triggerItem struct {
	ID          string `json:"id" col:"ID"`
	Name        string `json:"name" col:"Name"`
	Description string `json:"description,omitempty" col:"Description"`
	Disabled    bool   `json:"disabled" col:"Disabled"`
	Triggered   bool   `json:"triggered" col:"Triggered"`
	AlertType   string `json:"alert_type,omitempty" col:"Alert Type"`
	Threshold   string `json:"threshold,omitempty" col:"Threshold"`
}

type triggerDetail struct {
	ID          string            `json:"id" detail:"ID"`
	Name        string            `json:"name" detail:"Name"`
	Description string            `json:"description,omitempty" detail:"Description"`
	DatasetSlug string            `json:"dataset_slug,omitempty" detail:"Dataset Slug"`
	Disabled    bool              `json:"disabled" detail:"Disabled"`
	Triggered   bool              `json:"triggered" detail:"Triggered"`
	AlertType   string            `json:"alert_type,omitempty" detail:"Alert Type"`
	Frequency   int               `json:"frequency,omitempty" detail:"Frequency"`
	Threshold   *triggerThreshold `json:"threshold,omitempty"`
	QueryID     string            `json:"query_id,omitempty"`
	HasQuery    bool              `json:"has_query,omitempty"`
	Recipients  []recipientItem   `json:"recipients,omitempty"`
	Tags        []tagItem         `json:"tags,omitempty"`
	CreatedAt   string            `json:"created_at,omitempty" detail:"Created At"`
	UpdatedAt   string            `json:"updated_at,omitempty" detail:"Updated At"`
}

type triggerThreshold struct {
	Op            string  `json:"op"`
	Value         float64 `json:"value"`
	ExceededLimit int     `json:"exceeded_limit,omitempty"`
}

type recipientItem struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type,omitempty"`
	Target string `json:"target,omitempty"`
}

type tagItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func formatThreshold(t *api.TriggerResponse) string {
	if t.Threshold == nil {
		return ""
	}
	return fmt.Sprintf("%s %g", t.Threshold.Op, t.Threshold.Value)
}

func toItem(t api.TriggerResponse) triggerItem {
	return triggerItem{
		ID:          deref.String(t.Id),
		Name:        deref.String(t.Name),
		Description: deref.String(t.Description),
		Disabled:    deref.Bool(t.Disabled),
		Triggered:   deref.Bool(t.Triggered),
		AlertType:   deref.Enum(t.AlertType),
		Threshold:   formatThreshold(&t),
	}
}

func toDetail(t api.TriggerResponse) triggerDetail {
	d := triggerDetail{
		ID:          deref.String(t.Id),
		Name:        deref.String(t.Name),
		Description: deref.String(t.Description),
		DatasetSlug: deref.String(t.DatasetSlug),
		Disabled:    deref.Bool(t.Disabled),
		Triggered:   deref.Bool(t.Triggered),
		AlertType:   deref.Enum(t.AlertType),
		Frequency:   deref.Int(t.Frequency),
		QueryID:     deref.String(t.QueryId),
		HasQuery:    t.Query != nil,
		CreatedAt:   deref.Time(t.CreatedAt),
		UpdatedAt:   deref.Time(t.UpdatedAt),
	}
	if t.Threshold != nil {
		d.Threshold = &triggerThreshold{
			Op:            string(t.Threshold.Op),
			Value:         float64(t.Threshold.Value),
			ExceededLimit: deref.Int(t.Threshold.ExceededLimit),
		}
	}
	if t.Recipients != nil {
		for _, r := range *t.Recipients {
			d.Recipients = append(d.Recipients, recipientItem{
				ID:     deref.String(r.Id),
				Type:   deref.Enum(r.Type),
				Target: deref.String(r.Target),
			})
		}
	}
	if t.Tags != nil {
		for _, tag := range *t.Tags {
			d.Tags = append(d.Tags, tagItem{Key: tag.Key, Value: tag.Value})
		}
	}
	return d
}

func writeTriggerDetail(opts *options.RootOptions, detail triggerDetail) error {
	fields := output.FieldsFromTags(detail)

	fields = append(fields, output.Field{Label: "Threshold", Value: formatThresholdDetail(detail.Threshold)})

	if detail.QueryID != "" {
		fields = append(fields, output.Field{Label: "Query ID", Value: detail.QueryID})
	} else if detail.HasQuery {
		fields = append(fields, output.Field{Label: "Query ID", Value: "(inline)"})
	}

	if len(detail.Recipients) > 0 {
		targets := make([]string, len(detail.Recipients))
		for i, r := range detail.Recipients {
			targets[i] = r.Target
		}
		fields = append(fields, output.Field{Label: "Recipients", Value: strings.Join(targets, ", ")})
	}

	if len(detail.Tags) > 0 {
		tags := make([]string, len(detail.Tags))
		for i, t := range detail.Tags {
			tags[i] = t.Key + "=" + t.Value
		}
		fields = append(fields, output.Field{Label: "Tags", Value: strings.Join(tags, ", ")})
	}

	return opts.OutputWriter().WriteFields(detail, fields)
}

func formatThresholdDetail(t *triggerThreshold) string {
	if t == nil {
		return ""
	}
	return fmt.Sprintf("%s %g", t.Op, t.Value)
}
