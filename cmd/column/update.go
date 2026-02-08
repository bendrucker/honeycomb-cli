package column

import (
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
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	ctx := cmd.Context()

	getResp, err := client.GetColumnWithResponse(ctx, dataset, columnID, keyEditor(key))
	if err != nil {
		return fmt.Errorf("getting column: %w", err)
	}

	if err := api.CheckResponse(getResp.StatusCode(), getResp.Body); err != nil {
		return err
	}

	if getResp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", getResp.Status())
	}

	col := *getResp.JSON200

	if cmd.Flags().Changed("description") {
		col.Description = &description
	}
	if cmd.Flags().Changed("hidden") {
		col.Hidden = &hidden
	}

	resp, err := client.UpdateColumnWithResponse(ctx, dataset, columnID, col, keyEditor(key))
	if err != nil {
		return fmt.Errorf("updating column: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeColumnDetail(opts, *resp.JSON200)
}
