package trigger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
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
		Example: `  # Disable a trigger
  honeycomb trigger update abc123 --dataset my-dataset --disabled

  # Update a trigger from a file
  honeycomb trigger update abc123 --dataset my-dataset --file trigger.json`,
		Args: cobra.ExactArgs(1),
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

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file with full trigger definition (- for stdin)")
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
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	var body api.TriggerResponse

	if flags.hasFile {
		data, err := command.ReadDefinitionFile(opts.IOStreams, flags.file)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &body); err != nil {
			return fmt.Errorf("parsing trigger JSON: %w", err)
		}
	} else {
		resp, err := client.GetTriggerWithResponse(ctx, dataset, triggerID)
		if err != nil {
			return fmt.Errorf("getting trigger: %w", err)
		}
		current, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
		if err != nil {
			return err
		}
		body = *current

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

	// A GET returns both query_id and the resolved inline query, but the update
	// endpoint rejects a body carrying both (HTTP 400). When query_id is set it
	// is authoritative, so drop the redundant inline query before sending.
	if body.QueryId != nil {
		body.Query = nil
	}

	data, err := api.MarshalStrippingReadOnly(body, "TriggerResponse")
	if err != nil {
		return fmt.Errorf("encoding trigger: %w", err)
	}

	resp, err := client.UpdateTriggerWithBodyWithResponse(ctx, dataset, triggerID, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("updating trigger: %w", err)
	}

	updated, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeTriggerDetail(opts, toDetail(*updated))
}
