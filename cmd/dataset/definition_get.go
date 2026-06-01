package dataset

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewDefinitionGetCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <dataset-slug>",
		Short: "Get dataset definitions",
		Example: `  # Get the definitions for a dataset
  honeycomb dataset definition get my-dataset`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDefinitionGet(cmd.Context(), opts, args[0])
		},
	}
}

func runDefinitionGet(ctx context.Context, opts *options.RootOptions, slug string) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	resp, err := client.ListDatasetDefinitionsWithResponse(ctx, slug)
	if err != nil {
		return fmt.Errorf("getting dataset definitions: %w", err)
	}

	defs, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeDefinitions(opts, defs)
}

func writeDefinitions(opts *options.RootOptions, defs *api.DatasetDefinitions) error {
	rv := reflect.ValueOf(*defs)
	rt := rv.Type()

	var rows [][]string
	for i := range rt.NumField() {
		field := rv.Field(i)
		jsonTag := rt.Field(i).Tag.Get("json")
		name, _, _ := strings.Cut(jsonTag, ",")

		column, colType := "—", "—"
		if !field.IsNil() {
			def := field.Interface().(*api.DatasetDefinition)
			if def.Name != "" {
				column = def.Name
			}
			if def.ColumnType != nil {
				colType = string(*def.ColumnType)
			}
		}
		rows = append(rows, []string{name, column, colType})
	}

	return opts.OutputWriter().WriteDynamic(defs, output.DynamicTableDef{
		Headers: []string{"Field", "Column", "Type"},
		Rows:    rows,
	})
}
