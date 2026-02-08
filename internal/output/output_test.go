package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
)

type testItem struct {
	Name  string `json:"name" yaml:"name"`
	Count int    `json:"count" yaml:"count"`
}

var testTable = TableDef{
	Columns: []Column{
		{Header: "Name", Value: func(v any) string { return v.(testItem).Name }},
		{Header: "Count", Value: func(v any) string { return fmt.Sprintf("%d", v.(testItem).Count) }},
	},
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

func TestWrite_YAML(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatYAML)

	items := []testItem{{Name: "a", Count: 1}}
	if err := w.Write(items, testTable); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "name: a") {
		t.Errorf("yaml missing 'name: a': %s", out)
	}
	if !strings.Contains(out, "count: 1") {
		t.Errorf("yaml missing 'count: 1': %s", out)
	}
}

func TestWrite_Table(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	items := []testItem{{Name: "a", Count: 1}, {Name: "b", Count: 2}}
	if err := w.Write(items, testTable); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3 (header + 2 rows)", len(lines))
	}
	if !strings.Contains(lines[0], "NAME") || !strings.Contains(lines[0], "COUNT") {
		t.Errorf("header = %q", lines[0])
	}
}

func TestWrite_Table_Empty(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	var items []testItem
	if err := w.Write(items, testTable); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1 (header only)", len(lines))
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

func TestWriteValue_JSON(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatJSON)

	item := testItem{Name: "a", Count: 1}
	if err := w.WriteValue(item, nil); err != nil {
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

func TestWriteValue_YAML(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatYAML)

	item := testItem{Name: "b", Count: 2}
	if err := w.WriteValue(item, nil); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "name: b") {
		t.Errorf("yaml missing 'name: b': %s", out)
	}
}

func TestWriteValue_Table(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, FormatTable)

	item := testItem{Name: "a", Count: 1}
	err := w.WriteValue(item, func(out io.Writer) error {
		_, err := fmt.Fprintf(out, "Name:\t%s\nCount:\t%d\n", item.Name, item.Count)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Name:") || !strings.Contains(out, "a") {
		t.Errorf("table output = %q", out)
	}
}

func TestWriteValue_UnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "xml")

	err := w.WriteValue(testItem{}, nil)
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("err = %v, want unsupported format", err)
	}
}
