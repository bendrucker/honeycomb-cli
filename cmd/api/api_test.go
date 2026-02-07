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

	ts := iostreams.Test()
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
	ts := iostreams.Test()
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
	ts := iostreams.Test()
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
	ts := iostreams.Test()
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

	ts := iostreams.Test()
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

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
