package jsonutil

import "encoding/json"

// Sanitize validates data as JSON and returns it unchanged if valid. If
// parsing fails, it sanitizes invalid escape sequences and returns the
// corrected bytes. Returns the original parse error if sanitization does
// not produce valid JSON.
func Sanitize(data []byte) ([]byte, error) {
	if json.Valid(data) {
		return data, nil
	}
	fixed := SanitizeEscapes(data)
	if !json.Valid(fixed) {
		// Return a descriptive error from the original data
		var js json.RawMessage
		return nil, json.Unmarshal(data, &js)
	}
	return fixed, nil
}

// SanitizeEscapes removes invalid JSON escape sequences from within JSON
// string values. Shells like zsh escape characters such as ! to \! via
// history expansion, producing invalid JSON when piped to stdin. This
// function replaces \X with X inside strings when X is not a valid JSON
// escape character (", \, /, b, f, n, r, t, u).
func SanitizeEscapes(data []byte) []byte {
	out := make([]byte, 0, len(data))
	inString := false

	for i := 0; i < len(data); i++ {
		b := data[i]

		if !inString {
			if b == '"' {
				inString = true
			}
			out = append(out, b)
			continue
		}

		// Inside a JSON string
		if b == '"' {
			inString = false
			out = append(out, b)
			continue
		}

		if b == '\\' && i+1 < len(data) {
			next := data[i+1]
			if isValidJSONEscape(next) {
				out = append(out, b, next)
			} else {
				out = append(out, next)
			}
			i++
			continue
		}

		out = append(out, b)
	}

	return out
}

func isValidJSONEscape(b byte) bool {
	switch b {
	case '"', '\\', '/', 'b', 'f', 'n', 'r', 't', 'u':
		return true
	}
	return false
}
