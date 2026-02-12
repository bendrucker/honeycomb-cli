package dataset

import (
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dataset",
		Short:   "Manage datasets",
		Aliases: []string{"datasets"},
	}

	cmd.AddCommand(NewListCmd(opts))
	cmd.AddCommand(NewGetCmd(opts))
	cmd.AddCommand(NewCreateCmd(opts))
	cmd.AddCommand(NewUpdateCmd(opts))
	cmd.AddCommand(NewDeleteCmd(opts))
	cmd.AddCommand(NewDefinitionCmd(opts))

	return cmd
}
