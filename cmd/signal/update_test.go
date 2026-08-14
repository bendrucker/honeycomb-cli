package signal

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestUpdate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		args     []string
		expected map[string]any
	}{
		{
			name:     "disable",
			args:     []string{"--enabled=false"},
			expected: map[string]any{"enabled": false},
		},
		{
			name:     "sensitivity",
			args:     []string{"--sensitivity", "low"},
			expected: map[string]any{"sensitivity": "low"},
		},
		{
			name:     "recipients",
			args:     []string{"--recipient", "r1", "--recipient", "r2"},
			expected: map[string]any{"recipients": []any{map[string]any{"id": "r1"}, map[string]any{"id": "r2"}}},
		},
		{
			name:     "clear recipients",
			args:     []string{"--clear-recipients"},
			expected: map[string]any{"recipients": []any{}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("method = %q, want PUT", r.Method)
				}
				if r.URL.Path != "/1/signals/sig-1" {
					t.Errorf("path = %q, want /1/signals/sig-1", r.URL.Path)
				}

				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				want, err := json.Marshal(tc.expected)
				if err != nil {
					t.Fatal(err)
				}
				got, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != string(want) {
					t.Errorf("body = %s, want %s", got, want)
				}

				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(signalJSON("sig-1", "checkout"))
			}))

			cmd := NewCmd(opts)
			cmd.SetArgs(append([]string{"update", "sig-1"}, tc.args...))
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUpdate_NoChanges(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("unexpected request")
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "sig-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no flags are set")
	}
	if !strings.Contains(err.Error(), "--enabled") {
		t.Errorf("error = %q, want mention of --enabled", err.Error())
	}
}

func TestUpdate_InvalidSensitivity(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("unexpected request")
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "sig-1", "--sensitivity", "extreme"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid sensitivity")
	}
	if !strings.Contains(err.Error(), "sensitivity") {
		t.Errorf("error = %q, want mention of sensitivity", err.Error())
	}
}

func TestUpdate_RecipientsMutuallyExclusive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("unexpected request")
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"update", "sig-1", "--recipient", "r1", "--clear-recipients"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
}
