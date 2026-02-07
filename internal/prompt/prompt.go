package prompt

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"golang.org/x/term"
)

func Line(out io.Writer, in io.Reader, prompt string) (string, error) {
	_, _ = fmt.Fprint(out, prompt)
	return ReadLine(in)
}

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
