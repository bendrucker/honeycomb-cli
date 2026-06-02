package board

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
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
		Example: `  # Create a board with a name
  honeycomb board create --name "Latency"

  # Create a board from a JSON file
  honeycomb board create --file board.json`,
		Long: `Create a board.

Use --file to provide a full board definition as JSON. The preset_filters array
requires both "column" (the column name) and "alias" (a display label, max 50
characters) for each entry:

  {"preset_filters": [{"column": "service.name", "alias": "Service"}]}`,
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
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	if file != "" {
		return createFromFile(ctx, client, opts, file)
	}

	if name == "" {
		promptable := opts.IOStreams.CanPrompt()
		name, err = command.Resolve(opts.IOStreams, name, command.Field{
			Prompt:            "Board name: ",
			Required:          true,
			NonInteractiveErr: fmt.Errorf("--name or --file is required in non-interactive mode"),
			EmptyErr:          fmt.Errorf("board name is required"),
		})
		if err != nil {
			return err
		}
		if promptable && !cmd.Flags().Changed("description") {
			desc, err = command.Resolve(opts.IOStreams, desc, command.Field{
				Prompt: "Description (optional): ",
			})
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

	resp, err := client.CreateBoardWithResponse(ctx, board)
	if err != nil {
		return fmt.Errorf("creating board: %w", err)
	}

	created, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON201)
	if err != nil {
		return err
	}

	return writeBoardDetail(opts, boardToDetail(*created))
}

func createFromFile(ctx context.Context, client *api.ClientWithResponses, opts *options.RootOptions, file string) error {
	raw, err := command.ReadDefinitionFile(opts.IOStreams, file)
	if err != nil {
		return err
	}

	data, err := api.StripReadOnly(raw, "Board")
	if err != nil {
		return fmt.Errorf("stripping read-only fields: %w", err)
	}

	data, err = stripPanelDataset(data)
	if err != nil {
		return fmt.Errorf("stripping panel dataset: %w", err)
	}

	resp, err := client.CreateBoardWithBodyWithResponse(ctx, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating board: %w", err)
	}

	created, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON201)
	if err != nil {
		return err
	}

	return writeBoardDetail(opts, boardToDetail(*created))
}
