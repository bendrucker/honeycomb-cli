//go:build integration

package integration

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestAuth(t *testing.T) {
	const profile = "integration-test-auth"

	t.Cleanup(func() {
		args := authArgs(profile, "auth", "logout")
		cmd := exec.Command(binary, args...)
		_ = cmd.Run()
	})

	t.Run("login", func(t *testing.T) {
		args := authArgs(profile, "auth", "login",
			"--key-type", "config",
			"--key-secret", configKeySecret,
			"--no-verify",
		)
		cmd := exec.Command(binary, args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("auth login failed: %v\nstderr: %s", err, stderr.String())
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
		args := authArgs(profile, "auth", "status")
		cmd := exec.Command(binary, args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err == nil {
			t.Errorf("expected auth status to fail after logout, got success\nstdout: %s", stdout.String())
		}
	})
}

func authArgs(profile string, args ...string) []string {
	flags := []string{"--no-interactive", "--profile", profile, "--format", "json"}
	if apiURL != "" {
		flags = append(flags, "--api-url", apiURL)
	}
	return append(args, flags...)
}

func authRun(t *testing.T, profile string, args ...string) result {
	t.Helper()
	all := authArgs(profile, args...)
	cmd := exec.Command(binary, all...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v\nargs: %v\nstderr: %s", err, args, stderr.String())
	}
	return result{stdout: stdout.Bytes(), stderr: stderr.Bytes()}
}
