package key

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/deref"
	"github.com/spf13/cobra"
)

func NewCreateCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an API key",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runKeyCreate(cmd.Context(), opts, *team, file)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON:API request body (- for stdin)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func runKeyCreate(ctx context.Context, opts *options.RootOptions, team, file string) error {
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

	resp, err := client.CreateApiKeyWithBodyWithResponse(ctx, api.TeamSlug(team), "application/vnd.api+json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("creating API key: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	if resp.ApplicationvndApiJSON201 == nil {
		return fmt.Errorf("unexpected response: %s", resp.Status())
	}

	detail := createResponseToDetail(resp.ApplicationvndApiJSON201)

	if detail.Secret != "" {
		_, _ = fmt.Fprintf(opts.IOStreams.Err, "Save this secret now — it cannot be retrieved again\n")
	}

	return writeKeyDetail(opts, detail)
}

func createResponseToDetail(resp *api.ApiKeyCreateResponse) keyDetail {
	detail := keyDetail{
		ID:     deref.String(resp.Data.Id),
		Secret: deref.String(resp.Data.Attributes.Secret),
	}

	if ingest, err := resp.Data.Attributes.AsIngestKeyAttributes(); err == nil {
		detail.Name = ingest.Name
		detail.KeyType = string(ingest.KeyType)
		detail.Disabled = deref.Bool(ingest.Disabled)
		return detail
	}
	if cfg, err := resp.Data.Attributes.AsConfigurationKeyAttributes(); err == nil {
		detail.Name = cfg.Name
		detail.KeyType = string(cfg.KeyType)
		detail.Disabled = deref.Bool(cfg.Disabled)
		return detail
	}

	var raw struct {
		Name     string `json:"name"`
		KeyType  string `json:"key_type"`
		Disabled bool   `json:"disabled"`
	}
	rawBytes, _ := json.Marshal(resp.Data.Attributes)
	_ = json.Unmarshal(rawBytes, &raw)
	detail.Name = raw.Name
	detail.KeyType = raw.KeyType
	detail.Disabled = raw.Disabled

	return detail
}

func openFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	return f, nil
}
