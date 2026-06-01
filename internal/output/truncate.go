package output

// Truncate shortens s to at most max runes, appending "..." when it must cut.
// It slices on rune boundaries, so the result is always valid UTF-8 even when s
// contains multibyte characters.
func Truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	const ellipsis = "..."
	keep := max - len(ellipsis)
	if keep < 0 {
		keep = 0
	}
	return string(r[:keep]) + ellipsis
}
