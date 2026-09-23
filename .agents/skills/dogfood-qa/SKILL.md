---
name: dogfood-qa
description: QA the honeycomb CLI against a live account to find bugs and feature gaps. Target one command or resource family for manual testing, or sweep everything before a release. Use when dogfooding the CLI, reproducing a bug, or producing a findings report mapped to issues.
user-invocable: true
---

# Dogfood QA

Exercise honeycomb CLI commands against a live account to find bugs and feature gaps. Treat the account as disposable: mutate only throwaway resources.

## Scope the Run

State the target before testing: one command (`trigger update`), one resource family (`board` plus `board view`), or a full sweep. Test only what is in scope. A sweep is this same method fanned out across every family; see [references/sweep.md](references/sweep.md).

## Test Matrix

For each command in scope:

- `--help` matches actual behavior
- Happy-path CRUD
- `--format json` and `--format table`
- `--no-interactive`
- Error paths: missing required flags, bad IDs, invalid enums, malformed JSON (`-f -`), nonexistent dataset
- Exit codes
- Flag names consistent with sibling commands
- Pagination

Trace each finding to `cmd/<resource>/` or `internal/` and cite `file:line`. File cosmetics too.

## Capture Output

Do not judge exit codes or JSON through a pipe. `cmd | head` reports the pipe's exit status, not the command's, and `2>&1 | tee` merges stderr into stdout. Both fabricate "exits 0 on error" and "corrupt JSON" findings. Capture the streams separately and read `$?` directly:

```sh
honeycomb <cmd> >out 2>err; echo "exit=$?"
jq -e . <out >/dev/null && echo valid || echo invalid
```

## Safety

- Mutate only resources you create. Never touch existing account data.
- Run `create`/`update`/`delete` against a throwaway environment and dataset.
- Use a throwaway profile for `auth login`/`logout`.

[references/setup.md](references/setup.md) covers the sandbox bypass, throwaway-environment provisioning, and teardown.

## Record Findings

Write one markdown file per finding from [assets/finding-template.md](assets/finding-template.md). For a sweep, group findings that share a root cause and build the report; see [references/sweep.md](references/sweep.md).
