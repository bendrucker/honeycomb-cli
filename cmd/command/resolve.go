package command

import (
	"io"

	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
)

// Field describes a single create-command input that is supplied by a flag but
// falls back to an interactive prompt when the flag is empty. Resolve consumes
// it to collapse the flag-or-prompt branch that every create command would
// otherwise hand-write.
//
// Prompt is the text shown when prompting. Choices, when non-empty, switches
// the prompt to prompt.Choice (the value must match one of the options).
// Required controls whether an empty result is an error: a required field that
// cannot be prompted returns NonInteractiveErr, and a required field that is
// prompted but left blank returns EmptyErr. An optional field (Required false)
// prompts only when possible and accepts an empty result.
//
// Stream selects which writer the prompt text is written to. Most call sites
// prompt on Err; a zero Stream defaults to Err so only the call sites that
// prompt on Out (dataset create, query run) need to set it.
type Field struct {
	Prompt   string
	Required bool
	Choices  []string

	NonInteractiveErr error
	EmptyErr          error

	Stream Stream
}

// Stream identifies which IOStreams writer a prompt is written to. The zero
// value is StreamErr, matching the majority of call sites.
type Stream int

const (
	StreamErr Stream = iota
	StreamOut
)

// Resolve returns value when it is already set, otherwise prompts for it.
//
//   - value already set: returned as-is, no prompt.
//   - value empty, required, cannot prompt: returns field.NonInteractiveErr.
//   - value empty, can prompt: prompts (prompt.Choice when Choices is set,
//     otherwise prompt.Line) and returns the answer.
//   - value empty, required, prompted but still empty: returns field.EmptyErr.
//   - value empty, optional, cannot prompt: returns "" with no error.
func Resolve(ios *iostreams.IOStreams, value string, field Field) (string, error) {
	if value != "" {
		return value, nil
	}

	if !ios.CanPrompt() {
		if field.Required {
			return "", field.NonInteractiveErr
		}
		return "", nil
	}

	out := promptWriter(ios, field.Stream)

	var (
		answer string
		err    error
	)
	if len(field.Choices) > 0 {
		answer, err = prompt.Choice(out, ios.In, field.Prompt, field.Choices)
	} else {
		answer, err = prompt.Line(out, ios.In, field.Prompt)
	}
	if err != nil {
		return "", err
	}

	if field.Required && answer == "" {
		return "", field.EmptyErr
	}

	return answer, nil
}

func promptWriter(ios *iostreams.IOStreams, s Stream) io.Writer {
	if s == StreamOut {
		return ios.Out
	}
	return ios.Err
}
