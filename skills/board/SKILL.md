---
name: board
description: >-
  Manage Honeycomb boards: create, update panels, configure visualizations,
  and review in Chrome. Activated by explicit mention of Honeycomb boards
  or /board invocation.
user-invocable: true
---

# Honeycomb Board Management

Manage Honeycomb boards using the CLI. Covers the full lifecycle: creating boards, adding query/text/SLO panels, configuring visualization settings, and reviewing boards visually in Chrome.

## Prerequisites

Verify auth before any board operation:

```
honeycomb auth status
```

A `config` key is required. If not configured, run `honeycomb auth login`.

## Board Lifecycle

### Create a board

```
honeycomb board create --name "Service Health" --description "Key metrics for the API gateway"
```

Save the returned board ID for subsequent operations.

### Get current board state

```
honeycomb board get <board-id> --format json
```

Always fetch the current board state before updating. The update command merges JSON fields, but panels are replaced as a whole array.

### Update a board

Pipe JSON to update. Non-specified top-level fields are preserved (name, description, tags, preset_filters, layout_generation). Panels are replaced entirely when included.

```
jq -n '{panels: [...]}' | honeycomb board update <board-id> --file -
```

To replace the board completely (discarding all existing fields):

```
honeycomb board update <board-id> --file board.json --replace
```

## Adding Query Panels

Adding a query panel requires three steps: create a query, create a query annotation, then add the panel to the board.

### Step 1: Create the query

Use the Honeycomb MCP server to create and test queries when available. This works on all plans and doesn't require Enterprise API permissions.

```
honeycomb mcp call run_query -f dataset=<dataset> -f query_json='<json>'
```

To save a query via the API (requires appropriate key permissions):

```
honeycomb api -X POST /1/queries/<dataset> --input <query-file>.json
```

Note: piping JSON via stdin can fail if string values contain special characters (e.g., `!=` in filter operators). Use `--input` with a file instead.

Extract the query `id` from the response.

### Step 2: Create a query annotation

```
jq -n '{name: "Panel Title", query_id: "<query-id>"}' | honeycomb query create --dataset <dataset> --file -
```

Extract the annotation `id` from the response.

### Step 3: Add the panel to the board

Fetch the current board, append the new panel to the panels array, and update:

```
honeycomb board get <board-id> --format json | \
  jq '.panels += [{"type": "query", "query_panel": {"query_id": "<qid>", "query_annotation_id": "<aid>", "query_style": "graph", "visualization_settings": {"charts": [{"chart_type": "tsbar"}]}}}]' | \
  honeycomb board update <board-id> --file - --replace
```

When adding panels to an existing board, fetch the board JSON first, modify the panels array, and send the full board back with `--replace`. The merge behavior without `--replace` replaces the panels array wholesale if the `panels` key is present.

**Important**: The `dataset` field appears in `board get` output on query panels but is rejected by the update API. Strip it before sending: `jq 'walk(if type == "object" and has("dataset") and has("query_id") then del(.dataset) else . end)'`

## Panel Types

### Query panel

```json
{
  "type": "query",
  "query_panel": {
    "query_id": "<id>",
    "query_annotation_id": "<id>",
    "query_style": "graph",
    "visualization_settings": { ... }
  },
  "position": {"x_coordinate": 0, "y_coordinate": 0, "width": 6, "height": 4}
}
```

`query_style`: `"graph"` (default), `"table"`, `"combo"` (graph + table).

### Text panel

Use for section headers, documentation, and context. Supports Markdown (max 10,000 chars).

```json
{
  "type": "text",
  "text_panel": {"content": "## Request Performance\nKey latency and throughput metrics."},
  "position": {"x_coordinate": 0, "y_coordinate": 0, "width": 12, "height": 1}
}
```

### SLO panel

```json
{
  "type": "slo",
  "slo_panel": {"slo_id": "<id>"},
  "position": {"x_coordinate": 0, "y_coordinate": 0, "width": 6, "height": 4}
}
```

## Visualization Settings

Per-panel settings in `query_panel.visualization_settings`:

```json
{
  "charts": [
    {
      "chart_index": 0,
      "chart_type": "tsbar",
      "log_scale": false,
      "omit_missing_values": false
    }
  ],
  "hide_compare": false,
  "hide_hovers": false,
  "hide_markers": false,
  "overlaid_charts": false,
  "utc_xaxis": false
}
```

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

Choose chart types based on the data pattern, not the Honeycomb default (line graph).

