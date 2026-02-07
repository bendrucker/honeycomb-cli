package slo

import (
	"context"
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewViewCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var detailed bool

	cmd := &cobra.Command{
		Use:   "view <slo-id>",
		Short: "View an SLO",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSLOView(cmd.Context(), opts, *dataset, args[0], detailed)
		},
	}

	cmd.Flags().BoolVar(&detailed, "detailed", false, "Include compliance and budget data (Enterprise)")

	return cmd
}

func runSLOView(ctx context.Context, opts *options.RootOptions, dataset, sloID string, detailed bool) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	params := &api.GetSloParams{}
	if detailed {
		params.Detailed = ptr(true)
	}

	resp, err := client.GetSloWithResponse(ctx, dataset, sloID, params, keyEditor(key))
	if err != nil {
		return fmt.Errorf("getting SLO: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	// GetSloResp.JSON200 is a union type (unusable). Unmarshal resp.Body instead.
	var sloResp sloDetailedResponse
	if err := json.Unmarshal(resp.Body, &sloResp); err != nil {
		return fmt.Errorf("parsing SLO response: %w", err)
	}

	detail := detailedToDetail(sloResp)

	format := opts.ResolveFormat()
	if format != "table" {
		return opts.OutputWriter().Write(detail, output.TableDef{})
	}

	tw := tabwriter.NewWriter(opts.IOStreams.Out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "ID:\t%s\n", detail.ID)
	_, _ = fmt.Fprintf(tw, "Name:\t%s\n", detail.Name)
	_, _ = fmt.Fprintf(tw, "Description:\t%s\n", detail.Description)
	_, _ = fmt.Fprintf(tw, "SLI Alias:\t%s\n", detail.SLIAlias)
	_, _ = fmt.Fprintf(tw, "Target:\t%s\n", formatTarget(detail.TargetPerMillion))
	_, _ = fmt.Fprintf(tw, "Time Period:\t%s\n", formatTimePeriod(detail.TimePeriodDays))
	if len(detail.DatasetSlugs) > 0 {
		_, _ = fmt.Fprintf(tw, "Datasets:\t%s\n", joinStrings(detail.DatasetSlugs))
	}
	if detail.CreatedAt != "" {
		_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", detail.CreatedAt)
	}
	if detail.UpdatedAt != "" {
		_, _ = fmt.Fprintf(tw, "Updated At:\t%s\n", detail.UpdatedAt)
	}
	if detail.ResetAt != "" {
		_, _ = fmt.Fprintf(tw, "Reset At:\t%s\n", detail.ResetAt)
	}
	if detail.Compliance != nil {
		_, _ = fmt.Fprintf(tw, "Compliance:\t%g%%\n", *detail.Compliance)
	}
	if detail.BudgetRemaining != nil {
		_, _ = fmt.Fprintf(tw, "Budget Remaining:\t%g\n", *detail.BudgetRemaining)
	}

	return tw.Flush()
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

func ptr[T any](v T) *T {
	return &v
}
