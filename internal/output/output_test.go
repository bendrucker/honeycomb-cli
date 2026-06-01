package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type testItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

var testTable = TableDef{
	Columns: []Column{
		{Header: "Name", Value: func(v any) string { return v.(testItem).Name }},
		{Header: "Count", Value: func(v any) string { return fmt.Sprintf("%d", v.(testItem).Count) }},
	},
}

func TestValidateFormat(t *testing.T) {
	for _, tc := range []struct {
		name    string
		format  string
		wantErr bool
	}{
		{name: "json", format: "json", wantErr: false},
		{name: "table", format: "table", wantErr: false},
		{name: "empty is unset default", format: "", wantErr: false},
		{name: "unknown", format: "xml", wantErr: true},
		{name: "case sensitive", format: "JSON", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFormat(tc.format)
			if tc.wantErr && err == nil {
				t.Errorf("ValidateFormat(%q) = nil, want error", tc.format)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("ValidateFormat(%q) = %v, want nil", tc.format, err)
			}
		})
	}
}

func TestWrite_JSON(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatJSON)

	items := []testItem{{Name: "a", Count: 1}, {Name: "b", Count: 2}}
	if err := w.Write(items, testTable); err != nil {
		t.Fatal(err)
	}

	var got []testItem
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 2 || got[0].Name != "a" || got[1].Count != 2 {
		t.Errorf("got %+v", got)
	}
}

func TestWrite_Table(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	items := []testItem{{Name: "a", Count: 1}, {Name: "b", Count: 2}}
	if err := w.Write(items, testTable); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "COUNT") {
		t.Errorf("missing headers in output:\n%s", out)
	}
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Errorf("missing data in output:\n%s", out)
	}
	if !strings.Contains(out, "╭") {
		t.Errorf("expected rounded border in output:\n%s", out)
	}
}

func TestWrite_Table_Empty(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	var items []testItem
	if err := w.Write(items, testTable); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected headers in empty table:\n%s", out)
	}
}

func TestWriteList_Table_Empty(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	var items []testItem
	if err := w.WriteList(items, testTable, "No items found."); err != nil {
		t.Fatal(err)
	}

	if out := buf.String(); out != "No items found.\n" {
		t.Errorf("table output = %q, want %q", out, "No items found.\n")
	}
	if strings.Contains(buf.String(), "NAME") {
		t.Errorf("expected no header table for empty list:\n%s", buf.String())
	}
}

func TestWriteList_Table_NonEmpty(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	items := []testItem{{Name: "a", Count: 1}}
	if err := w.WriteList(items, testTable, "No items found."); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "a") {
		t.Errorf("expected header table for non-empty list:\n%s", out)
	}
	if strings.Contains(out, "No items found.") {
		t.Errorf("unexpected empty message for non-empty list:\n%s", out)
	}
}

func TestWriteList_JSON_Empty(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatJSON)

	items := []testItem{}
	if err := w.WriteList(items, testTable, "No items found."); err != nil {
		t.Fatal(err)
	}

	if out := strings.TrimSpace(buf.String()); out != "[]" {
		t.Errorf("JSON output = %q, want %q", out, "[]")
	}
}

func TestWrite_UnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "xml")

	err := w.Write([]testItem{}, testTable)
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("err = %v, want unsupported format", err)
	}
}

func TestWrite_Table_NonSlice(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	err := w.Write("not a slice", testTable)
	if err == nil || !strings.Contains(err.Error(), "requires a slice") {
		t.Errorf("err = %v, want requires a slice", err)
	}
}

func TestWrite_Table_NoColumns(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	err := w.Write([]testItem{{Name: "a", Count: 1}}, TableDef{})
	if err == nil || !strings.Contains(err.Error(), "at least one column") {
		t.Errorf("err = %v, want column definition error", err)
	}
}

func TestWriteMessage_JSON(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatJSON)

	item := testItem{Name: "a", Count: 1}
	if err := w.WriteMessage(item, "ignored in JSON mode"); err != nil {
		t.Fatal(err)
	}

	var got testItem
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Name != "a" || got.Count != 1 {
		t.Errorf("got %+v", got)
	}
}

func TestWriteMessage_Table(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	if err := w.WriteMessage(testItem{Name: "a", Count: 1}, "Authenticated as acme"); err != nil {
		t.Fatal(err)
	}

	if out := buf.String(); out != "Authenticated as acme\n" {
		t.Errorf("table output = %q, want %q", out, "Authenticated as acme\n")
	}
}

func TestWriteMessage_TableEmptyLine(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	if err := w.WriteMessage(testItem{Name: "a", Count: 1}, ""); err != nil {
		t.Fatal(err)
	}

	if out := buf.String(); out != "" {
		t.Errorf("table output = %q, want empty", out)
	}
}

func TestWriteMessage_UnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "xml")

	err := w.WriteMessage(testItem{}, "msg")
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("err = %v, want unsupported format", err)
	}
}

func TestWriteFields_JSON(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatJSON)

	item := testItem{Name: "a", Count: 1}
	fields := []Field{
		{"Name", item.Name},
		{"Count", fmt.Sprintf("%d", item.Count)},
	}
	if err := w.WriteFields(item, fields); err != nil {
		t.Fatal(err)
	}

	var got testItem
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Name != "a" || got.Count != 1 {
		t.Errorf("got %+v", got)
	}
}

func TestWriteFields_Table(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	fields := []Field{
		{"Name", "a"},
		{"Count", "1"},
	}
	if err := w.WriteFields(testItem{Name: "a", Count: 1}, fields); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Name") || !strings.Contains(out, "a") {
		t.Errorf("missing Name field in output:\n%s", out)
	}
	if !strings.Contains(out, "Count") || !strings.Contains(out, "1") {
		t.Errorf("missing Count field in output:\n%s", out)
	}
	if !strings.Contains(out, "╭") {
		t.Errorf("expected rounded border in output:\n%s", out)
	}
}

func TestWriteFields_TableClosingBorder(t *testing.T) {
	for _, tc := range []struct {
		name   string
		fields []Field
	}{
		{
			name:   "single row",
			fields: []Field{{"Name", "a"}},
		},
		{
			name:   "multiple rows",
			fields: []Field{{"Name", "a"}, {"Count", "1"}},
		},
		{
			name:   "multi line value",
			fields: []Field{{"Name", "a"}, {"Body", "line one\nline two"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := New(&buf, FormatTable)
			if err := w.WriteFields(testItem{}, tc.fields); err != nil {
				t.Fatal(err)
			}

			out := strings.TrimRight(buf.String(), "\n")
			if !strings.HasPrefix(out, "╭") {
				t.Errorf("expected top border, got:\n%s", out)
			}
			if !strings.HasSuffix(out, "╯") {
				t.Errorf("expected closing bottom border, got:\n%s", out)
			}
		})
	}
}

func TestWriteFields_UnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "xml")

	err := w.WriteFields(testItem{}, nil)
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("err = %v, want unsupported format", err)
	}
}

func TestWriteDynamic_Table(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	td := DynamicTableDef{
		Headers: []string{"Col A", "Col B"},
		Rows:    [][]string{{"1", "2"}, {"3", "4"}},
	}
	if err := w.WriteDynamic(nil, td); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "COL A") || !strings.Contains(out, "COL B") {
		t.Errorf("missing headers in output:\n%s", out)
	}
	if !strings.Contains(out, "╭") {
		t.Errorf("expected rounded border in output:\n%s", out)
	}
}
