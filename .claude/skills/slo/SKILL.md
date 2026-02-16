---
name: slo
description: Working with Honeycomb SLOs, burn alerts, and SLO history
---

# SLO Commands

Dependency chain: derived column (SLI) → SLO → burn alert. All commands require `--dataset`.

`target_per_million`: 99.9% = 999000, 99% = 990000, 95% = 950000.

## SLO CRUD

- `slo create -f -` (file-only, no flags yet — see #109)
- `slo update <id> --name/--description/--target/--time-period` (flag-based, read-modify-write) or `-f` (mutually exclusive)
- `slo get <id>` / `slo get <id> --detailed` (Enterprise-only)
- `slo list` / `slo delete <id> --yes`

Create body: `{"name":"...","sli":{"alias":"<derived-column-alias>"},"time_period_days":30,"target_per_million":999000}`

## Burn Alerts

Create/update require `recipients` (non-empty array of `{"id":"..."}`) and `slo` as `{"id":"..."}` (nested, not flat `slo_id`).

Exhaustion time: `{"alert_type":"exhaustion_time","exhaustion_minutes":240,"slo":{"id":"..."},"recipients":[{"id":"..."}]}`

Budget rate: `{"alert_type":"budget_rate","budget_rate_window_minutes":60,"budget_rate_decrease_threshold_per_million":50000,"slo":{"id":"..."},"recipients":[{"id":"..."}]}`

All burn alert commands are file-only (`-f`), no flags yet (#110, #111).

## SLO History

`slo history --slo-id <id> --start-time <unix> --end-time <unix>` — `--slo-id` is repeatable. Response is a map keyed by SLO ID.
