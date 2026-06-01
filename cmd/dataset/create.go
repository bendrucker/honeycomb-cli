package dataset

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions) *cobra.Command {
	var (
		name            string
		description     string
		expandJsonDepth int
		deleteProtected bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a dataset",
		Example: `  # Create a dataset
  honeycomb dataset create --name "My Dataset"

  # Create a dataset with a description and JSON unpacking
  honeycomb dataset create --name "My Dataset" \
    --description "Production traffic" --expand-json-depth 2`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var ejd *int
			if cmd.Flags().Changed("expand-json-depth") {
				ejd = &expandJsonDepth
			}
			var clearProtection bool
			if cmd.Flags().Changed("delete-protected") && !deleteProtected {
				clearProtection = true
			}
			return runDatasetCreate(cmd.Context(), opts, name, description, ejd, clearProtection)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Dataset name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Dataset description")
	cmd.Flags().IntVar(&expandJsonDepth, "expand-json-depth", 0, "Maximum unpacking depth of nested JSON fields")
	cmd.Flags().BoolVar(&deleteProtected, "delete-protected", true, "Protect dataset from deletion")

	return cmd
}

func runDatasetCreate(ctx context.Context, opts *options.RootOptions, name, description string, expandJsonDepth *int, clearProtection bool) error {
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

	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
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

	resp, err := client.CreateDatasetWithResponse(ctx, body)
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

	if clearProtection {
		dataset, err = clearDatasetProtection(ctx, client, dataset)
		if err != nil {
			return err
		}
	}

	detail := mapDatasetDetail(dataset)
	return writeDatasetDetail(opts, detail)
}

func clearDatasetProtection(ctx context.Context, client *api.ClientWithResponses, created *api.Dataset) (*api.Dataset, error) {
	protected := false
	body := api.DatasetUpdatePayload{
		Description: deref.String(created.Description),
		Settings: &struct {
			DeleteProtected *bool `json:"delete_protected,omitempty"`
		}{
			DeleteProtected: &protected,
		},
	}
	if created.ExpandJsonDepth != nil {
		body.ExpandJsonDepth = *created.ExpandJsonDepth
	}

	resp, err := client.UpdateDatasetWithResponse(ctx, deref.String(created.Slug), body)
	if err != nil {
		return nil, fmt.Errorf("clearing delete protection: %w", err)
	}

	dataset, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return nil, fmt.Errorf("clearing delete protection: %w", err)
	}
	return dataset, nil
}
