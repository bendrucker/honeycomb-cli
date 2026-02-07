package agent

import "testing"

func TestDetect(t *testing.T) {
	a := Detect()
	if a != nil {
		t.Fatalf("expected nil agent in test, got %q", a.Name)
	}

	t.Setenv("CLAUDE_CODE", "1")
	a = Detect()
	if a == nil || a.Name != "claude-code" {
		t.Fatalf("expected claude-code agent, got %v", a)
	}
}
