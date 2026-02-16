package recipient

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewUpdateCmd(opts *options.RootOptions) *cobra.Command {
	var (
		file           string
		recipientType  string
		target         string
		channel        string
		integrationKey string
		name           string
		url            string
	)

	cmd := &cobra.Command{
		Use:   "update <recipient-id>",
		Short: "Update a recipient",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd, opts, args[0], file, recipientType, target, channel, integrationKey, name, url)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to JSON file (- for stdin)")
	cmd.Flags().StringVar(&recipientType, "type", "", "Recipient type (email, slack, pagerduty, msteams, msteams_workflow, webhook)")
	cmd.Flags().StringVar(&target, "target", "", "Email address (for email type)")
	cmd.Flags().StringVar(&channel, "channel", "", "Slack channel (for slack type)")
	cmd.Flags().StringVar(&integrationKey, "integration-key", "", "PagerDuty integration key (for pagerduty type)")
	cmd.Flags().StringVar(&name, "name", "", "Recipient name (for pagerduty, msteams, msteams_workflow, webhook)")
	cmd.Flags().StringVar(&url, "url", "", "Webhook URL (for msteams, msteams_workflow, webhook)")

	for _, flag := range []string{"type", "target", "channel", "integration-key", "name", "url"} {
		cmd.MarkFlagsMutuallyExclusive("file", flag)
	}

	return cmd
}

func runUpdate(cmd *cobra.Command, opts *options.RootOptions, recipientID, file, recipientType, target, channel, integrationKey, name, url string) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	ctx := cmd.Context()

	var data []byte
	if file != "" {
		data, err = readFile(opts, file)
		if err != nil {
			return err
		}
	} else if hasAnyFlag(cmd, "type", "target", "channel", "integration-key", "name", "url") {
		resp, err := client.GetRecipientWithResponse(ctx, recipientID, auth)
		if err != nil {
			return fmt.Errorf("getting recipient: %w", err)
		}
		if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		var current map[string]any
		if err := json.Unmarshal(resp.Body, &current); err != nil {
			return fmt.Errorf("parsing recipient: %w", err)
		}

		applyRecipientFlags(cmd, current, recipientType, target, channel, integrationKey, name, url)

		data, err = json.Marshal(current)
		if err != nil {
			return fmt.Errorf("encoding recipient: %w", err)
		}
	} else {
		return fmt.Errorf("--file, --type, --target, --channel, --integration-key, --name, or --url is required")
	}

	resp, err := client.UpdateRecipientWithBodyWithResponse(ctx, recipientID, "application/json", bytes.NewReader(data), auth)
	if err != nil {
		return fmt.Errorf("updating recipient: %w", err)
	}

	if err := api.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}

	detail, err := unmarshalRecipient(resp.Body)
	if err != nil {
		return err
	}

	return writeRecipientDetail(opts, detail)
}

func hasAnyFlag(cmd *cobra.Command, names ...string) bool {
	for _, n := range names {
		if cmd.Flags().Changed(n) {
			return true
		}
	}
	return false
}

func applyRecipientFlags(cmd *cobra.Command, current map[string]any, recipientType, target, channel, integrationKey, name, url string) {
	if cmd.Flags().Changed("type") {
		current["type"] = recipientType
	}

	details, _ := current["details"].(map[string]any)
	if details == nil {
		details = map[string]any{}
		current["details"] = details
	}

	if cmd.Flags().Changed("target") {
		details["email_address"] = target
	}
	if cmd.Flags().Changed("channel") {
		details["slack_channel"] = channel
	}
	if cmd.Flags().Changed("integration-key") {
		details["pagerduty_integration_key"] = integrationKey
	}
	if cmd.Flags().Changed("name") {
		t, _ := current["type"].(string)
		switch t {
		case "pagerduty":
			details["pagerduty_integration_name"] = name
		default:
			details["webhook_name"] = name
		}
	}
	if cmd.Flags().Changed("url") {
		details["webhook_url"] = url
	}

	// Remove read-only fields
	delete(current, "id")
	delete(current, "created_at")
	delete(current, "updated_at")
}
