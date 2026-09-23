# Setup and Teardown

Needed before any test that mutates resources. Read-only testing of an existing profile needs none of this.

## Let the Binary Bypass the Sandbox

The CLI calls `api.honeycomb.io` over HTTPS. Under the Bash sandbox, a MITM proxy presents a CA the Go TLS stack rejects, so every call fails with `tls: failed to verify certificate: x509: OSStatus -26276`. The main agent can pass `dangerouslyDisableSandbox`, but subagents cannot (the harness returns "Run outside of the sandbox" with no prompt), which blocks parallel testing.

Fix it once: add the binary to `sandbox.excludedCommands` in `.claude/settings.local.json`, then restart Claude Code.

```json
{
  "sandbox": {
    "excludedCommands": ["tmp/honeycomb *", "./tmp/honeycomb *"]
  }
}
```

A matching `permissions.allow` rule (`Bash(tmp/honeycomb:*)`) runs it without a prompt for the main agent and subagents alike. Build first: `go build -o tmp/honeycomb ./cmd/honeycomb`.

## Provision a Throwaway Environment

Mirror `TestMain` in `integration/integration_test.go`. Record the IDs for teardown.

1. Environment: `tmp/honeycomb environment create --team <team> --name qa-scratch --format json`. Keep the `id`.
2. Config key scoped to it: `tmp/honeycomb key create --team <team> --key-type configuration --all-permissions --environment <env-id> --name qa-key --format json`. Keep the `id` and `secret`.
3. Store the key under a dedicated profile: `tmp/honeycomb auth login --profile qa --key-type config --key-secret <secret> --verify`.
4. Dataset: `tmp/honeycomb dataset create --profile qa --name qa-ds --format json`.

Config-API commands then use `--profile qa`. Management commands (`environment`, `key`) need the management key in the `default` profile plus `--team`.

## Teardown

```sh
# dataset (disable protection first)
tmp/honeycomb dataset update qa-ds --profile qa --delete-protected=false
tmp/honeycomb dataset delete qa-ds --profile qa --yes
# scoped key
tmp/honeycomb key delete <key-id> --team <team> --yes
# environment (disable protection first)
tmp/honeycomb environment update <env-id> --team <team> --delete-protected=false
tmp/honeycomb environment delete <env-id> --team <team> --yes
# profile
tmp/honeycomb auth logout --profile qa
```

Revert the `sandbox.excludedCommands` change if you added it only for this run.
