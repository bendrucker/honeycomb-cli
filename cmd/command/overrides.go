package command

import (
	"encoding/json"
	"fmt"
)

// ApplyOverrides merges overrides into a JSON object body and re-encodes it.
// It returns data unchanged when there are no overrides, so callers can build
// the map from whichever flags the user actually set and pass it through
// unconditionally.
func ApplyOverrides(data []byte, overrides map[string]any) ([]byte, error) {
	if len(overrides) == 0 {
		return data, nil
	}

	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	for k, v := range overrides {
		body[k] = v
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encoding JSON: %w", err)
	}

	return encoded, nil
}
