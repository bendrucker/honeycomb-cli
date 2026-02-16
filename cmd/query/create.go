package query

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file    string
		name    string
		desc    string
		queryID string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a query annotation",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAnnotationCreate(cmd, opts, *dataset, file, name, desc, queryID)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&name, "name", "", "Annotation name")
	cmd.Flags().StringVar(&desc, "description", "", "Annotation description")
	cmd.Flags().StringVar(&queryID, "query-id", "", "Query ID to annotate")

	cmd.MarkFlagsMutuallyExclusive("file", "name")
	cmd.MarkFlagsMutuallyExclusive("file", "description")
	cmd.MarkFlagsMutuallyExclusive("file", "query-id")

	return cmd
}

func runAnnotationCreate(cmd *cobra.Command, opts *options.RootOptions, dataset, file, name, desc, queryID string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	var data []byte
	if file != "" {
		data, err = readFile(opts, file)
		if err != nil {
			return err
		}
	} else if cmd.Flags().Changed("name") || cmd.Flags().Changed("query-id") {
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		if queryID == "" {
			return fmt.Errorf("--query-id is required")
		}

		annotation := api.QueryAnnotation{
			Name:    name,
			QueryId: queryID,
		}
		if cmd.Flags().Changed("description") {
			annotation.Description = &desc
		}

		data, err = api.MarshalStrippingReadOnly(annotation, "QueryAnnotation")
		if err != nil {
			return fmt.Errorf("encoding query annotation: %w", err)
		}
	} else {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--file is required in non-interactive mode")
		}
		file, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Path to query annotation JSON file: ")
		if err != nil {
			return err
		}
		if file == "" {
			return fmt.Errorf("file path is required")
		}
		data, err = readFile(opts, file)
		if err != nil {
			return err
		}
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.CreateQueryAnnotationWithBodyWithResponse(ctx, dataset, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("creating query annotation: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	var annotation api.QueryAnnotation
	if err := json.Unmarshal(resp.Body, &annotation); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	return writeAnnotationDetail(opts, annotationToDetail(annotation))
}