Boards are **launchpads, not wallboards** — each panel should answer a question or launch a deeper inquiry. A panel that shows *what* but not a path to *why* is a dead end. Every query panel is clickable in Honeycomb, so design panels as starting points for investigation, not final answers.

### Latency

- **Calculations**: Use percentiles, never AVG. AVG is skewed by outliers and doesn't represent typical user experience. Use `P50` as the baseline (median) and `P99` for worst-case.
- **Chart type**: `line` with `overlaid_charts: true` for P50/P99 overlay. The gap between P50 and P99 is the signal — a widening gap indicates tail latency problems.
- **Units**: Always milliseconds. Note conversion in the panel title if the column uses other units.
- **Title pattern**: "Latency: P50 and P99 (ms)"
- **Granularity**: Match to request volume. Low-traffic services need larger buckets (300s+) to avoid noisy graphs.
- **Heatmap complement**: Add a full-width `HEATMAP` panel below the percentile chart for distribution visibility. Heatmaps reveal bimodal distributions and outlier clusters that percentile lines hide.

### Error rate

- **Calculation**: `AVG` on a boolean error column. In Honeycomb, `AVG(bool)` computes the proportion of `true` values — this IS the error rate (0.0–1.0).
- **Chart type**: `line` — shows trend clearly without visual clutter
- **Breakdowns**: Break down by `service.name` or `http.route` to answer "where are errors coming from?" Do NOT break down by `http.status_code` — individual status codes are too granular and produce noisy, unactionable charts.
- **Granularity**: Use 300s (5min) for per-service breakdowns to smooth out noise from sparse buckets. At 2-min granularity, a single error in a low-volume bucket reads as 100% error rate.
- **Title pattern**: "Error Rate" (aggregate) or "Error Rate by Service"

### Throughput / request volume

- **Calculations**: `COUNT`
- **Chart type**: `stacked` bar when broken down by service or route — clearly shows both total volume and composition. Use `tsbar` only for total count without breakdown.
- **Breakdowns**: `service.name`, `http.route`, `rpc.method`
- **Granularity**: Produce meaningful bars. For a 2-hour window, 60–120s. For 24 hours, 300–600s.
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

### Anti-patterns

- **AVG for latency**: Hides outliers. Always use percentiles (P50/P99).
- **Stacked bar for latency**: Stacking percentiles is meaningless — percentiles are not additive.
- **Status code breakdowns**: Too many series, not actionable. Break down by service or route instead.
- **Line charts on sparse data**: Interpolation creates false continuity. Use `tsbar`.
- **Too many breakdowns**: More than 5–7 series on one chart becomes unreadable. Use `limit` in the query or aggregate differently.

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

## Board Styling

### Recommended board structure

Organize panels top-to-bottom by importance:

1. **Header text panel** — board title and purpose (full width, height 1)
2. **Key SLIs** — error rate and latency side by side (half width each). These are the first thing someone checks during an incident.
3. **Distribution** — latency heatmap (full width). Shows what percentile lines hide.
4. **Breakdowns** — error rate by service + throughput by service side by side. Answers "where is the problem?"
5. **Deep dives** — top routes, slow queries, etc. (full width). Investigation starting points.

### Panel titles (query annotation names)

- Be descriptive: include what is measured and the key dimension
- Include units when the value isn't obvious: "Latency: P50 and P99 (ms)"
- Include the percentile when relevant: "Latency: P50 and P99 (ms)" not "Request Latency"
- Match the team's existing naming conventions when updating existing boards
- Avoid abbreviations unless universally understood in context

### Text panels for structure

Use text panels as section dividers to organize large boards:

```json
{"type": "text", "text_panel": {"content": "## Downstream Dependencies"}}
```

Add context where metrics need interpretation:

```json
{"type": "text", "text_panel": {"content": "**Note**: Latency spikes >500ms typically correlate with database connection pool exhaustion. Check the connection pool metrics below."}}
```

### Preset filters

Add up to 5 preset filters for cross-board filtering. These create dropdown filters at the top of the board.

```json
{
  "preset_filters": [
    {"column": "service.name", "alias": "Service"},
    {"column": "environment", "alias": "Environment"},
    {"column": "http.route", "alias": "Route"}
  ]
}
```

- `alias` is the display label (max 50 chars)
- Choose columns that meaningfully partition the data across all panels
- Common choices: `service.name`, `environment`, `k8s.namespace.name`, `http.route`

### Tags

Categorize boards with up to 10 key:value tags (lowercase only):

