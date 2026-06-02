package trigger

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
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
			if !command.AnyChanged(cmd, "file", "name", "description", "disabled", "enabled") {
				return fmt.Errorf("provide --file or at least one of --name, --description, --disabled, --enabled")
			}

			return runUpdate(cmd, opts, *dataset, args[0], file, name, description, disabled)
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

func runUpdate(cmd *cobra.Command, opts *options.RootOptions, dataset, triggerID, file, name, description string, disabled bool) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	var body api.TriggerResponse

	if cmd.Flags().Changed("file") {
		data, err := command.ReadDefinitionFile(opts.IOStreams, file)
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

		if cmd.Flags().Changed("name") {
			body.Name = &name
		}
		if cmd.Flags().Changed("description") {
			body.Description = &description
		}
		if cmd.Flags().Changed("disabled") {
			body.Disabled = &disabled
		}
		if cmd.Flags().Changed("enabled") {
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
