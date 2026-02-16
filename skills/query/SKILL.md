---
name: query
description: >-
  Design and build Honeycomb queries: calculations, filters, breakdowns,
  visualization selection, granularity tuning, and anti-patterns. Activated
  by explicit mention of Honeycomb queries or /query invocation.
user-invocable: true
---

# Honeycomb Query Design

Design effective Honeycomb queries for dashboards, investigations, and saved queries. Covers query specification, visualization selection, SLI-oriented patterns, and common anti-patterns.

## Query Specification

### Structure

```json
{
  "calculations": [{"op": "COUNT"}, {"op": "P99", "column": "duration_ms"}],
  "filters": [{"column": "service.name", "op": "=", "value": "api-gateway"}],
  "filter_combination": "AND",
  "breakdowns": ["http.route"],
  "orders": [{"column": "http.route", "order": "ascending"}],
  "limit": 20,
  "time_range": 7200,
  "granularity": 60
}
```

### Calculation operators

`COUNT`, `SUM`, `AVG`, `MAX`, `MIN`, `COUNT_DISTINCT`, `HEATMAP`, `CONCURRENCY`, `P001`, `P01`, `P05`, `P10`, `P20`, `P25`, `P50`, `P75`, `P80`, `P90`, `P95`, `P99`, `P999`, `RATE_AVG`, `RATE_SUM`, `RATE_MAX`

### Filter operators

`=`, `!=`, `>`, `>=`, `<`, `<=`, `starts-with`, `does-not-start-with`, `ends-with`, `does-not-end-with`, `exists`, `does-not-exist`, `contains`, `does-not-contain`, `in`, `not-in`

### Time range

- `time_range`: Relative window in seconds (default: 7200 = 2 hours)
- `start_time` / `end_time`: Absolute UNIX timestamps
- Combine one timestamp with `time_range`, but not all three

### Granularity

Seconds per time bucket. Valid range: `time_range / 1000` to `time_range`.

| Time range | Recommended granularity | Buckets |
|-----------|------------------------|---------|
| 10 minutes (600) | 10-30s | 20-60 |
| 2 hours (7200) | 60s | 120 |
| 24 hours (86400) | 300-600s | 144-288 |
| 7 days (604800) | 1800-3600s | 168-336 |
| 28 days (2419200) | 3600-14400s | 168-672 |

## Visualization Selection

Choose chart types based on the data pattern, not the Honeycomb default (line graph).

### Chart types

| Value | Name | Best for |
|-------|------|----------|
| `line` | Line graph | Continuous metrics with high cardinality over time (CPU, memory) |
| `tsbar` | Time series bar | Time-bucketed counts and sparse data (errors, deploys, events) |
| `stacked` | Stacked area | Showing composition/proportion over time (errors by type, traffic by service) |
| `stat` | Stat card | Single headline number with sparkline (current p99, error rate) |
| `cbar` | Categorical bar | Comparing values across groups, non-time-series (latency by endpoint) |
| `cpie` | Categorical pie | Proportional breakdown of a total (traffic share by region) |

### `chart_index`

When a query has multiple calculations, `chart_index` maps to the 0-based index in the `calculations` array. Set chart type per calculation:

```json
{
  "charts": [
    {"chart_index": 0, "chart_type": "tsbar"},
    {"chart_index": 1, "chart_type": "line"}
  ]
}
```

## Visualization Recommendations

### Latency

- **Calculations**: Use percentiles, never AVG. AVG is skewed by outliers and doesn't represent typical user experience. Use `P50` as the baseline (median) and `P99` for worst-case.
- **Chart type**: `line` with `overlaid_charts: true` for P50/P99 overlay. The gap between P50 and P99 is the signal -- a widening gap indicates tail latency problems.
- **Units**: Always milliseconds. Note conversion in the panel title if the column uses other units.
- **Title pattern**: "Latency: P50 and P99 (ms)"
- **Granularity**: Match to request volume. Low-traffic services need larger buckets (300s+) to avoid noisy graphs.
- **Heatmap complement**: Add a full-width `HEATMAP` panel below the percentile chart for distribution visibility. Heatmaps reveal bimodal distributions and outlier clusters that percentile lines hide.

### Error rate

