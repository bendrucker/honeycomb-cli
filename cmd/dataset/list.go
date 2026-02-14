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
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description string  `json:"description,omitempty"`
	Columns     *int    `json:"columns,omitempty"`
	LastWritten *string `json:"last_written,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

var datasetListTable = output.TableDef{
	Columns: []output.Column{
		{Header: "Name", Value: func(v any) string { return v.(datasetItem).Name }},
		{Header: "Slug", Value: func(v any) string { return v.(datasetItem).Slug }},
		{Header: "Description", Value: func(v any) string { return v.(datasetItem).Description }},
		{Header: "Columns", Value: func(v any) string {
			if c := v.(datasetItem).Columns; c != nil {
				return fmt.Sprintf("%d", *c)
			}
			return "—"
		}},
		{Header: "Last Written", Value: func(v any) string {
			if lw := v.(datasetItem).LastWritten; lw != nil {
				return *lw
			}
			return "—"
		}},
	},
}

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
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.ListDatasetsWithResponse(ctx, auth)
	if err != nil {
		return fmt.Errorf("listing datasets: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	items := make([]datasetItem, len(*resp.JSON200))
	for i, d := range *resp.JSON200 {
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

	return opts.OutputWriterList().Write(items, datasetListTable)
}
