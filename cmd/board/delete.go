package board

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(opts *options.RootOptions) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <board-id>",
		Short: "Delete a board",
		Example: `  # Delete a board, prompting for confirmation
  honeycomb board delete abc123

  # Delete without confirmation
  honeycomb board delete abc123 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBoardDelete(cmd.Context(), opts, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runBoardDelete(ctx context.Context, opts *options.RootOptions, boardID string, yes bool) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	proceed, err := command.ConfirmDelete(opts.IOStreams, yes, "board", boardID, func() (string, error) {
		board, err := getBoard(ctx, client, boardID)
		if err != nil {
			return "", err
		}
		return board.Name, nil
	})
	if err != nil {
		return err
	}
	if !proceed {
		return fmt.Errorf("aborted")
	}

	resp, err := client.DeleteBoardWithResponse(ctx, boardID)
	if err != nil {
		return fmt.Errorf("deleting board: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(boardID, fmt.Sprintf("Board %s deleted", boardID))
}
