package column

import (
	"bytes"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		description string
		hidden      bool
	)

	cmd := &cobra.Command{
		Use:   "update <column-id>",
		Short: "Update a column",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runColumnUpdate(cmd, opts, *dataset, args[0], description, hidden)
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "Column description")
	cmd.Flags().BoolVar(&hidden, "hidden", false, "Hide column from autocomplete")

	return cmd
}

func runColumnUpdate(cmd *cobra.Command, opts *options.RootOptions, dataset, columnID, description string, hidden bool) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	getResp, err := client.GetColumnWithResponse(ctx, dataset, columnID)
	if err != nil {
		return fmt.Errorf("getting column: %w", err)
	}

	current, err := api.Decode(getResp.StatusCode(), getResp.Status(), getResp.Body, getResp.JSON200)
	if err != nil {
		return err
	}

	col := *current

	if cmd.Flags().Changed("description") {
		col.Description = &description
	}
	if cmd.Flags().Changed("hidden") {
		col.Hidden = &hidden
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
