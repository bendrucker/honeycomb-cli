---
title: <concise title>
family: <resource family key>
severity: crash | bug | gap | cosmetic
mode: interactive | non-interactive | both | n/a
format: json | table | both | n/a
status: needs-triage
commands:
  - <verbatim repro command>
---

## Summary
One to three sentences: what is wrong and why it matters.

## Reproduce
```
<exact commands and relevant output>
```

## Expected
What should happen.

## Actual
What happened.

## Root cause
file:line and a brief explanation.

## Suggested fix
Concise proposal.

<!--
Severity guide:
  crash    panic, hang, or data loss
  bug      wrong behavior, output, exit code, or invalid JSON
  gap      missing capability or poor UX that blocks a real workflow
  cosmetic wording, spacing, header-only tables, minor inconsistency

Filename: tmp/qa/findings/<family>-<short-slug>.md (kebab-case, unique).
Link related findings with [[other-slug]].
-->
