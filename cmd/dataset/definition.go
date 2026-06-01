package dataset

import (
	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/spf13/cobra"
)

func NewDefinitionCmd(opts *options.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "definition",
		Short:   "Manage dataset definitions",
		Aliases: []string{"definitions"},
		Example: `  # Get dataset definitions
  honeycomb dataset definition get my-dataset

  # Update dataset definitions from a file
  honeycomb dataset definition update my-dataset --file definitions.json`,
	}

	cmd.AddCommand(NewDefinitionGetCmd(opts))
	cmd.AddCommand(NewDefinitionUpdateCmd(opts))

	return command.Group(cmd)
}
