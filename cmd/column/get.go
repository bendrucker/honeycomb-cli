package column

import (
	"context"
	"fmt"
	"regexp"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

// columnIDPattern matches Honeycomb base58 column IDs, which use the Bitcoin
// base58 alphabet (no 0, O, I, or l, and no separators). Anything outside this
// set, including key names with underscores or dots, is treated as a key name
// and resolved via the list endpoint.
var columnIDPattern = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]+$`)

func NewGetCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <column-id-or-key-name>",
		Short: "Get a column by ID or key name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runColumnGet(cmd.Context(), opts, *dataset, args[0])
		},
	}
}

func runColumnGet(ctx context.Context, opts *options.RootOptions, dataset, column string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	if !columnIDPattern.MatchString(column) {
		return getColumnByKeyName(ctx, opts, client, dataset, column)
	}

	resp, err := client.GetColumnWithResponse(ctx, dataset, column)
	if err != nil {
		return fmt.Errorf("getting column: %w", err)
	}

	col, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeColumnDetail(opts, *col)
}

func getColumnByKeyName(ctx context.Context, opts *options.RootOptions, client api.ClientInterface, dataset, keyName string) error {
	columns, err := listColumns(ctx, client, dataset, &api.ListColumnsParams{KeyName: &keyName})
	if err != nil {
		return err
	}

	for _, col := range columns {
		if col.KeyName == keyName {
			return writeColumnDetail(opts, col)
		}
	}

	return fmt.Errorf("no column found with key name %q in dataset %q", keyName, dataset)
}
