#!/bin/sh
set -e

file_path=$(echo "$TOOL_INPUT" | jq -r '.file_path // empty')
[ -z "$file_path" ] && exit 0

case "$file_path" in
  *.gen.go) exit 0 ;;
  *.go) golangci-lint run --fix "$(dirname "$file_path")/..." ;;
esac
