package api

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func parseFields(raw, typed []string, stdin io.Reader) (map[string]any, error) {
	result := make(map[string]any)

	for _, f := range raw {
		key, val, ok := strings.Cut(f, "=")
		if !ok {
			return nil, fmt.Errorf("invalid field %q (must be key=value)", f)
		}
		setField(result, key, val)
	}

	for _, f := range typed {
		key, val, ok := strings.Cut(f, "=")
		if !ok {
			return nil, fmt.Errorf("invalid field %q (must be key=value)", f)
		}
		coerced, err := coerceValue(val, stdin)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", key, err)
		}
		setField(result, key, coerced)
	}

	return result, nil
}

func setField(m map[string]any, key string, val any) {
	bracket := strings.IndexByte(key, '[')
	if bracket < 0 {
		m[key] = val
		return
	}

	name := key[:bracket]
	rest := key[bracket:]

	close := strings.IndexByte(rest, ']')
	if close < 0 {
		m[key] = val
		return
	}

	inner := rest[1:close]

	if inner == "" {
		arr, _ := m[name].([]any)
		m[name] = append(arr, val)
		return
	}

	nested, _ := m[name].(map[string]any)
	if nested == nil {
		nested = make(map[string]any)
		m[name] = nested
	}

	suffix := rest[close+1:]
	setField(nested, inner+suffix, val)
}

func coerceValue(s string, stdin io.Reader) (any, error) {
	switch s {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "null":
		return nil, nil
	}

	if strings.HasPrefix(s, "@") {
		return readFileValue(s[1:], stdin)
	}

	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i, nil
	}
	if n, err := strconv.ParseFloat(s, 64); err == nil {
		return n, nil
	}

	return s, nil
}

func readFileValue(path string, stdin io.Reader) (string, error) {
	var r io.Reader
	if path == "-" {
		r = stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("reading @%s: %w", path, err)
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("reading @%s: %w", path, err)
	}
	return string(b), nil
}
