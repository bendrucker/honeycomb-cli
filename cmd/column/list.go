package column

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewListCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List columns in a dataset",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runColumnList(cmd.Context(), opts, *dataset)
		},
	}
}

func runColumnList(ctx context.Context, opts *options.RootOptions, dataset string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	// Use the raw ListColumns method because the generated ListColumnsWithResponse
	// parser cannot unmarshal the response into its union type (JSON200 is
	// struct { union json.RawMessage } which fails on the JSON array body).
	resp, err := client.ListColumns(ctx, dataset, nil, keyEditor(key))
	if err != nil {
		return fmt.Errorf("listing columns: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode, body); err != nil {
		return err
	}

	var columns []api.Column
	if err := json.Unmarshal(body, &columns); err != nil {
		return fmt.Errorf("parsing columns: %w", err)
	}

	items := make([]columnItem, len(columns))
	for i, c := range columns {
		item := columnItem{
			KeyName: c.KeyName,
		}
		if c.Id != nil {
			item.ID = *c.Id
		}
		if c.Type != nil {
			item.Type = string(*c.Type)
		}
		if c.Description != nil {
			item.Description = *c.Description
		}
		if c.Hidden != nil {
			item.Hidden = *c.Hidden
		}
		if c.LastWritten != nil {
			item.LastWritten = *c.LastWritten
		}
		items[i] = item
	}

	return opts.OutputWriter().Write(items, columnListTable)
}
