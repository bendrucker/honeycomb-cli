package column

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
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
			if !command.AnyChanged(cmd, "file", "description", "hidden") {
				return fmt.Errorf("provide --file or at least one of --description, --hidden")
			}

			return runColumnUpdate(cmd, opts, *dataset, args[0], file, description, hidden)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&description, "description", "", "Column description")
	cmd.Flags().BoolVar(&hidden, "hidden", false, "Hide column from autocomplete")

	cmd.MarkFlagsMutuallyExclusive("file", "description")
	cmd.MarkFlagsMutuallyExclusive("file", "hidden")

	return cmd
}

func runColumnUpdate(cmd *cobra.Command, opts *options.RootOptions, dataset, columnID, file, description string, hidden bool) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	var col api.Column

	if cmd.Flags().Changed("file") {
		data, err := command.ReadDefinitionFile(opts.IOStreams, file)
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

		if cmd.Flags().Changed("description") {
			col.Description = &description
		}
		if cmd.Flags().Changed("hidden") {
			col.Hidden = &hidden
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
