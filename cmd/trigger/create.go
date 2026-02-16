package trigger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !cmd.Flags().Changed("file") {
				return fmt.Errorf("--file is required")
			}

			return runCreate(cmd, opts, *dataset, createFlags{
				file:        file,
				name:        name,
				hasName:     cmd.Flags().Changed("name"),
				description: description,
				hasDesc:     cmd.Flags().Changed("description"),
				disabled:    disabled,
				hasDisabled: cmd.Flags().Changed("disabled"),
				enabled:     enabled,
				hasEnabled:  cmd.Flags().Changed("enabled"),
			})
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

type createFlags struct {
	file        string
	name        string
	hasName     bool
	description string
	hasDesc     bool
	disabled    bool
	hasDisabled bool
	enabled     bool
	hasEnabled  bool
}

func runCreate(cmd *cobra.Command, opts *options.RootOptions, dataset string, flags createFlags) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	var r io.Reader
	if flags.file == "-" {
		r = opts.IOStreams.In
	} else {
		f, err := os.Open(flags.file)
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	if flags.hasName || flags.hasDesc || flags.hasDisabled || flags.hasEnabled {
		var body map[string]any
		if err := json.Unmarshal(data, &body); err != nil {
			return fmt.Errorf("parsing trigger JSON: %w", err)
		}

		if flags.hasName {
			body["name"] = flags.name
		}
		if flags.hasDesc {
			body["description"] = flags.description
		}
		if flags.hasDisabled {
			body["disabled"] = flags.disabled
		}
		if flags.hasEnabled {
			body["disabled"] = false
		}

		data, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding trigger: %w", err)
		}
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.CreateTriggerWithBodyWithResponse(cmd.Context(), dataset, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("creating trigger: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	detail := toDetail(*resp.JSON201)
	return writeTriggerDetail(opts, detail)
}
