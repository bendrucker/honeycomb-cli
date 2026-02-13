package column

import (
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions, dataset *string) *cobra.Command {
	var (
		keyName     string
		colType     string
		description string
		hidden      bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a column",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if keyName == "" {
				if !opts.IOStreams.CanPrompt() {
					return fmt.Errorf("--key-name is required in non-interactive mode")
				}
				var err error
				keyName, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Key name: ")
				if err != nil {
					return err
				}
			}

			if colType == "" && opts.IOStreams.CanPrompt() {
				var err error
				colType, err = prompt.Choice(opts.IOStreams.Err, opts.IOStreams.In, "Type (string/float/integer/boolean): ", []string{"string", "float", "integer", "boolean"})
				if err != nil {
					return err
				}
			}

			return runColumnCreate(cmd, opts, *dataset, keyName, colType, description, hidden)
		},
	}

	cmd.Flags().StringVar(&keyName, "key-name", "", "Column key name (required)")
	cmd.Flags().StringVar(&colType, "type", "", "Column type: string, float, integer, boolean")
	cmd.Flags().StringVar(&description, "description", "", "Column description")
	cmd.Flags().BoolVar(&hidden, "hidden", false, "Hide column from autocomplete")

	return cmd
}

func runColumnCreate(cmd *cobra.Command, opts *options.RootOptions, dataset, keyName, colType, description string, hidden bool) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	body := api.CreateColumnJSONRequestBody{
		KeyName: keyName,
	}
	if colType != "" {
		t := api.CreateColumnType(colType)
		body.Type = &t
	}
	if cmd.Flags().Changed("description") {
		body.Description = &description
	}
	if cmd.Flags().Changed("hidden") {
		body.Hidden = &hidden
	}

	resp, err := client.CreateColumnWithResponse(cmd.Context(), dataset, body, auth)
	if err != nil {
		return fmt.Errorf("creating column: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.JSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeColumnDetail(opts, *resp.JSON201)
}
