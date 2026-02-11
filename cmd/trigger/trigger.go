package trigger

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	var dataset string

	cmd := &cobra.Command{
		Use:     "trigger",
		Short:   "Manage triggers",
		Aliases: []string{"triggers"},
	}

	cmd.PersistentFlags().StringVar(&dataset, "dataset", "", "Dataset slug (required)")
	_ = cmd.MarkPersistentFlagRequired("dataset")

	cmd.AddCommand(NewListCmd(opts, &dataset))
	cmd.AddCommand(NewGetCmd(opts, &dataset))
	cmd.AddCommand(NewCreateCmd(opts, &dataset))
	cmd.AddCommand(NewUpdateCmd(opts, &dataset))
	cmd.AddCommand(NewDeleteCmd(opts, &dataset))

	return cmd
}

type triggerItem struct {
	ID          string `json:"id"                       yaml:"id"`
	Name        string `json:"name"                     yaml:"name"`
	Description string `json:"description,omitempty"    yaml:"description,omitempty"`
	Disabled    bool   `json:"disabled"                 yaml:"disabled"`
	Triggered   bool   `json:"triggered"                yaml:"triggered"`
	AlertType   string `json:"alert_type,omitempty"     yaml:"alert_type,omitempty"`
	Threshold   string `json:"threshold,omitempty"      yaml:"threshold,omitempty"`
}

type triggerDetail struct {
	ID          string            `json:"id"                       yaml:"id"`
	Name        string            `json:"name"                     yaml:"name"`
	Description string            `json:"description,omitempty"    yaml:"description,omitempty"`
	DatasetSlug string            `json:"dataset_slug,omitempty"   yaml:"dataset_slug,omitempty"`
	Disabled    bool              `json:"disabled"                 yaml:"disabled"`
	Triggered   bool              `json:"triggered"                yaml:"triggered"`
	AlertType   string            `json:"alert_type,omitempty"     yaml:"alert_type,omitempty"`
	Frequency   int               `json:"frequency,omitempty"      yaml:"frequency,omitempty"`
	Threshold   *triggerThreshold `json:"threshold,omitempty"     yaml:"threshold,omitempty"`
	QueryID     string            `json:"query_id,omitempty"       yaml:"query_id,omitempty"`
	HasQuery    bool              `json:"has_query,omitempty"      yaml:"has_query,omitempty"`
	Recipients  []recipientItem   `json:"recipients,omitempty"     yaml:"recipients,omitempty"`
	Tags        []tagItem         `json:"tags,omitempty"           yaml:"tags,omitempty"`
	CreatedAt   string            `json:"created_at,omitempty"     yaml:"created_at,omitempty"`
	UpdatedAt   string            `json:"updated_at,omitempty"     yaml:"updated_at,omitempty"`
}

type triggerThreshold struct {
	Op            string  `json:"op"                        yaml:"op"`
	Value         float64 `json:"value"                     yaml:"value"`
	ExceededLimit int     `json:"exceeded_limit,omitempty"  yaml:"exceeded_limit,omitempty"`
}

type recipientItem struct {
	ID     string `json:"id,omitempty"     yaml:"id,omitempty"`
	Type   string `json:"type,omitempty"   yaml:"type,omitempty"`
	Target string `json:"target,omitempty" yaml:"target,omitempty"`
}

