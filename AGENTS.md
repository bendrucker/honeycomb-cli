# Honeycomb CLI

CLI for [Honeycomb](https://www.honeycomb.io/), modeled after the GitHub CLI (`gh`). `honeycomb --help` lists the command tree.

## Commands

```
go build -o /dev/null ./cmd/honeycomb
go test ./...
go vet ./...
golangci-lint run ./...
go generate ./internal/api/...
```

`CONTRIBUTING.md` covers code generation, refreshing `api.json`, integration tests, and releases.

## Go Conventions

- Return errors instead of panicking. Wrap with `fmt.Errorf("context: %w", err)`.
- Write table-driven tests with `t.Run` and the stdlib `testing` package. Use `t.Setenv` for env vars.
- Parse config with `encoding/json`.
- Never edit `internal/api/client.gen.go`. Change `api.json` or `overlay.yaml` and regenerate.

## Command Design

Every command works in both modes:

- Interactive (TTY): prompt for missing inputs, show rich output.
- Non-interactive (piped, CI, or a detected agent): require every input as a flag and never prompt.

`internal/agent` detects coding agents from env vars and forces non-interactive mode without changing the output format. `--no-interactive` forces it manually.

Build each command as a `New*Cmd` factory (no `init()`) that takes the parent options struct and registers with `AddCommand`. The `add-command` skill walks through a new one.

## Output

`--format` accepts `json` and `table`. Detail commands use `opts.OutputWriter()`, which defaults to `table` on a TTY and `json` otherwise. List commands use `opts.OutputWriterList()`, which always defaults to `table`.

Derive columns and fields from struct tags on the projection struct: `col:"Header"` for `output.TableFromTags[T]()` and `detail:"Label"` for `output.FieldsFromTags(v)`. A column that needs a join, pointer deref, conditional, or custom number/time formatting gets no tag. Append it as an explicit `output.Column` or `output.Field` after the derived set.

## API Requests

Build the client with `opts.ClientFor(team, kind)`, declaring the command's `options.AuthKind`. It bakes in auth, so call sites pass no request editor. Pass the `--team` flag pointer for `AuthManagement`, which resolves the team first, and `nil` otherwise:

```go
client, err := opts.ClientFor(nil, options.AuthConfig)
if err != nil {
	return err
}
resp, err := client.ListTriggersWithResponse(ctx, dataset)
if err != nil {
	return fmt.Errorf("listing triggers: %w", err)
}
triggers, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
if err != nil {
	return err
}
```

Use `api.Decode` wherever a `JSON200`/`JSON201` field is read. It checks the status and guards the typed-nil case. Responses decoded by hand (SLO union types, the raw `ListColumns` body) and delete commands call `api.CheckResponse` directly.

`cmd/command` holds the CRUD helpers:

- `ConfirmDelete(ios, yes, noun, fallbackName, fetchName)`: handles `--yes`, the non-interactive guard, and the y/N prompt. `fetchName` resolves a display name only when prompting. Pass `nil` to show `fallbackName`. The caller decides what a declined delete means.
- `ReadDefinitionFile(ios, path)`: reads a JSON definition from a file or `-` (stdin) and sanitizes it.
- `ApplyOverrides(data, overrides)`: merges flag overrides into a `map[string]any` body. Commands that build a typed request struct set fields directly.

## Authentication

Keys live in the OS keyring under service `honeycomb-cli`, account `{profile}:{type}`:

| Type | Header | Used For |
|------|--------|----------|
| `config` | `X-Honeycomb-Team` | Configuration API (boards, SLOs, triggers, columns, queries) |
| `ingest` | `X-Honeycomb-Team` | Sending events |
| `management` | `Authorization: Bearer KEY_ID:KEY_SECRET` | Management API v2 (environments, keys) |

A v2 key from `key create` returns an `id` and a `secret`. Only the `secret` works as a config or ingest key. `auth login --key-id --key-secret` stores `id:secret`, which the v1 Configuration API rejects, so store a v2 secret directly:

```
security add-generic-password -s honeycomb-cli -a default:config -w '<KEY_SECRET>'
```

## Testing

Unit tests use `keyring.MockInit()` for an in-memory keyring and `httptest.NewServer` for API stubs.

To exercise the binary by hand, store a key as above, then build to `tmp/`:

```
go build -o tmp/honeycomb ./cmd/honeycomb
tmp/honeycomb auth status
```

Remove the key afterward with `security delete-generic-password -s honeycomb-cli -a default:config`.

Integration tests (`integration/`, `-tags integration`) hit the live API. Run them with the sandbox disabled and target the command under change with `-run`. Setup modes and env vars are in `CONTRIBUTING.md`.

The Honeycomb MCP server in `.mcp.json` is a reference implementation. Compare CLI output against its `run_query`, `get_dataset_columns`, and `find_columns` results.
