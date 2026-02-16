package board

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/jsonutil"
	"github.com/spf13/cobra"
)

func NewViewUpdateCmd(opts *options.RootOptions, board *string) *cobra.Command {
	var (
		file       string
		name       string
		filterArgs []string
	)

	cmd := &cobra.Command{
		Use:   "update <view-id>",
		Short: "Update a board view",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runViewUpdate(cmd, opts, *board, args[0], file, name, filterArgs)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&name, "name", "", "View name")
	cmd.Flags().StringArrayVar(&filterArgs, "filter", nil, "Filter: column:operation:value (repeatable)")

	cmd.MarkFlagsMutuallyExclusive("file", "name")
	cmd.MarkFlagsMutuallyExclusive("file", "filter")

	return cmd
}

func runViewUpdate(cmd *cobra.Command, opts *options.RootOptions, boardID, viewID, file, name string, filterArgs []string) error {
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
		return updateViewFromFile(ctx, client, opts, auth, boardID, viewID, file)
	}

	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("filter") {
		return fmt.Errorf("--file, --name, or --filter is required")
	}

	current, err := getView(ctx, client, auth, boardID, viewID)
	if err != nil {
		return err
	}

	body := api.UpdateBoardViewJSONRequestBody{
		Name:    name,
		Filters: []api.BoardViewFilter{},
	}

	if !cmd.Flags().Changed("name") {
		body.Name = deref.String(current.Name)
	}

	if cmd.Flags().Changed("filter") {
		body.Filters, err = parseViewFilters(filterArgs)
		if err != nil {
			return err
		}
	} else if current.Filters != nil {
		body.Filters = *current.Filters
	}

	resp, err := client.UpdateBoardViewWithResponse(ctx, boardID, viewID, body, auth)
	if err != nil {
		return fmt.Errorf("updating board view: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeViewDetail(opts, viewResponseToDetail(*resp.JSON200))
}

func updateViewFromFile(ctx context.Context, client *api.ClientWithResponses, opts *options.RootOptions, auth api.RequestEditorFn, boardID, viewID, file string) error {
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

	data, err = jsonutil.Sanitize(data)
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	resp, err := client.UpdateBoardViewWithBodyWithResponse(ctx, boardID, viewID, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("updating board view: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeViewDetail(opts, viewResponseToDetail(*resp.JSON200))
}

func getView(ctx context.Context, client *api.ClientWithResponses, auth api.RequestEditorFn, boardID, viewID string) (*api.BoardViewResponse, error) {
	resp, err := client.GetBoardViewWithResponse(ctx, boardID, viewID, auth)
	if err != nil {
		return nil, fmt.Errorf("getting board view: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return resp.JSON200, nil
}
