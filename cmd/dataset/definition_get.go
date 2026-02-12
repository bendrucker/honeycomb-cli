package dataset

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewDefinitionGetCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <dataset-slug>",
		Short: "Get dataset definitions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDefinitionGet(cmd.Context(), opts, args[0])
		},
	}
}

func runDefinitionGet(ctx context.Context, opts *options.RootOptions, slug string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.ListDatasetDefinitionsWithResponse(ctx, slug, keyEditor(key))
	if err != nil {
		return fmt.Errorf("getting dataset definitions: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeDefinitions(opts, resp.JSON200)
}

func writeDefinitions(opts *options.RootOptions, defs *api.DatasetDefinitions) error {
	return opts.OutputWriter().WriteValue(defs, func(out io.Writer) error {
		tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "FIELD\tCOLUMN\tTYPE\n")

		rv := reflect.ValueOf(*defs)
		rt := rv.Type()
		for i := range rt.NumField() {
			field := rv.Field(i)
			if field.IsNil() {
				continue
			}
			def := field.Interface().(*api.DatasetDefinition)
			jsonTag := rt.Field(i).Tag.Get("json")
			name, _, _ := strings.Cut(jsonTag, ",")
			colType := ""
			if def.ColumnType != nil {
				colType = string(*def.ColumnType)
			}
			_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", name, def.Name, colType)
		}

		return tw.Flush()
	})
}
