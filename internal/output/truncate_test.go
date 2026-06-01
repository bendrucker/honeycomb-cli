package output

import (
	"testing"
	"unicode/utf8"
)

func TestTruncate(t *testing.T) {
	for _, tc := range []struct {
		name      string
		input     string
		max       int
		expected  string
		truncated bool
	}{
		{
			name:     "short ascii",
			input:    "hello",
			max:      10,
			expected: "hello",
		},
		{
			name:     "exact max",
			input:    "hello",
			max:      5,
			expected: "hello",
		},
		{
			name:      "truncated ascii",
			input:     "hello world",
			max:       8,
			expected:  "hello...",
			truncated: true,
		},
		{
			name:     "multibyte within limit",
			input:    "αβγδ",
			max:      10,
			expected: "αβγδ",
		},
		{
			name:      "multibyte truncated",
			input:     "αβγδεζηθ",
			max:       5,
			expected:  "αβ...",
			truncated: true,
		},
		{
			name:      "emoji truncated",
			input:     "🎉🎊🎈🎁🎀",
			max:       4,
			expected:  "🎉...",
			truncated: true,
		},
		{
			name:      "mixed truncated",
			input:     "hello αβγ world",
			max:       10,
			expected:  "hello α...",
			truncated: true,
		},
		{
			name:      "max zero",
			input:     "hello world",
			max:       0,
			expected:  "...",
			truncated: true,
		},
		{
			name:      "max one",
			input:     "hello world",
			max:       1,
			expected:  "...",
			truncated: true,
		},
		{
			name:      "max three",
			input:     "hello world",
			max:       3,
			expected:  "...",
			truncated: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := Truncate(tc.input, tc.max)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
			if tc.truncated && !utf8.ValidString(result) {
				t.Errorf("result %q is not valid UTF-8", result)
			}
		})
	}
}
