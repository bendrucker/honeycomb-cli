package column

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCalculatedCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file        string
		alias       string
		expression  string
		description string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a calculated column",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if file != "" {
				return runCalculatedCreateFromFile(cmd.Context(), opts, *dataset, file)
			}

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

			body := api.CreateCalculatedFieldJSONRequestBody{
				Alias:      alias,
				Expression: expression,
			}
			if cmd.Flags().Changed("description") || description != "" {
				body.Description = &description
			}

			return runCalculatedCreate(cmd.Context(), opts, *dataset, body)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&alias, "alias", "", "Calculated column alias")
	cmd.Flags().StringVar(&expression, "expression", "", "Calculated column expression")
	cmd.Flags().StringVar(&description, "description", "", "Calculated column description")

	cmd.MarkFlagsMutuallyExclusive("file", "alias")
	cmd.MarkFlagsMutuallyExclusive("file", "expression")
	cmd.MarkFlagsMutuallyExclusive("file", "description")

	return cmd
}

func runCalculatedCreateFromFile(ctx context.Context, opts *options.RootOptions, dataset, file string) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	data, err := command.ReadDefinitionFile(opts.IOStreams, file)
	if err != nil {
		return err
	}

	resp, err := client.CreateCalculatedFieldWithBodyWithResponse(ctx, dataset, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating calculated column: %w", err)
	}

	field, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON201)
	if err != nil {
		return err
	}

	return writeCalculatedDetail(opts, *field)
}

func runCalculatedCreate(ctx context.Context, opts *options.RootOptions, dataset string, body api.CreateCalculatedFieldJSONRequestBody) error {
	client, err := opts.Client(config.KeyConfig)
	if err != nil {
		return err
	}

	resp, err := client.CreateCalculatedFieldWithResponse(ctx, dataset, body)
	if err != nil {
		return fmt.Errorf("creating calculated column: %w", err)
	}

	field, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON201)
	if err != nil {
		return err
	}

	return writeCalculatedDetail(opts, *field)
}
