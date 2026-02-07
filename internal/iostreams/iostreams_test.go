package iostreams

import "testing"

func TestSystem(t *testing.T) {
	ios := System()
	if ios.In == nil || ios.Out == nil || ios.Err == nil {
		t.Fatal("System() returned nil streams")
	}
}

func TestTest(t *testing.T) {
	ts := Test()
	if ts.CanPrompt() {
		t.Fatal("Test streams should not be promptable")
	}
	if ts.IsStdoutTTY() {
		t.Fatal("Test streams should not be TTY")
	}
}

func TestCanPrompt(t *testing.T) {
	ios := &IOStreams{stdinIsTTY: true, stdoutIsTTY: true}
	if !ios.CanPrompt() {
		t.Fatal("expected CanPrompt true with both TTYs")
	}

	ios.SetNeverPrompt(true)
	if ios.CanPrompt() {
		t.Fatal("expected CanPrompt false after SetNeverPrompt")
	}
}

func TestColorEnabled(t *testing.T) {
	ts := Test()
	if ts.ColorEnabled() {
		t.Fatal("Test streams should not have color enabled")
	}
}
