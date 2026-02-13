package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

type KeyStatus struct {
	Type        string `json:"type"`
	Status      string `json:"status"`
	Team        string `json:"team,omitempty"`
	Environment string `json:"environment,omitempty"`
	KeyID       string `json:"key_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Error       string `json:"error,omitempty"`
}

var statusTable = output.TableDef{
	Columns: []output.Column{
		{Header: "Type", Value: func(v any) string { return v.(KeyStatus).Type }},
		{Header: "Status", Value: func(v any) string { return v.(KeyStatus).Status }},
		{Header: "Team", Value: func(v any) string { return v.(KeyStatus).Team }},
		{Header: "Environment", Value: func(v any) string { return v.(KeyStatus).Environment }},
		{Header: "Key ID", Value: func(v any) string { return v.(KeyStatus).KeyID }},
	},
}

func NewStatusCmd(opts *options.RootOptions) *cobra.Command {
	var offline bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAuthStatus(cmd.Context(), opts, offline)
		},
	}

	cmd.Flags().BoolVar(&offline, "offline", false, "Skip API verification")

	return cmd
}

var keyTypes = []config.KeyType{config.KeyConfig, config.KeyIngest, config.KeyManagement}

func runAuthStatus(ctx context.Context, opts *options.RootOptions, offline bool) error {
	profile := opts.ActiveProfile()

	type storedKey struct {
		keyType config.KeyType
		value   string
	}

	var keys []storedKey
	for _, kt := range keyTypes {
		val, err := config.GetKey(profile, kt)
		if err == keyring.ErrNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("reading %s key: %w", kt, err)
		}
		keys = append(keys, storedKey{keyType: kt, value: val})
	}

	if len(keys) == 0 {
		return fmt.Errorf("no keys configured for profile %q (run honeycomb auth login)", profile)
	}

	var statuses []KeyStatus
	if offline {
		for _, k := range keys {
			statuses = append(statuses, KeyStatus{
				Type:   string(k.keyType),
				Status: "stored",
			})
		}
	} else {
		client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
		if err != nil {
			return fmt.Errorf("creating API client: %w", err)
		}

		for _, k := range keys {
			ks, err := verifyKey(ctx, client, k.keyType, k.value)
			if err != nil {
				return err
			}
			statuses = append(statuses, ks)
		}
	}

	return opts.OutputWriter().Write(statuses, statusTable)
}

func verifyKey(ctx context.Context, client *api.ClientWithResponses, kt config.KeyType, value string) (KeyStatus, error) {
	ks := KeyStatus{Type: string(kt)}

	switch kt {
	case config.KeyConfig, config.KeyIngest:
		resp, err := client.GetAuthWithResponse(ctx, keyEditor(kt, value))
		if err != nil {
			return ks, fmt.Errorf("verifying %s key: %w", kt, err)
		}
		switch resp.StatusCode() {
		case http.StatusOK:
			ks.Status = "valid"
			ks.KeyID = resp.JSON200.Id
			if resp.JSON200.Team.Name != nil {
				ks.Team = *resp.JSON200.Team.Name
			}
			if resp.JSON200.Environment.Name != nil {
				ks.Environment = *resp.JSON200.Environment.Name
			}
		case http.StatusUnauthorized:
			ks.Status = "invalid"
		default:
			ks.Status = "error"
			ks.Error = resp.Status()
		}

	case config.KeyManagement:
		resp, err := client.GetV2AuthWithResponse(ctx, keyEditor(kt, value))
		if err != nil {
			return ks, fmt.Errorf("verifying %s key: %w", kt, err)
		}
		switch resp.StatusCode() {
		case http.StatusOK:
			ks.Status = "valid"
			if resp.ApplicationvndApiJSON200 != nil {
				if resp.ApplicationvndApiJSON200.Data.Id != nil {
					ks.KeyID = *resp.ApplicationvndApiJSON200.Data.Id
				}
				if resp.ApplicationvndApiJSON200.Data.Attributes != nil && resp.ApplicationvndApiJSON200.Data.Attributes.Name != nil {
					ks.Name = *resp.ApplicationvndApiJSON200.Data.Attributes.Name
				}
			}
		case http.StatusUnauthorized:
			ks.Status = "invalid"
		default:
			ks.Status = "error"
			ks.Error = resp.Status()
		}
	}

	return ks, nil
}

func keyEditor(kt config.KeyType, key string) api.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		config.ApplyAuth(req, kt, key)
		return nil
	}
}
