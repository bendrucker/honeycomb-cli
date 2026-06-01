package dataset

import (
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type datasetItem struct {
	Name        string  `json:"name" col:"Name"`
	Slug        string  `json:"slug" col:"Slug"`
	Description string  `json:"description,omitempty" col:"Description"`
	Columns     *int    `json:"columns,omitempty"`
	LastWritten *string `json:"last_written,omitempty"`
	CreatedAt   string  `json:"created_at" col:"Created"`
}

var datasetListTable = func() output.TableDef {
	table := output.TableFromTags[datasetItem]()
	table.Columns = append(table.Columns,
		output.Col("Columns", func(d datasetItem) string {
			if c := d.Columns; c != nil {
				return fmt.Sprintf("%d", *c)
			}
			return "—"
		}),
		output.Col("Last Written", func(d datasetItem) string {
			if lw := d.LastWritten; lw != nil {
				return *lw
			}
			return "—"
		}),
	)
	return table
}()

func NewListCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List datasets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDatasetList(cmd.Context(), opts)
		},
	}
}

func runDatasetList(ctx context.Context, opts *options.RootOptions) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	resp, err := client.ListDatasetsWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("listing datasets: %w", err)
	}

	datasets, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	items := make([]datasetItem, len(*datasets))
	for i, d := range *datasets {
		item := datasetItem{
			Name:        d.Name,
			Slug:        deref.String(d.Slug),
			Description: deref.String(d.Description),
			CreatedAt:   deref.String(d.CreatedAt),
		}
		if d.RegularColumnsCount.IsSpecified() && !d.RegularColumnsCount.IsNull() {
			v := d.RegularColumnsCount.MustGet()
			item.Columns = &v
		}
		if d.LastWrittenAt.IsSpecified() && !d.LastWrittenAt.IsNull() {
			v := d.LastWrittenAt.MustGet()
			item.LastWritten = &v
		}
		items[i] = item
	}

	return opts.OutputWriterList().WriteList(items, datasetListTable, "No datasets found.")
}
