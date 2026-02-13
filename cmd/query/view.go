package query

import (
	"context"
	"fmt"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type annotationWithQuery struct {
	annotationDetail
	Query *api.Query `json:"query,omitempty"`
}

func NewViewCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "view <annotation-id>",
		Short: "View a query annotation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnnotationView(cmd.Context(), opts, *dataset, args[0])
		},
	}
}

func runAnnotationView(ctx context.Context, opts *options.RootOptions, dataset, annotationID string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetQueryAnnotationWithResponse(ctx, dataset, annotationID, auth)
	if err != nil {
		return fmt.Errorf("getting query annotation: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	annotation := resp.JSON200

	queryResp, err := client.GetQueryWithResponse(ctx, dataset, annotation.QueryId, auth)
	if err != nil {
		return fmt.Errorf("getting query: %w", err)
	}

	if err := api.CheckResponse(queryResp.StatusCode(), queryResp.Body); err != nil {
		return err
	}

	combined := annotationWithQuery{
		annotationDetail: annotationToDetail(*annotation),
		Query:            queryResp.JSON200,
	}

	fields := []output.Field{
		{Label: "ID", Value: combined.ID},
		{Label: "Name", Value: combined.Name},
		{Label: "Description", Value: combined.Description},
		{Label: "Query ID", Value: combined.QueryID},
	}
	if combined.Source != "" {
		fields = append(fields, output.Field{Label: "Source", Value: combined.Source})
	}
	if combined.CreatedAt != "" {
		fields = append(fields, output.Field{Label: "Created At", Value: combined.CreatedAt})
	}
	if combined.UpdatedAt != "" {
		fields = append(fields, output.Field{Label: "Updated At", Value: combined.UpdatedAt})
	}

	if q := combined.Query; q != nil {
		if q.TimeRange != nil {
			fields = append(fields, output.Field{Label: "Time Range", Value: fmt.Sprintf("%ds", *q.TimeRange)})
		}
		if q.Breakdowns != nil && len(*q.Breakdowns) > 0 {
			fields = append(fields, output.Field{Label: "Breakdowns", Value: strings.Join(*q.Breakdowns, ", ")})
		}
		if q.Calculations != nil {
			for _, calc := range *q.Calculations {
				col := ""
				if calc.Column.IsSpecified() && !calc.Column.IsNull() {
					col = calc.Column.MustGet()
				}
				fields = append(fields, output.Field{Label: "Calculation", Value: calcColumnName(string(calc.Op), col)})
			}
		}
		if q.Filters != nil {
			for _, f := range *q.Filters {
				col := ""
				if f.Column.IsSpecified() && !f.Column.IsNull() {
					col = f.Column.MustGet()
				}
				val := ""
				if f.Value != nil {
					val = fmt.Sprintf("%v", f.Value)
				}
				fields = append(fields, output.Field{Label: "Filter", Value: fmt.Sprintf("%s %s %s", col, f.Op, val)})
			}
		}
	}

	return opts.OutputWriter().WriteFields(combined, fields)
}
