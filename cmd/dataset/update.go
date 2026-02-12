package dataset

import (
	"context"
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
			var dp *bool
			if cmd.Flags().Changed("delete-protected") {
				dp = &deleteProtected
			}
			var ejd *int
			if cmd.Flags().Changed("expand-json-depth") {
				ejd = &expandJsonDepth
			}
			var desc *string
			if cmd.Flags().Changed("description") {
				desc = &description
			}
			return runDatasetUpdate(cmd.Context(), opts, args[0], desc, ejd, dp)
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "Dataset description")
	cmd.Flags().IntVar(&expandJsonDepth, "expand-json-depth", 0, "Maximum unpacking depth of nested JSON fields")
	cmd.Flags().BoolVar(&deleteProtected, "delete-protected", false, "Protect dataset from deletion")

	return cmd
}

func runDatasetUpdate(ctx context.Context, opts *options.RootOptions, slug string, description *string, expandJsonDepth *int, deleteProtected *bool) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	body := api.DatasetUpdatePayload{}
	if description != nil {
		body.Description = *description
	}
	if expandJsonDepth != nil {
		body.ExpandJsonDepth = *expandJsonDepth
	}
	if deleteProtected != nil {
		body.Settings = &struct {
			DeleteProtected *bool `json:"delete_protected,omitempty"`
		}{
			DeleteProtected: deleteProtected,
		}
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

	detail := mapDatasetDetail(resp.JSON200)
	return writeDatasetDetail(opts, detail)
}
