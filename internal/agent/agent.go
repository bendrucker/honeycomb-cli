// Package agent detects which AI coding agent, if any, the CLI is running
// under.
package agent

import "os"

// Agent identifies the AI coding agent running the current process.
type Agent struct {
	Name string
}

var checks = []struct {
	env  string
	name string
}{
	{"CLAUDE_CODE", "claude-code"},
	{"CURSOR_SESSION_ID", "cursor"},
	{"CODEX", "codex"},
	{"GITHUB_COPILOT", "github-copilot"},
	{"WINDSURF_SESSION_ID", "windsurf"},
	{"CLINE", "cline"},
}

// Detect returns the running AI coding agent, or nil if none is detected.
func Detect() *Agent {
	for _, c := range checks {
		if _, ok := os.LookupEnv(c.env); ok {
			return &Agent{Name: c.name}
		}
	}
	return nil
}
