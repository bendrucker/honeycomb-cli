package auth

import (
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}

	cmd.AddCommand(NewLoginCmd(opts))
	cmd.AddCommand(NewLogoutCmd(opts))
	cmd.AddCommand(NewStatusCmd(opts))

	return cmd
}
