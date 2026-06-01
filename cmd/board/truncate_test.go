package board

import "testing"

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  string
	}{
		{"short ASCII", "hello", 10, "hello"},
		{"exact max", "hello", 5, "hello"},
		{"truncated ASCII", "hello world", 8, "hello..."},
		{"multibyte within limit", "αβγδ", 10, "αβγδ"},
		{"multibyte truncated", "αβγδεζηθ", 5, "αβ..."},
		{"emoji truncated", "🎉🎊🎈🎁🎀", 4, "🎉..."},
		{"mixed ASCII and multibyte", "hello αβγ world", 10, "hello α..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
		})
	}
}
