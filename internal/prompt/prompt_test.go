package prompt

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// lineWriter returns a pipe where each line is written sequentially.
// This avoids bufio.Scanner reading ahead and consuming multiple lines at once.
func lineWriter(lines ...string) (io.Reader, io.WriteCloser) {
	r, w := io.Pipe()
	go func() {
		for _, line := range lines {
			_, _ = w.Write([]byte(line))
		}
		_ = w.Close()
	}()
	return r, w
}

func TestReadLine(t *testing.T) {
	r := strings.NewReader("hello\n")
	got, err := ReadLine(r)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Errorf("ReadLine = %q, want %q", got, "hello")
	}
}

func TestReadLine_EOF(t *testing.T) {
	r := strings.NewReader("")
	_, err := ReadLine(r)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if !strings.Contains(err.Error(), "unexpected end of input") {
		t.Errorf("err = %q, want unexpected end of input", err.Error())
	}
}

func TestLine(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("world\n")

	got, err := Line(&out, in, "Enter: ")
	if err != nil {
		t.Fatal(err)
	}
	if got != "world" {
		t.Errorf("Line = %q, want %q", got, "world")
	}
	if out.String() != "Enter: " {
		t.Errorf("prompt output = %q, want %q", out.String(), "Enter: ")
	}
}

func TestChoice(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("YES\n")

	got, err := Choice(&out, in, "Pick: ", []string{"yes", "no"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "yes" {
		t.Errorf("Choice = %q, want %q", got, "yes")
	}
}

func TestChoice_InvalidThenValid(t *testing.T) {
	var out bytes.Buffer
	r, w := lineWriter("maybe\n", "no\n")
	defer func() { _ = w.Close() }()

	got, err := Choice(&out, r, "Pick: ", []string{"yes", "no"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "no" {
		t.Errorf("Choice = %q, want %q", got, "no")
	}
	if !strings.Contains(out.String(), "Invalid choice") {
		t.Errorf("output = %q, want Invalid choice message", out.String())
	}
}

func TestChoice_EOF(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("")

	_, err := Choice(&out, in, "Pick: ", []string{"yes", "no"})
	if err == nil {
		t.Fatal("expected error for EOF")
	}
}

func TestSecret_FallbackToReadLine(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("secret-value\n")

	got, err := Secret(&out, in, 0, "Password: ")
	if err != nil {
		t.Fatal(err)
	}
	if got != "secret-value" {
		t.Errorf("Secret = %q, want %q", got, "secret-value")
	}
	if out.String() != "Password: " {
		t.Errorf("prompt output = %q, want %q", out.String(), "Password: ")
	}
}
