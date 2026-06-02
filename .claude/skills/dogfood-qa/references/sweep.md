# Full Sweep

A pre-release pass over every command. Same test matrix as a targeted run (see `SKILL.md`), fanned out and consolidated into a report.

## Fan Out by Resource Family

Run a `Workflow` with one subagent per family:

- auth + api
- board + board view
- column + calculated
- dataset + definition
- environment
- key
- marker + setting
- query + annotation
- recipient
- trigger
- mcp

`slo` is plan-gated; skip it unless the account has SLO access.

Give each agent the account facts, the test matrix, the safety rules, and `assets/finding-template.md`. Each writes one finding per file to `tmp/qa/findings/` and returns a structured summary. Subagents cannot disable the sandbox, so the binary must already be in `excludedCommands` (see [setup.md](setup.md)).

## Consolidate

Many findings share one root cause across resources, such as a missing table border from a single `internal/output` helper. Group them into clusters so the review maps to a handful of issues, not one per occurrence.

## Build the Report

```sh
bun scripts/report.ts   # reads tmp/qa/findings/, writes report.html + findings-index.md
```

`report.html` is an offline, filterable dashboard. `findings-index.md` is a parseable table plus the proposed cluster-to-issue mapping. Edit the cluster map at the top of `report.ts` for the current run.

## Map to Integration Tests

Confirmed bugs and shared-code refactors make good integration-test targets. The existing suite covers happy-path CRUD in JSON; error paths, `--format table`, and flag-based updates are the usual gaps. See `integration/integration_test.go` for the managed-mode harness and its plan-gate skip guards.
