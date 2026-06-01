package column

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewCalculatedListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List calculated columns in a dataset",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCalculatedList(cmd.Context(), opts, *dataset)
		},
	}
}

func runCalculatedList(ctx context.Context, opts *options.RootOptions, dataset string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	// Use the raw ListCalculatedFields method because the generated
	// ListCalculatedFieldsWithResponse parser cannot unmarshal the response
	// into its union type (JSON200 is struct { union json.RawMessage }).
	resp, err := client.ListCalculatedFields(ctx, dataset, nil)
	if err != nil {
		return fmt.Errorf("listing calculated columns: %w", err)
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
		return fmt.Errorf("parsing calculated columns: %w", err)
	}

	items := make([]calculatedItem, len(fields))
	for i, f := range fields {
		items[i] = toCalculatedItem(f)
	}

	return opts.OutputWriterList().WriteList(items, calculatedListTable, "No calculated columns found.")
}
