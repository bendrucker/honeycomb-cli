package board

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewViewCreateCmd(opts *options.RootOptions, board *string) *cobra.Command {
	var (
		file       string
		name       string
		filterArgs []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a board view",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runViewCreate(cmd, opts, *board, file, name, filterArgs)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&name, "name", "", "View name")
	cmd.Flags().StringArrayVar(&filterArgs, "filter", nil, "Filter: column:operation:value (repeatable, required)")

	cmd.MarkFlagsMutuallyExclusive("file", "name")
	cmd.MarkFlagsMutuallyExclusive("file", "filter")

	return cmd
}

func runViewCreate(cmd *cobra.Command, opts *options.RootOptions, boardID, file, name string, filterArgs []string) error {
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
		return createViewFromFile(ctx, client, opts, auth, boardID, file)
	}

	if name == "" {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--name or --file is required in non-interactive mode")
		}
		name, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "View name: ")
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("view name is required")
		}
	}

	filters, err := parseViewFilters(filterArgs)
	if err != nil {
		return err
	}

	if len(filters) == 0 {
		return fmt.Errorf("at least one --filter is required")
	}

	body := api.CreateBoardViewJSONRequestBody{
		Name:    name,
		Filters: filters,
	}

	resp, err := client.CreateBoardViewWithResponse(ctx, boardID, body, auth)
	if err != nil {
		return fmt.Errorf("creating board view: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeViewDetail(opts, viewResponseToDetail(*resp.JSON201))
}

func parseViewFilters(args []string) ([]api.BoardViewFilter, error) {
	filters := make([]api.BoardViewFilter, 0, len(args))
	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 3)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid filter %q: expected column:operation or column:operation:value", arg)
		}
		f := api.BoardViewFilter{
			Column:    parts[0],
			Operation: api.BoardViewFilterOperation(parts[1]),
		}
		if len(parts) == 3 {
			f.Value = parts[2]
		}
		filters = append(filters, f)
	}
	return filters, nil
}

func createViewFromFile(ctx context.Context, client *api.ClientWithResponses, opts *options.RootOptions, auth api.RequestEditorFn, boardID, file string) error {
	var r io.Reader
	if file == "-" {
		r = opts.IOStreams.In
	} else {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	resp, err := client.CreateBoardViewWithBodyWithResponse(ctx, boardID, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("creating board view: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeViewDetail(opts, viewResponseToDetail(*resp.JSON201))
}
