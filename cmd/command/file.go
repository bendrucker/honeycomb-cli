package command

import (
	"fmt"
	"io"
	"os"

	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/bendrucker/honeycomb-cli/internal/jsonutil"
)

// ReadDefinitionFile reads a JSON definition from a file path, or from stdin
// when path is "-", and sanitizes shell-mangled escapes (see jsonutil). It is
// the single home for the file/stdin intake that the create and update
// commands repeat.
func ReadDefinitionFile(ios *iostreams.IOStreams, path string) ([]byte, error) {
	var r io.Reader
	if path == "-" {
		r = ios.In
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	data, err = jsonutil.Sanitize(data)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return data, nil
}
