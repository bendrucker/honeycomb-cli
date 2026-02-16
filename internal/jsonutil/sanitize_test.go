package jsonutil

import (
	"strings"
	"testing"
)

func TestSanitizeEscapes(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid JSON unchanged",
			input: `{"op":"!="}`,
			want:  `{"op":"!="}`,
		},
		{
			name:  "backslash bang from zsh",
			input: `{"op":"\!="}`,
			want:  `{"op":"!="}`,
		},
		{
			name:  "valid escape preserved",
			input: `{"msg":"line\nbreak"}`,
			want:  `{"msg":"line\nbreak"}`,
		},
		{
			name:  "escaped quote preserved",
			input: `{"msg":"say \"hello\""}`,
			want:  `{"msg":"say \"hello\""}`,
		},
		{
			name:  "escaped backslash preserved",
			input: `{"path":"C:\\Users"}`,
			want:  `{"path":"C:\\Users"}`,
		},
		{
			name:  "unicode escape preserved",
			input: `{"char":"\u0041"}`,
			want:  `{"char":"\u0041"}`,
		},
		{
			name:  "multiple invalid escapes",
			input: `{"a":"\!","b":"\="}`,
			want:  `{"a":"!","b":"="}`,
		},
		{
			name:  "no strings",
			input: `[1, 2, 3]`,
			want:  `[1, 2, 3]`,
		},
		{
			name:  "backslash outside string unchanged",
			input: `{"key":"val"}`,
			want:  `{"key":"val"}`,
		},
		{
			name:  "nested objects",
			input: `{"filters":[{"op":"\!=","value":"test"}]}`,
			want:  `{"filters":[{"op":"!=","value":"test"}]}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := string(SanitizeEscapes([]byte(tc.input)))
			if got != tc.want {
				t.Errorf("SanitizeEscapes(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid JSON passes through",
			input: `{"op":"!="}`,
			want:  `{"op":"!="}`,
		},
		{
			name:  "invalid escape gets sanitized",
			input: `{"op":"\!="}`,
			want:  `{"op":"!="}`,
		},
		{
			name:  "nested invalid escapes",
			input: `{"filters":[{"op":"\!=","value":"test"}]}`,
			want:  `{"filters":[{"op":"!=","value":"test"}]}`,
		},
		{
			name:    "unfixable JSON returns error",
			input:   `{invalid`,
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Sanitize([]byte(tc.input))
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("Sanitize(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitize_ErrorFromOriginal(t *testing.T) {
	_, err := Sanitize([]byte(`not json at all`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid character") {
		t.Errorf("error = %q, want original parse error", err.Error())
	}
}
