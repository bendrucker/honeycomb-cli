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

// ValidFormats lists the output formats accepted by the --format flag.
var ValidFormats = []string{FormatJSON, FormatTable}

// ValidateFormat reports whether format is an accepted --format value. The
// empty string is allowed: it represents the unset flag, which callers resolve
// to a concrete format based on TTY detection and command type.
func ValidateFormat(format string) error {
	switch format {
	case "", FormatJSON, FormatTable:
		return nil
	default:
		return fmt.Errorf("invalid --format %q: must be one of %s", format, strings.Join(ValidFormats, ", "))
	}
}

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

// WriteMessage emits data as JSON, or a single human-readable line in table
// mode. An empty line writes nothing in table mode. It covers the "JSON
// object, or one status line" shape that previously required callers to pass
// a table-building closure.
func (w *Writer) WriteMessage(data any, line string) error {
	switch w.format {
	case FormatJSON:
		return w.writeJSON(data)
	case FormatTable:
		if line == "" {
			return nil
		}
		_, err := fmt.Fprintln(w.out, line)
		return err
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
	// This table sets no Headers, so the default BorderHeader has no separator
	// to draw. Leaving it enabled is also what makes lipgloss render the closing
	// bottom rule; BorderHeader(false) suppresses the bottom border entirely.
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		StyleFunc(fieldsStyleFunc)

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
	return w.WriteMessage(map[string]string{"id": id}, msg)
}
