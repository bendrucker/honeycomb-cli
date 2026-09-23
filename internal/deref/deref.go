// Package deref provides nil-safe dereferencing helpers for pointer fields.
package deref

import "time"

// Val returns the dereferenced value of p, or the zero value if p is nil.
func Val[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// String returns the dereferenced value of p, or an empty string if p is nil.
func String(p *string) string {
	return Val(p)
}

// Bool returns the dereferenced value of p, or false if p is nil.
func Bool(p *bool) bool {
	return Val(p)
}

// Int returns the dereferenced value of p, or 0 if p is nil.
func Int(p *int) int {
	return Val(p)
}

// Time formats p as RFC3339, or returns an empty string if p is nil.
func Time(p *time.Time) string {
	if p == nil {
		return ""
	}
	return p.Format(time.RFC3339)
}

// Enum returns the string value of p, or an empty string if p is nil.
func Enum[T ~string](p *T) string {
	if p == nil {
		return ""
	}
	return string(*p)
}
