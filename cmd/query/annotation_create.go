package query

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
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
		Long: `Create a query annotation using one of two input modes:

Flag mode: pass --name and --query-id together (--description is optional) to
build the annotation from flags.

File mode: pass --file with a path to a JSON definition (- for stdin). The file
flag is mutually exclusive with the flag-mode inputs.`,
		Example: `  # Flag mode
  honeycomb query annotation create --dataset my-dataset \
    --name "Latency Query" --query-id q-abc --description "p99 latency"

  # File mode
  honeycomb query annotation create --dataset my-dataset --file annotation.json
  echo '{"name":"...","query_id":"..."}' | honeycomb query annotation create \
    --dataset my-dataset --file -`,
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
	cmd.MarkFlagsRequiredTogether("name", "query-id")

	return cmd
}

func runAnnotationCreate(cmd *cobra.Command, opts *options.RootOptions, dataset, file, name, desc, queryID string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	var data []byte
	if file != "" {
		data, err = command.ReadDefinitionFile(opts.IOStreams, file)
		if err != nil {
			return err
		}
	} else if cmd.Flags().Changed("name") || cmd.Flags().Changed("query-id") || cmd.Flags().Changed("description") {
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
		file, err = command.Resolve(opts.IOStreams, file, command.Field{
			Prompt:            "Path to query annotation JSON file: ",
			Required:          true,
			NonInteractiveErr: fmt.Errorf("--file is required in non-interactive mode"),
			EmptyErr:          fmt.Errorf("file path is required"),
		})
		if err != nil {
			return err
		}
		data, err = command.ReadDefinitionFile(opts.IOStreams, file)
		if err != nil {
			return err
		}
	}

	resp, err := client.CreateQueryAnnotationWithBodyWithResponse(ctx, dataset, "application/json", bytes.NewReader(data))
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