type tagItem struct {
	Key   string `json:"key"   yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

func formatThreshold(t *api.TriggerResponse) string {
	if t.Threshold == nil {
		return ""
	}
	return fmt.Sprintf("%s %g", t.Threshold.Op, t.Threshold.Value)
}

func toItem(t api.TriggerResponse) triggerItem {
	item := triggerItem{
		Threshold: formatThreshold(&t),
	}
	if t.Id != nil {
		item.ID = *t.Id
	}
	if t.Name != nil {
		item.Name = *t.Name
	}
	if t.Description != nil {
		item.Description = *t.Description
	}
	if t.Disabled != nil {
		item.Disabled = *t.Disabled
	}
	if t.Triggered != nil {
		item.Triggered = *t.Triggered
	}
	if t.AlertType != nil {
		item.AlertType = string(*t.AlertType)
	}
	return item
}

func toDetail(t api.TriggerResponse) triggerDetail {
	d := triggerDetail{}
	if t.Id != nil {
		d.ID = *t.Id
	}
	if t.Name != nil {
		d.Name = *t.Name
	}
	if t.Description != nil {
		d.Description = *t.Description
	}
	if t.DatasetSlug != nil {
		d.DatasetSlug = *t.DatasetSlug
	}
	if t.Disabled != nil {
		d.Disabled = *t.Disabled
	}
	if t.Triggered != nil {
		d.Triggered = *t.Triggered
	}
	if t.AlertType != nil {
		d.AlertType = string(*t.AlertType)
	}
	if t.Frequency != nil {
		d.Frequency = *t.Frequency
	}
	if t.Threshold != nil {
		d.Threshold = &triggerThreshold{
			Op:    string(t.Threshold.Op),
			Value: float64(t.Threshold.Value),
		}
		if t.Threshold.ExceededLimit != nil {
			d.Threshold.ExceededLimit = *t.Threshold.ExceededLimit
		}
	}
	if t.QueryId != nil {
		d.QueryID = *t.QueryId
	}
	if t.Query != nil {
		d.HasQuery = true
	}
	if t.Recipients != nil {
		for _, r := range *t.Recipients {
			ri := recipientItem{}
			if r.Id != nil {
				ri.ID = *r.Id
			}
			if r.Type != nil {
				ri.Type = string(*r.Type)
			}
			if r.Target != nil {
				ri.Target = *r.Target
			}
			d.Recipients = append(d.Recipients, ri)
		}
	}
	if t.Tags != nil {
		for _, tag := range *t.Tags {
			d.Tags = append(d.Tags, tagItem{Key: tag.Key, Value: tag.Value})
		}
	}
	if t.CreatedAt != nil {
		d.CreatedAt = t.CreatedAt.Format(time.RFC3339)
	}
	if t.UpdatedAt != nil {
		d.UpdatedAt = t.UpdatedAt.Format(time.RFC3339)
	}
	return d
}

func writeTriggerDetail(opts *options.RootOptions, detail triggerDetail) error {
	return opts.OutputWriter().WriteValue(detail, func(out io.Writer) error {
		tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID:\t%s\n", detail.ID)
		_, _ = fmt.Fprintf(tw, "Name:\t%s\n", detail.Name)
		_, _ = fmt.Fprintf(tw, "Description:\t%s\n", detail.Description)
		_, _ = fmt.Fprintf(tw, "Disabled:\t%s\n", strconv.FormatBool(detail.Disabled))
		_, _ = fmt.Fprintf(tw, "Triggered:\t%s\n", strconv.FormatBool(detail.Triggered))
		_, _ = fmt.Fprintf(tw, "Alert Type:\t%s\n", detail.AlertType)
		_, _ = fmt.Fprintf(tw, "Dataset Slug:\t%s\n", detail.DatasetSlug)
		_, _ = fmt.Fprintf(tw, "Frequency:\t%d\n", detail.Frequency)
		_, _ = fmt.Fprintf(tw, "Threshold:\t%s\n", formatThresholdDetail(detail.Threshold))
		if detail.QueryID != "" {
			_, _ = fmt.Fprintf(tw, "Query ID:\t%s\n", detail.QueryID)
		} else if detail.HasQuery {
			_, _ = fmt.Fprintf(tw, "Query ID:\t(inline)\n")
		}
		_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", detail.CreatedAt)
		_, _ = fmt.Fprintf(tw, "Updated At:\t%s\n", detail.UpdatedAt)
		if len(detail.Recipients) > 0 {
			targets := make([]string, len(detail.Recipients))
			for i, r := range detail.Recipients {
				targets[i] = r.Target
			}
			_, _ = fmt.Fprintf(tw, "Recipients:\t%s\n", strings.Join(targets, ", "))
		}
		if len(detail.Tags) > 0 {
			tags := make([]string, len(detail.Tags))
			for i, t := range detail.Tags {
				tags[i] = t.Key + "=" + t.Value
			}
			_, _ = fmt.Fprintf(tw, "Tags:\t%s\n", strings.Join(tags, ", "))
		}
		return tw.Flush()
	})
}

func formatThresholdDetail(t *triggerThreshold) string {
	if t == nil {
		return ""
	}
	return fmt.Sprintf("%s %g", t.Op, t.Value)
}

func keyEditor(key string) api.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		config.ApplyAuth(req, config.KeyConfig, key)
		return nil
	}
}
