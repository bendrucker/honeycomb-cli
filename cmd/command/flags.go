package command

import "github.com/spf13/cobra"

// AnyChanged reports whether the user set any of the named flags on cmd. It
// backs the "provide --file or at least one of ..." guards that several update
// commands share, reading cobra's own change tracking instead of a parallel
// set of has-X booleans.
func AnyChanged(cmd *cobra.Command, names ...string) bool {
	for _, n := range names {
		if cmd.Flags().Changed(n) {
			return true
		}
	}
	return false
}
