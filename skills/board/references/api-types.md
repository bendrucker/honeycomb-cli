# Board API Types

Quick reference for the board-related API types used by the Honeycomb CLI.

## Board

```
name              string          (required, 1-255 chars)
description       *string         (0-1024 chars)
type              BoardType       (required, always "flexible")
layout_generation *string         (write-only: "auto" | "manual")
panels            *[]BoardPanel   (union: QueryPanel | SLOPanel | TextPanel)
preset_filters    *[]PresetFilter (max 5)
tags              *[]Tag          (max 10)
id                *string         (read-only)
links.board_url   *string         (read-only)
```

## BoardPanel (discriminated union on `type`)

### QueryPanel (`type: "query"`)

```
position                     *BoardPanelPosition
query_panel.query_id         string   (required)
query_panel.query_annotation_id string (required)
query_panel.dataset          *string  (read-only)
query_panel.query_style      *string  ("graph" | "table" | "combo")
query_panel.visualization_settings *BoardQueryVisualizationSettings
```

### TextPanel (`type: "text"`)

```
position                *BoardPanelPosition
text_panel.content      string (required, max 10000 chars, Markdown)
```

### SLOPanel (`type: "slo"`)

```
position           *BoardPanelPosition
slo_panel.slo_id   *string
```

## BoardPanelPosition

```
x_coordinate  *int   (0-based column in 12-column grid)
y_coordinate  *int   (0-based row)
width         *int   (1-12, grid columns)
height        *int   (grid rows, typically 1-8)
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

## PresetFilter

```
column  string  (required, column name)
alias   string  (required, display label, max 50 chars)
```

## Tag

```
key    string  (lowercase letters only)
value  string  (lowercase alphanumeric, `/`, `-`)
```

## BoardViewFilter

```
column     string                    (required)
operation  BoardViewFilterOperation  (required)
value      interface{}               (optional, can be string/number/array)
```

Operations: `=`, `!=`, `>`, `>=`, `<`, `<=`, `starts-with`, `does-not-start-with`, `ends-with`, `does-not-end-with`, `exists`, `does-not-exist`, `contains`, `does-not-contain`, `in`, `not-in`

For Query and QueryAnnotation types, see the [query skill API types](../../query/references/api-types.md).
