package api

import (
	"bytes"
	"strings"
	"testing"
)

func TestFilterJQ(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		expr    string
		want    string
		wantErr bool
	}{
		{
			name:  "extract field",
			input: `{"name":"test","id":123}`,
			expr:  ".name",
			want:  "test\n",
		},
		{
			name:  "extract number",
			input: `{"count":42}`,
			expr:  ".count",
			want:  "42\n",
		},
		{
			name:  "array map",
			input: `[{"name":"a"},{"name":"b"}]`,
			expr:  ".[].name",
			want:  "a\nb\n",
		},
		{
			name:  "nested object",
			input: `{"a":{"b":"c"}}`,
			expr:  ".a.b",
			want:  "c\n",
		},
		{
			name:  "null value",
			input: `{"a":null}`,
			expr:  ".a",
			want:  "null\n",
		},
		{
			name:  "object output",
			input: `{"a":1,"b":2}`,
			expr:  "{x: .a}",
			want:  "{\"x\":1}\n",
		},
		{
			name:    "invalid expression",
			input:   `{}`,
			expr:    ".[invalid",
			wantErr: true,
		},
		{
			name:    "invalid json",
			input:   `not json`,
			expr:    ".",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := filterJQ(strings.NewReader(tt.input), &out, tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("filterJQ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && out.String() != tt.want {
				t.Errorf("filterJQ() = %q, want %q", out.String(), tt.want)
			}
		})
	}
}
