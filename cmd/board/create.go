package board

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
	var (
		file string
		name string
		desc string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a board",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runBoardCreate(cmd, opts, file, name, desc)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&name, "name", "", "Board name")
	cmd.Flags().StringVar(&desc, "description", "", "Board description")

	cmd.MarkFlagsMutuallyExclusive("file", "name")
	cmd.MarkFlagsMutuallyExclusive("file", "description")

	return cmd
}

func runBoardCreate(cmd *cobra.Command, opts *options.RootOptions, file, name, desc string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	ctx := cmd.Context()

	if file != "" {
		return createFromFile(ctx, client, opts, auth, file)
	}

	if name == "" {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--name or --file is required in non-interactive mode")
		}
		name, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Board name: ")
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("board name is required")
		}
		if !cmd.Flags().Changed("description") {
			desc, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Description (optional): ")
			if err != nil {
				return err
			}
		}
	}

	board := api.Board{
		Name: name,
		Type: api.Flexible,
	}
	if desc != "" {
		board.Description = &desc
	}

	resp, err := client.CreateBoardWithResponse(ctx, board, auth)
	if err != nil {
		return fmt.Errorf("creating board: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeBoardDetail(opts, boardToDetail(*resp.JSON201))
}

func createFromFile(ctx context.Context, client *api.ClientWithResponses, opts *options.RootOptions, auth api.RequestEditorFn, file string) error {
	var r io.Reader
	if file == "-" {
		r = opts.IOStreams.In
	} else {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	raw, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	data, err := api.StripReadOnly(raw, "Board")
	if err != nil {
		return fmt.Errorf("stripping read-only fields: %w", err)
	}

	resp, err := client.CreateBoardWithBodyWithResponse(ctx, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("creating board: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeBoardDetail(opts, boardToDetail(*resp.JSON201))
}
