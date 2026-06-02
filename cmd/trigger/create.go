package trigger

import (
	"bytes"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file        string
		name        string
		description string
		disabled    bool
		enabled     bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trigger",
		Example: `  # Create a trigger from a file
  honeycomb trigger create --dataset my-dataset --file trigger.json

  # Create from a file, overriding the name
  honeycomb trigger create --dataset my-dataset --file trigger.json \
    --name "High latency"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !cmd.Flags().Changed("file") {
				return fmt.Errorf("--file is required")
			}

			return runCreate(cmd, opts, *dataset, file, name, description, disabled)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file with trigger definition (- for stdin)")
	cmd.Flags().StringVar(&name, "name", "", "Trigger name (overrides file)")
	cmd.Flags().StringVar(&description, "description", "", "Trigger description (overrides file)")
	cmd.Flags().BoolVar(&disabled, "disabled", false, "Disable the trigger (overrides file)")
	cmd.Flags().BoolVar(&enabled, "enabled", false, "Enable the trigger (overrides file)")

	cmd.MarkFlagsMutuallyExclusive("disabled", "enabled")

	return cmd
}

func runCreate(cmd *cobra.Command, opts *options.RootOptions, dataset, file, name, description string, disabled bool) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	data, err := command.ReadDefinitionFile(opts.IOStreams, file)
	if err != nil {
		return err
	}

	overrides := map[string]any{}
	if cmd.Flags().Changed("name") {
		overrides["name"] = name
	}
	if cmd.Flags().Changed("description") {
		overrides["description"] = description
	}
	if cmd.Flags().Changed("disabled") {
		overrides["disabled"] = disabled
	}
	if cmd.Flags().Changed("enabled") {
		overrides["disabled"] = false
	}

	data, err = command.ApplyOverrides(data, overrides)
	if err != nil {
		return err
	}

	resp, err := client.CreateTriggerWithBodyWithResponse(cmd.Context(), dataset, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating trigger: %w", err)
	}

	trigger, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON201)
	if err != nil {
		return err
	}

	return writeTriggerDetail(opts, toDetail(*trigger))
}
