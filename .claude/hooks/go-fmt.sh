#!/bin/sh
set -e

file_path=$(echo "$TOOL_INPUT" | jq -r '.file_path // empty')
[ -z "$file_path" ] && exit 0

case "$file_path" in
  *.go) gofmt -w "$file_path" ;;
esac
