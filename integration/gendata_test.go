//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestGenData(t *testing.T) {
	type event struct {
		Time string         `json:"time"`
		Data map[string]any `json:"data"`
	}

	now := time.Now()
	routes := []string{"/api/users", "/api/orders", "/api/products", "/api/health", "/api/search"}
	methods := []string{"GET", "POST", "GET", "GET", "GET"}
	services := []string{"api-gateway", "user-service", "order-service"}
	statusCodes := []int{200, 200, 200, 200, 200, 200, 200, 200, 201, 204, 400, 404, 500, 502, 503}

	var events []event
	for range 200 {
		routeIdx := rand.Intn(len(routes))
		status := statusCodes[rand.Intn(len(statusCodes))]

		var duration float64
		switch {
		case status >= 500:
			duration = 200 + rand.Float64()*1800
		case routes[routeIdx] == "/api/search":
			duration = 50 + rand.Float64()*450
		default:
			duration = 5 + rand.Float64()*145
		}

		ts := now.Add(-time.Duration(rand.Intn(7200)) * time.Second)

		events = append(events, event{
			Time: ts.UTC().Format(time.RFC3339),
			Data: map[string]any{
				"service.name":     services[rand.Intn(len(services))],
				"http.method":      methods[routeIdx],
				"http.route":       routes[routeIdx],
				"http.status_code": status,
				"duration_ms":      float64(int(duration*100)) / 100,
				"error":            status >= 400,
				"trace.trace_id":   fmt.Sprintf("trace-%05d", rand.Intn(100000)),
				"user.id":          fmt.Sprintf("user-%d", rand.Intn(50)+1),
			},
		})
	}

	body, err := json.Marshal(events)
	if err != nil {
		t.Fatalf("encoding events: %v", err)
	}

	url := "https://api.honeycomb.io"
	if apiURL != "" {
		url = apiURL
	}

	req, err := http.NewRequest("POST", url+"/1/batch/"+dataset, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	req.Header.Set("X-Honeycomb-Team", configKeySecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("sending events: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		var result json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("HTTP %d (failed to read body: %v)", resp.StatusCode, err)
		}
		t.Fatalf("HTTP %d: %s", resp.StatusCode, result)
	}

	t.Logf("sent %d events to %s", len(events), dataset)
}
