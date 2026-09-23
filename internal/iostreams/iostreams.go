// Package iostreams provides IO stream abstractions for reading input and
// writing output.
package iostreams

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mattn/go-isatty"
)

// IOStreams holds the input, output, and error streams used by commands.
type IOStreams struct {
	In  io.ReadCloser
	Out io.Writer
	Err io.Writer

	stdinIsTTY  bool
	stdoutIsTTY bool
	neverPrompt bool
}

// System returns IOStreams backed by the process's stdin, stdout, and stderr.
func System() *IOStreams {
	stdinFd := os.Stdin.Fd()
	stdoutFd := os.Stdout.Fd()

	return &IOStreams{
		In:          os.Stdin,
		Out:         os.Stdout,
		Err:         os.Stderr,
		stdinIsTTY:  isatty.IsTerminal(stdinFd) || isatty.IsCygwinTerminal(stdinFd),
		stdoutIsTTY: isatty.IsTerminal(stdoutFd) || isatty.IsCygwinTerminal(stdoutFd),
	}
}

// TestStreams holds IOStreams backed by in-memory buffers for tests.
type TestStreams struct {
	*IOStreams
	InBuf  *bytes.Buffer
	OutBuf *bytes.Buffer
	ErrBuf *bytes.Buffer
}

// Test returns TestStreams backed by in-memory buffers for tests.
func Test(tb testing.TB) *TestStreams {
	tb.Helper()
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	return &TestStreams{
		IOStreams: &IOStreams{
			In:  io.NopCloser(in),
			Out: out,
			Err: io.MultiWriter(errBuf, &testLogWriter{tb: tb}),
		},
		InBuf:  in,
		OutBuf: out,
		ErrBuf: errBuf,
	}
}

// TestPromptable returns test streams whose CanPrompt reports true, so tests
// can exercise interactive prompt paths. Write prompt answers to InBuf.
func TestPromptable(tb testing.TB) *TestStreams {
	tb.Helper()
	ts := Test(tb)
	ts.stdinIsTTY = true
	ts.stdoutIsTTY = true
	return ts
}

type testLogWriter struct {
	tb testing.TB
}

func (w *testLogWriter) Write(p []byte) (int, error) {
	w.tb.Helper()
	w.tb.Log(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}

// CanPrompt reports whether both stdin and stdout are TTYs and prompting is
// not disabled.
func (s *IOStreams) CanPrompt() bool {
	return s.stdinIsTTY && s.stdoutIsTTY && !s.neverPrompt
}

// SetNeverPrompt disables interactive prompting when v is true.
func (s *IOStreams) SetNeverPrompt(v bool) {
	s.neverPrompt = v
}

// IsStdinTTY reports whether stdin is a TTY.
func (s *IOStreams) IsStdinTTY() bool {
	return s.stdinIsTTY
}

// IsStdoutTTY reports whether stdout is a TTY.
func (s *IOStreams) IsStdoutTTY() bool {
	return s.stdoutIsTTY
}

// ColorEnabled reports whether color output should be used, honoring the
// NO_COLOR environment variable.
func (s *IOStreams) ColorEnabled() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	return s.stdoutIsTTY
}
