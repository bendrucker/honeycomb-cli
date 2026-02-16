package iostreams

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mattn/go-isatty"
)

type IOStreams struct {
	In  io.ReadCloser
	Out io.Writer
	Err io.Writer

	stdinIsTTY  bool
	stdoutIsTTY bool
	neverPrompt bool
}

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

type TestStreams struct {
	*IOStreams
	InBuf  *bytes.Buffer
	OutBuf *bytes.Buffer
	ErrBuf *bytes.Buffer
}

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

type testLogWriter struct {
	tb testing.TB
}

func (w *testLogWriter) Write(p []byte) (int, error) {
	w.tb.Helper()
	w.tb.Log(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}

func (s *IOStreams) CanPrompt() bool {
	return s.stdinIsTTY && s.stdoutIsTTY && !s.neverPrompt
}

func (s *IOStreams) SetNeverPrompt(v bool) {
	s.neverPrompt = v
}

func (s *IOStreams) IsStdinTTY() bool {
	return s.stdinIsTTY
}

func (s *IOStreams) IsStdoutTTY() bool {
	return s.stdoutIsTTY
}

func (s *IOStreams) ColorEnabled() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	return s.stdoutIsTTY
}
