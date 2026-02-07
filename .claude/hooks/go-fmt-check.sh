#!/bin/sh

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')
[ -z "$file_path" ] && exit 0

case "$file_path" in
  *.go) ;;
  *) exit 0 ;;
esac

diff=$(gofmt -d "$file_path")
[ -z "$diff" ] && exit 0

echo "$diff" | jq -Rs '{hookSpecificOutput:{hookEventName:"PostToolUse",additionalContext:("gofmt diff:\n" + .)}}'
