package query

import (
	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	var dataset string

	cmd := &cobra.Command{
		Use:     "query",
		Short:   "Manage queries and saved queries",
		Aliases: []string{"queries"},
	}

	cmd.PersistentFlags().StringVar(&dataset, "dataset", "", "Dataset slug (required)")
	_ = cmd.MarkPersistentFlagRequired("dataset")

	cmd.AddCommand(NewRunCmd(opts, &dataset))
	cmd.AddCommand(NewAnnotationCmd(opts, &dataset))

	return command.Group(cmd)
}
