//go:build integration

package integration

import (
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
)

func TestAuth(t *testing.T) {
	const profile = "integration-test-auth"

	t.Cleanup(func() {
		execAuthCmd(t, profile, nil, "auth", "logout")
	})

	t.Run("login", func(t *testing.T) {
		_, err := execAuthCmd(t, profile, nil,
			"auth", "login",
			"--key-type", "config",
			"--key-secret", configKeySecret,
			"--no-verify",
		)
		if err != nil {
			t.Fatalf("auth login failed: %v", err)
		}
	})

	t.Run("status", func(t *testing.T) {
		r := authRun(t, profile, "auth", "status")
		statuses := parseJSON[[]map[string]any](t, r.stdout)
		if len(statuses) == 0 {
			t.Fatal("expected at least one key status")
		}
		found := false
		for _, s := range statuses {
			if s["type"] == "config" {
				found = true
				if s["status"] != "valid" {
					t.Errorf("expected status %q, got %q", "valid", s["status"])
				}
			}
		}
		if !found {
			t.Errorf("config key not found in status output")
		}
	})

	t.Run("status offline", func(t *testing.T) {
		r := authRun(t, profile, "auth", "status", "--offline")
		statuses := parseJSON[[]map[string]any](t, r.stdout)
		if len(statuses) == 0 {
			t.Fatal("expected at least one key status")
		}
		found := false
		for _, s := range statuses {
			if s["type"] == "config" {
				found = true
				if s["status"] != "stored" {
					t.Errorf("expected status %q, got %q", "stored", s["status"])
				}
			}
		}
		if !found {
			t.Errorf("config key not found in status output")
		}
	})

	t.Run("logout", func(t *testing.T) {
		r := authRun(t, profile, "auth", "logout")
		deleted := parseJSON[[]map[string]any](t, r.stdout)
		if len(deleted) == 0 {
			t.Errorf("expected at least one deleted key")
		}
	})

	t.Run("status after logout", func(t *testing.T) {
		_, err := execAuthCmd(t, profile, nil, "auth", "status")
		if err == nil {
			t.Errorf("expected auth status to fail after logout")
		}
	})
}

func execAuthCmd(tb testing.TB, profile string, stdin []byte, args ...string) (result, error) {
	flags := []string{"--no-interactive", "--profile", profile, "--format", "json"}
	if apiURL != "" {
		flags = append(flags, "--api-url", apiURL)
	}
	allArgs := append(args, flags...)

	ts := iostreams.Test(tb)
	if stdin != nil {
		ts.InBuf.Write(stdin)
	}

	rootCmd := cmd.NewRootCmd(ts.IOStreams)
	rootCmd.SetArgs(allArgs)

	err := rootCmd.Execute()
	return result{stdout: ts.OutBuf.Bytes(), stderr: ts.ErrBuf.Bytes()}, err
}

func authRun(t *testing.T, profile string, args ...string) result {
	t.Helper()
	r, err := execAuthCmd(t, profile, nil, args...)
	if err != nil {
		t.Fatalf("command failed: %v\nargs: %v\nstderr: %s", err, args, r.stderr)
	}
	return r
}
