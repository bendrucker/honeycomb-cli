# Honeycomb CLI

CLI for [Honeycomb](https://www.honeycomb.io/), modeled after the GitHub CLI (`gh`).

## Commands

```
go build -o /dev/null .
go test ./...
go vet ./...
go generate ./internal/api/...
golangci-lint run ./...
```

## Project Structure

```
main.go                          Entry point
cmd/
  root.go                        Root cobra command, global flags
  options/options.go             RootOptions shared across command packages
  auth/
    auth.go                      auth parent command
    login.go                     auth login command
    status.go                    auth status command with key verification
internal/
  api/
    generate.go                  go:generate directives for oapi-codegen
    client.gen.go                Generated API client and types (do not edit)
  iostreams/iostreams.go         IO abstraction with TTY detection
  agent/agent.go                 AI coding agent detection
  config/
    config.go                    JSON config (~/.config/honeycomb/config.json)
    keyring.go                   OS keyring with timeout wrapper
api.json                         Honeycomb OpenAPI 3.1 source spec
overlay.yaml                     OpenAPI overlay for 3.1→3.0 compatibility
oapi-codegen.yaml                Code generation config
```

## Code Generation

The API client is generated from `api.json` (Honeycomb's OpenAPI 3.1 spec) using `oapi-codegen`. Since `oapi-codegen` doesn't support OpenAPI 3.1, an [overlay](https://github.com/OAI/Overlay-Specification) converts 3.1 patterns to 3.0 equivalents at generation time:

- `type: ['null', T]` → `type: T` + `nullable: true`
- `type: [T]` → `type: T`
- `type: [T1, T2, T3]` → remove type (any)
- `oneOf`/`anyOf` with `{type: null}` → `nullable: true` on parent

Run `go generate ./internal/api/...` to regenerate. The generated file is committed; the overlay is applied automatically via the oapi-codegen config.

## Go Conventions

- **Go 1.25** — use `go tool` for codegen dependencies, range-over-func, etc.
- **Error handling** — return errors, don't panic. Wrap with `fmt.Errorf("context: %w", err)`.
- **Naming** — `NewXxx` constructors, unexported fields, `opts` for option structs.
- **Testing** — table-driven tests with `t.Run`. Use `t.Setenv` for env vars. No testify.
- **No Viper** — config is parsed with `encoding/json` directly.

## Dependencies

| Package | Purpose |
|---------|---------|
| `spf13/cobra` | CLI framework |
| `encoding/json` | Config parsing (stdlib) |
| `zalando/go-keyring` | OS keyring for secrets |
| `mattn/go-isatty` | TTY detection |
| `oapi-codegen/runtime` | Generated client runtime |
| `oapi-codegen/nullable` | Three-state nullable for OpenAPI |
| `charmbracelet/huh` | Terminal forms and prompts |

## Authentication

Three key types, stored in the OS keyring keyed by `{profile}:{type}`:

| Type | Header | Used For |
|------|--------|----------|
| `config` | `X-Honeycomb-Team` | Configuration API (boards, SLOs, triggers, columns, queries, etc.) |
| `ingest` | `X-Honeycomb-Team` | Sending events |
| `management` | `Authorization: Bearer KEY_ID:KEY_SECRET` | Management API v2 (environments, keys) |

V2 API keys created via `key create` produce an `id` + `secret`. The `secret` alone is used as the `X-Honeycomb-Team` value for config/ingest keys. The `auth login` command combines `--key-id` and `--key-secret` into `id:secret` format, which does not work for v1 API authentication. For v2-created keys, store the secret directly via the OS keyring (`security add-generic-password`).

## Interactive vs Non-Interactive

Every command must work in both modes:

- **Interactive** (TTY): prompt for missing inputs, show rich output, use color
- **Non-interactive** (piped/CI/agent): require all inputs as flags, output structured data (JSON)

Agent detection (`internal/agent`) auto-enables non-interactive mode and defaults to JSON output. The `--no-interactive` flag provides manual control.

## Agent Detection

When an AI coding agent is detected via environment variables (`CLAUDE_CODE`, `CURSOR_SESSION_ID`, `CODEX`, `GITHUB_COPILOT`, `WINDSURF_SESSION_ID`, `CLINE`), the CLI:
- Forces non-interactive mode
- Defaults output format to JSON

## Honeycomb MCP Server

A Honeycomb MCP server is configured and available as a reference implementation. Use its tools (`run_query`, `get_dataset_columns`, `find_columns`) to validate CLI behavior against real API responses.

## Output Formats

The `--format` flag supports `json` and `table`. Default is `table` in TTY, `json` otherwise.

## Command Design

Commands follow a consistent pattern:
- `New*Cmd` factory function, no `init()`
- Accept an `*iostreams.IOStreams` (or parent options struct)
- Register with parent via `AddCommand`
- Support `--format` for output
- Both interactive and non-interactive paths

## Testing

**Unit tests** use `keyring.MockInit()` for an in-memory keyring and `httptest.NewServer` for API stubs. These run in `go test` with no external dependencies.

**Interactive testing** requires a real OS keyring and a live Honeycomb API key. Build the binary and store a key manually:

```
go build -o tmp/honeycomb .
security add-generic-password -s honeycomb-cli -a default:config -w '<KEY_SECRET>'
tmp/honeycomb auth status
tmp/honeycomb auth status --offline
tmp/honeycomb auth status --format json
```

Remove the key afterward with `security delete-generic-password -s honeycomb-cli -a default:config`.

V2-created keys (from `key create`) should have their `secret` stored directly as shown above. Do not use `auth login` for v2 keys -- it combines `--key-id` and `--key-secret` into `id:secret` format, which the v1 Configuration API does not accept.

## TUI Ideas (Future)

- **Query results table** — run a query, display results in a rich table with sorting/filtering
- **Board browser** — list boards, preview panels, open in browser
- **Trace viewer** — navigate spans in a trace tree
- **SLO dashboard** — live burn rate, budget remaining, recent breaches
- **Dataset explorer** — browse columns, see types and descriptions
- **Trigger status** — live view of trigger states and recent firings

## MCP Client (Future)

The CLI will include an `mcp` subcommand that acts as an MCP client to the Honeycomb MCP server. This provides access to features like query execution without requiring Enterprise-tier API key permissions (the Query Data API's "Run Queries" permission is Enterprise-only). The `query` command would still use the API directly and require the appropriate key permissions. The `mcp` command offers an alternative path using Honeycomb's own MCP server, which is available on all plans.

## Planned Command Hierarchy

```
honeycomb auth login/logout/status
honeycomb query run/list/get
honeycomb dataset list/get/create
honeycomb board list/get/create/delete
honeycomb slo list/get/create/delete
honeycomb trigger list/get/create/delete
honeycomb marker create/list
honeycomb column list/get
honeycomb mcp query/...                # MCP client, works on all plans
honeycomb api <method> <path>          # arbitrary API escape hatch
```
