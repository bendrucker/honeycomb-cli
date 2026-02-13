package deref

import (
	"testing"
	"time"
)

func TestVal(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := Val[string](nil); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
	t.Run("non-nil", func(t *testing.T) {
		s := "hello"
		if got := Val(&s); got != "hello" {
			t.Errorf("got %q, want hello", got)
		}
	})
}

func TestString(t *testing.T) {
	if got := String(nil); got != "" {
		t.Errorf("got %q, want empty", got)
	}
	s := "hello"
	if got := String(&s); got != "hello" {
		t.Errorf("got %q, want hello", got)
	}
}

func TestBool(t *testing.T) {
	if got := Bool(nil); got != false {
		t.Errorf("got %v, want false", got)
	}
	b := true
	if got := Bool(&b); got != true {
		t.Errorf("got %v, want true", got)
	}
}

func TestInt(t *testing.T) {
	if got := Int(nil); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
	n := 42
	if got := Int(&n); got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func TestFloat64(t *testing.T) {
	if got := Float64(nil); got != 0 {
		t.Errorf("got %f, want 0", got)
	}
	f := 3.14
	if got := Float64(&f); got != 3.14 {
		t.Errorf("got %f, want 3.14", got)
	}
}

func TestTime(t *testing.T) {
	if got := Time(nil); got != "" {
		t.Errorf("got %q, want empty", got)
	}
	ts := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	if got := Time(&ts); got != "2024-01-15T12:00:00Z" {
		t.Errorf("got %q, want 2024-01-15T12:00:00Z", got)
	}
}

type color string

func TestEnum(t *testing.T) {
	if got := Enum[color](nil); got != "" {
		t.Errorf("got %q, want empty", got)
	}
	c := color("red")
	if got := Enum(&c); got != "red" {
		t.Errorf("got %q, want red", got)
	}
}
