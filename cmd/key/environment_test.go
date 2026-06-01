package key

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/internal/api"
)

func TestResolveEnvironment(t *testing.T) {
	for _, tc := range []struct {
		name      string
		value     string
		pages     [][]envStub
		wantID    string
		wantCalls int
		wantErr   string
	}{
		{
			name:      "id passes through without api call",
			value:     "hcaen_01abc",
			wantID:    "hcaen_01abc",
			wantCalls: 0,
		},
		{
			name:      "name resolves to id",
			value:     "production",
			pages:     [][]envStub{{{ID: "hcaen_01prod", Name: "production"}, {ID: "hcaen_01stg", Name: "staging"}}},
			wantID:    "hcaen_01prod",
			wantCalls: 1,
		},
		{
			name:  "name resolves on second page",
			value: "production",
			pages: [][]envStub{
				{{ID: "hcaen_01stg", Name: "staging"}},
				{{ID: "hcaen_01prod", Name: "production"}},
			},
			wantID:    "hcaen_01prod",
			wantCalls: 2,
		},
		{
			name:      "name with no match errors",
			value:     "missing",
			pages:     [][]envStub{{{ID: "hcaen_01stg", Name: "staging"}}},
			wantErr:   `no environment found with name "missing"`,
			wantCalls: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls int
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				page := 0
				if after := r.URL.Query().Get("page[after]"); after != "" {
					page = pageIndex(after)
				}
				w.Header().Set("Content-Type", "application/vnd.api+json")
				_, _ = w.Write(envListBody(tc.pages, page))
			}))
			t.Cleanup(srv.Close)

			client, err := api.NewClientWithResponses(srv.URL)
			if err != nil {
				t.Fatal(err)
			}

			gotID, err := resolveEnvironment(context.Background(), client, "my-team", tc.value)

			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %v, want to contain %q", err, tc.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotID != tc.wantID {
				t.Errorf("id = %q, want %q", gotID, tc.wantID)
			}
			if calls != tc.wantCalls {
				t.Errorf("api calls = %d, want %d", calls, tc.wantCalls)
			}
		})
	}
}

type envStub struct {
	ID   string
	Name string
}

func pageIndex(cursor string) int {
	switch cursor {
	case "p1":
		return 1
	default:
		return 0
	}
}

func envListBody(pages [][]envStub, page int) []byte {
	var data []map[string]any
	if page < len(pages) {
		for _, e := range pages[page] {
			data = append(data, map[string]any{
				"id":         e.ID,
				"type":       "environments",
				"attributes": map[string]any{"name": e.Name},
			})
		}
	}

	body := map[string]any{"data": data}
	if page+1 < len(pages) {
		body["links"] = map[string]any{"next": "https://api.honeycomb.io/2/teams/my-team/environments?page[after]=p1"}
	}

	out, _ := json.Marshal(body)
	return out
}
