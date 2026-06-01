package query

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var includeBoardAnnotations bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List query annotations",
		Example: `  # List annotations for a dataset
  honeycomb query annotation list --dataset my-dataset

  # Include annotations created from boards
  honeycomb query annotation list --dataset my-dataset \
    --include-board-annotations`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAnnotationList(cmd.Context(), opts, *dataset, includeBoardAnnotations)
		},
	}

	cmd.Flags().BoolVar(&includeBoardAnnotations, "include-board-annotations", false, "Include annotations created from boards")

	return cmd
}

func runAnnotationList(ctx context.Context, opts *options.RootOptions, dataset string, includeBoardAnnotations bool) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	params := &api.ListQueryAnnotationsParams{}
	if includeBoardAnnotations {
		params.IncludeBoardAnnotations = ptr(true)
	}

	resp, err := client.ListQueryAnnotationsWithResponse(ctx, dataset, params)
	if err != nil {
		return fmt.Errorf("listing query annotations: %w", err)
	}

	list, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	items := make([]annotationItem, len(*list))
	for i, a := range *list {
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

	return opts.OutputWriterList().WriteList(items, annotationListTable, "No query annotations found.")
}

func ptr[T any](v T) *T {
	return &v
}
