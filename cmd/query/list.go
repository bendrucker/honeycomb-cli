package query

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

var annotationListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "ID", Value: func(v any) string { return v.(annotationItem).ID }},
		{Header: "Name", Value: func(v any) string { return v.(annotationItem).Name }},
		{Header: "Query ID", Value: func(v any) string { return v.(annotationItem).QueryID }},
		{Header: "Source", Value: func(v any) string { return v.(annotationItem).Source }},
	},
}

func NewListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var includeBoardAnnotations bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List query annotations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAnnotationList(cmd.Context(), opts, *dataset, includeBoardAnnotations)
		},
	}

	cmd.Flags().BoolVar(&includeBoardAnnotations, "include-board-annotations", false, "Include annotations created from boards")

	return cmd
}

func runAnnotationList(ctx context.Context, opts *options.RootOptions, dataset string, includeBoardAnnotations bool) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	params := &api.ListQueryAnnotationsParams{}
	if includeBoardAnnotations {
		params.IncludeBoardAnnotations = ptr(true)
	}

	resp, err := client.ListQueryAnnotationsWithResponse(ctx, dataset, params, auth)
	if err != nil {
		return fmt.Errorf("listing query annotations: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	items := make([]annotationItem, len(*resp.JSON200))
	for i, a := range *resp.JSON200 {
		item := annotationItem{
			Name:    a.Name,
			QueryID: a.QueryId,
		}
		if a.Id != nil {
			item.ID = *a.Id
		}
		if a.Source != nil {
			item.Source = string(*a.Source)
		}
		items[i] = item
	}

	return opts.OutputWriterList().Write(items, annotationListTable)
}

func ptr[T any](v T) *T {
	return &v
}
