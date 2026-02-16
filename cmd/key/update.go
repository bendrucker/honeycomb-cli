package key

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
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
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.RequireTeam(team); err != nil {
				return err
			}
			if file != "" {
				return runKeyUpdateFromFile(cmd.Context(), opts, *team, args[0], file)
			}
			return runKeyUpdateFromFlags(cmd, opts, *team, args[0], name)
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

func runKeyUpdateFromFile(ctx context.Context, opts *options.RootOptions, team, id, file string) error {
	auth, err := opts.KeyEditor(config.KeyManagement)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	data, err := readBodyFile(opts, file)
	if err != nil {
		return err
	}

	resp, err := client.UpdateApiKeyWithBodyWithResponse(ctx, api.TeamSlug(team), api.ID(id), "application/vnd.api+json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("updating API key: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.ApplicationvndApiJSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeKeyDetail(opts, objectToDetail(resp.ApplicationvndApiJSON200.Data))
}

func runKeyUpdateFromFlags(cmd *cobra.Command, opts *options.RootOptions, team, id, name string) error {
	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("disabled") && !cmd.Flags().Changed("enabled") {
		return fmt.Errorf("--file, --name, --disabled, or --enabled is required")
	}

	auth, err := opts.KeyEditor(config.KeyManagement)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	ctx := cmd.Context()

	getResp, err := client.GetApiKeyWithResponse(ctx, api.TeamSlug(team), api.ID(id), auth)
	if err != nil {
		return fmt.Errorf("getting API key: %w", err)
	}
	if err := api.CheckResponse(getResp.StatusCode(), getResp.Body); err != nil {
		return err
	}
	if getResp.ApplicationvndApiJSON200 == nil {
		return fmt.Errorf("unexpected response: %s", getResp.Status())
	}

	current := objectToDetail(getResp.ApplicationvndApiJSON200.Data)

	if cmd.Flags().Changed("name") {
		current.Name = name
	}
	if cmd.Flags().Changed("disabled") {
		current.Disabled = true
	}
	if cmd.Flags().Changed("enabled") {
		current.Disabled = false
	}

	body, err := buildKeyUpdateRequest(id, current.Name, current.Disabled)
	if err != nil {
		return err
	}

	resp, err := client.UpdateApiKeyWithApplicationVndAPIPlusJSONBodyWithResponse(ctx, api.TeamSlug(team), api.ID(id), body, auth)
	if err != nil {
		return fmt.Errorf("updating API key: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.ApplicationvndApiJSON200 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	return writeKeyDetail(opts, objectToDetail(resp.ApplicationvndApiJSON200.Data))
}

func buildKeyUpdateRequest(id, name string, disabled bool) (api.ApiKeyUpdateRequest, error) {
	var body api.ApiKeyUpdateRequest

	if strings.HasPrefix(id, "hcxik_") {
		req := api.IngestKeyRequest{
			Id:   id,
			Type: api.IngestKeyRequestTypeApiKeys,
		}
		req.Attributes.Name = &name
		req.Attributes.Disabled = &disabled
		if err := body.Data.FromIngestKeyRequest(req); err != nil {
			return body, fmt.Errorf("building request: %w", err)
		}
	} else if strings.HasPrefix(id, "hcxlk_") {
		req := api.ConfigurationKeyRequest{
			Id:   id,
			Type: api.ConfigurationKeyRequestTypeApiKeys,
		}
		req.Attributes.Name = &name
		req.Attributes.Disabled = &disabled
		if err := body.Data.FromConfigurationKeyRequest(req); err != nil {
			return body, fmt.Errorf("building request: %w", err)
		}
	} else {
		return body, fmt.Errorf("unrecognized key ID prefix: %s (expected hcxik_ or hcxlk_)", id)
	}

	return body, nil
}
