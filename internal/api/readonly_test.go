package api

import (
	"encoding/json"
	"testing"
)

func TestStripReadOnly(t *testing.T) {
	t.Run("strips known fields", func(t *testing.T) {
		input := `{"id":"abc","name":"My Board","links":{"self":"url"},"description":"desc"}`
		got, err := StripReadOnly([]byte(input), "Board")
		if err != nil {
			t.Fatal(err)
		}

		var m map[string]json.RawMessage
		if err := json.Unmarshal(got, &m); err != nil {
			t.Fatal(err)
		}

		if _, ok := m["id"]; ok {
			t.Error("expected id to be stripped")
		}
		if _, ok := m["links"]; ok {
			t.Error("expected links to be stripped")
		}
		if _, ok := m["name"]; !ok {
			t.Error("expected name to be preserved")
		}
		if _, ok := m["description"]; !ok {
			t.Error("expected description to be preserved")
		}
	})

	t.Run("unknown schema returns input unchanged", func(t *testing.T) {
		input := `{"id":"abc","name":"test"}`
		got, err := StripReadOnly([]byte(input), "UnknownSchema")
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != input {
			t.Errorf("expected unchanged input, got %s", got)
		}
	})
}

func TestMarshalStrippingReadOnly(t *testing.T) {
	board := Board{
		Id:          strPtr("abc"),
		Name:        "My Board",
		Description: strPtr("desc"),
	}

	got, err := MarshalStrippingReadOnly(board, "Board")
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatal(err)
	}

	if _, ok := m["id"]; ok {
		t.Error("expected id to be stripped")
	}
	if _, ok := m["name"]; !ok {
		t.Error("expected name to be preserved")
	}
	if _, ok := m["description"]; !ok {
		t.Error("expected description to be preserved")
	}
}

func strPtr(s string) *string { return &s }
