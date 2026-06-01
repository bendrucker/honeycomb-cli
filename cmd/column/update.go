package column

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file        string
		description string
		hidden      bool
	)

	cmd := &cobra.Command{
		Use:   "update <column-id>",
		Short: "Update a column",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hasFile := cmd.Flags().Changed("file")
			hasDesc := cmd.Flags().Changed("description")
			hasHidden := cmd.Flags().Changed("hidden")

			if !hasFile && !hasDesc && !hasHidden {
				return fmt.Errorf("provide --file or at least one of --description, --hidden")
			}

			return runColumnUpdate(cmd.Context(), opts, *dataset, args[0], columnUpdateFlags{
				file:        file,
				hasFile:     hasFile,
				description: description,
				hasDesc:     hasDesc,
				hidden:      hidden,
				hasHidden:   hasHidden,
			})
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&description, "description", "", "Column description")
	cmd.Flags().BoolVar(&hidden, "hidden", false, "Hide column from autocomplete")

	cmd.MarkFlagsMutuallyExclusive("file", "description")
	cmd.MarkFlagsMutuallyExclusive("file", "hidden")

	return cmd
}

type columnUpdateFlags struct {
	file        string
	hasFile     bool
	description string
	hasDesc     bool
	hidden      bool
	hasHidden   bool
}

func runColumnUpdate(ctx context.Context, opts *options.RootOptions, dataset, columnID string, flags columnUpdateFlags) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	var col api.Column

	if flags.hasFile {
		data, err := command.ReadDefinitionFile(opts.IOStreams, flags.file)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &col); err != nil {
			return fmt.Errorf("parsing column JSON: %w", err)
		}
	} else {
		getResp, err := client.GetColumnWithResponse(ctx, dataset, columnID)
		if err != nil {
			return fmt.Errorf("getting column: %w", err)
		}
		current, err := api.Decode(getResp.StatusCode(), getResp.Status(), getResp.Body, getResp.JSON200)
		if err != nil {
			return err
		}
		col = *current

		if flags.hasDesc {
			col.Description = &flags.description
		}
		if flags.hasHidden {
			col.Hidden = &flags.hidden
		}
	}

	data, err := api.MarshalStrippingReadOnly(col, "Column")
	if err != nil {
		return fmt.Errorf("encoding column: %w", err)
	}

	resp, err := client.UpdateColumnWithBodyWithResponse(ctx, dataset, columnID, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("updating column: %w", err)
	}

	updated, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeColumnDetail(opts, *updated)
}
