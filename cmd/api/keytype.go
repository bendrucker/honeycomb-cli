package api

import (
	"strings"

	"github.com/bendrucker/honeycomb-cli/internal/config"
)

func inferKeyType(path string) config.KeyType {
	if strings.HasPrefix(path, "/2/") {
		return config.KeyManagement
	}

	parts := strings.SplitN(path, "/", 4)
	if len(parts) >= 3 {
		switch parts[2] {
		case "events", "batch", "kinesis_events":
			return config.KeyIngest
		}
	}

	return config.KeyConfig
}
