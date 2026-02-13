package slo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewBurnAlertUpdateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "update <burn-alert-id>",
		Short: "Update a burn alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBurnAlertUpdate(cmd.Context(), opts, *dataset, args[0], file)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func runBurnAlertUpdate(ctx context.Context, opts *options.RootOptions, dataset, burnAlertID, file string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	data, err := readFile(opts, file)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.UpdateBurnAlertWithBodyWithResponse(ctx, dataset, burnAlertID, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("updating burn alert: %w", err)
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
