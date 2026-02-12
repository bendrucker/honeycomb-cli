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
	"github.com/spf13/cobra"
)

func NewViewUpdateCmd(opts *options.RootOptions, board *string) *cobra.Command {
	var (
		file string
		name string
	)

	cmd := &cobra.Command{
		Use:   "update <view-id>",
		Short: "Update a board view",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runViewUpdate(cmd, opts, *board, args[0], file, name)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&name, "name", "", "View name")

	cmd.MarkFlagsMutuallyExclusive("file", "name")

	return cmd
}

func runViewUpdate(cmd *cobra.Command, opts *options.RootOptions, boardID, viewID, file, name string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	ctx := cmd.Context()

	if file != "" {
		return updateViewFromFile(ctx, client, opts, key, boardID, viewID, file)
	}

	if !cmd.Flags().Changed("name") {
		return fmt.Errorf("--file or --name is required")
	}

	current, err := getView(ctx, client, key, boardID, viewID)
	if err != nil {
		return err
	}

	body := api.UpdateBoardViewJSONRequestBody{
		Name:    name,
		Filters: []api.BoardViewFilter{},
	}
	if current.Filters != nil {
		body.Filters = *current.Filters
	}

	resp, err := client.UpdateBoardViewWithResponse(ctx, boardID, viewID, body, keyEditor(key))
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

func updateViewFromFile(ctx context.Context, client *api.ClientWithResponses, opts *options.RootOptions, key, boardID, viewID, file string) error {
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

	resp, err := client.UpdateBoardViewWithBodyWithResponse(ctx, boardID, viewID, "application/json", bytes.NewReader(data), keyEditor(key))
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

func getView(ctx context.Context, client *api.ClientWithResponses, key, boardID, viewID string) (*api.BoardViewResponse, error) {
	resp, err := client.GetBoardViewWithResponse(ctx, boardID, viewID, keyEditor(key))
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
