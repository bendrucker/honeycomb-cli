package query

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <annotation-id>",
		Short: "Delete a query annotation",
		Example: `  # Delete an annotation, prompting for confirmation
  honeycomb query annotation delete q-abc --dataset my-dataset

  # Delete without confirmation
  honeycomb query annotation delete q-abc --dataset my-dataset --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnnotationDelete(cmd.Context(), opts, *dataset, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runAnnotationDelete(ctx context.Context, opts *options.RootOptions, dataset, annotationID string, yes bool) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	proceed, err := command.ConfirmDelete(opts.IOStreams, yes, "query annotation", annotationID, func() (string, error) {
		a, err := getAnnotation(ctx, client, dataset, annotationID)
		if err != nil {
			return "", err
		}
		return a.Name, nil
	})
	if err != nil {
		return err
	}
	if !proceed {
		return fmt.Errorf("aborted")
	}

	resp, err := client.DeleteQueryAnnotationWithResponse(ctx, dataset, annotationID)
	if err != nil {
		return fmt.Errorf("deleting query annotation: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(annotationID, fmt.Sprintf("Query annotation %s deleted", annotationID))
}
