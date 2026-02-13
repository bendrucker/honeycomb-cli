package slo

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
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an SLO",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSLOCreate(cmd.Context(), opts, *dataset, file)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")

	return cmd
}

func runSLOCreate(ctx context.Context, opts *options.RootOptions, dataset, file string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	if file == "" {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--file is required in non-interactive mode")
		}
		file, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Path to SLO JSON file: ")
		if err != nil {
			return err
		}
		if file == "" {
			return fmt.Errorf("file path is required")
		}
	}

	data, err := readFile(opts, file)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.CreateSloWithBodyWithResponse(ctx, dataset, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("creating SLO: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	var slo api.SLO
	if err := json.Unmarshal(resp.Body, &slo); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	return writeSloDetail(opts, sloToDetail(slo))
}

func readFile(opts *options.RootOptions, file string) ([]byte, error) {
	var r io.Reader
	if file == "-" {
		r = opts.IOStreams.In
	} else {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		defer f.Close() //nolint:errcheck // best-effort close on read-only file
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var js json.RawMessage
	if err := json.Unmarshal(data, &js); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return data, nil
}
