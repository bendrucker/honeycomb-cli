package agent

import "os"

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

func Detect() *Agent {
	for _, c := range checks {
		if _, ok := os.LookupEnv(c.env); ok {
			return &Agent{Name: c.name}
		}
	}
	return nil
}
