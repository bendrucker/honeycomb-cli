package dataset

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions) *cobra.Command {
	var (
		name            string
		description     string
		expandJsonDepth int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a dataset",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var ejd *int
			if cmd.Flags().Changed("expand-json-depth") {
				ejd = &expandJsonDepth
			}
			return runDatasetCreate(cmd.Context(), opts, name, description, ejd)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Dataset name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Dataset description")
	cmd.Flags().IntVar(&expandJsonDepth, "expand-json-depth", 0, "Maximum unpacking depth of nested JSON fields")

	return cmd
}

func runDatasetCreate(ctx context.Context, opts *options.RootOptions, name, description string, expandJsonDepth *int) error {
	ios := opts.IOStreams

	if ios.CanPrompt() {
		var err error
		if name == "" {
			name, err = prompt.Line(ios.Out, ios.In, "Dataset name: ")
			if err != nil {
				return fmt.Errorf("reading name: %w", err)
			}
		}
		if description == "" {
			description, err = prompt.Line(ios.Out, ios.In, "Description (optional): ")
			if err != nil {
				return fmt.Errorf("reading description: %w", err)
			}
		}
	} else {
		if name == "" {
			return fmt.Errorf("--name is required in non-interactive mode")
		}
	}

	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	body := api.DatasetCreationPayload{
		Name: name,
	}
	if description != "" {
		body.Description = &description
	}
	if expandJsonDepth != nil {
		body.ExpandJsonDepth = expandJsonDepth
	}

	resp, err := client.CreateDatasetWithResponse(ctx, body, keyEditor(key))
	if err != nil {
		return fmt.Errorf("creating dataset: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	var dataset *api.Dataset
	switch {
	case resp.JSON201 != nil:
		dataset = resp.JSON201
	case resp.JSON200 != nil:
		dataset = resp.JSON200
	default:
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	detail := mapDatasetDetail(dataset)
	return writeDatasetDetail(opts, detail)
}
