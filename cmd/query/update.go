package query

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
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
		Args:  cobra.ExactArgs(1),
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
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	ctx := cmd.Context()

	if file != "" {
		return updateAnnotationFromFile(ctx, client, opts, auth, dataset, annotationID, file)
	}

	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("description") {
		return fmt.Errorf("--file, --name, or --description is required")
	}

	current, err := getAnnotation(ctx, client, auth, dataset, annotationID)
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

	resp, err := client.UpdateQueryAnnotationWithBodyWithResponse(ctx, dataset, annotationID, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("updating query annotation: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeAnnotationDetail(opts, annotationToDetail(*resp.JSON200))
}

func updateAnnotationFromFile(ctx context.Context, client *api.ClientWithResponses, opts *options.RootOptions, auth api.RequestEditorFn, dataset, annotationID, file string) error {
	raw, err := readFile(opts, file)
	if err != nil {
		return err
	}

	data, err := api.StripReadOnly(raw, "QueryAnnotation")
	if err != nil {
		return fmt.Errorf("stripping read-only fields: %w", err)
	}

	resp, err := client.UpdateQueryAnnotationWithBodyWithResponse(ctx, dataset, annotationID, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("updating query annotation: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeAnnotationDetail(opts, annotationToDetail(*resp.JSON200))
}

func getAnnotation(ctx context.Context, client *api.ClientWithResponses, auth api.RequestEditorFn, dataset, annotationID string) (*api.QueryAnnotation, error) {
	resp, err := client.GetQueryAnnotationWithResponse(ctx, dataset, annotationID, auth)
	if err != nil {
		return nil, fmt.Errorf("getting query annotation: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return resp.JSON200, nil
}
