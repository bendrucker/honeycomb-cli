package mcp

import (
	"reflect"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestSchemaType(t *testing.T) {
	for _, tc := range []struct {
		name string
		prop map[string]any
		want string
	}{
		{
			name: "plain string",
			prop: map[string]any{"type": "object"},
			want: "object",
		},
		{
			name: "nullable list picks non null",
			prop: map[string]any{"type": []any{"object", "null"}},
			want: "object",
		},
		{
			name: "null first in list",
			prop: map[string]any{"type": []any{"null", "integer"}},
			want: "integer",
		},
		{
			name: "absent type",
			prop: map[string]any{"description": "no type here"},
			want: "",
		},
		{
			name: "oneof not recognized",
			prop: map[string]any{"oneOf": []any{}},
			want: "",
		},
		{
			name: "list of only null",
			prop: map[string]any{"type": []any{"null"}},
			want: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := schemaType(tc.prop); got != tc.want {
				t.Errorf("schemaType = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCoerceValue(t *testing.T) {
	for _, tc := range []struct {
		name         string
		value        string
		declaredType string
		want         any
		wantErr      bool
	}{
		{
			name:         "object",
			value:        `{"a":1}`,
			declaredType: "object",
			want:         map[string]any{"a": float64(1)},
		},
		{
			name:         "array",
			value:        `[1,2]`,
			declaredType: "array",
			want:         []any{float64(1), float64(2)},
		},
		{
			name:         "number",
			value:        "1.5",
			declaredType: "number",
			want:         1.5,
		},
		{
			name:         "integer",
			value:        "42",
			declaredType: "integer",
			want:         int64(42),
		},
		{
			name:         "boolean",
			value:        "true",
			declaredType: "boolean",
			want:         true,
		},
		{
			name:         "string left as is",
			value:        "hello",
			declaredType: "string",
			want:         "hello",
		},
		{
			name:         "unknown type left as is",
			value:        "hello",
			declaredType: "",
			want:         "hello",
		},
		{
			name:         "json string stays string when typed string",
			value:        `{"a":1}`,
			declaredType: "string",
			want:         `{"a":1}`,
		},
		{
			name:         "invalid object",
			value:        "not json",
			declaredType: "object",
			wantErr:      true,
		},
		{
			name:         "scalar rejected for object",
			value:        "42",
			declaredType: "object",
			wantErr:      true,
		},
		{
			name:         "object rejected for array",
			value:        `{"a":1}`,
			declaredType: "array",
			wantErr:      true,
		},
		{
			name:         "nan rejected for number",
			value:        "NaN",
			declaredType: "number",
			wantErr:      true,
		},
		{
			name:         "inf rejected for number",
			value:        "Inf",
			declaredType: "number",
			wantErr:      true,
		},
		{
			name:         "invalid integer",
			value:        "abc",
			declaredType: "integer",
			wantErr:      true,
		},
		{
			name:         "invalid boolean",
			value:        "maybe",
			declaredType: "boolean",
			wantErr:      true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := coerceValue(tc.value, tc.declaredType)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %#v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("coerceValue = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestCoerceArgs(t *testing.T) {
	schema := mcp.ToolInputSchema{
		Properties: map[string]any{
			"query_json": map[string]any{"type": "object"},
			"limit":      map[string]any{"type": "integer"},
			"dataset":    map[string]any{"type": "string"},
		},
	}

	t.Run("coerces string values by declared type", func(t *testing.T) {
		args := map[string]any{
			"query_json": `{"calculations":[{"op":"COUNT"}]}`,
			"limit":      "10",
			"dataset":    "api",
		}
		if err := coerceArgs(args, schema); err != nil {
			t.Fatal(err)
		}

		if _, ok := args["query_json"].(map[string]any); !ok {
			t.Errorf("query_json = %#v, want map[string]any", args["query_json"])
		}
		if args["limit"] != int64(10) {
			t.Errorf("limit = %#v, want int64(10)", args["limit"])
		}
		if args["dataset"] != "api" {
			t.Errorf("dataset = %#v, want string %q", args["dataset"], "api")
		}
	})

	t.Run("leaves already typed values untouched", func(t *testing.T) {
		nested := map[string]any{"sub": "value"}
		args := map[string]any{"query_json": nested}
		if err := coerceArgs(args, schema); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(args["query_json"], nested) {
			t.Errorf("query_json = %#v, want it left as the original map", args["query_json"])
		}
	})

	t.Run("leaves unknown keys as strings", func(t *testing.T) {
		args := map[string]any{"unknown": `{"a":1}`}
		if err := coerceArgs(args, schema); err != nil {
			t.Fatal(err)
		}
		if args["unknown"] != `{"a":1}` {
			t.Errorf("unknown = %#v, want it left as a string", args["unknown"])
		}
	})

	t.Run("returns error for unparseable value", func(t *testing.T) {
		args := map[string]any{"limit": "not-a-number"}
		err := coerceArgs(args, schema)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
