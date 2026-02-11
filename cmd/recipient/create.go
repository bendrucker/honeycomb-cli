package recipient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a recipient",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRecipientCreate(cmd.Context(), opts, file)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")

	return cmd
}

func runRecipientCreate(ctx context.Context, opts *options.RootOptions, file string) error {
	if file == "" {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--file is required in non-interactive mode")
		}
		var err error
		file, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "File path: ")
		if err != nil {
			return err
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

	data, err := readFile(opts, file)
	if err != nil {
		return err
	}

	resp, err := client.CreateRecipientWithBodyWithResponse(ctx, "application/json", bytes.NewReader(data), keyEditor(key))
	if err != nil {
		return fmt.Errorf("creating recipient: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	detail, err := parseRecipientBody(resp.Body)
	if err != nil {
		return err
	}

	return writeRecipientDetail(opts, detail)
}

func readFile(opts *options.RootOptions, file string) ([]byte, error) {
	var r io.Reader
	if file == "-" {
		r = opts.IOStreams.In
	} else {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return data, nil
}
