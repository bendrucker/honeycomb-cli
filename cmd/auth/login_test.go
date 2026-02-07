package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/bendrucker/honeycomb-cli/internal/output"
)

func TestAuthLogin_Success(t *testing.T) {
	tests := []struct {
		name      string
		keyType   string
		keyID     string
		keySecret string
		verify    bool
		handler   http.Handler
		want      loginResult
	}{
		{
			name:      "config key verified",
			keyType:   "config",
			keyID:     "myid",
			keySecret: "mysecret",
			verify:    true,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/1/auth" {
					http.NotFound(w, r)
					return
				}
				if r.Header.Get("X-Honeycomb-Team") != "myid:mysecret" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":             "myid",
					"type":           "configuration",
					"team":           map[string]string{"name": "My Team", "slug": "my-team"},
					"environment":    map[string]string{"name": "production", "slug": "production"},
					"api_key_access": map[string]bool{"events": true},
				})
			}),
			want: loginResult{
				Type:        "config",
				Team:        "My Team",
				Environment: "production",
				KeyID:       "myid",
				Verified:    true,
			},
		},
		{
			name:      "management key verified",
			keyType:   "management",
			keyID:     "mgmtid",
			keySecret: "mgmtsecret",
			verify:    true,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/2/auth" {
					http.NotFound(w, r)
					return
				}
				if r.Header.Get("Authorization") != "Bearer mgmtid:mgmtsecret" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.Header().Set("Content-Type", "application/vnd.api+json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"id":   "mgmt-id",
						"type": "api-keys",
						"attributes": map[string]any{
							"name":     "My Management Key",
							"key_type": "management",
						},
					},
				})
			}),
			want: loginResult{
				Type:     "management",
				KeyID:    "mgmt-id",
				Name:     "My Management Key",
				Verified: true,
			},
		},
		{
			name:      "no verify",
			keyType:   "ingest",
			keyID:     "myid",
			keySecret: "mysecret",
			verify:    false,
			want: loginResult{
				Type: "ingest",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var apiURL string
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				t.Cleanup(srv.Close)
				apiURL = srv.URL
			}

			ts := iostreams.Test()
			opts := &options.RootOptions{
				IOStreams: ts.IOStreams,
				Config:    &config.Config{},
				APIUrl:    apiURL,
				Format:    output.FormatJSON,
			}

			kt := config.KeyType(tt.keyType)
			t.Cleanup(func() { _ = config.DeleteKey("default", kt) })

			err := runAuthLogin(t.Context(), opts, tt.keyType, tt.keyID, tt.keySecret, tt.verify)
			if err != nil {
				t.Fatal(err)
			}

			var result loginResult
			if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
				t.Fatalf("unmarshal output: %v", err)
			}
			if result != tt.want {
				t.Errorf("got %+v, want %+v", result, tt.want)
			}

			stored, err := config.GetKey("default", kt)
			if err != nil {
				t.Fatal(err)
			}
			wantStored := tt.keyID + ":" + tt.keySecret
			if stored != wantStored {
				t.Errorf("stored key = %q, want %q", stored, wantStored)
			}
		})
	}
}

func TestAuthLogin_InvalidKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(t, w, map[string]string{"error": "unauthorized"})
	}))
	t.Cleanup(srv.Close)

	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
		Format:    output.FormatJSON,
	}

	err := runAuthLogin(t.Context(), opts, "config", "badid", "badsecret", true)
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
	if err.Error() != "invalid key" {
		t.Errorf("got error %q, want %q", err.Error(), "invalid key")
	}

	_, err = config.GetKey("default", config.KeyConfig)
	if err == nil {
		t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })
		t.Error("key should not be stored after failed verification")
	}
}

func TestAuthLogin_MissingKeyType(t *testing.T) {
	ts := iostreams.Test()
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		Format:    output.FormatJSON,
	}

	err := runAuthLogin(t.Context(), opts, "", "myid", "mysecret", false)
	if err == nil {
		t.Fatal("expected error for missing key type")
	}
	want := "--key-type is required in non-interactive mode"
	if err.Error() != want {
		t.Errorf("got error %q, want %q", err.Error(), want)
	}
}

func TestAuthLogin_StdinSecret(t *testing.T) {
	ts := iostreams.Test()
	ts.InBuf.WriteString("stdin-secret\n")

	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		Format:    output.FormatJSON,
	}

	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	err := runAuthLogin(t.Context(), opts, "config", "myid", "", false)
	if err != nil {
		t.Fatal(err)
	}

	stored, err := config.GetKey("default", config.KeyConfig)
	if err != nil {
		t.Fatal(err)
	}
	if stored != "myid:stdin-secret" {
		t.Errorf("stored key = %q, want %q", stored, "myid:stdin-secret")
	}
}
