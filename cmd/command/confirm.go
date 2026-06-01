// Package command holds helpers shared by the resource CRUD commands: delete
// confirmation, definition-file intake, and flag-override merging.
package command

import (
	"fmt"
	"strings"

	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
)

// ConfirmDelete reports whether a delete should proceed. When yes is set it
// returns true without prompting. Otherwise it requires an interactive
// terminal, resolves a display name (calling fetchName only when a prompt is
// actually needed, so a --yes delete makes no extra API call), and prompts for
// y/N. fetchName may be nil, in which case fallbackName is shown. The caller
// decides what a declined delete means (return nil, or an error).
func ConfirmDelete(ios *iostreams.IOStreams, yes bool, noun, fallbackName string, fetchName func() (string, error)) (bool, error) {
	if yes {
		return true, nil
	}

	if !ios.CanPrompt() {
		return false, fmt.Errorf("--yes is required in non-interactive mode")
	}

	name := fallbackName
	if fetchName != nil {
		fetched, err := fetchName()
		if err != nil {
			return false, err
		}
		if fetched != "" {
			name = fetched
		}
	}

	answer, err := prompt.Line(ios.Err, ios.In, fmt.Sprintf("Delete %s %q? (y/N): ", noun, name))
	if err != nil {
		return false, err
	}

	return strings.EqualFold(answer, "y"), nil
}
