# Query Examples

Common query patterns for board panels. Each example shows the query JSON spec and recommended visualization settings.

## Latency: P50 and P99

Always use percentiles for latency, never AVG. P50 shows the typical experience, P99 shows the worst case. The gap between them is the signal.

```json
{
  "calculations": [
    {"op": "P50", "column": "duration_ms"},
    {"op": "P99", "column": "duration_ms"}
  ],
  "filters": [{"column": "http.route", "op": "!=", "value": "/api/health"}],
  "time_range": 7200,
  "granularity": 120
}
```

Visualization: `line` with `overlaid_charts: true`. Title: "Latency: P50 and P99 (ms)"

Filter out health checks — they skew latency downward and aren't user-facing.

## Error rate (aggregate)

Use `AVG` on a boolean error column. Honeycomb computes the proportion of `true` values, giving you the error rate as a 0.0–1.0 decimal.

```json
{
  "calculations": [{"op": "AVG", "column": "error"}],
  "time_range": 7200,
  "granularity": 120
}
```

Visualization: `line`. Title: "Error Rate"

## Error rate by service

Break down by service to answer "where are errors coming from?" Use 5-min granularity to smooth noise from sparse buckets.

```json
{
  "calculations": [{"op": "AVG", "column": "error"}],
  "breakdowns": ["service.name"],
  "time_range": 7200,
  "granularity": 300
}
```

Visualization: `line`. Title: "Error Rate by Service"

Do NOT break down by `http.status_code` — individual codes produce noisy, unactionable charts.

## Throughput by service

```json
{
  "calculations": [{"op": "COUNT"}],
  "breakdowns": ["service.name"],
  "time_range": 7200,
  "granularity": 120,
  "limit": 20
}
```

Visualization: `stacked` bar — shows both total volume and per-service composition. Title: "Request Volume by Service"

## Slow requests (HAVING filter)

```json
{
  "calculations": [{"op": "P99", "column": "duration_ms"}, {"op": "COUNT"}],
  "breakdowns": ["http.route"],
  "havings": [{"calculate_op": "P99", "column": "duration_ms", "op": ">", "value": 500}],
  "time_range": 7200,
  "granularity": 60,
  "limit": 10
}
```

Visualization: `cbar` for the P99 calculation. Title: "Slow Endpoints (p99 > 500ms)"

## Unique users

```json
{
  "calculations": [{"op": "COUNT_DISTINCT", "column": "user.id"}],
  "time_range": 86400,
  "granularity": 300
}
```

Visualization: `stat` for headline number. Title: "Unique Users (24h)"

## Database query duration

```json
{
  "calculations": [{"op": "HEATMAP", "column": "db.duration_ms"}],
  "filters": [{"column": "db.system", "op": "=", "value": "postgresql"}],
  "time_range": 7200,
  "granularity": 60
}
```

Visualization: default (heatmap renders natively). Title: "Database Query Duration Distribution"

## Error budget burn rate (comparison)

```json
{
  "calculations": [{"op": "AVG", "column": "error"}],
  "time_range": 86400,
  "granularity": 300,
  "compare_time_offset_seconds": 86400
}
```

Visualization: `line`. Title: "Error Rate vs. 24h Ago"

## Queue depth / concurrent operations

```json
{
  "calculations": [{"op": "CONCURRENCY"}],
  "filters": [{"column": "span.kind", "op": "=", "value": "consumer"}],
  "time_range": 7200,
  "granularity": 60
}
```

Visualization: `line`. Title: "Concurrent Queue Consumers"

## Top endpoints by traffic share

```json
{
  "calculations": [{"op": "COUNT"}],
  "breakdowns": ["http.route"],
  "time_range": 3600,
  "limit": 10
}
```

Visualization: `cpie`. Title: "Traffic Share by Endpoint"

## Sparse events (webhooks, batch jobs)

```json
{
  "calculations": [{"op": "COUNT"}],
  "filters": [{"column": "event.type", "op": "=", "value": "webhook.received"}],
  "time_range": 86400,
  "granularity": 600
}
```

Visualization: `tsbar` with `omit_missing_values: true`. Title: "Webhook Events Received"

Wide granularity (600s = 10min buckets) ensures sparse events produce visible bars rather than isolated dots on a line chart.

## Full board panel JSON example

Complete panel ready to insert into a board's panels array:

```json
{
  "type": "query",
  "query_panel": {
    "query_id": "abc123",
    "query_annotation_id": "def456",
    "query_style": "graph",
    "visualization_settings": {
      "charts": [
        {"chart_index": 0, "chart_type": "tsbar", "omit_missing_values": false}
      ],
      "hide_markers": false,
      "overlaid_charts": false
    }
  },
  "position": {
    "x_coordinate": 0,
    "y_coordinate": 0,
    "width": 6,
    "height": 4
  }
}
```
