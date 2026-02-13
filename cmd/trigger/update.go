package trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file        string
		name        string
		description string
		disabled    bool
		enabled     bool
	)

	cmd := &cobra.Command{
		Use:   "update <trigger-id>",
		Short: "Update a trigger",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hasFile := cmd.Flags().Changed("file")
			hasName := cmd.Flags().Changed("name")
			hasDesc := cmd.Flags().Changed("description")
			hasDisabled := cmd.Flags().Changed("disabled")
			hasEnabled := cmd.Flags().Changed("enabled")

			if !hasFile && !hasName && !hasDesc && !hasDisabled && !hasEnabled {
				return fmt.Errorf("provide --file or at least one of --name, --description, --disabled, --enabled")
			}

			return runUpdate(cmd.Context(), opts, *dataset, args[0], updateFlags{
				file:        file,
				hasFile:     hasFile,
				name:        name,
				hasName:     hasName,
				description: description,
				hasDesc:     hasDesc,
				disabled:    disabled,
				hasDisabled: hasDisabled,
				enabled:     enabled,
				hasEnabled:  hasEnabled,
			})
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Path to JSON file with full trigger definition")
	cmd.Flags().StringVar(&name, "name", "", "Trigger name")
	cmd.Flags().StringVar(&description, "description", "", "Trigger description")
	cmd.Flags().BoolVar(&disabled, "disabled", false, "Disable the trigger")
	cmd.Flags().BoolVar(&enabled, "enabled", false, "Enable the trigger")

	cmd.MarkFlagsMutuallyExclusive("disabled", "enabled")
	cmd.MarkFlagsMutuallyExclusive("file", "name")
	cmd.MarkFlagsMutuallyExclusive("file", "description")
	cmd.MarkFlagsMutuallyExclusive("file", "disabled")
	cmd.MarkFlagsMutuallyExclusive("file", "enabled")

	return cmd
}

type updateFlags struct {
	file        string
	hasFile     bool
	name        string
	hasName     bool
	description string
	hasDesc     bool
	disabled    bool
	hasDisabled bool
	enabled     bool
	hasEnabled  bool
}

func runUpdate(ctx context.Context, opts *options.RootOptions, dataset, triggerID string, flags updateFlags) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	var body api.TriggerResponse

	if flags.hasFile {
		data, err := os.ReadFile(flags.file)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}
		if err := json.Unmarshal(data, &body); err != nil {
			return fmt.Errorf("parsing trigger JSON: %w", err)
		}
	} else {
		resp, err := client.GetTriggerWithResponse(ctx, dataset, triggerID, auth)
		if err != nil {
			return fmt.Errorf("getting trigger: %w", err)
		}
		if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		if resp.JSON200 == nil {
			return fmt.Errorf("unexpected response: %s", resp.Status())
		}
		body = *resp.JSON200

		if flags.hasName {
			body.Name = &flags.name
		}
		if flags.hasDesc {
			body.Description = &flags.description
		}
		if flags.hasDisabled {
			body.Disabled = &flags.disabled
		}
		if flags.hasEnabled {
			v := false
			body.Disabled = &v
		}
	}

	resp, err := client.UpdateTriggerWithResponse(ctx, dataset, triggerID, body, auth)
	if err != nil {
		return fmt.Errorf("updating trigger: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	detail := toDetail(*resp.JSON200)
	return writeTriggerDetail(opts, detail)
}