- **Calculation**: `AVG` on a boolean error column. In Honeycomb, `AVG(bool)` computes the proportion of `true` values -- this IS the error rate (0.0-1.0).
- **Chart type**: `line` -- shows trend clearly without visual clutter
- **Breakdowns**: Break down by `service.name` or `http.route` to answer "where are errors coming from?" Do NOT break down by `http.status_code` -- individual status codes are too granular and produce noisy, unactionable charts.
- **Granularity**: Use 300s (5min) for per-service breakdowns to smooth out noise from sparse buckets. At 2-min granularity, a single error in a low-volume bucket reads as 100% error rate.
- **Title pattern**: "Error Rate" (aggregate) or "Error Rate by Service"

### Throughput / request volume

- **Calculations**: `COUNT`
- **Chart type**: `stacked` bar when broken down by service or route -- clearly shows both total volume and composition. Use `tsbar` only for total count without breakdown.
- **Breakdowns**: `service.name`, `http.route`, `rpc.method`
- **Granularity**: Produce meaningful bars. For a 2-hour window, 60-120s. For 24 hours, 300-600s.
- **Title pattern**: "Request Volume by Service"

### Cardinality / unique values

- **Calculations**: `COUNT_DISTINCT` on the column of interest
- **Chart type**: `cbar` for comparing across groups, `stat` for a single headline number
- **Use case**: Unique users, distinct trace IDs, unique error messages

### Distribution / heatmap

- **Calculations**: `HEATMAP` on a numeric column
- **Chart type**: Leave as default (heatmap renders natively)
- **Use case**: Latency distribution, request size distribution. Especially valuable for spotting bimodal patterns (e.g., cache hit vs miss) that aggregates hide.

### Rate of change

- **Calculations**: `RATE_SUM`, `RATE_AVG`, or `RATE_MAX`
- **Chart type**: `line` (shows trend direction clearly)
- **Use case**: Detecting acceleration/deceleration in metrics

### Sparse data

When data arrives infrequently (webhooks, batch jobs, rare errors):

- **Always use `tsbar`** instead of line charts. Line graphs interpolate between points, creating misleading visual continuity.
- Set `omit_missing_values: true` to avoid flat zero lines between events.
- Use wider granularity to aggregate sparse events into visible bars.

## Anti-Patterns

- **AVG for latency**: Hides outliers. Always use percentiles (P50/P99).
- **Stacked bar for latency**: Stacking percentiles is meaningless -- percentiles are not additive.
- **Status code breakdowns**: Too many series, not actionable. Break down by service or route instead.
- **Line charts on sparse data**: Interpolation creates false continuity. Use `tsbar`.
- **Too many breakdowns**: More than 5-7 series on one chart becomes unreadable. Use `limit` in the query or aggregate differently.

## SLI-Oriented Query Design

Structure queries around Service Level Indicators. An SLI is a ratio: good events / total events. Keep to five or fewer per service.

| SLI | Honeycomb calculation | Visualization |
|-----|----------------------|---------------|
| Availability | `AVG(error)` inverted (1 - error rate) | `line` |
| Latency | `P50`/`P99` on `duration_ms` | `line` overlaid |
| Error rate | `AVG(error)` on boolean column | `line` |
| Throughput | `COUNT` | `stacked` by service |
| Freshness | `MAX(age_seconds)` or custom | `line` |

For SLO-aware dashboards, use `compare_time_offset_seconds` (e.g., `86400`) to overlay current error rate against the previous day. This shows burn trajectory without requiring SLO-specific API access.

## Running Queries

### Via MCP server (all plans)

Use the Honeycomb MCP server to create and test queries. This works on all plans and doesn't require Enterprise API permissions.

```
honeycomb mcp call run_query -f dataset=<dataset> -f query_json='<json>'
```

### Via API (requires key permissions)

```
honeycomb api -X POST /1/queries/<dataset> --input <query-file>.json
```

Note: piping JSON via stdin can fail if string values contain special characters (e.g., `!=` in filter operators). Use `--input` with a file instead.

## Saved Queries (Query Annotations)

A "saved query" in the Honeycomb UI is a query annotation that references a query ID.

```
jq -n '{name: "Panel Title", query_id: "<query-id>"}' | honeycomb query create --dataset <dataset> --file -
```

## Reference

- [Query examples](references/examples.md)
- [API types](references/api-types.md)
