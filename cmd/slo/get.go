package slo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewGetCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var detailed bool

	cmd := &cobra.Command{
		Use:     "get <slo-id>",
		Aliases: []string{"view"},
		Short:   "Get an SLO",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSLOGet(cmd.Context(), opts, *dataset, args[0], detailed)
		},
	}

	cmd.Flags().BoolVar(&detailed, "detailed", false, "Include compliance and budget data (Enterprise)")

	return cmd
}

func runSLOGet(ctx context.Context, opts *options.RootOptions, dataset, sloID string, detailed bool) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	params := &api.GetSloParams{}
	if detailed {
		params.Detailed = ptr(true)
	}

	resp, err := client.GetSloWithResponse(ctx, dataset, sloID, params, keyEditor(key))
	if err != nil {
		return fmt.Errorf("getting SLO: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	// GetSloResp.JSON200 is a union type (unusable). Unmarshal resp.Body instead.
	var sloResp sloDetailedResponse
	if err := json.Unmarshal(resp.Body, &sloResp); err != nil {
		return fmt.Errorf("parsing SLO response: %w", err)
	}

	return writeSloDetail(opts, detailedToDetail(sloResp))
}

func ptr[T any](v T) *T {
	return &v
}
