package column

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/spf13/cobra"
)

func NewListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var keyName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List columns in a dataset",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runColumnList(cmd.Context(), opts, *dataset, keyName)
		},
	}

	cmd.Flags().StringVar(&keyName, "key-name", "", "Filter to the column with this key name")

	return cmd
}

func runColumnList(ctx context.Context, opts *options.RootOptions, dataset, keyName string) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	var params *api.ListColumnsParams
	if keyName != "" {
		params = &api.ListColumnsParams{KeyName: &keyName}
	}

	columns, err := listColumns(ctx, client, dataset, params)
	if err != nil {
		return err
	}

	items := make([]columnItem, len(columns))
	for i, c := range columns {
		items[i] = columnItem{
			ID:          deref.String(c.Id),
			KeyName:     c.KeyName,
			Type:        deref.Enum(c.Type),
			Description: deref.String(c.Description),
			Hidden:      deref.Bool(c.Hidden),
			LastWritten: deref.String(c.LastWritten),
		}
	}

	return opts.OutputWriterList().Write(items, columnListTable)
}

// listColumns fetches columns via the raw ListColumns method because the
// generated ListColumnsWithResponse parser cannot unmarshal the response into
// its union type (JSON200 is struct { union json.RawMessage } which fails on
// the JSON array body).
func listColumns(ctx context.Context, client api.ClientInterface, dataset string, params *api.ListColumnsParams) ([]api.Column, error) {
	resp, err := client.ListColumns(ctx, dataset, params)
	if err != nil {
		return nil, fmt.Errorf("listing columns: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode, body); err != nil {
		return nil, err
	}

	var columns []api.Column
	if err := json.Unmarshal(body, &columns); err != nil {
		return nil, fmt.Errorf("parsing columns: %w", err)
	}

	return columns, nil
}
