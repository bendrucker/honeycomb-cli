package slo

import (
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/spf13/cobra"
)

func NewCmd(opts *options.RootOptions) *cobra.Command {
	var dataset string

	cmd := &cobra.Command{
		Use:     "slo",
		Short:   "Manage SLOs",
		Aliases: []string{"slos"},
	}

	cmd.PersistentFlags().StringVar(&dataset, "dataset", "", "Dataset slug (required)")
	_ = cmd.MarkPersistentFlagRequired("dataset")

	cmd.AddCommand(NewListCmd(opts, &dataset))
	cmd.AddCommand(NewGetCmd(opts, &dataset))
	cmd.AddCommand(NewCreateCmd(opts, &dataset))
	cmd.AddCommand(NewUpdateCmd(opts, &dataset))
	cmd.AddCommand(NewDeleteCmd(opts, &dataset))
	cmd.AddCommand(NewBurnAlertCmd(opts, &dataset))
	cmd.AddCommand(NewHistoryCmd(opts, &dataset))

	return cmd
}
