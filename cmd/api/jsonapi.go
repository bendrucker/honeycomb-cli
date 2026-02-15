package api

import (
	"encoding/json"
	"maps"
	"net/http"
	"net/url"
	"strings"
)

func isV2Path(path string) bool {
	return strings.HasPrefix(extractPath(path), "/2/")
}

func extractPath(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		if u, err := url.Parse(path); err == nil {
			return u.Path
		}
	}
	return path
}

func contentTypeForPath(path string) string {
	if isV2Path(path) {
		return "application/vnd.api+json"
	}
	return "application/json"
}

func inferResourceType(method, path string) string {
	path = extractPath(path)
	if idx := strings.IndexByte(path, '?'); idx >= 0 {
		path = path[:idx]
	}
	path = strings.TrimRight(path, "/")

	segments := strings.Split(path, "/")

	if (method == http.MethodPatch || method == http.MethodPut) && len(segments) >= 2 {
		return segments[len(segments)-2]
	}

	return segments[len(segments)-1]
}

type jsonAPIResource struct {
	ID         string         `json:"id,omitempty"`
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes"`
}

func wrapJSONAPI(fields map[string]any, resourceType string) map[string]any {
	return map[string]any{
		"data": jsonAPIResource{
			Type:       resourceType,
			Attributes: fields,
		},
	}
}

func unwrapJSONAPI(data []byte) ([]byte, error) {
	var probe struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return data, nil
	}
	if probe.Data == nil {
		return data, nil
	}

	// Try single resource
	var single jsonAPIResource
	if err := json.Unmarshal(probe.Data, &single); err == nil && single.Type != "" {
		return json.Marshal(flattenResource(single))
	}

	// Try list of resources
	var list []jsonAPIResource
	if err := json.Unmarshal(probe.Data, &list); err == nil {
		flat := make([]map[string]any, len(list))
		for i, r := range list {
			flat[i] = flattenResource(r)
		}
		return json.Marshal(flat)
	}

	return data, nil
}

func flattenResource(r jsonAPIResource) map[string]any {
	flat := make(map[string]any)
	maps.Copy(flat, r.Attributes)

	// id and type from the envelope take precedence over attributes
	if r.ID != "" {
		flat["id"] = r.ID
	}
	flat["type"] = r.Type
	return flat
}
