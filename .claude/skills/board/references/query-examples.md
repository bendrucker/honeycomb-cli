# Query Examples

Common query patterns for board panels. Each example shows the query JSON spec and recommended visualization settings.

## Request latency (p99)

```json
{
  "calculations": [{"op": "P99", "column": "duration_ms"}],
  "filters": [{"column": "service.name", "op": "=", "value": "api-gateway"}],
  "breakdowns": ["http.route"],
  "time_range": 7200,
  "granularity": 60,
  "limit": 10
}
```

Visualization: `tsbar` for single percentile. Title: "Request Latency by Route (p99, ms)"

## Multiple percentiles overlaid

```json
{
  "calculations": [
    {"op": "P50", "column": "duration_ms"},
    {"op": "P95", "column": "duration_ms"},
    {"op": "P99", "column": "duration_ms"}
  ],
  "time_range": 7200,
  "granularity": 60
}
```

Visualization: `line` with `overlaid_charts: true`. Title: "Request Latency Distribution (p50/p95/p99, ms)"

## Error rate by status code

```json
{
  "calculations": [{"op": "COUNT"}],
  "filters": [{"column": "http.status_code", "op": ">=", "value": 400}],
  "breakdowns": ["http.status_code"],
  "time_range": 7200,
  "granularity": 60
}
```

Visualization: `stacked`. Title: "Errors by Status Code"

## Throughput by service

```json
{
  "calculations": [{"op": "COUNT"}],
  "breakdowns": ["service.name"],
  "time_range": 7200,
  "granularity": 60,
  "limit": 20
}
```

Visualization: `tsbar`. Title: "Request Volume by Service"

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
