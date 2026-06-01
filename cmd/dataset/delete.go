package dataset

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
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
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	proceed, err := command.ConfirmDelete(opts.IOStreams, yes, "dataset", slug, func() (string, error) {
		resp, err := client.GetDatasetWithResponse(ctx, slug)
		if err != nil {
			return "", fmt.Errorf("getting dataset: %w", err)
		}
		if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
			return "", err
		}
		if resp.JSON200 == nil {
			return "", fmt.Errorf("unexpected response: %s", resp.Status())
		}
		return resp.JSON200.Name, nil
	})
	if err != nil {
		return err
	}
	if !proceed {
		return fmt.Errorf("aborted")
	}

	httpResp, err := client.DeleteDataset(ctx, slug)
	if err != nil {
		return fmt.Errorf("deleting dataset: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if err := api.CheckResponse(httpResp.StatusCode, body); err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 409 && strings.Contains(apiErr.Message, "delete protected") {
			return fmt.Errorf("dataset %s is delete protected; disable protection first with: honeycomb dataset update %s --delete-protected=false", slug, slug)
		}
		return err
	}

	return opts.OutputWriter().WriteDeleted(slug, fmt.Sprintf("Dataset %s deleted", slug))
}
