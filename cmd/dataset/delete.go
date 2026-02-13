package dataset

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(opts *options.RootOptions) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <slug>",
		Short: "Delete a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDatasetDelete(cmd.Context(), opts, args[0], yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runDatasetDelete(ctx context.Context, opts *options.RootOptions, slug string, yes bool) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
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

		resp, err := client.GetDatasetWithResponse(ctx, slug, auth)
		if err != nil {
			return fmt.Errorf("getting dataset: %w", err)
		}
		if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		if resp.JSON200 == nil {
			return fmt.Errorf("unexpected response: %s", resp.Status())
		}

		answer, err := prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, fmt.Sprintf("Delete dataset %q? (y/N): ", resp.JSON200.Name))
		if err != nil {
			return err
		}
		if !strings.EqualFold(answer, "y") {
			return fmt.Errorf("aborted")
		}
	}

	httpResp, err := client.DeleteDataset(ctx, slug, auth)
	if err != nil {
		return fmt.Errorf("deleting dataset: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if err := api.CheckResponse(httpResp.StatusCode, body); err != nil {
		return err
	}

	return opts.OutputWriter().WriteDeleted(slug, fmt.Sprintf("Dataset %s deleted", slug))
}
