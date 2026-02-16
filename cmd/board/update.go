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
	"github.com/bendrucker/honeycomb-cli/internal/jsonutil"
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

	data, err = stripPanelDataset(data)
	if err != nil {
		return fmt.Errorf("stripping panel dataset: %w", err)
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
		raw, readErr := io.ReadAll(r)
		if readErr != nil {
			return fmt.Errorf("reading file: %w", readErr)
		}

		var fillErr error
		data, fillErr = fillRequiredFields(ctx, client, auth, boardID, raw)
		if fillErr != nil {
			return fillErr
		}
		sanitized, sanitizeErr := jsonutil.Sanitize(data)
		if sanitizeErr != nil {
			return fmt.Errorf("invalid JSON: %w", sanitizeErr)
		}
		data = sanitized
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

	data, panelErr := stripPanelDataset(data)
	if panelErr != nil {
		return fmt.Errorf("stripping panel dataset: %w", panelErr)
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

// fillRequiredFields ensures that "name" and "type" are present in the board
// JSON, fetching the current board to fill them in when missing. This allows
// --replace to work without redundantly specifying fields that are already known.
func fillRequiredFields(ctx context.Context, client *api.ClientWithResponses, auth api.RequestEditorFn, boardID string, data []byte) ([]byte, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing board JSON: %w", err)
	}

	_, hasName := m["name"]
	_, hasType := m["type"]
	if hasName && hasType {
		return data, nil
	}

	current, err := getBoard(ctx, client, auth, boardID)
	if err != nil {
		return nil, err
	}

	if !hasName {
		raw, _ := json.Marshal(current.Name)
		m["name"] = raw
	}
	if !hasType {
		raw, _ := json.Marshal(current.Type)
		m["type"] = raw
	}

	return json.Marshal(m)
}

func encodeJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// stripPanelDataset removes the "dataset" field from "query_panel" objects
// within the "panels" array. The API returns dataset in query panels on read
// but rejects it on write.
func stripPanelDataset(data []byte) ([]byte, error) {
	var board map[string]json.RawMessage
	if err := json.Unmarshal(data, &board); err != nil {
		return nil, err
	}

	raw, ok := board["panels"]
	if !ok {
		return data, nil
	}

	var panels []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &panels); err != nil {
		return data, err
	}

	changed := false
	for i, panel := range panels {
		qp, ok := panel["query_panel"]
		if !ok {
			continue
		}
		var qpMap map[string]json.RawMessage
		if err := json.Unmarshal(qp, &qpMap); err != nil {
			continue
		}
		if _, has := qpMap["dataset"]; !has {
			continue
		}
		delete(qpMap, "dataset")
		changed = true
		reencoded, err := json.Marshal(qpMap)
		if err != nil {
			return nil, err
		}
		panels[i]["query_panel"] = reencoded
	}

	if !changed {
		return data, nil
	}

	reencoded, err := json.Marshal(panels)
	if err != nil {
		return nil, err
	}
	board["panels"] = reencoded
	return json.Marshal(board)
}
