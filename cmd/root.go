package cmd

import (
	"github.com/bendrucker/honeycomb-cli/internal/agent"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/spf13/cobra"
)

type RootOptions struct {
	IOStreams *iostreams.IOStreams
	Config    *config.Config
	Agent     *agent.Agent

	NoInteractive bool
	Format        string
	APIUrl        string
	Profile       string
}

func NewRootCmd(ios *iostreams.IOStreams) *cobra.Command {
	opts := &RootOptions{IOStreams: ios}

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

			if a := agent.Detect(); a != nil {
				opts.Agent = a
				opts.NoInteractive = true
				if opts.Format == "" {
					opts.Format = "json"
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

	return cmd
}
