package deref

import "time"

func Val[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func String(p *string) string {
	return Val(p)
}

func Bool(p *bool) bool {
	return Val(p)
}

func Int(p *int) int {
	return Val(p)
}

func Time(p *time.Time) string {
	if p == nil {
		return ""
	}
	return p.Format(time.RFC3339)
}

func Enum[T ~string](p *T) string {
	if p == nil {
		return ""
	}
	return string(*p)
}
