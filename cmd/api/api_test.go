package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func setupTest(t *testing.T, handler http.Handler, kt config.KeyType, key string) (*options.RootOptions, *iostreams.TestStreams) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    srv.URL,
	}

	if err := config.SetKey("default", kt, key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", kt) })

	return opts, ts
}

func TestRun_GET(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.Header.Get("X-Honeycomb-Team") != "test-key" {
			t.Errorf("missing auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}), config.KeyConfig, "test-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/auth"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(ts.OutBuf.String(), `"status":"ok"`) {
		t.Errorf("output = %q, want JSON with status", ts.OutBuf.String())
	}
}

func TestRun_POST_WithFields(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "My Board" {
			t.Errorf("body name = %v, want My Board", body["name"])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "new-id"})
	}), config.KeyConfig, "test-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/boards", "-f", "name=My Board"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(ts.OutBuf.String(), `"id":"new-id"`) {
		t.Errorf("output = %q, want JSON with id", ts.OutBuf.String())
	}
}

func TestRun_ManagementKey(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer mgmt-key" {
			t.Errorf("auth = %q, want Bearer mgmt-key", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}), config.KeyManagement, "mgmt-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/2/teams"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRun_JQ(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"name": "test-team"})
	}), config.KeyConfig, "test-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/auth", "--jq", ".name"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(ts.OutBuf.String())
	if got != "test-team" {
		t.Errorf("jq output = %q, want %q", got, "test-team")
	}
}

func TestRun_Include(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}), config.KeyConfig, "test-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/auth", "--include"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	errOut := ts.ErrBuf.String()
	if !strings.Contains(errOut, "200 OK") {
		t.Errorf("stderr = %q, want status line", errOut)
	}
	if !strings.Contains(errOut, "Content-Type") {
		t.Errorf("stderr = %q, want headers", errOut)
	}
}

func TestRun_NonSuccess(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}), config.KeyConfig, "test-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/boards/missing"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("error = %q, want HTTP 404", err.Error())
	}
	if !strings.Contains(err.Error(), "GET /1/boards/missing") {
		t.Errorf("error = %q, want method and path", err.Error())
	}
	if !strings.Contains(ts.OutBuf.String(), "not found") {
		t.Errorf("output = %q, want body written despite error", ts.OutBuf.String())
	}
}

func TestRun_NoKey(t *testing.T) {
	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/auth"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "no config key configured") {
		t.Errorf("error = %q, want missing key message", err.Error())
	}
}

func TestRun_InvalidKeyType(t *testing.T) {
	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/auth", "--key-type", "bogus"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid key type")
	}
	if !strings.Contains(err.Error(), "invalid key type") {
		t.Errorf("error = %q, want invalid key type message", err.Error())
	}
}

func TestRun_KeyTypeOverride(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Honeycomb-Team") != "ingest-key" {
			t.Errorf("auth = %q, want ingest-key", r.Header.Get("X-Honeycomb-Team"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}), config.KeyIngest, "ingest-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/auth", "--key-type", "ingest"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRun_Paginate(t *testing.T) {
	callCount := 0
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			nextURL := "http://" + r.Host + "/1/columns/ds?cursor=page2"
			w.Header().Set("Link", `<`+nextURL+`>; rel="next"`)
			_, _ = w.Write([]byte(`[{"name":"col1"}]`))
		} else {
			_, _ = w.Write([]byte(`[{"name":"col2"}]`))
		}
	}), config.KeyConfig, "test-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/columns/ds", "--paginate"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if callCount != 2 {
		t.Errorf("request count = %d, want 2", callCount)
	}
	out := ts.OutBuf.String()
	if !strings.Contains(out, "col1") || !strings.Contains(out, "col2") {
		t.Errorf("output = %q, want both pages", out)
	}
}

