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
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	if !yes {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--yes is required in non-interactive mode")
		}

		view, err := getView(ctx, client, key, boardID, viewID)
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

	resp, err := client.DeleteBoardViewWithResponse(ctx, boardID, viewID, keyEditor(key))
	if err != nil {
		return fmt.Errorf("deleting board view: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(opts.IOStreams.Err, "View %s deleted\n", viewID)
	return nil
}
