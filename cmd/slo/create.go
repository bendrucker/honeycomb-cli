package slo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/jsonutil"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file       string
		name       string
		sliAlias   string
		target     int
		timePeriod int
		desc       string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an SLO",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSLOCreate(cmd, opts, *dataset, file, name, sliAlias, target, timePeriod, desc)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&name, "name", "", "SLO name")
	cmd.Flags().StringVar(&sliAlias, "sli-alias", "", "SLI calculated field alias")
	cmd.Flags().IntVar(&target, "target", 0, "Target per million")
	cmd.Flags().IntVar(&timePeriod, "time-period", 0, "Time period in days")
	cmd.Flags().StringVar(&desc, "description", "", "SLO description")

	cmd.MarkFlagsMutuallyExclusive("file", "name")
	cmd.MarkFlagsMutuallyExclusive("file", "sli-alias")
	cmd.MarkFlagsMutuallyExclusive("file", "target")
	cmd.MarkFlagsMutuallyExclusive("file", "time-period")
	cmd.MarkFlagsMutuallyExclusive("file", "description")

	return cmd
}

func runSLOCreate(cmd *cobra.Command, opts *options.RootOptions, dataset, file, name, sliAlias string, target, timePeriod int, desc string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	var data []byte

	if cmd.Flags().Changed("file") {
		data, err = readFile(opts, file)
		if err != nil {
			return err
		}
	} else {
		if name == "" || sliAlias == "" || target == 0 || timePeriod == 0 {
			return fmt.Errorf("--file or all of --name, --sli-alias, --target, --time-period are required")
		}

		body := api.SLO{
			Name: name,
			Sli: struct {
				Alias string `json:"alias"`
			}{Alias: sliAlias},
			TargetPerMillion: target,
			TimePeriodDays:   timePeriod,
		}
		if cmd.Flags().Changed("description") {
			body.Description = &desc
		}

		data, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding SLO: %w", err)
		}
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.CreateSloWithBodyWithResponse(cmd.Context(), dataset, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("creating SLO: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	var slo api.SLO
	if err := json.Unmarshal(resp.Body, &slo); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	return writeSloDetail(opts, sloToDetail(slo))
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

	data, err = jsonutil.Sanitize(data)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return data, nil
}
