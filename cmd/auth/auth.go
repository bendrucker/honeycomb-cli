package auth

import (
	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Example: `  # Authenticate with Honeycomb
  honeycomb auth login

  # Check the current authentication status
  honeycomb auth status

  # Remove stored authentication keys
  honeycomb auth logout`,
	}

	cmd.AddCommand(NewLoginCmd(opts))
	cmd.AddCommand(NewLogoutCmd(opts))
	cmd.AddCommand(NewStatusCmd(opts))
	cmd.AddCommand(NewProfileCmd(opts))

	return command.Group(cmd)
}
