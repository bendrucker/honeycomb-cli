package fields

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		raw     []string
		typed   []string
		want    string
		wantErr bool
	}{
		{
			name: "simple string field",
			raw:  []string{"name=My Board"},
			want: `{"name":"My Board"}`,
		},
		{
			name:  "typed boolean",
			typed: []string{"active=true"},
			want:  `{"active":true}`,
		},
		{
			name:  "typed null",
			typed: []string{"value=null"},
			want:  `{"value":null}`,
		},
		{
			name:  "typed number",
			typed: []string{"count=42"},
			want:  `{"count":42}`,
		},
		{
			name:  "typed float",
			typed: []string{"rate=3.14"},
			want:  `{"rate":3.14}`,
		},
		{
			name:  "typed string passthrough",
			typed: []string{"name=hello"},
			want:  `{"name":"hello"}`,
		},
		{
			name: "nested bracket notation",
			raw:  []string{"foo[bar]=baz"},
			want: `{"foo":{"bar":"baz"}}`,
		},
		{
			name: "array bracket notation",
			raw:  []string{"foo[]=1", "foo[]=2"},
			want: `{"foo":["1","2"]}`,
		},
		{
			name:  "mixed raw and typed",
			raw:   []string{"name=test"},
			typed: []string{"count=5"},
			want:  `{"count":5,"name":"test"}`,
		},
		{
			name:    "invalid raw field",
			raw:     []string{"noequals"},
			wantErr: true,
		},
		{
			name:    "invalid typed field",
			typed:   []string{"noequals"},
			wantErr: true,
		},
		{
			name: "deeply nested brackets",
			raw:  []string{"a[b][c]=deep"},
			want: `{"a":{"b":{"c":"deep"}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.raw, tt.typed, strings.NewReader(""))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			b, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != tt.want {
				t.Errorf("Parse() = %s, want %s", b, tt.want)
			}
		})
	}
}

func TestParse_AtFile(t *testing.T) {
	tmp := t.TempDir() + "/data.txt"
	if err := os.WriteFile(tmp, []byte("file-content"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Parse(nil, []string{"data=@" + tmp}, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}

	if got["data"] != "file-content" {
		t.Errorf("data = %q, want %q", got["data"], "file-content")
	}
}

func TestParse_AtStdin(t *testing.T) {
	stdin := strings.NewReader("stdin-content")
	got, err := Parse(nil, []string{"data=@-"}, stdin)
	if err != nil {
		t.Fatal(err)
	}

	if got["data"] != "stdin-content" {
		t.Errorf("data = %q, want %q", got["data"], "stdin-content")
	}
}

func TestParse_AtFileMissing(t *testing.T) {
	_, err := Parse(nil, []string{"data=@/nonexistent/file.txt"}, strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "reading @") {
		t.Errorf("error = %q, want @file context", err.Error())
	}
}
