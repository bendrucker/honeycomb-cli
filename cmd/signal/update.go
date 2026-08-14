package signal

import (
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions) *cobra.Command {
	var (
		enabled         bool
		sensitivity     string
		recipients      []string
		clearRecipients bool
	)

	cmd := &cobra.Command{
		Use:   "update <signal-id>",
		Short: "Update a signal",
		Long: "Update a signal.\n\n" +
			"Sensitivity applies to error_rate signals that have finished training. " +
			"Recipients are replaced as a set: pass --recipient for each recipient to keep, " +
			"or --clear-recipients to remove them all.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd, opts, args[0], enabled, sensitivity, recipients, clearRecipients)
		},
	}

	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable or disable the signal")
	cmd.Flags().StringVar(&sensitivity, "sensitivity", "", "Sensitivity: "+command.EnumUsage(sensitivities))
	cmd.Flags().StringSliceVar(&recipients, "recipient", nil, "Recipient ID to notify (repeatable, replaces the current set)")
	cmd.Flags().BoolVar(&clearRecipients, "clear-recipients", false, "Remove all recipients from the signal")

	cmd.MarkFlagsMutuallyExclusive("recipient", "clear-recipients")

	return cmd
}

func runUpdate(cmd *cobra.Command, opts *options.RootOptions, id string, enabled bool, sensitivity string, recipients []string, clearRecipients bool) error {
	if !command.AnyChanged(cmd, "enabled", "sensitivity", "recipient", "clear-recipients") {
		return fmt.Errorf("--enabled, --sensitivity, --recipient, or --clear-recipients is required")
	}
	if err := command.ValidateEnum("sensitivity", sensitivity, sensitivities); err != nil {
		return err
	}

	body := api.UpdateSignalRequest{}
	if cmd.Flags().Changed("enabled") {
		body.Enabled = &enabled
	}
	if cmd.Flags().Changed("sensitivity") {
		s := api.AnomalySignalSensitivity(sensitivity)
		body.Sensitivity = &s
	}
	if clearRecipients {
		body.Recipients = &[]api.SignalRecipient{}
	} else if len(recipients) > 0 {
		assigned := make([]api.SignalRecipient, len(recipients))
		for i, r := range recipients {
			assigned[i] = api.SignalRecipient{Id: r}
		}
		body.Recipients = &assigned
	}

	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	resp, err := client.UpdateSignalWithResponse(cmd.Context(), id, body)
	if err != nil {
		return fmt.Errorf("updating signal: %w", err)
	}

	signal, err := api.Decode(resp.StatusCode(), resp.Status(), resp.Body, resp.JSON200)
	if err != nil {
		return err
	}

	return writeSignalDetail(opts, toDetail(*signal))
}
