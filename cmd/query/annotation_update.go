package query

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file string
		name string
		desc string
	)

	cmd := &cobra.Command{
		Use:   "update <annotation-id>",
		Short: "Update a query annotation",
		Example: `  # Update an annotation's name
  honeycomb query annotation update q-abc --dataset my-dataset \
    --name "p99 latency"

  # Update from a JSON file
  honeycomb query annotation update q-abc --dataset my-dataset \
    --file annotation.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnnotationUpdate(cmd, opts, *dataset, args[0], file, name, desc)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&name, "name", "", "Annotation name")
	cmd.Flags().StringVar(&desc, "description", "", "Annotation description")

	cmd.MarkFlagsMutuallyExclusive("file", "name")
	cmd.MarkFlagsMutuallyExclusive("file", "description")

	return cmd
}

func runAnnotationUpdate(cmd *cobra.Command, opts *options.RootOptions, dataset, annotationID, file, name, desc string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	if file != "" {
		return updateAnnotationFromFile(ctx, client, opts, dataset, annotationID, file)
	}

	if !command.AnyChanged(cmd, "name", "description") {
		return fmt.Errorf("--file, --name, or --description is required")
	}

	current, err := getAnnotation(ctx, client, dataset, annotationID)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("name") {
		current.Name = name
	}
	if cmd.Flags().Changed("description") {
		current.Description = &desc
	}

	data, err := api.MarshalStrippingReadOnly(current, "QueryAnnotation")
	if err != nil {
		return fmt.Errorf("encoding query annotation: %w", err)
	}

	resp, err := client.UpdateQueryAnnotationWithBodyWithResponse(ctx, dataset, annotationID, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("updating query annotation: %w", err)
	}

	annotation, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeAnnotationDetail(opts, annotationToDetail(*annotation))
}

func updateAnnotationFromFile(ctx context.Context, client *api.ClientWithResponses, opts *options.RootOptions, dataset, annotationID, file string) error {
	raw, err := command.ReadDefinitionFile(opts.IOStreams, file)
	if err != nil {
		return err
	}

	data, err := api.StripReadOnly(raw, "QueryAnnotation")
	if err != nil {
		return fmt.Errorf("stripping read-only fields: %w", err)
	}

	resp, err := client.UpdateQueryAnnotationWithBodyWithResponse(ctx, dataset, annotationID, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("updating query annotation: %w", err)
	}

	annotation, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeAnnotationDetail(opts, annotationToDetail(*annotation))
}

func getAnnotation(ctx context.Context, client *api.ClientWithResponses, dataset, annotationID string) (*api.QueryAnnotation, error) {
	resp, err := client.GetQueryAnnotationWithResponse(ctx, dataset, annotationID)
	if err != nil {
		return nil, fmt.Errorf("getting query annotation: %w", err)
	}

	return api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
}
