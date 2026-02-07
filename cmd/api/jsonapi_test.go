package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestIsV2Path(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/1/auth", false},
		{"/1/boards", false},
		{"/2/teams", true},
		{"/2/environments", true},
		{"https://api.honeycomb.io/2/teams?cursor=abc", true},
		{"https://api.honeycomb.io/1/boards", false},
		{"http://localhost:8080/2/environments", true},
		{"/other", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isV2Path(tt.path)
			if got != tt.want {
				t.Errorf("isV2Path(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestContentTypeForPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/1/boards", "application/json"},
		{"/2/teams", "application/vnd.api+json"},
		{"https://api.honeycomb.io/2/teams", "application/vnd.api+json"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := contentTypeForPath(tt.path)
			if got != tt.want {
				t.Errorf("contentTypeForPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestInferResourceType(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		want   string
	}{
		{"POST collection", http.MethodPost, "/2/teams/slug/environments", "environments"},
		{"PATCH resource", http.MethodPatch, "/2/teams/slug/environments/env-id", "environments"},
		{"PUT resource", http.MethodPut, "/2/teams/slug/environments/env-id", "environments"},
		{"POST with trailing slash", http.MethodPost, "/2/teams/slug/environments/", "environments"},
		{"POST with query", http.MethodPost, "/2/teams/slug/environments?foo=bar", "environments"},
		{"GET collection", http.MethodGet, "/2/teams/slug/environments", "environments"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferResourceType(tt.method, tt.path)
			if got != tt.want {
				t.Errorf("inferResourceType(%q, %q) = %q, want %q", tt.method, tt.path, got, tt.want)
			}
		})
	}
}

func TestWrapJSONAPI(t *testing.T) {
	fields := map[string]any{"name": "prod", "slug": "prod"}
	wrapped := wrapJSONAPI(fields, "environments")

	data, ok := wrapped["data"].(jsonAPIResource)
	if !ok {
		t.Fatal("missing data key")
	}
	if data.Type != "environments" {
		t.Errorf("type = %v, want environments", data.Type)
	}
	if data.Attributes["name"] != "prod" {
		t.Errorf("attributes.name = %v, want prod", data.Attributes["name"])
	}
}

func TestUnwrapJSONAPI(t *testing.T) {
	t.Run("single resource", func(t *testing.T) {
		input := `{"data":{"id":"abc","type":"environments","attributes":{"name":"prod","slug":"prod"}}}`
		got, err := unwrapJSONAPI([]byte(input))
		if err != nil {
			t.Fatal(err)
		}

		var flat map[string]any
		if err := json.Unmarshal(got, &flat); err != nil {
			t.Fatal(err)
		}
		if flat["id"] != "abc" {
			t.Errorf("id = %v, want abc", flat["id"])
		}
		if flat["type"] != "environments" {
			t.Errorf("type = %v, want environments", flat["type"])
		}
		if flat["name"] != "prod" {
			t.Errorf("name = %v, want prod", flat["name"])
		}
	})

	t.Run("list of resources", func(t *testing.T) {
		input := `{"data":[{"id":"a","type":"environments","attributes":{"name":"prod"}},{"id":"b","type":"environments","attributes":{"name":"staging"}}]}`
		got, err := unwrapJSONAPI([]byte(input))
		if err != nil {
			t.Fatal(err)
		}

		var flat []map[string]any
		if err := json.Unmarshal(got, &flat); err != nil {
			t.Fatal(err)
		}
		if len(flat) != 2 {
			t.Fatalf("len = %d, want 2", len(flat))
		}
		if flat[0]["name"] != "prod" {
			t.Errorf("first name = %v, want prod", flat[0]["name"])
		}
		if flat[1]["id"] != "b" {
			t.Errorf("second id = %v, want b", flat[1]["id"])
		}
	})

	t.Run("envelope id takes precedence over attribute id", func(t *testing.T) {
		input := `{"data":{"id":"envelope-id","type":"environments","attributes":{"id":"attr-id","name":"prod"}}}`
		got, err := unwrapJSONAPI([]byte(input))
		if err != nil {
			t.Fatal(err)
		}

		var flat map[string]any
		if err := json.Unmarshal(got, &flat); err != nil {
			t.Fatal(err)
		}
		if flat["id"] != "envelope-id" {
			t.Errorf("id = %v, want envelope-id (envelope should take precedence)", flat["id"])
		}
		if flat["type"] != "environments" {
			t.Errorf("type = %v, want environments", flat["type"])
		}
		if flat["name"] != "prod" {
			t.Errorf("name = %v, want prod", flat["name"])
		}
	})

	t.Run("non-jsonapi passthrough", func(t *testing.T) {
		input := `{"name":"plain json"}`
		got, err := unwrapJSONAPI([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != input {
			t.Errorf("got %q, want passthrough", string(got))
		}
	})

	t.Run("invalid json passthrough", func(t *testing.T) {
		input := `not json`
		got, err := unwrapJSONAPI([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != input {
			t.Errorf("got %q, want passthrough", string(got))
		}
	})
}
