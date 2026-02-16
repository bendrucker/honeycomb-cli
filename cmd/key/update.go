package key

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions, team *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an API key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.RequireTeam(team); err != nil {
				return err
			}
			return runKeyUpdate(cmd.Context(), opts, *team, args[0], file)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON:API request body (- for stdin)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func runKeyUpdate(ctx context.Context, opts *options.RootOptions, team, id, file string) error {
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
