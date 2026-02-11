package column

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewCalculatedListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List calculated fields in a dataset",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCalculatedList(cmd.Context(), opts, *dataset)
		},
	}
}

func runCalculatedList(ctx context.Context, opts *options.RootOptions, dataset string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.ListCalculatedFields(ctx, api.DatasetSlugOrAll(dataset), nil, keyEditor(key))
	if err != nil {
		return fmt.Errorf("listing calculated fields: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode, body); err != nil {
		return err
	}

	var fields []api.CalculatedField
	if err := json.Unmarshal(body, &fields); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	items := make([]calculatedItem, len(fields))
	for i, f := range fields {
		items[i] = toCalculatedItem(f)
	}

	return opts.OutputWriter().Write(items, calculatedListTable)
}
