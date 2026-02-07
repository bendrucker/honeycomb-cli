package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

type KeyStatus struct {
	Type        string `json:"type" yaml:"type"`
	Status      string `json:"status" yaml:"status"`
	Team        string `json:"team,omitempty" yaml:"team,omitempty"`
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`
	KeyID       string `json:"key_id,omitempty" yaml:"key_id,omitempty"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Error       string `json:"error,omitempty" yaml:"error,omitempty"`
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

	return writeStatuses(opts, statuses)
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

func writeStatuses(opts *options.RootOptions, statuses []KeyStatus) error {
	out := opts.IOStreams.Out

	switch opts.ResolveFormat() {
	case "json":
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(statuses)
	case "yaml":
		return yaml.NewEncoder(out).Encode(statuses)
	case "table":
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "TYPE\tSTATUS\tTEAM\tENVIRONMENT\tKEY ID")
		for _, s := range statuses {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.Type, s.Status, s.Team, s.Environment, s.KeyID)
		}
		return w.Flush()
	default:
		return fmt.Errorf("unsupported format: %s", opts.ResolveFormat())
	}
}

func keyEditor(kt config.KeyType, key string) api.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		config.ApplyAuth(req, kt, key)
		return nil
	}
}
