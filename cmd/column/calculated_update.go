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

func NewCalculatedUpdateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		file        string
		alias       string
		expression  string
		description string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a calculated column",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !command.AnyChanged(cmd, "file", "alias", "expression", "description") {
				return fmt.Errorf("provide --file or at least one of --alias, --expression, --description")
			}

			return runCalculatedUpdate(cmd, opts, *dataset, args[0], file, alias, expression, description)
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

func runCalculatedUpdate(cmd *cobra.Command, opts *options.RootOptions, dataset, id, file, alias, expression, description string) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	var body api.CalculatedField

	if cmd.Flags().Changed("file") {
		data, err := command.ReadDefinitionFile(opts.IOStreams, file)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &body); err != nil {
			return fmt.Errorf("parsing calculated column JSON: %w", err)
		}
	} else {
		getResp, err := client.GetCalculatedFieldWithResponse(ctx, dataset, id)
		if err != nil {
			return fmt.Errorf("getting calculated column: %w", err)
		}
		current, err := api.Decode(getResp.StatusCode(), getResp.Status(), getResp.Body, getResp.JSON200)
		if err != nil {
			return err
		}
		body = *current

		if cmd.Flags().Changed("alias") {
			body.Alias = alias
		}
		if cmd.Flags().Changed("expression") {
			body.Expression = expression
		}
		if cmd.Flags().Changed("description") {
			body.Description = &description
		}
	}

	data, err := api.MarshalStrippingReadOnly(body, "CalculatedField")
	if err != nil {
		return fmt.Errorf("encoding calculated column: %w", err)
	}

	resp, err := client.UpdateCalculatedFieldWithBodyWithResponse(ctx, dataset, id, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("updating calculated column: %w", err)
	}

	if _, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200); err != nil {
		return err
	}

	updatedResp, err := client.GetCalculatedFieldWithResponse(ctx, dataset, id)
	if err != nil {
		return fmt.Errorf("getting calculated column: %w", err)
	}

	updated, err := api.Decode(updatedResp.StatusCode(), updatedResp.Status(), updatedResp.Body, updatedResp.JSON200)
	if err != nil {
		return err
	}

	return writeCalculatedDetail(opts, *updated)
}
