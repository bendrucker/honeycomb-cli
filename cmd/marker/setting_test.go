package marker

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestSettingList(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/marker_settings/test" {
			t.Errorf("path = %q, want /1/marker_settings/test", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":         "ms1",
				"type":       "deploys",
				"color":      "#F96E11",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-02T00:00:00Z",
			},
			{
				"id":    "ms2",
				"type":  "incidents",
				"color": "#FF0000",
			},
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "setting", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []settingItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "ms1" {
		t.Errorf("items[0].ID = %q, want %q", items[0].ID, "ms1")
	}
	if items[0].Type != "deploys" {
		t.Errorf("items[0].Type = %q, want %q", items[0].Type, "deploys")
	}
	if items[0].Color != "#F96E11" {
		t.Errorf("items[0].Color = %q, want %q", items[0].Color, "#F96E11")
	}
	if items[0].UpdatedAt != "2024-01-02T00:00:00Z" {
		t.Errorf("items[0].UpdatedAt = %q, want %q", items[0].UpdatedAt, "2024-01-02T00:00:00Z")
	}
}

func TestSettingList_Empty(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "setting", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var items []settingItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestSettingCreate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["type"] != "deploys" {
			t.Errorf("body type = %q, want %q", body["type"], "deploys")
		}
		if body["color"] != "#F96E11" {
			t.Errorf("body color = %q, want %q", body["color"], "#F96E11")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "ms1",
			"type":       "deploys",
			"color":      "#F96E11",
			"created_at": "2024-01-01T00:00:00Z",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "setting", "create", "--type", "deploys", "--color", "#F96E11"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var item settingItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &item); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if item.ID != "ms1" {
		t.Errorf("ID = %q, want %q", item.ID, "ms1")
	}
	if item.Type != "deploys" {
		t.Errorf("Type = %q, want %q", item.Type, "deploys")
	}
}

func TestSettingUpdate(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/ms1") {
			t.Errorf("path = %q, want suffix /ms1", r.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["color"] != "#00FF00" {
			t.Errorf("body color = %q, want %q", body["color"], "#00FF00")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":    "ms1",
			"type":  "deploys",
			"color": "#00FF00",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "setting", "update", "ms1", "--color", "#00FF00"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var item settingItem
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &item); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if item.Color != "#00FF00" {
		t.Errorf("Color = %q, want %q", item.Color, "#00FF00")
	}
}

func TestSettingUpdate_NoFlags(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "ms1", "type": "deploys", "color": "#F96E11",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "setting", "update", "ms1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestSettingDelete_WithYes(t *testing.T) {
	opts, ts := setupTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":    "ms1",
			"type":  "deploys",
			"color": "#F96E11",
		})
	}))

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "setting", "delete", "ms1", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := ts.ErrBuf.String()
	if !strings.Contains(output, "Deleted marker setting ms1") {
		t.Errorf("stderr = %q, want deletion message", output)
	}
}

func TestSettingDelete_NoYesNonInteractive(t *testing.T) {
	opts, _ := setupTest(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	opts.IOStreams.SetNeverPrompt(true)

	cmd := NewCmd(opts)
	cmd.SetArgs([]string{"--dataset", "test", "setting", "delete", "ms1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --yes")
	}
	if !strings.Contains(err.Error(), "--yes is required") {
		t.Errorf("error = %q, want --yes required message", err.Error())
	}
}