```json
{
  "tags": [
    {"key": "team", "value": "platform"},
    {"key": "service", "value": "api-gateway"}
  ]
}
```

## Panel Layout

### Auto layout (default)

Set `layout_generation: "auto"` and omit position fields. Honeycomb arranges panels automatically.

### Manual grid positioning

The board uses a 12-column grid. Specify `position` on each panel:

```json
{"x_coordinate": 0, "y_coordinate": 0, "width": 6, "height": 4}
```

- **Full width**: `width: 12`
- **Half width (side by side)**: `width: 6`, second panel at `x_coordinate: 6`
- **Third width**: `width: 4`
- **Height**: Typically 3-5 for graphs, 1-2 for text panels, 4-5 for heatmaps

Common layouts:

- **Two-column**: Pairs of `width: 6` panels
- **Dashboard header**: Full-width text panel (`width: 12, height: 1`) followed by stat cards (`width: 3, height: 3`)
- **Detail section**: Full-width graph (`width: 12, height: 5`) for deep-dive queries

When adding panels to an existing board with manual positioning, fetch the current board to understand the y_coordinate of the last panel, then place new panels below.

## Board Views

Board views are saved filter configurations that provide different perspectives on the same board.

### Create a view

```
honeycomb board view create --board <board-id> --name "Production" --filter "environment:=:production"
```

Multiple filters:

```
honeycomb board view create --board <board-id> --name "API Errors" \
  --filter "service.name:=:api-gateway" \
  --filter "http.status_code:>=:500"
```

Filter format: `column:operation:value` (value optional for `exists`/`does-not-exist`).

### List views

```
honeycomb board view list --board <board-id>
```

### Common view patterns

- **By environment**: "Production", "Staging", "Development"
- **By service**: One view per key service
- **By severity**: "Errors Only" (status >= 500), "Slow Requests" (duration > threshold)

## Chrome Visual Review

After creating or updating a board, open it in Chrome for visual assessment.

### Workflow

1. Get the board URL:
   ```
   honeycomb board get <board-id> --format json | jq -r '.links.board_url'
   ```

2. Open the board in Chrome using `mcp__claude-in-chrome__navigate`

3. Ask the user to confirm they are logged into Honeycomb and can see the board

4. Use `mcp__claude-in-chrome__read_page` to capture the board state

5. Assess the board for:
   - Empty panels (no data in the time range)
   - Line charts on sparse data (should be bar charts)
   - Missing breakdowns that would add context
   - Panels that are too small to read
   - Inconsistent naming across panel titles
   - Missing section headers for logical groupings

6. Report findings and suggest specific changes using the update workflow above

### Active review mode

When reviewing automatically, check:

- Do all panels have data in the current time range?
- Are line charts appropriate for the data density?
- Are panel titles descriptive and consistent?
- Would preset filters improve the board's usability?
- Are related panels grouped with text panel section headers?

### User-driven review mode

When the user is reviewing:

1. Open the board in Chrome
2. Ask what looks wrong or could be improved
3. Apply requested changes via the CLI
4. Refresh the board and confirm the changes look correct

## Removing Panels

Fetch the board, filter out the panel by index or query_annotation_id, and update:

```
honeycomb board get <board-id> --format json | \
  jq 'del(.panels[2])' | \
  honeycomb board update <board-id> --file - --replace
```

## Reordering Panels

Fetch the board and rearrange the panels array or update position coordinates:

```
honeycomb board get <board-id> --format json | \
  jq '.panels |= [.[2], .[0], .[1]] + .[3:]' | \
  honeycomb board update <board-id> --file - --replace
```

## SLI-Oriented Board Design

Structure boards around Service Level Indicators. An SLI is a ratio: good events / total events. Keep to five or fewer per service.

| SLI | Honeycomb calculation | Visualization |
|-----|----------------------|---------------|
| Availability | `AVG(error)` inverted (1 - error rate) | `line` |
| Latency | `P50`/`P99` on `duration_ms` | `line` overlaid |
| Error rate | `AVG(error)` on boolean column | `line` |
| Throughput | `COUNT` | `stacked` by service |
| Freshness | `MAX(age_seconds)` or custom | `line` |

For SLO-aware boards, use `compare_time_offset_seconds` (e.g., `86400`) to overlay current error rate against the previous day. This shows burn trajectory without requiring SLO-specific API access.

## Reference

- [Board API types](references/api-types.md)
- [Query examples](references/query-examples.md)
