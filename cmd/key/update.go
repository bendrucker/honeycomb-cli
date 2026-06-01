package key

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var (
		file     string
		name     string
		disabled bool
		enabled  bool
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an API key",
		Example: `  # Rename an API key
  honeycomb key update abc123 --name "renamed"

  # Disable an API key
  honeycomb key update abc123 --disabled`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.ClientFor(team, options.AuthManagement)
			if err != nil {
				return err
			}
			if file != "" {
				return runKeyUpdateFromFile(cmd.Context(), opts, client, *team, args[0], file)
			}
			return runKeyUpdateFromFlags(cmd, opts, client, *team, args[0], name)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON:API request body (- for stdin)")
	cmd.Flags().StringVar(&name, "name", "", "Key name")
	cmd.Flags().BoolVar(&disabled, "disabled", false, "Disable the key")
	cmd.Flags().BoolVar(&enabled, "enabled", false, "Enable the key")

	cmd.MarkFlagsMutuallyExclusive("file", "name")
	cmd.MarkFlagsMutuallyExclusive("file", "disabled")
	cmd.MarkFlagsMutuallyExclusive("file", "enabled")
	cmd.MarkFlagsMutuallyExclusive("disabled", "enabled")

	return cmd
}

func runKeyUpdateFromFile(ctx context.Context, opts *options.RootOptions, client *api.ClientWithResponses, team, id, file string) error {
	data, err := command.ReadDefinitionFile(opts.IOStreams, file)
	if err != nil {
		return err
	}

	resp, err := client.UpdateApiKeyWithBodyWithResponse(ctx, api.TeamSlug(team), api.ID(id), "application/vnd.api+json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("updating API key: %w", err)
	}

	updated, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON200)
	if err != nil {
		return err
	}

	return writeKeyDetail(opts, objectToDetail(updated.Data))
}

func runKeyUpdateFromFlags(cmd *cobra.Command, opts *options.RootOptions, client *api.ClientWithResponses, team, id, name string) error {
	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("disabled") && !cmd.Flags().Changed("enabled") {
		return fmt.Errorf("--file, --name, --disabled, or --enabled is required")
	}

	ctx := cmd.Context()

	getResp, err := client.GetApiKeyWithResponse(ctx, api.TeamSlug(team), api.ID(id))
	if err != nil {
		return fmt.Errorf("getting API key: %w", err)
	}
	existing, err := api.Decode(getResp.StatusCode(), getResp.Status(), getResp.Body, getResp.ApplicationvndApiJSON200)
	if err != nil {
		return err
	}

	current := objectToDetail(existing.Data)

	if cmd.Flags().Changed("name") {
		current.Name = name
	}
	if cmd.Flags().Changed("disabled") {
		current.Disabled = true
	}
	if cmd.Flags().Changed("enabled") {
		current.Disabled = false
	}

	body, err := buildKeyUpdateRequest(id, current.KeyType, current.Name, current.Disabled)
	if err != nil {
		return err
	}

	resp, err := client.UpdateApiKeyWithApplicationVndAPIPlusJSONBodyWithResponse(ctx, api.TeamSlug(team), api.ID(id), body)
	if err != nil {
		return fmt.Errorf("updating API key: %w", err)
	}

	updated, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.ApplicationvndApiJSON200)
	if err != nil {
		return err
	}

	return writeKeyDetail(opts, objectToDetail(updated.Data))
}

func buildKeyUpdateRequest(id, keyType, name string, disabled bool) (api.ApiKeyUpdateRequest, error) {
	var body api.ApiKeyUpdateRequest

	switch keyType {
	case string(api.IngestKeyAttributesKeyTypeIngest):
		req := api.IngestKeyRequest{
			Id:   id,
			Type: api.IngestKeyRequestTypeApiKeys,
		}
		req.Attributes.Name = &name
		req.Attributes.Disabled = &disabled
		if err := body.Data.FromIngestKeyRequest(req); err != nil {
			return body, fmt.Errorf("building request: %w", err)
		}
	case string(api.ConfigurationKeyAttributesKeyTypeConfiguration):
		req := api.ConfigurationKeyRequest{
			Id:   id,
			Type: api.ConfigurationKeyRequestTypeApiKeys,
		}
		req.Attributes.Name = &name
		req.Attributes.Disabled = &disabled
		if err := body.Data.FromConfigurationKeyRequest(req); err != nil {
			return body, fmt.Errorf("building request: %w", err)
		}
	default:
		return body, fmt.Errorf("unrecognized key type: %q", keyType)
	}

	return body, nil
}
