package slo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file       string
		name       string
		desc       string
		target     int
		timePeriod int
	)

	cmd := &cobra.Command{
		Use:   "update <slo-id>",
		Short: "Update an SLO",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSLOUpdate(cmd, opts, *dataset, args[0], file, name, desc, target, timePeriod)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&name, "name", "", "SLO name")
	cmd.Flags().StringVar(&desc, "description", "", "SLO description")
	cmd.Flags().IntVar(&target, "target", 0, "Target per million")
	cmd.Flags().IntVar(&timePeriod, "time-period", 0, "Time period in days")

	cmd.MarkFlagsMutuallyExclusive("file", "name")
	cmd.MarkFlagsMutuallyExclusive("file", "description")
	cmd.MarkFlagsMutuallyExclusive("file", "target")
	cmd.MarkFlagsMutuallyExclusive("file", "time-period")

	return cmd
}

func runSLOUpdate(cmd *cobra.Command, opts *options.RootOptions, dataset, sloID, file, name, desc string, target, timePeriod int) error {
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
		return updateFromFile(ctx, client, opts, key, dataset, sloID, file)
	}

	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("description") && !cmd.Flags().Changed("target") && !cmd.Flags().Changed("time-period") {
		return fmt.Errorf("--file, --name, --description, --target, or --time-period is required")
	}

	current, err := getSLO(ctx, client, key, dataset, sloID)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("name") {
		current.Name = name
	}
	if cmd.Flags().Changed("description") {
		current.Description = &desc
	}
	if cmd.Flags().Changed("target") {
		current.TargetPerMillion = target
	}
	if cmd.Flags().Changed("time-period") {
		current.TimePeriodDays = timePeriod
	}

	resp, err := client.UpdateSloWithResponse(ctx, dataset, sloID, *current, keyEditor(key))
	if err != nil {
		return fmt.Errorf("updating SLO: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return opts.OutputWriter().Write(sloToDetail(*resp.JSON200), output.TableDef{})
}

func updateFromFile(ctx context.Context, client *api.ClientWithResponses, opts *options.RootOptions, key, dataset, sloID, file string) error {
	data, err := readFile(opts, file)
	if err != nil {
		return err
	}

	resp, err := client.UpdateSloWithBodyWithResponse(ctx, dataset, sloID, "application/json", bytes.NewReader(data), keyEditor(key))
	if err != nil {
		return fmt.Errorf("updating SLO: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return opts.OutputWriter().Write(sloToDetail(*resp.JSON200), output.TableDef{})
}

func getSLO(ctx context.Context, client *api.ClientWithResponses, key, dataset, sloID string) (*api.SLO, error) {
	resp, err := client.GetSloWithResponse(ctx, dataset, sloID, &api.GetSloParams{}, keyEditor(key))
	if err != nil {
		return nil, fmt.Errorf("getting SLO: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return nil, err
	}

	// GetSloResp.JSON200 is a union type (unusable). Unmarshal resp.Body instead.
	var s api.SLO
	if err := json.Unmarshal(resp.Body, &s); err != nil {
		return nil, fmt.Errorf("parsing SLO response: %w", err)
	}

	return &s, nil
}
