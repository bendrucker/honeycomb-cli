package dataset

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/jsonutil"
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
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	data, err := readDefinitionFile(opts, file)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.PatchDatasetDefinitionsWithBodyWithResponse(ctx, slug, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("updating dataset definitions: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeDefinitions(opts, resp.JSON200)
}

func readDefinitionFile(opts *options.RootOptions, file string) ([]byte, error) {
	var r io.Reader
	if file == "-" {
		r = opts.IOStreams.In
	} else {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		defer f.Close() //nolint:errcheck // best-effort close on read-only file
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	data, err = jsonutil.Sanitize(data)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return data, nil
}
