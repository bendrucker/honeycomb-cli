// Package prompt provides interactive command-line prompt helpers.
package prompt

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"golang.org/x/term"
)

// Line writes prompt to out and reads a line of input from in.
func Line(out io.Writer, in io.Reader, prompt string) (string, error) {
	_, _ = fmt.Fprint(out, prompt)
	return ReadLine(in)
}

// Choice prompts until the input matches one of choices (case-insensitive),
// reprompting on an invalid entry.
func Choice(out io.Writer, in io.Reader, prompt string, choices []string) (string, error) {
	for {
		line, err := Line(out, in, prompt)
		if err != nil {
			return "", err
		}
		for _, c := range choices {
			if strings.EqualFold(line, c) {
				return c, nil
			}
		}
		_, _ = fmt.Fprintf(out, "Invalid choice. Options: %s\n", strings.Join(choices, ", "))
	}
}

// Secret writes prompt to out and reads a value without echoing it when fd
// is a terminal, falling back to a plain read from in otherwise.
func Secret(out io.Writer, in io.Reader, fd uintptr, prompt string) (string, error) {
	_, _ = fmt.Fprint(out, prompt)
	if fd != 0 {
		b, err := term.ReadPassword(int(fd))
		_, _ = fmt.Fprintln(out)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return ReadLine(in)
}

// ReadLine reads a single line from r, stripping the trailing newline. It
// returns an error if r has no more input.
func ReadLine(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		return strings.TrimRight(scanner.Text(), "\r\n"), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("unexpected end of input")
}
