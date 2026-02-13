package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

const (
	FormatJSON  = "json"
	FormatYAML  = "yaml"
	FormatTable = "table"
)

type Column struct {
	// Header is the column title, written in Title Case (e.g., "Key Name").
	// It is automatically uppercased when rendered in a table.
	Header string
	Value  func(any) string
}

type TableDef struct {
	Columns []Column
}

type Writer struct {
	out    io.Writer
	format string
}

func New(out io.Writer, format string) *Writer {
	return &Writer{out: out, format: format}
}

func (w *Writer) Write(data any, table TableDef) error {
	switch w.format {
	case FormatJSON:
		enc := json.NewEncoder(w.out)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case FormatYAML:
		return yaml.NewEncoder(w.out).Encode(data)
	case FormatTable:
		return w.writeTable(data, table)
	default:
		return fmt.Errorf("unsupported format: %s", w.format)
	}
}

func (w *Writer) WriteValue(data any, writeTable func(io.Writer) error) error {
	switch w.format {
	case FormatJSON:
		enc := json.NewEncoder(w.out)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case FormatYAML:
		return yaml.NewEncoder(w.out).Encode(data)
	case FormatTable:
		return writeTable(w.out)
	default:
		return fmt.Errorf("unsupported format: %s", w.format)
	}
}

func (w *Writer) writeTable(data any, table TableDef) error {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("table format requires a slice, got %s", rv.Kind())
	}

	if len(table.Columns) == 0 {
		return fmt.Errorf("table format requires at least one column definition")
	}

	tw := tabwriter.NewWriter(w.out, 0, 0, 2, ' ', 0)
	for i, col := range table.Columns {
		if i > 0 {
			_, _ = fmt.Fprint(tw, "\t")
		}
		_, _ = fmt.Fprint(tw, strings.ToUpper(col.Header))
	}
	_, _ = fmt.Fprintln(tw)

	for i := range rv.Len() {
		elem := rv.Index(i).Interface()
		for j, col := range table.Columns {
			if j > 0 {
				_, _ = fmt.Fprint(tw, "\t")
			}
			_, _ = fmt.Fprint(tw, col.Value(elem))
		}
		_, _ = fmt.Fprintln(tw)
	}

	return tw.Flush()
}

type DynamicTableDef struct {
	Headers []string
	Rows    [][]string
}

func (w *Writer) WriteDynamic(data any, table DynamicTableDef) error {
	switch w.format {
	case FormatJSON:
		enc := json.NewEncoder(w.out)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case FormatYAML:
		return yaml.NewEncoder(w.out).Encode(data)
	case FormatTable:
		return w.writeDynamicTable(table)
	default:
		return fmt.Errorf("unsupported format: %s", w.format)
	}
}

func (w *Writer) writeDynamicTable(table DynamicTableDef) error {
	if len(table.Headers) == 0 {
		return fmt.Errorf("table format requires at least one column definition")
	}

	tw := tabwriter.NewWriter(w.out, 0, 0, 2, ' ', 0)
	for i, h := range table.Headers {
		if i > 0 {
			_, _ = fmt.Fprint(tw, "\t")
		}
		_, _ = fmt.Fprint(tw, strings.ToUpper(h))
	}
	_, _ = fmt.Fprintln(tw)

	for _, row := range table.Rows {
		for i, cell := range row {
			if i > 0 {
				_, _ = fmt.Fprint(tw, "\t")
			}
			_, _ = fmt.Fprint(tw, cell)
		}
		_, _ = fmt.Fprintln(tw)
	}

	return tw.Flush()
}

func (w *Writer) WriteDeleted(id, msg string) error {
	return w.WriteValue(map[string]string{"id": id}, func(out io.Writer) error {
		_, err := fmt.Fprintln(out, msg)
		return err
	})
}
