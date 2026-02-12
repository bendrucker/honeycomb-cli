package environment

import (
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
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
			return runEnvironmentUpdate(cmd, opts, *team, args[0], desc, color, deleteProtected)
		},
	}

	cmd.Flags().StringVar(&desc, "description", "", "Environment description")
	cmd.Flags().StringVar(&color, "color", "", "Environment color")
	cmd.Flags().BoolVar(&deleteProtected, "delete-protected", false, "Protect environment from deletion")

	return cmd
}

func runEnvironmentUpdate(cmd *cobra.Command, opts *options.RootOptions, team, envID, desc, color string, deleteProtected bool) error {
	if !cmd.Flags().Changed("description") && !cmd.Flags().Changed("color") && !cmd.Flags().Changed("delete-protected") {
		return fmt.Errorf("--description, --color, or --delete-protected is required")
	}

	key, err := opts.RequireKey(config.KeyManagement)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
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

	resp, err := client.UpdateEnvironmentWithApplicationVndAPIPlusJSONBodyWithResponse(cmd.Context(), team, envID, body, keyEditor(key))
	if err != nil {
		return fmt.Errorf("updating environment: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.ApplicationvndApiJSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeEnvironmentDetail(opts, envToDetail(resp.ApplicationvndApiJSON200.Data))
}
