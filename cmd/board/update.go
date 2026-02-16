package board

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions) *cobra.Command {
	var (
		file    string
		replace bool
		name    string
		desc    string
	)

	cmd := &cobra.Command{
		Use:   "update <board-id>",
		Short: "Update a board",
		Long: `Update a board.

Use --file to provide a full or partial board definition as JSON. The
preset_filters array requires both "column" (the column name) and "alias"
(a display label, max 50 characters) for each entry:

  {"preset_filters": [{"column": "service.name", "alias": "Service"}]}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBoardUpdate(cmd, opts, args[0], file, replace, name, desc)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().BoolVar(&replace, "replace", false, "Replace the board entirely (with --file)")
	cmd.Flags().StringVar(&name, "name", "", "Board name")
	cmd.Flags().StringVar(&desc, "description", "", "Board description")

	cmd.MarkFlagsMutuallyExclusive("file", "name")
	cmd.MarkFlagsMutuallyExclusive("file", "description")

	return cmd
}

func runBoardUpdate(cmd *cobra.Command, opts *options.RootOptions, boardID, file string, replace bool, name, desc string) error {
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
		return updateFromFile(ctx, client, opts, auth, boardID, file, replace)
	}

	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("description") {
		return fmt.Errorf("--file, --name, or --description is required")
	}

	current, err := getBoard(ctx, client, auth, boardID)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("name") {
		current.Name = name
	}
	if cmd.Flags().Changed("description") {
		current.Description = &desc
	}

	data, err := api.MarshalStrippingReadOnly(current, "Board")
	if err != nil {
		return fmt.Errorf("encoding board: %w", err)
	}

	resp, err := client.UpdateBoardWithBodyWithResponse(ctx, boardID, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("updating board: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeBoardDetail(opts, boardToDetail(*resp.JSON200))
}

func updateFromFile(ctx context.Context, client *api.ClientWithResponses, opts *options.RootOptions, auth api.RequestEditorFn, boardID, file string, replace bool) error {
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

	var data []byte

	if replace {
		var err error
		data, err = io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}
	} else {
		incoming, err := readBoardJSON(r)
		if err != nil {
			return err
		}

		current, err := getBoard(ctx, client, auth, boardID)
		if err != nil {
			return err
		}

		mergeBoard(current, &incoming)

		data, err = encodeJSON(current)
		if err != nil {
			return fmt.Errorf("encoding board: %w", err)
		}
	}

	data, stripErr := api.StripReadOnly(data, "Board")
	if stripErr != nil {
		return fmt.Errorf("stripping read-only fields: %w", stripErr)
	}

	resp, err := client.UpdateBoardWithBodyWithResponse(ctx, boardID, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("updating board: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeBoardDetail(opts, boardToDetail(*resp.JSON200))
}

func getBoard(ctx context.Context, client *api.ClientWithResponses, auth api.RequestEditorFn, boardID string) (*api.Board, error) {
	resp, err := client.GetBoardWithResponse(ctx, boardID, auth)
	if err != nil {
		return nil, fmt.Errorf("getting board: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// mergeBoard applies non-zero fields from src onto dst.
// Type is intentionally not merged because it is always "flexible".
func mergeBoard(dst *api.Board, src *api.Board) {
	if src.Name != "" {
		dst.Name = src.Name
	}
	if src.Description != nil {
		dst.Description = src.Description
	}
	if src.LayoutGeneration != nil {
		dst.LayoutGeneration = src.LayoutGeneration
	}
	if src.Panels != nil {
		dst.Panels = src.Panels
	}
	if src.PresetFilters != nil {
		dst.PresetFilters = src.PresetFilters
	}
	if src.Tags != nil {
		dst.Tags = src.Tags
	}
}

func encodeJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
