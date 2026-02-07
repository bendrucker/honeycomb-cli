package trigger

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trigger",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCreate(cmd.Context(), opts, *dataset, file)
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Path to JSON file with trigger definition")

	return cmd
}

func runCreate(ctx context.Context, opts *options.RootOptions, dataset, file string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	if file == "" {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--file is required in non-interactive mode")
		}
		var promptErr error
		file, promptErr = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "File path: ")
		if promptErr != nil {
			return promptErr
		}
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.CreateTriggerWithBodyWithResponse(ctx, dataset, "application/json", bytes.NewReader(data), keyEditor(key))
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
	return opts.OutputWriter().Write(detail, triggerViewTable)
}
