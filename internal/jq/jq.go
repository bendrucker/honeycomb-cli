package jq

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/itchyny/gojq"
)

func Filter(input io.Reader, output io.Writer, expr string) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("parsing jq expression: %w", err)
	}

	var data any
	if err := json.NewDecoder(input).Decode(&data); err != nil {
		return fmt.Errorf("decoding JSON for jq: %w", err)
	}

	iter := query.Run(data)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			return fmt.Errorf("jq: %w", err)
		}

		switch val := v.(type) {
		case string:
			_, _ = fmt.Fprintln(output, val)
		default:
			b, err := json.Marshal(val)
			if err != nil {
				return fmt.Errorf("encoding jq result: %w", err)
			}
			_, _ = fmt.Fprintln(output, string(b))
		}
	}

	return nil
}
