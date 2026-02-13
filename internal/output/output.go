package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

const (
	FormatJSON  = "json"
	FormatTable = "table"
)

var cellStyle = lipgloss.NewStyle().Padding(0, 1)

var styleFunc = func(row, col int) lipgloss.Style {
	if row == table.HeaderRow {
		return cellStyle.Align(lipgloss.Center)
	}
	return cellStyle
}

var fieldsStyleFunc = func(row, col int) lipgloss.Style {
	return cellStyle
}

type Column struct {
	// Header is the column title, written in Title Case (e.g., "Key Name").
	// It is automatically uppercased when rendered in a table.
	Header string
	Value  func(any) string
}

type TableDef struct {
	Columns []Column
}

type Field struct {
	Label string
	Value string
}

type Writer struct {
	out    io.Writer
	format string
}

func New(out io.Writer, format string) *Writer {
	return &Writer{out: out, format: format}
}

func (w *Writer) writeJSON(data any) error {
	enc := json.NewEncoder(w.out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (w *Writer) Write(data any, td TableDef) error {
	switch w.format {
	case FormatJSON:
		return w.writeJSON(data)
	case FormatTable:
		return w.writeTable(data, td)
	default:
		return fmt.Errorf("unsupported format: %s", w.format)
	}
}

func (w *Writer) WriteValue(data any, writeTable func(io.Writer) error) error {
	switch w.format {
	case FormatJSON:
		return w.writeJSON(data)
	case FormatTable:
		return writeTable(w.out)
	default:
		return fmt.Errorf("unsupported format: %s", w.format)
	}
}

func (w *Writer) WriteFields(data any, fields []Field) error {
	switch w.format {
	case FormatJSON:
		return w.writeJSON(data)
	case FormatTable:
		return w.writeFieldsTable(fields)
	default:
		return fmt.Errorf("unsupported format: %s", w.format)
	}
}

func (w *Writer) writeTable(data any, td TableDef) error {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("table format requires a slice, got %s", rv.Kind())
	}

	if len(td.Columns) == 0 {
		return fmt.Errorf("table format requires at least one column definition")
	}

	headers := make([]string, len(td.Columns))
	for i, col := range td.Columns {
		headers[i] = strings.ToUpper(col.Header)
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		StyleFunc(styleFunc).
		Headers(headers...)

	for i := range rv.Len() {
		elem := rv.Index(i).Interface()
		row := make([]string, len(td.Columns))
		for j, col := range td.Columns {
			row[j] = col.Value(elem)
		}
		t.Row(row...)
	}

	_, err := fmt.Fprintln(w.out, t)
	return err
}

func (w *Writer) writeFieldsTable(fields []Field) error {
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		StyleFunc(fieldsStyleFunc).
		BorderHeader(false)

	for _, f := range fields {
		t.Row(f.Label, f.Value)
	}

	_, err := fmt.Fprintln(w.out, t)
	return err
}

type DynamicTableDef struct {
	Headers []string
	Rows    [][]string
}

func (w *Writer) WriteDynamic(data any, td DynamicTableDef) error {
	switch w.format {
	case FormatJSON:
		return w.writeJSON(data)
	case FormatTable:
		return w.writeDynamicTable(td)
	default:
		return fmt.Errorf("unsupported format: %s", w.format)
	}
}

func (w *Writer) writeDynamicTable(td DynamicTableDef) error {
	if len(td.Headers) == 0 {
		return fmt.Errorf("table format requires at least one column definition")
	}

	headers := make([]string, len(td.Headers))
	for i, h := range td.Headers {
		headers[i] = strings.ToUpper(h)
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		StyleFunc(styleFunc).
		Headers(headers...)

	for _, row := range td.Rows {
		t.Row(row...)
	}

	_, err := fmt.Fprintln(w.out, t)
	return err
}

func (w *Writer) WriteDeleted(id, msg string) error {
	return w.WriteValue(map[string]string{"id": id}, func(out io.Writer) error {
		_, err := fmt.Fprintln(out, msg)
		return err
	})
}
