package slo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewBurnAlertCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a burn alert",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runBurnAlertCreate(cmd.Context(), opts, *dataset, file)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")

	return cmd
}

func runBurnAlertCreate(ctx context.Context, opts *options.RootOptions, dataset, file string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	if file == "" {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--file is required in non-interactive mode")
		}
		file, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Path to burn alert JSON file: ")
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

	resp, err := client.CreateBurnAlertWithBodyWithResponse(ctx, dataset, "application/json", bytes.NewReader(data), keyEditor(key))
	if err != nil {
		return fmt.Errorf("creating burn alert: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	var detail burnAlertDetail
	if err := json.Unmarshal(resp.Body, &detail); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	return writeBurnAlertDetail(opts, detail)
}
