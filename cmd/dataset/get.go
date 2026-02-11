package dataset

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

type datasetDetail struct {
	Name            string  `json:"name"                          yaml:"name"`
	Slug            string  `json:"slug"                          yaml:"slug"`
	Description     string  `json:"description,omitempty"         yaml:"description,omitempty"`
	ExpandJsonDepth *int    `json:"expand_json_depth,omitempty"   yaml:"expand_json_depth,omitempty"`
	Columns         *int    `json:"columns,omitempty"             yaml:"columns,omitempty"`
	LastWritten     *string `json:"last_written,omitempty"        yaml:"last_written,omitempty"`
	DeleteProtected bool    `json:"delete_protected"              yaml:"delete_protected"`
	CreatedAt       string  `json:"created_at"                    yaml:"created_at"`
}

func mapDatasetDetail(d *api.Dataset) datasetDetail {
	detail := datasetDetail{
		Name: d.Name,
	}
	if d.Slug != nil {
		detail.Slug = *d.Slug
	}
	if d.Description != nil {
		detail.Description = *d.Description
	}
	if d.ExpandJsonDepth != nil {
		detail.ExpandJsonDepth = d.ExpandJsonDepth
	}
	if d.CreatedAt != nil {
		detail.CreatedAt = *d.CreatedAt
	}
	if d.RegularColumnsCount.IsSpecified() && !d.RegularColumnsCount.IsNull() {
		v := d.RegularColumnsCount.MustGet()
		detail.Columns = &v
	}
	if d.LastWrittenAt.IsSpecified() && !d.LastWrittenAt.IsNull() {
		v := d.LastWrittenAt.MustGet()
		detail.LastWritten = &v
	}
	if d.Settings != nil && d.Settings.DeleteProtected != nil {
		detail.DeleteProtected = *d.Settings.DeleteProtected
	}
	return detail
}

func NewGetCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <slug>",
		Short: "Get a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDatasetGet(cmd.Context(), opts, args[0])
		},
	}
}

func runDatasetGet(ctx context.Context, opts *options.RootOptions, slug string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetDatasetWithResponse(ctx, slug, keyEditor(key))
	if err != nil {
		return fmt.Errorf("getting dataset: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	detail := mapDatasetDetail(resp.JSON200)
	return writeDatasetDetail(opts, detail)
}

func writeDatasetDetail(opts *options.RootOptions, detail datasetDetail) error {
	return opts.OutputWriter().WriteValue(detail, func(out io.Writer) error {
		tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "Name:\t%s\n", detail.Name)
		_, _ = fmt.Fprintf(tw, "Slug:\t%s\n", detail.Slug)
		if detail.Description != "" {
			_, _ = fmt.Fprintf(tw, "Description:\t%s\n", detail.Description)
		}
		if detail.ExpandJsonDepth != nil {
			_, _ = fmt.Fprintf(tw, "Expand JSON Depth:\t%d\n", *detail.ExpandJsonDepth)
		}
		if detail.Columns != nil {
			_, _ = fmt.Fprintf(tw, "Columns:\t%d\n", *detail.Columns)
		}
		if detail.LastWritten != nil {
			_, _ = fmt.Fprintf(tw, "Last Written:\t%s\n", *detail.LastWritten)
		}
		_, _ = fmt.Fprintf(tw, "Delete Protected:\t%t\n", detail.DeleteProtected)
		_, _ = fmt.Fprintf(tw, "Created At:\t%s\n", detail.CreatedAt)
		return tw.Flush()
	})
}
