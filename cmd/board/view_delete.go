package board

import (
	"context"
	"fmt"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewViewDeleteCmd(opts *options.RootOptions, board *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <view-id>",
		Short: "Delete a board view",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runViewDelete(cmd.Context(), opts, *board, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runViewDelete(ctx context.Context, opts *options.RootOptions, boardID, viewID string, yes bool) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	if !yes {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--yes is required in non-interactive mode")
		}

		view, err := getView(ctx, client, boardID, viewID)
		if err != nil {
			return err
		}

		name := viewID
		if view.Name != nil {
			name = *view.Name
		}

		answer, err := prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, fmt.Sprintf("Delete view %q? (y/N): ", name))
		if err != nil {
			return err
		}
		if !strings.EqualFold(answer, "y") {
			return fmt.Errorf("aborted")
		}
	}

	resp, err := client.DeleteBoardViewWithResponse(ctx, boardID, viewID)
	if err != nil {
		return fmt.Errorf("deleting board view: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(viewID, fmt.Sprintf("View %s deleted", viewID))
}
