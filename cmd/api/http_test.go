package api

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestBuildRequest_GET_WithFields(t *testing.T) {
	fields := map[string]any{"key": "value", "n": 42}
	req, err := buildRequest(http.MethodGet, "https://api.honeycomb.io", "/1/boards", fields, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != http.MethodGet {
		t.Errorf("method = %q, want GET", req.Method)
	}

	q := req.URL.Query()
	if q.Get("key") != "value" {
		t.Errorf("query key = %q, want %q", q.Get("key"), "value")
	}
	if q.Get("n") != "42" {
		t.Errorf("query n = %q, want %q", q.Get("n"), "42")
	}
	if req.Body != nil {
		t.Error("GET with fields should have nil body")
	}
}

func TestBuildRequest_POST_WithFields(t *testing.T) {
	fields := map[string]any{"name": "test"}
	req, err := buildRequest(http.MethodPost, "https://api.honeycomb.io", "/1/boards", fields, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if req.Method != http.MethodPost {
		t.Errorf("method = %q, want POST", req.Method)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("content-type = %q, want application/json", req.Header.Get("Content-Type"))
	}

	body, _ := io.ReadAll(req.Body)
	if !strings.Contains(string(body), `"name":"test"`) {
		t.Errorf("body = %q, want JSON with name field", body)
	}
}

func TestBuildRequest_WithBody(t *testing.T) {
	body := strings.NewReader(`{"data":[]}`)
	req, err := buildRequest(http.MethodPost, "https://api.honeycomb.io", "/1/events/ds", nil, body, nil)
	if err != nil {
		t.Fatal(err)
	}

	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("content-type = %q, want application/json", req.Header.Get("Content-Type"))
	}
}

func TestBuildRequest_CustomHeaders(t *testing.T) {
	headers := []string{"Accept: text/plain", "X-Custom: foo"}
	req, err := buildRequest(http.MethodGet, "https://api.honeycomb.io", "/1/boards", nil, nil, headers)
	if err != nil {
		t.Fatal(err)
	}

	if req.Header.Get("Accept") != "text/plain" {
		t.Errorf("Accept = %q, want text/plain", req.Header.Get("Accept"))
	}
	if req.Header.Get("X-Custom") != "foo" {
		t.Errorf("X-Custom = %q, want foo", req.Header.Get("X-Custom"))
	}
}

func TestBuildRequest_InvalidHeader(t *testing.T) {
	_, err := buildRequest(http.MethodGet, "https://api.honeycomb.io", "/1/boards", nil, nil, []string{"no-colon"})
	if err == nil {
		t.Fatal("expected error for invalid header")
	}
}

func TestNextPageURL(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "has next",
			header: `<https://api.honeycomb.io/1/columns/ds?cursor=abc>; rel="next"`,
			want:   "https://api.honeycomb.io/1/columns/ds?cursor=abc",
		},
		{
			name:   "no next",
			header: `<https://api.honeycomb.io/1/columns/ds?cursor=abc>; rel="prev"`,
			want:   "",
		},
		{
			name:   "empty",
			header: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{Header: http.Header{}}
			if tt.header != "" {
				resp.Header.Set("Link", tt.header)
			}
			got := nextPageURL(resp)
			if got != tt.want {
				t.Errorf("nextPageURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteResponseHeaders(t *testing.T) {
	resp := &http.Response{
		Proto:  "HTTP/1.1",
		Status: "200 OK",
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
	}

	var buf strings.Builder
	writeResponseHeaders(&buf, resp)
	out := buf.String()

	if !strings.Contains(out, "HTTP/1.1 200 OK") {
		t.Errorf("missing status line in %q", out)
	}
	if !strings.Contains(out, "Content-Type: application/json") {
		t.Errorf("missing content-type in %q", out)
	}
}

func TestBuildRequest_PaginationFullURL(t *testing.T) {
	req, err := buildRequest(http.MethodGet, "", "https://api.honeycomb.io/1/columns/ds?cursor=abc", nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if req.URL.String() != "https://api.honeycomb.io/1/columns/ds?cursor=abc" {
		t.Errorf("url = %q, want full pagination URL", req.URL.String())
	}
}
