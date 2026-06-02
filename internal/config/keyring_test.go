package config

import (
	"net/http"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestParseKeyType(t *testing.T) {
	tests := []struct {
		input   string
		want    KeyType
		wantErr bool
	}{
		{"config", KeyConfig, false},
		{"ingest", KeyIngest, false},
		{"management", KeyManagement, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseKeyType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseKeyType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseKeyType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestKeyTypes(t *testing.T) {
	want := []KeyType{KeyConfig, KeyIngest, KeyManagement}
	got := KeyTypes()
	if len(got) != len(want) {
		t.Fatalf("KeyTypes() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("KeyTypes()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestManagementKey(t *testing.T) {
	if got := ManagementKey("id-123", "secret-456"); got != "id-123:secret-456" {
		t.Errorf("ManagementKey() = %q, want %q", got, "id-123:secret-456")
	}
}

func TestSetManagementKey(t *testing.T) {
	keyring.MockInit()

	if err := SetManagementKey("test-profile", "id-123", "secret-456"); err != nil {
		t.Fatalf("SetManagementKey() error = %v", err)
	}
	t.Cleanup(func() { _ = DeleteKey("test-profile", KeyManagement) })

	got, err := GetKey("test-profile", KeyManagement)
	if err != nil {
		t.Fatalf("GetKey() error = %v", err)
	}
	if got != "id-123:secret-456" {
		t.Errorf("stored value = %q, want %q", got, "id-123:secret-456")
	}
}

func TestApplyAuth(t *testing.T) {
	tests := []struct {
		name       string
		kt         KeyType
		key        string
		wantHeader string
		wantValue  string
	}{
		{
			name:       "config key",
			kt:         KeyConfig,
			key:        "my-config-key",
			wantHeader: "X-Honeycomb-Team",
			wantValue:  "my-config-key",
		},
		{
			name:       "ingest key",
			kt:         KeyIngest,
			key:        "my-ingest-key",
			wantHeader: "X-Honeycomb-Team",
			wantValue:  "my-ingest-key",
		},
		{
			name:       "management key",
			kt:         KeyManagement,
			key:        "my-mgmt-key",
			wantHeader: "Authorization",
			wantValue:  "Bearer my-mgmt-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
			ApplyAuth(req, tt.kt, tt.key)
			got := req.Header.Get(tt.wantHeader)
			if got != tt.wantValue {
				t.Errorf("header %q = %q, want %q", tt.wantHeader, got, tt.wantValue)
			}
		})
	}
}
