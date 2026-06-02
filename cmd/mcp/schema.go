package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// findTool fetches the server's tool list and returns the tool with the given
// name. ListTools follows pagination internally. Resolving the tool up front
// validates the name (and surfaces its input schema) before the call is built,
// so an unknown tool fails with a clear error instead of a server-side rejection.
func findTool(ctx context.Context, c *mcpclient.Client, name string) (mcp.Tool, error) {
	result, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return mcp.Tool{}, fmt.Errorf("listing tools: %w", err)
	}
	for _, t := range result.Tools {
		if t.Name == name {
			return t, nil
		}
	}
	return mcp.Tool{}, fmt.Errorf("unknown tool %q", name)
}

// coerceArgs converts string-valued arguments to the type declared for them in
// the tool's input schema, so a value like -f query_json='{...}' reaches the
// server as the object the tool expects rather than a string it cannot
// interpret. Only top-level string values are coerced: non-string values (the
// already-typed results of -F, and nested maps/arrays built by key[sub]=
// syntax) are skipped, and a property the schema does not type stays a string.
func coerceArgs(args map[string]any, schema mcp.ToolInputSchema) error {
	for key, val := range args {
		s, ok := val.(string)
		if !ok {
			continue
		}

		prop, ok := schema.Properties[key].(map[string]any)
		if !ok {
			continue
		}

		coerced, err := coerceValue(s, schemaType(prop))
		if err != nil {
			return fmt.Errorf("argument %q: %w", key, err)
		}
		args[key] = coerced
	}
	return nil
}

// schemaType returns the JSON Schema type for a property. A plain string type
// is returned as-is; a type expressed as a list (e.g. ["string", "null"] for a
// nullable property) yields the first non-null entry. Anything else (a type
// declared through oneOf/anyOf, or absent entirely) returns the empty string,
// which coerceValue treats as leaving the value as a string.
func schemaType(prop map[string]any) string {
	switch t := prop["type"].(type) {
	case string:
		return t
	case []any:
		for _, item := range t {
			if s, ok := item.(string); ok && s != "null" {
				return s
			}
		}
	}
	return ""
}

// coerceValue converts a string argument to the Go value matching a declared
// JSON Schema type. An unknown type, the empty type, or "string" leaves the
// value as a string. A value that cannot be parsed as the declared type is an
// error: the user asked for a type the value does not satisfy.
func coerceValue(s, declaredType string) (any, error) {
	switch declaredType {
	case "object", "array":
		var v any
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			return nil, fmt.Errorf("invalid JSON for %s: %w", declaredType, err)
		}
		// json.Unmarshal accepts any valid JSON, including scalars, so verify
		// the parsed value is the declared kind rather than silently sending a
		// number/string/null where the tool expects an object or array.
		if !matchesJSONKind(v, declaredType) {
			return nil, fmt.Errorf("expected a JSON %s", declaredType)
		}
		return v, nil
	case "number":
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %w", err)
		}
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return nil, fmt.Errorf("invalid number: %q is not a finite JSON number", s)
		}
		return n, nil
	case "integer":
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer: %w", err)
		}
		return i, nil
	case "boolean":
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, fmt.Errorf("invalid boolean: %w", err)
		}
		return b, nil
	default:
		return s, nil
	}
}

// matchesJSONKind reports whether a value decoded by json.Unmarshal is the
// declared JSON Schema container kind: a map for "object", a slice for "array".
func matchesJSONKind(v any, declaredType string) bool {
	switch declaredType {
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "array":
		_, ok := v.([]any)
		return ok
	default:
		return true
	}
}
