package output

import (
	"strings"
	"testing"
)

type tagged struct {
	ID      string `col:"ID" detail:"ID"`
	Name    string `col:"Name" detail:"Name"`
	Enabled bool   `col:"Enabled" detail:"Enabled"`
	Count   int    `col:"Count" detail:"Count"`
	Ratio   float64
	Hidden  string
}

func TestTableFromTags(t *testing.T) {
	td := TableFromTags[tagged]()

	wantHeaders := []string{"ID", "Name", "Enabled", "Count"}
	if len(td.Columns) != len(wantHeaders) {
		t.Fatalf("got %d columns, want %d", len(td.Columns), len(wantHeaders))
	}

	item := tagged{ID: "abc", Name: "widget", Enabled: true, Count: 7, Ratio: 0.5, Hidden: "x"}
	wantValues := []string{"abc", "widget", "true", "7"}
	for i, col := range td.Columns {
		if col.Header != wantHeaders[i] {
			t.Errorf("column %d header = %q, want %q", i, col.Header, wantHeaders[i])
		}
		if got := col.Value(item); got != wantValues[i] {
			t.Errorf("column %q value = %q, want %q", col.Header, got, wantValues[i])
		}
	}
}

func TestFieldsFromTags(t *testing.T) {
	item := tagged{ID: "abc", Name: "widget", Enabled: false, Count: 0, Hidden: "x"}
	fields := FieldsFromTags(item)

	want := []Field{
		{Label: "ID", Value: "abc"},
		{Label: "Name", Value: "widget"},
		{Label: "Enabled", Value: "false"},
		{Label: "Count", Value: "0"},
	}
	if len(fields) != len(want) {
		t.Fatalf("got %d fields, want %d", len(fields), len(want))
	}
	for i, f := range fields {
		if f != want[i] {
			t.Errorf("field %d = %+v, want %+v", i, f, want[i])
		}
	}
}

func TestFieldsFromTags_Float(t *testing.T) {
	type floaty struct {
		Ratio float64 `detail:"Ratio"`
	}
	fields := FieldsFromTags(floaty{Ratio: 99.5})
	if len(fields) != 1 || fields[0].Value != "99.5" {
		t.Errorf("got %+v, want Ratio=99.5", fields)
	}
}

func TestTableFromTags_RendersThroughWrite(t *testing.T) {
	var buf strings.Builder
	w := New(&buf, FormatTable)
	if err := w.Write([]tagged{{ID: "abc", Name: "widget", Enabled: true, Count: 7}}, TableFromTags[tagged]()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"ID", "NAME", "abc", "widget", "true", "7"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
}
