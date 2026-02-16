# Query API Types

Quick reference for query-related API types used by the Honeycomb CLI.

## Query

```
calculations[]
  op                QueryOp   (required)
  column            *string   (required for most ops, optional for COUNT)
  name              *string   (display name, Metrics Beta)
  filters[]                   (per-calculation filters, Metrics Beta)
  filter_combination *string  ("AND" | "OR")

filters[]
  column            string    (required)
  op                FilterOp  (required)
  value             *any      (required for most ops)

filter_combination  *string   ("AND" | "OR", default "AND")
breakdowns          *[]string (GROUP BY columns, max 100)
orders[]
  column            *string
  op                *QueryOp
  order             *string   ("ascending" | "descending")

limit               *int      (1-1000, default 100; up to 10000 with disable_series)
time_range          *int      (seconds, default 7200)
start_time          *int      (UNIX timestamp)
end_time            *int      (UNIX timestamp)
granularity         *int      (seconds per bucket, range: time_range/1000 to time_range)

havings[]
  calculate_op      HavingOp  (required)
  column            *string
  op                *string   ("=" | "!=" | ">" | ">=" | "<" | "<=")
  value             *float32

calculated_fields[]
  name              string    (required)
  expression        string    (required, derived column formula syntax)

compare_time_offset_seconds *int (1800|3600|7200|28800|86400|604800|2419200|15724800)
```

## QueryAnnotation

```
name       string  (required, 1-320 chars)
query_id   string  (required, immutable after creation)
description *string (max 1023 chars)
id         *string (read-only)
source     *string (read-only: "query" | "board")
```

## BoardQueryVisualizationSettings

```
charts[]
  chart_index          *int     (0-based, maps to calculations array)
  chart_type           *string  ("line" | "stacked" | "stat" | "tsbar" | "cbar" | "cpie")
  log_scale            *bool
  omit_missing_values  *bool

hide_compare     *bool   (hide time comparison overlay)
hide_hovers      *bool   (hide hover tooltips)
hide_markers     *bool   (hide deploy/event markers)
overlaid_charts  *bool   (overlay multiple calculations into one chart)
utc_xaxis        *bool   (display x-axis in UTC)
```
