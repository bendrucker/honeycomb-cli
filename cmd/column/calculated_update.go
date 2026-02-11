package column

import (
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewCalculatedUpdateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		alias       string
		expression  string
		description string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a calculated field",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("alias") && !cmd.Flags().Changed("expression") && !cmd.Flags().Changed("description") {
				return fmt.Errorf("at least one of --alias, --expression, or --description is required")
			}
			return runCalculatedUpdate(cmd, opts, *dataset, args[0], alias, expression, description)
		},
	}

	cmd.Flags().StringVar(&alias, "alias", "", "Human-readable name")
	cmd.Flags().StringVar(&expression, "expression", "", "Formula expression")
	cmd.Flags().StringVar(&description, "description", "", "Description")

	return cmd
}

func runCalculatedUpdate(cmd *cobra.Command, opts *options.RootOptions, dataset, fieldID, alias, expression, description string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	ctx := cmd.Context()
	slug := api.DatasetSlugOrAll(dataset)

	getResp, err := client.GetCalculatedFieldWithResponse(ctx, slug, fieldID, keyEditor(key))
	if err != nil {
		return fmt.Errorf("getting calculated field: %w", err)
	}

	if err := api.CheckResponse(getResp.StatusCode(), getResp.Body); err != nil {
		return err
	}

	if getResp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", getResp.Status())
	}

	field := *getResp.JSON200

	if cmd.Flags().Changed("alias") {
		field.Alias = alias
	}
	if cmd.Flags().Changed("expression") {
		field.Expression = expression
	}
	if cmd.Flags().Changed("description") {
		field.Description = &description
	}

	resp, err := client.UpdateCalculatedFieldWithResponse(ctx, slug, fieldID, field, keyEditor(key))
	if err != nil {
		return fmt.Errorf("updating calculated field: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeCalculatedDetail(opts, *resp.JSON200)
}
