package column

import (
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCalculatedCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		alias       string
		expression  string
		description string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a calculated field",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if alias == "" {
				if !opts.IOStreams.CanPrompt() {
					return fmt.Errorf("--alias is required in non-interactive mode")
				}
				var err error
				alias, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Alias: ")
				if err != nil {
					return err
				}
			}

			if expression == "" {
				if !opts.IOStreams.CanPrompt() {
					return fmt.Errorf("--expression is required in non-interactive mode")
				}
				var err error
				expression, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Expression: ")
				if err != nil {
					return err
				}
			}

			return runCalculatedCreate(cmd, opts, *dataset, alias, expression, description)
		},
	}

	cmd.Flags().StringVar(&alias, "alias", "", "Human-readable name (required)")
	cmd.Flags().StringVar(&expression, "expression", "", "Formula expression (required)")
	cmd.Flags().StringVar(&description, "description", "", "Description")

	return cmd
}

func runCalculatedCreate(cmd *cobra.Command, opts *options.RootOptions, dataset, alias, expression, description string) error {
	key, err := opts.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	body := api.CreateCalculatedFieldJSONRequestBody{
		Alias:      alias,
		Expression: expression,
	}
	if cmd.Flags().Changed("description") {
		body.Description = &description
	}

	resp, err := client.CreateCalculatedFieldWithResponse(cmd.Context(), api.DatasetSlugOrAll(dataset), body, keyEditor(key))
	if err != nil {
		return fmt.Errorf("creating calculated field: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeCalculatedDetail(opts, *resp.JSON201)
}
