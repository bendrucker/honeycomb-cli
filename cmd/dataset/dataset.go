package dataset

import (
	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dataset",
		Short:   "Manage datasets",
		Aliases: []string{"datasets"},
		Example: `  # List datasets
  honeycomb dataset list

  # Get a dataset by slug
  honeycomb dataset get my-dataset

  # Create a dataset
  honeycomb dataset create --name "My Dataset"`,
	}

	cmd.AddCommand(NewListCmd(opts))
	cmd.AddCommand(NewGetCmd(opts))
	cmd.AddCommand(NewCreateCmd(opts))
	cmd.AddCommand(NewUpdateCmd(opts))
	cmd.AddCommand(NewDeleteCmd(opts))
	cmd.AddCommand(NewDefinitionCmd(opts))

	return command.Group(cmd)
}
