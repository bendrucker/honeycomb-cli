package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Group configures a grouping command (one that only dispatches to
// subcommands) so that invoking it without a subcommand prints help and exits
// non-zero. Without this, cobra falls through to help with exit 0, making a
// bare or typo'd parent command read as success in CI guards.
//
// Cobra already rejects an unknown subcommand with a non-zero "unknown command"
// error. Group adds the missing case: a bare parent invocation, which sets
// cobra.NoArgs (so any positional arg is rejected) and a RunE that prints help
// and returns an error.
func Group(cmd *cobra.Command) *cobra.Command {
	cmd.Args = cobra.NoArgs
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		if err := cmd.Help(); err != nil {
			return err
		}
		return fmt.Errorf("%q requires a subcommand", cmd.CommandPath())
	}
	return cmd
}
