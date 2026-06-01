package options

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func newTestOptions(t *testing.T, apiURL string, cfg *config.Config) *RootOptions {
	t.Helper()
	ts := iostreams.Test(t)
	return &RootOptions{
		IOStreams: ts.IOStreams,
		Config:    cfg,
		APIUrl:    apiURL,
	}
}

func TestClientForAuthHeader(t *testing.T) {
	for _, tc := range []struct {
		name       string
		kind       AuthKind
		keyType    config.KeyType
		key        string
		team       string
		wantHeader string
		wantValue  string
	}{
		{
			name:       "config sets honeycomb team header",
			kind:       AuthConfig,
			keyType:    config.KeyConfig,
			key:        "cfg-secret",
			wantHeader: "X-Honeycomb-Team",
			wantValue:  "cfg-secret",
		},
		{
			name:       "ingest sets honeycomb team header",
			kind:       AuthIngest,
			keyType:    config.KeyIngest,
			key:        "ingest-secret",
			wantHeader: "X-Honeycomb-Team",
			wantValue:  "ingest-secret",
		},
		{
			name:       "management sets bearer header",
			kind:       AuthManagement,
			keyType:    config.KeyManagement,
			key:        "id:secret",
			team:       "my-team",
			wantHeader: "Authorization",
			wantValue:  "Bearer id:secret",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r.Header.Get(tc.wantHeader)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			t.Cleanup(srv.Close)

			if err := config.SetKey("default", tc.keyType, tc.key); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = config.DeleteKey("default", tc.keyType) })

			opts := newTestOptions(t, srv.URL, &config.Config{})
			team := tc.team
			client, err := opts.ClientFor(&team, tc.kind)
			if err != nil {
				t.Fatalf("ClientFor: %v", err)
			}

			if _, err := client.GetAuthWithResponse(t.Context()); err != nil {
				t.Fatalf("request: %v", err)
			}
			if got != tc.wantValue {
				t.Errorf("header %s = %q, want %q", tc.wantHeader, got, tc.wantValue)
			}
		})
	}
}

func TestClientForMCPOAuthRejected(t *testing.T) {
	opts := newTestOptions(t, "http://example.invalid", &config.Config{})
	if _, err := opts.ClientFor(nil, AuthMCPOAuth); err == nil {
		t.Fatal("expected error building a REST client for AuthMCPOAuth")
	}
}

func TestClientForManagementRequiresKey(t *testing.T) {
	opts := newTestOptions(t, "http://example.invalid", &config.Config{
		Profiles: map[string]*config.Profile{"default": {Team: "known-team"}},
	})
	team := ""
	// Team is inferred, but no management key is stored, so client construction
	// still fails on the missing credential rather than the missing team.
	_, err := opts.ClientFor(&team, AuthManagement)
	if err == nil {
		t.Fatal("expected error for missing management key")
	}
	if team != "known-team" {
		t.Errorf("team = %q, want inferred %q", team, "known-team")
	}
}

func TestRequireTeamInference(t *testing.T) {
	for _, tc := range []struct {
		name     string
		cfg      *config.Config
		flag     string
		wantTeam string
		wantErr  bool
	}{
		{
			name:     "explicit flag wins",
			cfg:      &config.Config{Profiles: map[string]*config.Profile{"default": {Team: "profile-team"}}},
			flag:     "explicit-team",
			wantTeam: "explicit-team",
		},
		{
			name:     "single known team is inferred",
			cfg:      &config.Config{Profiles: map[string]*config.Profile{"default": {Team: "profile-team"}}},
			flag:     "",
			wantTeam: "profile-team",
		},
		{
			name:    "no known team errors",
			cfg:     &config.Config{Profiles: map[string]*config.Profile{"default": {}}},
			flag:    "",
			wantErr: true,
		},
		{
			name:    "no profile errors",
			cfg:     &config.Config{},
			flag:    "",
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := newTestOptions(t, "http://example.invalid", tc.cfg)
			team := tc.flag
			err := opts.RequireTeam(&team)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("RequireTeam: %v", err)
			}
			if team != tc.wantTeam {
				t.Errorf("team = %q, want %q", team, tc.wantTeam)
			}
		})
	}
}

func TestAuthKindsEnumeration(t *testing.T) {
	kinds := AuthKinds()
	want := []config.KeyType{config.KeyConfig, config.KeyIngest, config.KeyManagement}
	if len(kinds) != len(want) {
		t.Fatalf("got %d kinds, want %d", len(kinds), len(want))
	}
	for i, k := range kinds {
		if k == AuthMCPOAuth {
			t.Errorf("AuthKinds must not include AuthMCPOAuth")
		}
		if k.KeyType() != want[i] {
			t.Errorf("kinds[%d].KeyType() = %q, want %q", i, k.KeyType(), want[i])
		}
	}
}
