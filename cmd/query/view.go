package query

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
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
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetQueryAnnotationWithResponse(ctx, dataset, annotationID, keyEditor(key))
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

	queryResp, err := client.GetQueryWithResponse(ctx, dataset, annotation.QueryId, keyEditor(key))
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

	return opts.OutputWriter().WriteValue(combined, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID:\t%s\n", combined.ID)
		_, _ = fmt.Fprintf(tw, "Name:\t%s\n", combined.Name)
		_, _ = fmt.Fprintf(tw, "Description:\t%s\n", combined.Description)
		_, _ = fmt.Fprintf(tw, "Query ID:\t%s\n", combined.QueryID)
		if combined.Source != "" {
			_, _ = fmt.Fprintf(tw, "Source:\t%s\n", combined.Source)
		}
		if combined.CreatedAt != "" {
			_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", combined.CreatedAt)
		}
		if combined.UpdatedAt != "" {
			_, _ = fmt.Fprintf(tw, "Updated At:\t%s\n", combined.UpdatedAt)
		}

		if q := combined.Query; q != nil {
			_, _ = fmt.Fprintln(tw)
			_, _ = fmt.Fprintln(tw, "Query:")
			if q.TimeRange != nil {
				_, _ = fmt.Fprintf(tw, "  Time Range:\t%ds\n", *q.TimeRange)
			}
			if q.Breakdowns != nil && len(*q.Breakdowns) > 0 {
				_, _ = fmt.Fprintf(tw, "  Breakdowns:\t%s\n", strings.Join(*q.Breakdowns, ", "))
			}
			if q.Calculations != nil {
				for _, calc := range *q.Calculations {
					col := ""
					if calc.Column.IsSpecified() && !calc.Column.IsNull() {
						col = calc.Column.MustGet()
					}
					if col != "" {
						_, _ = fmt.Fprintf(tw, "  Calculation:\t%s(%s)\n", calc.Op, col)
					} else {
						_, _ = fmt.Fprintf(tw, "  Calculation:\t%s\n", calc.Op)
					}
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
					_, _ = fmt.Fprintf(tw, "  Filter:\t%s %s %s\n", col, f.Op, val)
				}
			}
		}

		return tw.Flush()
	})
}
