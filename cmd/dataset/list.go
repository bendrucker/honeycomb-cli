package dataset

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

type datasetItem struct {
	Name        string  `json:"name"                    yaml:"name"`
	Slug        string  `json:"slug"                    yaml:"slug"`
	Description string  `json:"description,omitempty"   yaml:"description,omitempty"`
	Columns     *int    `json:"columns,omitempty"       yaml:"columns,omitempty"`
	LastWritten *string `json:"last_written,omitempty"  yaml:"last_written,omitempty"`
	CreatedAt   string  `json:"created_at"              yaml:"created_at"`
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
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.ListDatasetsWithResponse(ctx, keyEditor(key))
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
			Name: d.Name,
		}
		if d.Slug != nil {
			item.Slug = *d.Slug
		}
		if d.Description != nil {
			item.Description = *d.Description
		}
		if d.CreatedAt != nil {
			item.CreatedAt = *d.CreatedAt
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

	return opts.WriteFormatted(items, func(out io.Writer) error {
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tSLUG\tDESCRIPTION\tCOLUMNS\tLAST WRITTEN")
		for _, item := range items {
			cols := "—"
			if item.Columns != nil {
				cols = fmt.Sprintf("%d", *item.Columns)
			}
			lastWritten := "—"
			if item.LastWritten != nil {
				lastWritten = *item.LastWritten
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", item.Name, item.Slug, item.Description, cols, lastWritten)
		}
		return w.Flush()
	})
}

func keyEditor(key string) api.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		config.ApplyAuth(req, config.KeyConfig, key)
		return nil
	}
}
