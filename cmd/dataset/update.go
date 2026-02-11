package dataset

import (
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions) *cobra.Command {
	var (
		description     string
		expandJsonDepth int
		deleteProtected bool
	)

	cmd := &cobra.Command{
		Use:   "update <slug>",
		Short: "Update a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("description") && !cmd.Flags().Changed("expand-json-depth") && !cmd.Flags().Changed("delete-protected") {
				return fmt.Errorf("at least one of --description, --expand-json-depth, or --delete-protected is required")
			}
			return runDatasetUpdate(cmd, opts, args[0], description, expandJsonDepth, deleteProtected)
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "Dataset description")
	cmd.Flags().IntVar(&expandJsonDepth, "expand-json-depth", 0, "Maximum unpacking depth of nested JSON fields")
	cmd.Flags().BoolVar(&deleteProtected, "delete-protected", false, "Protect dataset from deletion")

	return cmd
}

func runDatasetUpdate(cmd *cobra.Command, opts *options.RootOptions, slug, description string, expandJsonDepth int, deleteProtected bool) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	ctx := cmd.Context()

	current, err := client.GetDatasetWithResponse(ctx, slug, keyEditor(key))
	if err != nil {
		return fmt.Errorf("getting dataset: %w", err)
	}
	if err := api.CheckResponse(current.StatusCode(), current.Body); err != nil {
		return err
	}
	if current.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", current.Status())
	}

	ds := current.JSON200
	body := api.DatasetUpdatePayload{}
	if ds.Description != nil {
		body.Description = *ds.Description
	}
	if ds.ExpandJsonDepth != nil {
		body.ExpandJsonDepth = *ds.ExpandJsonDepth
	}
	if ds.Settings != nil && ds.Settings.DeleteProtected != nil {
		body.Settings = &struct {
			DeleteProtected *bool `json:"delete_protected,omitempty"`
		}{DeleteProtected: ds.Settings.DeleteProtected}
	}

	if cmd.Flags().Changed("description") {
		body.Description = description
	}
	if cmd.Flags().Changed("expand-json-depth") {
		body.ExpandJsonDepth = expandJsonDepth
	}
	if cmd.Flags().Changed("delete-protected") {
		body.Settings = &struct {
			DeleteProtected *bool `json:"delete_protected,omitempty"`
		}{DeleteProtected: &deleteProtected}
	}

	resp, err := client.UpdateDatasetWithResponse(ctx, slug, body, keyEditor(key))
	if err != nil {
		return fmt.Errorf("updating dataset: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeDatasetDetail(opts, mapDatasetDetail(resp.JSON200))
}
