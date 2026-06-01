package board

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
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
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	if file != "" {
		return createViewFromFile(ctx, client, opts, boardID, file)
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

	resp, err := client.CreateBoardViewWithResponse(ctx, boardID, body)
	if err != nil {
		return fmt.Errorf("creating board view: %w", err)
	}

	created, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON201)
	if err != nil {
		return err
	}

	return writeViewDetail(opts, viewResponseToDetail(*created))
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

func createViewFromFile(ctx context.Context, client *api.ClientWithResponses, opts *options.RootOptions, boardID, file string) error {
	data, err := command.ReadDefinitionFile(opts.IOStreams, file)
	if err != nil {
		return err
	}

	resp, err := client.CreateBoardViewWithBodyWithResponse(ctx, boardID, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating board view: %w", err)
	}

	created, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON201)
	if err != nil {
		return err
	}

	return writeViewDetail(opts, viewResponseToDetail(*created))
}
