#!/bin/sh

input=$(cat)
stop_hook_active=$(echo "$input" | jq -r '.stop_hook_active // empty')
[ "$stop_hook_active" = "true" ] && exit 0

export GOLANGCI_LINT_CACHE="${GOLANGCI_LINT_CACHE:-/tmp/claude/golangci-lint}"

gofmt -w .
golangci-lint run --fix ./... 2>/dev/null

remaining=$(golangci-lint run ./... 2>&1 | grep -v '^level=')
if [ -n "$remaining" ]; then
  echo "$remaining" | jq -Rs '{hookSpecificOutput:{hookEventName:"Stop",additionalContext:("unfixed lint issues:\n" + .)}}'
fi
