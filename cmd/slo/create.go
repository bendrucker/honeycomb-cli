package slo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an SLO",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSLOCreate(cmd.Context(), opts, *dataset, file)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")

	return cmd
}

func runSLOCreate(ctx context.Context, opts *options.RootOptions, dataset, file string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	if file == "" {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--file is required in non-interactive mode")
		}
		file, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Path to SLO JSON file: ")
		if err != nil {
			return err
		}
		if file == "" {
			return fmt.Errorf("file path is required")
		}
	}

	data, err := readFile(opts, file)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.CreateSloWithBodyWithResponse(ctx, dataset, "application/json", bytes.NewReader(data), keyEditor(key))
	if err != nil {
		return fmt.Errorf("creating SLO: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	detail := sloCreateToDetail(*resp.JSON201)
	return opts.OutputWriter().Write(detail, output.TableDef{})
}

func sloCreateToDetail(s api.SLOCreate) sloDetail {
	d := sloDetail{
		Name:             s.Name,
		TargetPerMillion: s.TargetPerMillion,
		TimePeriodDays:   s.TimePeriodDays,
		SLIAlias:         s.Sli.Alias,
	}
	if s.Id != nil {
		d.ID = *s.Id
	}
	if s.Description != nil {
		d.Description = *s.Description
	}
	if s.CreatedAt != nil {
		d.CreatedAt = s.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if s.UpdatedAt != nil {
		d.UpdatedAt = s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if s.ResetAt.IsSpecified() && !s.ResetAt.IsNull() {
		d.ResetAt = s.ResetAt.MustGet().Format("2006-01-02T15:04:05Z07:00")
	}
	if s.DatasetSlugs != nil {
		for _, v := range *s.DatasetSlugs {
			if slug, ok := v.(string); ok {
				d.DatasetSlugs = append(d.DatasetSlugs, slug)
			}
		}
	}
	return d
}

func readFile(opts *options.RootOptions, file string) ([]byte, error) {
	var r io.Reader
	if file == "-" {
		r = opts.IOStreams.In
	} else {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		defer f.Close() //nolint:errcheck // best-effort close on read-only file
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var js json.RawMessage
	if err := json.Unmarshal(data, &js); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return data, nil
}
