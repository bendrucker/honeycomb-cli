---
name: honeycomb
description: Honeycomb CLI and MCP server usage — auth, queries, datasets, SLOs, boards, triggers, markers, columns, and raw API requests for observability workflows
---

# Honeycomb CLI

## Auth

Three key types, stored in the OS keyring per profile:

| Type | Used For |
|------|----------|
| `config` | Management API (boards, SLOs, triggers, columns, queries) |
| `ingest` | Sending events (`X-Honeycomb-Team` header) |
| `management` | Same as config, kept separate for granular access |

```bash
honeycomb auth login          # interactive key setup
honeycomb auth status         # verify stored key
honeycomb auth status --offline  # check keyring without API call
```

## Query

Requires a `config` key with query permissions.

```bash
# from a file
honeycomb query run --dataset my-dataset --file query.json

# from stdin
echo '{"calculations":[{"op":"COUNT"}],"time_range":3600}' | honeycomb query run --dataset my-dataset --file -

# list/view saved queries
honeycomb query list --dataset my-dataset
honeycomb query view <id> --dataset my-dataset
```

## Raw API Requests

`honeycomb api` makes authenticated requests to any Honeycomb API endpoint. Useful for endpoints without dedicated subcommands (events, query results, service maps).

```bash
honeycomb api /1/events/my-dataset -X POST --input events.json --key-type ingest
honeycomb api /1/query_results/my-dataset/abc123  # poll query results
honeycomb api /1/maps/dependencies/requests -q '.results'
honeycomb api /2/teams -q '.[].name' --paginate   # v2 with jq + pagination
```

Flags: `-X` method, `-f` string fields, `-F` typed fields, `-H` headers, `-q` jq filter, `--paginate`, `--input` body file, `--key-type` override, `--raw` skip JSON:API unwrapping (v2 paths).

## Global Flags

| Flag | Purpose |
|------|---------|
| `--format json\|table` | Output format (default: `table` in TTY, `json` in CI/agent) |
| `--profile` | Configuration profile |
| `--no-interactive` | Disable prompts |
| `--api-url` | Override Honeycomb API URL |

## Patterns

- **CRUD convention**: most resources support `list`, `get`, `create`, `update`, `delete` subcommands
- **Scoping**: use `--dataset` for dataset-scoped resources (columns, markers, queries), `--environment` where applicable
- **Agent auto-detection**: when `CLAUDE_CODE` is set, the CLI forces `--no-interactive` and defaults to `--format json`
- **Commands**: `auth`, `query`, `dataset`, `board`, `column`, `marker`, `slo`, `trigger`, `environment`, `key`, `recipient`, `mcp`, `api`

## MCP Server vs CLI `mcp` Subcommand

This plugin configures the Honeycomb MCP server (`.mcp.json`) for direct tool access via OAuth. Use its tools (`run_query`, `get_dataset_columns`, `find_columns`) when you need query results in context.

The CLI's `mcp` subcommand is a separate MCP *client* that authenticates with a config key (Bearer token). Use it when you need to write MCP results to disk:

```bash
honeycomb mcp call run_query -f dataset=prod -q '.content[].text' > results.json
honeycomb mcp tools  # list available MCP tools
```
