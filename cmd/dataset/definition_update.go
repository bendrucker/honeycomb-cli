package dataset

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewDefinitionUpdateCmd(opts *options.RootOptions) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "update <dataset-slug>",
		Short: "Update dataset definitions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDefinitionUpdate(cmd.Context(), opts, args[0], file)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func runDefinitionUpdate(ctx context.Context, opts *options.RootOptions, slug, file string) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	data, err := command.ReadDefinitionFile(opts.IOStreams, file)
	if err != nil {
		return err
	}

	resp, err := client.PatchDatasetDefinitionsWithBodyWithResponse(ctx, slug, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("updating dataset definitions: %w", err)
	}

	defs, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeDefinitions(opts, defs)
}
