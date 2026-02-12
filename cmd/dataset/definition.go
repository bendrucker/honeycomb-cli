package dataset

import (
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/spf13/cobra"
)

func NewDefinitionCmd(opts *options.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "definition",
		Short:   "Manage dataset definitions",
		Aliases: []string{"definitions"},
	}

	cmd.AddCommand(NewDefinitionGetCmd(opts))
	cmd.AddCommand(NewDefinitionUpdateCmd(opts))

	return cmd
}
