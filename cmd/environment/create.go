package environment

import (
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var (
		name            string
		desc            string
		color           string
		deleteProtected bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an environment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := command.ValidateEnum("color", color, environmentColors); err != nil {
				return err
			}
			client, err := opts.ClientFor(team, options.AuthManagement)
			if err != nil {
				return err
			}
			clearProtection := cmd.Flags().Changed("delete-protected") && !deleteProtected
			return runEnvironmentCreate(cmd, opts, client, *team, name, desc, color, clearProtection)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Environment name")
	cmd.Flags().StringVar(&desc, "description", "", "Environment description")
	cmd.Flags().StringVar(&color, "color", "", colorFlagUsage())
	cmd.Flags().BoolVar(&deleteProtected, "delete-protected", true, "Protect environment from deletion")

	return cmd
}

func runEnvironmentCreate(cmd *cobra.Command, opts *options.RootOptions, client *api.ClientWithResponses, team, name, desc, color string, clearProtection bool) error {
	if name == "" {
		promptable := opts.IOStreams.CanPrompt()
		var err error
		name, err = command.Resolve(opts.IOStreams, name, command.Field{
			Prompt:            "Environment name: ",
			Required:          true,
			NonInteractiveErr: fmt.Errorf("--name is required in non-interactive mode"),
			EmptyErr:          fmt.Errorf("environment name is required"),
		})
		if err != nil {
			return err
		}
		if promptable && !cmd.Flags().Changed("description") {
			desc, err = command.Resolve(opts.IOStreams, desc, command.Field{
				Prompt: "Description (optional): ",
			})
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

	resp, err := client.CreateEnvironmentWithApplicationVndAPIPlusJSONBodyWithResponse(cmd.Context(), team, body)
	if err != nil {
		return fmt.Errorf("creating environment: %w", err)
	}

	env, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON201)
	if err != nil {
		return err
	}

	data := env.Data
	if clearProtection {
		data, err = clearEnvironmentProtection(cmd, client, team, data.Id)
		if err != nil {
			return err
		}
	}

	return writeEnvironmentDetail(opts, envToDetail(data))
}

func clearEnvironmentProtection(cmd *cobra.Command, client *api.ClientWithResponses, team, envID string) (api.Environment, error) {
	protected := false
	body := api.UpdateEnvironmentRequest{}
	body.Data.Id = envID
	body.Data.Type = api.UpdateEnvironmentRequestDataTypeEnvironments
	body.Data.Attributes.Settings = &struct {
		DeleteProtected *bool `json:"delete_protected,omitempty"`
	}{
		DeleteProtected: &protected,
	}

	resp, err := client.UpdateEnvironmentWithApplicationVndAPIPlusJSONBodyWithResponse(cmd.Context(), team, envID, body)
	if err != nil {
		return api.Environment{}, fmt.Errorf("clearing delete protection: %w", err)
	}

	env, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON200)
	if err != nil {
		return api.Environment{}, fmt.Errorf("clearing delete protection: %w", err)
	}
	return env.Data, nil
}
