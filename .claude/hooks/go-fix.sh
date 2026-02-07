#!/bin/sh

input=$(cat)
stop_hook_active=$(echo "$input" | jq -r '.stop_hook_active // empty')
[ "$stop_hook_active" = "true" ] && exit 0

gofmt -w . 2>/dev/null || true
golangci-lint run --fix ./... 2>/dev/null || true
