package cmd

import (
	apiCmd "github.com/bendrucker/honeycomb-cli/cmd/api"
	"github.com/bendrucker/honeycomb-cli/cmd/auth"
	"github.com/bendrucker/honeycomb-cli/cmd/board"
	"github.com/bendrucker/honeycomb-cli/cmd/column"
	"github.com/bendrucker/honeycomb-cli/cmd/dataset"
	"github.com/bendrucker/honeycomb-cli/cmd/marker"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/cmd/recipient"
	"github.com/bendrucker/honeycomb-cli/cmd/slo"
	"github.com/bendrucker/honeycomb-cli/cmd/trigger"
	"github.com/bendrucker/honeycomb-cli/internal/agent"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

func NewRootCmd(ios *iostreams.IOStreams) *cobra.Command {
	opts := &options.RootOptions{IOStreams: ios}

	cmd := &cobra.Command{
		Use:           "honeycomb",
		Short:         "Honeycomb CLI",
		Long:          "Work with Honeycomb from the command line.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.Load(config.DefaultPath())
			if err != nil {
				return err
			}
			opts.Config = cfg

			if agent.Detect() != nil {
				opts.NoInteractive = true
				if opts.Format == "" {
					opts.Format = output.FormatJSON
				}
			}

			if opts.NoInteractive {
				ios.SetNeverPrompt(true)
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&opts.NoInteractive, "no-interactive", false, "Disable interactive prompts")
	cmd.PersistentFlags().StringVar(&opts.Format, "format", "", "Output format: json, table, yaml")
	cmd.PersistentFlags().StringVar(&opts.APIUrl, "api-url", "", "Honeycomb API URL")
	cmd.PersistentFlags().StringVar(&opts.Profile, "profile", "", "Configuration profile to use")

	cmd.AddCommand(apiCmd.NewCmd(opts))
	cmd.AddCommand(auth.NewCmd(opts))
	cmd.AddCommand(board.NewCmd(opts))
	cmd.AddCommand(column.NewCmd(opts))
	cmd.AddCommand(dataset.NewCmd(opts))
	cmd.AddCommand(marker.NewCmd(opts))
	cmd.AddCommand(recipient.NewCmd(opts))
	cmd.AddCommand(slo.NewCmd(opts))
	cmd.AddCommand(trigger.NewCmd(opts))

	return cmd
}
