package environment

import (
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var (
		desc            string
		color           string
		deleteProtected bool
	)

	cmd := &cobra.Command{
		Use:   "update <environment-id>",
		Short: "Update an environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := command.ValidateEnum("color", color, environmentColors); err != nil {
				return err
			}
			client, err := opts.ClientFor(team, options.AuthManagement)
			if err != nil {
				return err
			}
			return runEnvironmentUpdate(cmd, opts, client, *team, args[0], desc, color, deleteProtected)
		},
	}

	cmd.Flags().StringVar(&desc, "description", "", "Environment description")
	cmd.Flags().StringVar(&color, "color", "", colorFlagUsage())
	cmd.Flags().BoolVar(&deleteProtected, "delete-protected", false, "Protect environment from deletion")

	return cmd
}

func runEnvironmentUpdate(cmd *cobra.Command, opts *options.RootOptions, client *api.ClientWithResponses, team, envID, desc, color string, deleteProtected bool) error {
	if !cmd.Flags().Changed("description") && !cmd.Flags().Changed("color") && !cmd.Flags().Changed("delete-protected") {
		return fmt.Errorf("--description, --color, or --delete-protected is required")
	}

	body := api.UpdateEnvironmentRequest{}
	body.Data.Id = envID
	body.Data.Type = api.UpdateEnvironmentRequestDataTypeEnvironments
	if cmd.Flags().Changed("description") {
		body.Data.Attributes.Description = &desc
	}
	if cmd.Flags().Changed("color") {
		c := api.EnvironmentColor(color)
		body.Data.Attributes.Color = &c
	}
	if cmd.Flags().Changed("delete-protected") {
		body.Data.Attributes.Settings = &struct {
			DeleteProtected *bool `json:"delete_protected,omitempty"`
		}{
			DeleteProtected: &deleteProtected,
		}
	}

	resp, err := client.UpdateEnvironmentWithApplicationVndAPIPlusJSONBodyWithResponse(cmd.Context(), team, envID, body)
	if err != nil {
		return fmt.Errorf("updating environment: %w", err)
	}

	env, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON200)
	if err != nil {
		return err
	}

	return writeEnvironmentDetail(opts, envToDetail(env.Data))
}