func TestRun_PaginateNonGET(t *testing.T) {
	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		APIUrl:    "http://localhost",
	}

	if err := config.SetKey("default", config.KeyConfig, "test-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/boards", "--paginate", "-X", "POST"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for paginate with POST")
	}
	if !strings.Contains(err.Error(), "only supported with GET") {
		t.Errorf("error = %q, want paginate validation error", err.Error())
	}
}

func TestRun_InputStdin(t *testing.T) {
	var gotBody string
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"received":true}`))
	}), config.KeyIngest, "test-key")

	ts := iostreams.Test(t)
	ts.InBuf.WriteString(`{"data":[]}`)
	opts.IOStreams = ts.IOStreams

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/events/my-dataset", "--input", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if gotBody != `{"data":[]}` {
		t.Errorf("body = %q, want %q", gotBody, `{"data":[]}`)
	}
}

func TestRun_InputFile(t *testing.T) {
	var gotBody string
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}), config.KeyConfig, "test-key")

	tmpFile := t.TempDir() + "/body.json"
	if err := writeTestFile(tmpFile, `{"name":"from-file"}`); err != nil {
		t.Fatal(err)
	}

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/boards", "--input", tmpFile})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if gotBody != `{"name":"from-file"}` {
		t.Errorf("body = %q, want %q", gotBody, `{"name":"from-file"}`)
	}
}

func TestRun_StatusCodeBoundary(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"299 succeeds", 299, false},
		{"399 succeeds", 399, false},
		{"400 fails", 400, true},
		{"500 fails", 500, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(`{}`))
			}), config.KeyConfig, "test-key")

			cmd := NewCmd(opts)
			cmd.SetArgs([]string{"/1/auth"})
			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("status %d: error = %v, wantErr %v", tt.status, err, tt.wantErr)
			}
		})
	}
}

func TestRun_JQ_NonSuccess(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request","detail":"missing field"}`))
	}), config.KeyConfig, "test-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/1/boards", "--jq", ".error"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 400")
	}
	if !strings.Contains(err.Error(), "HTTP 400") {
		t.Errorf("error = %q, want HTTP 400", err.Error())
	}

	got := strings.TrimSpace(ts.OutBuf.String())
	if got != "bad request" {
		t.Errorf("jq output = %q, want %q", got, "bad request")
	}
}

func TestRun_V2_FieldWrapping(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantMethod string
		wantType   string
	}{
		{
			name:       "POST infers type from collection",
			args:       []string{"/2/teams/my-team/environments", "-f", "name=prod"},
			wantMethod: http.MethodPost,
			wantType:   "environments",
		},
		{
			name:       "PATCH infers type from parent segment",
			args:       []string{"/2/teams/my-team/environments/env-id", "-X", "PATCH", "-f", "name=updated"},
			wantMethod: http.MethodPatch,
			wantType:   "environments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBody map[string]any
			opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.wantMethod {
					t.Errorf("method = %q, want %q", r.Method, tt.wantMethod)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/vnd.api+json" {
					t.Errorf("Content-Type = %q, want application/vnd.api+json", ct)
				}
				_ = json.NewDecoder(r.Body).Decode(&gotBody)
				w.Header().Set("Content-Type", "application/vnd.api+json")
				_, _ = w.Write([]byte(`{"data":{"id":"1","type":"environments","attributes":{"name":"prod"}}}`))
			}), config.KeyManagement, "mgmt-key")

			cmd := NewCmd(opts)
			cmd.SetArgs(tt.args)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}

			data, ok := gotBody["data"].(map[string]any)
			if !ok {
				t.Fatal("request body missing data key")
			}
			if data["type"] != tt.wantType {
				t.Errorf("data.type = %v, want %s", data["type"], tt.wantType)
			}
			if _, ok := data["attributes"]; !ok {
				t.Error("request body missing data.attributes")
			}
		})
	}
}

