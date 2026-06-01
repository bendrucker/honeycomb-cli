# Contributing

## Development

Build, test, and lint:

```
go build -o /dev/null ./cmd/honeycomb
go test ./...
go vet ./...
golangci-lint run ./...
```

## Code Generation

The API client in `internal/api/client.gen.go` is generated from `api.json`,
Honeycomb's OpenAPI 3.1 spec. Because `oapi-codegen` does not support OpenAPI
3.1, `overlay.yaml` converts 3.1 patterns to 3.0 equivalents at generation time.

Regenerate after editing `api.json` or `overlay.yaml`:

```
go generate ./internal/api/...
```

The generated file is committed. Do not edit it by hand.

## Integration Tests

Integration tests live in `integration/` and run against the live Honeycomb API
with the `integration` build tag. Tests run in-process, so no binary build or
subprocess is needed. `HONEYCOMB_TEAM` is always required.

Disable the command sandbox when running integration tests—they make real API
calls and need network access.

### Managed Mode

Requires a management key, stored in the OS keyring under the `default` profile
or supplied via `HONEYCOMB_MANAGEMENT_KEY_ID` and
`HONEYCOMB_MANAGEMENT_KEY_SECRET`. `TestMain` creates a temporary environment,
config key, and dataset, then cleans them up afterward.

```
HONEYCOMB_TEAM=<slug> go test -tags integration -count=1 -v ./integration/
```

### Direct Mode

Set `HONEYCOMB_DATASET` to skip environment and key provisioning and use an
existing config key and dataset. The config key must already be stored in the OS
keyring. Optionally set `HONEYCOMB_PROFILE` (defaults to `default`). Direct mode
skips cleanup and cannot run tests that require provisioning managed resources;
use `-run` to target compatible tests.

```
HONEYCOMB_TEAM=<slug> HONEYCOMB_DATASET=<dataset> \
  go test -tags integration -count=1 -v -run TestSLO ./integration/
```

## Releases

Releases are cut by pushing a `vX.Y.Z` tag. GoReleaser builds the binaries and
generates release notes from commit messages, so there is no hand-maintained
changelog. Use [Conventional Commits](https://www.conventionalcommits.org/)
prefixes (`feat:`, `fix:`, etc.) so changes are grouped correctly.
