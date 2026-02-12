package environment

import (
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var (
		name  string
		desc  string
		color string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an environment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEnvironmentCreate(cmd, opts, *team, name, desc, color)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Environment name")
	cmd.Flags().StringVar(&desc, "description", "", "Environment description")
	cmd.Flags().StringVar(&color, "color", "", "Environment color")

	return cmd
}

func runEnvironmentCreate(cmd *cobra.Command, opts *options.RootOptions, team, name, desc, color string) error {
	key, err := opts.RequireKey(config.KeyManagement)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	if name == "" {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--name is required in non-interactive mode")
		}
		name, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Environment name: ")
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("environment name is required")
		}
		if !cmd.Flags().Changed("description") {
			desc, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Description (optional): ")
			if err != nil {
				return err
			}
		}
	}

	body := api.CreateEnvironmentRequest{}
	body.Data.Type = api.CreateEnvironmentRequestDataTypeEnvironments
	body.Data.Attributes.Name = name
	if desc != "" {
		body.Data.Attributes.Description = &desc
	}
	if color != "" {
		c := api.EnvironmentColor(color)
		body.Data.Attributes.Color = &c
	}

	resp, err := client.CreateEnvironmentWithApplicationVndAPIPlusJSONBodyWithResponse(cmd.Context(), team, body, keyEditor(key))
	if err != nil {
		return fmt.Errorf("creating environment: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.ApplicationvndApiJSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeEnvironmentDetail(opts, envToDetail(resp.ApplicationvndApiJSON201.Data))
}
