package api

import (
	"testing"

	"github.com/bendrucker/honeycomb-cli/internal/config"
)

func TestInferKeyType(t *testing.T) {
	tests := []struct {
		path string
		want config.KeyType
	}{
		{"/1/auth", config.KeyConfig},
		{"/1/boards", config.KeyConfig},
		{"/1/columns/my-dataset", config.KeyConfig},
		{"/1/events/my-dataset", config.KeyIngest},
		{"/1/batch/my-dataset", config.KeyIngest},
		{"/1/kinesis_events/my-dataset", config.KeyIngest},
		{"/2/teams", config.KeyManagement},
		{"/2/auth", config.KeyManagement},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := inferKeyType(tt.path)
			if got != tt.want {
				t.Errorf("inferKeyType(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
