package dataset

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type datasetDetail struct {
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	Description     string  `json:"description,omitempty"`
	ExpandJsonDepth *int    `json:"expand_json_depth,omitempty"`
	Columns         *int    `json:"columns,omitempty"`
	LastWritten     *string `json:"last_written,omitempty"`
	DeleteProtected bool    `json:"delete_protected"`
	CreatedAt       string  `json:"created_at"`
}

func mapDatasetDetail(d *api.Dataset) datasetDetail {
	detail := datasetDetail{
		Name:        d.Name,
		Slug:        deref.String(d.Slug),
		Description: deref.String(d.Description),
		CreatedAt:   deref.String(d.CreatedAt),
	}
	detail.ExpandJsonDepth = d.ExpandJsonDepth
	if d.RegularColumnsCount.IsSpecified() && !d.RegularColumnsCount.IsNull() {
		v := d.RegularColumnsCount.MustGet()
		detail.Columns = &v
	}
	if d.LastWrittenAt.IsSpecified() && !d.LastWrittenAt.IsNull() {
		v := d.LastWrittenAt.MustGet()
		detail.LastWritten = &v
	}
	if d.Settings != nil {
		detail.DeleteProtected = deref.Bool(d.Settings.DeleteProtected)
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
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.GetDatasetWithResponse(ctx, slug, auth)
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
	fields := []output.Field{
		{Label: "Name", Value: detail.Name},
		{Label: "Slug", Value: detail.Slug},
	}
	if detail.Description != "" {
		fields = append(fields, output.Field{Label: "Description", Value: detail.Description})
	}
	if detail.ExpandJsonDepth != nil {
		fields = append(fields, output.Field{Label: "Expand JSON Depth", Value: fmt.Sprintf("%d", *detail.ExpandJsonDepth)})
	}
	if detail.Columns != nil {
		fields = append(fields, output.Field{Label: "Columns", Value: fmt.Sprintf("%d", *detail.Columns)})
	}
	if detail.LastWritten != nil {
		fields = append(fields, output.Field{Label: "Last Written", Value: *detail.LastWritten})
	}
	fields = append(fields,
		output.Field{Label: "Delete Protected", Value: strconv.FormatBool(detail.DeleteProtected)},
		output.Field{Label: "Created At", Value: detail.CreatedAt},
	)
	return opts.OutputWriter().WriteFields(detail, fields)
}