func TestRun_V2_ResponseUnwrap(t *testing.T) {
	t.Run("single resource", func(t *testing.T) {
		opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_, _ = w.Write([]byte(`{"data":{"id":"abc","type":"environments","attributes":{"name":"prod"}}}`))
		}), config.KeyManagement, "mgmt-key")

		cmd := NewCmd(opts)
		cmd.SetArgs([]string{"/2/teams/my-team/environments/abc"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}

		var flat map[string]any
		if err := json.Unmarshal(ts.OutBuf.Bytes(), &flat); err != nil {
			t.Fatal(err)
		}
		if flat["id"] != "abc" {
			t.Errorf("id = %v, want abc", flat["id"])
		}
		if flat["name"] != "prod" {
			t.Errorf("name = %v, want prod", flat["name"])
		}
	})

	t.Run("list", func(t *testing.T) {
		opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.api+json")
			_, _ = w.Write([]byte(`{"data":[{"id":"a","type":"environments","attributes":{"name":"prod"}},{"id":"b","type":"environments","attributes":{"name":"staging"}}]}`))
		}), config.KeyManagement, "mgmt-key")

		cmd := NewCmd(opts)
		cmd.SetArgs([]string{"/2/teams/my-team/environments"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}

		var flat []map[string]any
		if err := json.Unmarshal(ts.OutBuf.Bytes(), &flat); err != nil {
			t.Fatal(err)
		}
		if len(flat) != 2 {
			t.Fatalf("len = %d, want 2", len(flat))
		}
		if flat[0]["name"] != "prod" {
			t.Errorf("[0].name = %v, want prod", flat[0]["name"])
		}
	})
}

func TestRun_V2_Raw(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":{"id":"abc","type":"environments","attributes":{"name":"prod"}}}`))
	}), config.KeyManagement, "mgmt-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/2/teams/my-team/environments/abc", "--raw"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var raw map[string]any
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["data"]; !ok {
		t.Error("--raw should preserve data envelope")
	}
}

func TestRun_V2_InputNotWrapped(t *testing.T) {
	var gotBody string
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":{"id":"1","type":"environments","attributes":{}}}`))
	}), config.KeyManagement, "mgmt-key")

	ts := iostreams.Test(t)
	ts.InBuf.WriteString(`{"data":{"type":"environments","attributes":{"name":"prod"}}}`)
	opts.IOStreams = ts.IOStreams

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/2/teams/my-team/environments", "--input", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(gotBody, `"data"`) {
		t.Errorf("--input body should be sent as-is, got %q", gotBody)
	}
}

func TestRun_V2_HeaderOverride(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json (overridden)", ct)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}), config.KeyManagement, "mgmt-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/2/teams/my-team/environments", "-f", "name=prod", "-H", "Content-Type: application/json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRun_V2_GET_FieldsAsQuery(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Query().Get("filter") != "prod" {
			t.Errorf("query filter = %q, want prod", r.URL.Query().Get("filter"))
		}
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			if len(b) > 0 {
				t.Errorf("GET should not have body, got %q", b)
			}
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}), config.KeyManagement, "mgmt-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/2/teams/my-team/environments", "-X", "GET", "-f", "filter=prod"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRun_V2_ErrorResponse(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"errors":[{"title":"Validation Error","detail":"name is required"}]}`))
	}), config.KeyManagement, "mgmt-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/2/teams/my-team/environments"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 422")
	}
	if !strings.Contains(err.Error(), "HTTP 422") {
		t.Errorf("error = %q, want HTTP 422", err.Error())
	}
	// Error body should be written to output even on failure
	if !strings.Contains(ts.OutBuf.String(), "name is required") {
		t.Errorf("output = %q, want error body written", ts.OutBuf.String())
	}
}

func TestRun_V2_JQ_Unwrap(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":{"id":"abc","type":"environments","attributes":{"name":"prod"}}}`))
	}), config.KeyManagement, "mgmt-key")

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"/2/teams/my-team/environments/abc", "--jq", ".name"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(ts.OutBuf.String())
	if got != "prod" {
		t.Errorf("jq output = %q, want %q (should operate on unwrapped data)", got, "prod")
	}
}

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
