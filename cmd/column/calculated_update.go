package column

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
	"github.com/bendrucker/honeycomb-cli/internal/jsonutil"
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
			hasFile := cmd.Flags().Changed("file")
			hasAlias := cmd.Flags().Changed("alias")
			hasExpr := cmd.Flags().Changed("expression")
			hasDesc := cmd.Flags().Changed("description")

			if !hasFile && !hasAlias && !hasExpr && !hasDesc {
				return fmt.Errorf("provide --file or at least one of --alias, --expression, --description")
			}

			return runCalculatedUpdate(cmd.Context(), opts, *dataset, args[0], calculatedUpdateFlags{
				file:        file,
				hasFile:     hasFile,
				alias:       alias,
				hasAlias:    hasAlias,
				expression:  expression,
				hasExpr:     hasExpr,
				description: description,
				hasDesc:     hasDesc,
			})
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

type calculatedUpdateFlags struct {
	file        string
	hasFile     bool
	alias       string
	hasAlias    bool
	expression  string
	hasExpr     bool
	description string
	hasDesc     bool
}

func runCalculatedUpdate(ctx context.Context, opts *options.RootOptions, dataset, id string, flags calculatedUpdateFlags) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	var body api.CalculatedField

	if flags.hasFile {
		var r io.Reader
		if flags.file == "-" {
			r = opts.IOStreams.In
		} else {
			f, err := os.Open(flags.file)
			if err != nil {
				return fmt.Errorf("opening file: %w", err)
			}
			defer func() { _ = f.Close() }()
			r = f
		}

		data, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}
		data, err = jsonutil.Sanitize(data)
		if err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		if err := json.Unmarshal(data, &body); err != nil {
			return fmt.Errorf("parsing calculated column JSON: %w", err)
		}
	} else {
		getResp, err := client.GetCalculatedFieldWithResponse(ctx, dataset, id, auth)
		if err != nil {
			return fmt.Errorf("getting calculated column: %w", err)
		}
		if err := api.CheckResponse(getResp.StatusCode(), getResp.Body); err != nil {
			return err
		}
		if getResp.JSON200 == nil {
			return fmt.Errorf("unexpected response: %s", getResp.Status())
		}
		body = *getResp.JSON200

		if flags.hasAlias {
			body.Alias = flags.alias
		}
		if flags.hasExpr {
			body.Expression = flags.expression
		}
		if flags.hasDesc {
			body.Description = &flags.description
		}
	}

	data, err := api.MarshalStrippingReadOnly(body, "CalculatedField")
	if err != nil {
		return fmt.Errorf("encoding calculated column: %w", err)
	}

	resp, err := client.UpdateCalculatedFieldWithBodyWithResponse(ctx, dataset, id, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("updating calculated column: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeCalculatedDetail(opts, *resp.JSON200)
}
