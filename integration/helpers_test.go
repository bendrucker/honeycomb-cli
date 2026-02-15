//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type result struct {
	stdout []byte
	stderr []byte
}

func run(t *testing.T, stdin []byte, args ...string) result {
	t.Helper()
	r, err := execBinary(stdin, args...)
	if err != nil {
		t.Fatalf("command failed: %v\nargs: %v\nstderr: %s", err, args, r.stderr)
	}
	return r
}

func runErr(stdin []byte, args ...string) (result, error) {
	return execBinary(stdin, args...)
}

func parseJSON[T any](t *testing.T, data []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("parsing JSON: %v\ndata: %s", err, data)
	}
	return v
}

func uniqueName(t *testing.T) string {
	t.Helper()
	name := t.Name()
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ToLower(name)
	if len(name) > 30 {
		name = name[:30]
	}
	return runID + "-" + name
}

func skipWithoutEnterprise(t *testing.T) {
	t.Helper()
	if !hasEnterprise {
		t.Skip("enterprise features not available")
	}
}

func skipWithoutPro(t *testing.T) {
	t.Helper()
	if !hasPro {
		t.Skip("pro features not available")
	}
}

func writeTemp(t *testing.T, data []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "integration-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.Write(data); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing temp file: %v", err)
	}
	return f.Name()
}
