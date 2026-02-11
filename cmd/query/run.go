package query

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/bendrucker/honeycomb-cli/internal/poll"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewRunCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file       string
		annotation string
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a query",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runQueryRun(cmd.Context(), opts, *dataset, file, annotation)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to query spec JSON file (- for stdin)")
	cmd.Flags().StringVarP(&annotation, "annotation", "a", "", "Annotation ID to re-run")
	cmd.MarkFlagsMutuallyExclusive("file", "annotation")

	return cmd
}

func runQueryRun(ctx context.Context, opts *options.RootOptions, dataset, file, annotation string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	queryID, err := resolveQueryID(ctx, opts, client, dataset, key, file, annotation)
	if err != nil {
		return err
	}

	resultResp, err := client.CreateQueryResultWithResponse(ctx, dataset, api.CreateQueryResultRequest{
		QueryId:       &queryID,
		DisableSeries: ptr(true),
	}, keyEditor(key))
	if err != nil {
		return fmt.Errorf("creating query result: %w", err)
	}
	if err := api.CheckResponse(resultResp.StatusCode(), resultResp.Body); err != nil {
		return err
	}
	if resultResp.JSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resultResp.Status())
	}
	if resultResp.JSON201.Id == nil {
		return fmt.Errorf("query result ID missing from response")
	}
	resultID := *resultResp.JSON201.Id

	cfg := poll.Config{
		Title:       "Running query...",
		Interactive: opts.IOStreams.CanPrompt(),
	}
	details, err := poll.Poll(ctx, cfg, func(ctx context.Context) (*api.QueryResultDetails, bool, error) {
		resp, err := client.GetQueryResultWithResponse(ctx, dataset, resultID, keyEditor(key))
		if err != nil {
			return nil, false, fmt.Errorf("getting query result: %w", err)
		}
		if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
			return nil, false, err
		}
		if resp.JSON200 == nil {
			return nil, false, fmt.Errorf("unexpected response: %s", resp.Status())
		}
		complete := resp.JSON200.Complete != nil && *resp.JSON200.Complete
		return resp.JSON200, complete, nil
	})
	if err != nil {
		return err
	}

	if details.Links != nil && details.Links.QueryUrl != nil {
		_, _ = fmt.Fprintf(opts.IOStreams.Err, "%s\n", *details.Links.QueryUrl)
	}

	return opts.OutputWriter().WriteDynamic(details, buildResultTable(details))
}

func resolveQueryID(ctx context.Context, opts *options.RootOptions, client *api.ClientWithResponses, dataset, key, file, annotation string) (string, error) {
	switch {
	case file != "":
		return createQueryFromFile(ctx, opts, client, dataset, key, file)
	case annotation != "":
		return queryIDFromAnnotation(ctx, client, dataset, key, annotation)
	case opts.IOStreams.CanPrompt():
		return promptQueryID(ctx, opts, client, dataset, key)
	default:
		return "", fmt.Errorf("either --file or --annotation is required")
	}
}

func createQueryFromFile(ctx context.Context, opts *options.RootOptions, client *api.ClientWithResponses, dataset, key, file string) (string, error) {
	data, err := readFile(opts, file)
	if err != nil {
		return "", err
	}

	resp, err := client.CreateQueryWithBodyWithResponse(ctx, dataset, "application/json", bytes.NewReader(data), keyEditor(key))
	if err != nil {
		return "", fmt.Errorf("creating query: %w", err)
	}
	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return "", err
	}
	if resp.JSON200 == nil {
		return "", fmt.Errorf("unexpected response: %s", resp.Status())
	}
	if resp.JSON200.Id == nil {
		return "", fmt.Errorf("query ID missing from response")
	}
	return *resp.JSON200.Id, nil
}

func queryIDFromAnnotation(ctx context.Context, client *api.ClientWithResponses, dataset, key, annotationID string) (string, error) {
	resp, err := client.GetQueryAnnotationWithResponse(ctx, dataset, annotationID, keyEditor(key))
	if err != nil {
		return "", fmt.Errorf("getting query annotation: %w", err)
	}
	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return "", err
	}
	if resp.JSON200 == nil {
		return "", fmt.Errorf("unexpected response: %s", resp.Status())
	}
	return resp.JSON200.QueryId, nil
}

func promptQueryID(ctx context.Context, opts *options.RootOptions, client *api.ClientWithResponses, dataset, key string) (string, error) {
	mode, err := prompt.Choice(opts.IOStreams.Out, opts.IOStreams.In, "Query source (file, annotation): ", []string{"file", "annotation"})
	if err != nil {
		return "", err
	}

	switch mode {
	case "file":
		path, err := prompt.Line(opts.IOStreams.Out, opts.IOStreams.In, "File path (- for stdin): ")
		if err != nil {
			return "", err
		}
		return createQueryFromFile(ctx, opts, client, dataset, key, path)
	case "annotation":
		id, err := prompt.Line(opts.IOStreams.Out, opts.IOStreams.In, "Annotation ID: ")
		if err != nil {
			return "", err
		}
		return queryIDFromAnnotation(ctx, client, dataset, key, id)
	default:
		return "", fmt.Errorf("unexpected mode: %s", mode)
	}
}

func buildResultTable(details *api.QueryResultDetails) output.DynamicTableDef {
	var headers []string
	if details.Query != nil {
		if details.Query.Breakdowns != nil {
			headers = append(headers, *details.Query.Breakdowns...)
		}
		if details.Query.Calculations != nil {
			for _, calc := range *details.Query.Calculations {
				var col string
				if calc.Column.IsSpecified() && !calc.Column.IsNull() {
					col = calc.Column.MustGet()
				}
				headers = append(headers, calcColumnName(string(calc.Op), col))
			}
		}
	}

	var rows [][]string
	if details.Data != nil && details.Data.Results != nil {
		for _, r := range *details.Data.Results {
			if r.Data == nil {
				continue
			}
			row := make([]string, len(headers))
			for i, h := range headers {
				if val, ok := (*r.Data)[h]; ok {
					row[i] = fmt.Sprintf("%v", val)
				}
			}
			rows = append(rows, row)
		}
	}

	return output.DynamicTableDef{
		Headers: headers,
		Rows:    rows,
	}
}

func calcColumnName(op, col string) string {
	if col == "" {
		return op
	}
	return fmt.Sprintf("%s(%s)", op, col)
}
