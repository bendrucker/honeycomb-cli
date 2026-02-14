# Honeycomb CLI

CLI for [Honeycomb](https://www.honeycomb.io/), modeled after the [GitHub CLI](https://cli.github.com/) (`gh`).

## Installation

```
go install github.com/bendrucker/honeycomb-cli@latest
```

## Authentication

```
honeycomb auth login
```

In interactive mode, the login command prompts for key type and credentials. Non-interactively, pass `--key-type` and `--key-secret` (and `--key-id` for management keys) as flags.

Check the status of stored credentials:

```
honeycomb auth status
```

### Key Types

| Type | Header | Used For |
|------|--------|----------|
| `config` | `X-Honeycomb-Team` | Configuration API (boards, SLOs, triggers, columns, queries) |
| `ingest` | `X-Honeycomb-Team` | Sending events |
| `management` | `Authorization: Bearer` | Management API (environments, API keys) |

### Security

API keys are stored in your operating system's keyring (macOS Keychain, GNOME Keyring, Windows Credential Manager) via [`zalando/go-keyring`](https://github.com/zalando/go-keyring). Keys are never written to disk or stored in environment variables. The keyring service name is `honeycomb-cli`, and each key is stored under `{profile}:{type}` (e.g. `default:config`).

All keyring operations have a 3-second timeout to prevent the CLI from hanging if the keyring is locked or unavailable.

## Usage

Commands follow a `honeycomb <resource> <action>` pattern:

```
honeycomb dataset list
honeycomb board get --slug my-board
honeycomb query run --dataset my-dataset --query-json '{"calculations": [{"op": "COUNT"}]}'
honeycomb trigger create --dataset my-dataset --name "Error rate" --threshold 100
```

Run `honeycomb help` or `honeycomb <resource> --help` for full details.

### Available Resources

`api`, `auth`, `board`, `column`, `dataset`, `environment`, `key`, `marker`, `mcp`, `query`, `recipient`, `slo`, `trigger`

### Global Flags

| Flag | Description |
|------|-------------|
| `--profile` | Configuration profile (default: `default`) |
| `--format` | Output format: `json` or `table` |
| `--no-interactive` | Disable interactive prompts |
| `--api-url` | Override the Honeycomb API URL |

### Output Formats

The `--format` flag supports `json` and `table`. Default is `table` in a TTY, `json` otherwise. List commands always default to `table` for compact, scannable output — even in non-TTY or agent contexts.

### Agent Detection

When running inside an AI coding agent (Claude Code, Cursor, Codex, GitHub Copilot, Windsurf, Cline), the CLI automatically disables interactive prompts.

## Development

```
go build -o /dev/null .
go test ./...
go vet ./...
golangci-lint run ./...
```

The API client is generated from `api.json` (Honeycomb's OpenAPI spec) using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen):

```
go generate ./internal/api/...
```
