package recipient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/command"
	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/spf13/cobra"
)

var recipientTypes = []string{"email", "slack", "pagerduty", "msteams", "msteams_workflow", "webhook"}

// detailFlags lists every type-specific detail flag and the recipient types it
// applies to. Flags set for a type not listed here are rejected before the body
// is built, so they cannot be silently dropped.
var detailFlags = map[string][]string{
	"target":          {"email"},
	"channel":         {"slack"},
	"integration-key": {"pagerduty"},
	"name":            {"pagerduty", "msteams", "msteams_workflow", "webhook"},
	"url":             {"msteams", "msteams_workflow", "webhook"},
}

func NewCreateCmd(opts *options.RootOptions) *cobra.Command {
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
		Use:   "create",
		Short: "Create a recipient",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if file != "" {
				return createFromFile(cmd.Context(), opts, file)
			}
			if recipientType != "" {
				if err := validateDetailFlags(cmd, recipientType); err != nil {
					return err
				}
			}
			return runCreate(cmd.Context(), opts, recipientType, target, channel, integrationKey, name, url)
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

// validateDetailFlags rejects any explicitly-set detail flag that does not
// apply to recipientType, so mismatched flags surface an error instead of being
// silently dropped.
func validateDetailFlags(cmd *cobra.Command, recipientType string) error {
	for flag, types := range detailFlags {
		if !cmd.Flags().Changed(flag) {
			continue
		}
		if slices.Contains(types, recipientType) {
			continue
		}
		return fmt.Errorf("--%s is not valid for %s recipients (valid for: %s)", flag, recipientType, strings.Join(types, ", "))
	}
	return nil
}

func runCreate(ctx context.Context, opts *options.RootOptions, recipientType, target, channel, integrationKey, name, url string) error {
	recipientType, err := command.Resolve(opts.IOStreams, recipientType, command.Field{
		Prompt:            "Type (email, slack, pagerduty, msteams, msteams_workflow, webhook): ",
		Required:          true,
		Choices:           recipientTypes,
		NonInteractiveErr: fmt.Errorf("--type or --file is required in non-interactive mode"),
	})
	if err != nil {
		return err
	}

	body, err := buildRecipientBody(opts, recipientType, target, channel, integrationKey, name, url)
	if err != nil {
		return err
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding recipient: %w", err)
	}

	return sendCreate(ctx, opts, data)
}

func buildRecipientBody(opts *options.RootOptions, recipientType, target, channel, integrationKey, name, url string) (map[string]any, error) {
	details := map[string]any{}
	var err error

	switch recipientType {
	case "email":
		target, err = command.Resolve(opts.IOStreams, target, command.Field{
			Prompt:            "Email address: ",
			Required:          true,
			NonInteractiveErr: fmt.Errorf("--target is required for email recipients in non-interactive mode"),
			EmptyErr:          fmt.Errorf("email address is required"),
		})
		if err != nil {
			return nil, err
		}
		details["email_address"] = target

	case "slack":
		channel, err = command.Resolve(opts.IOStreams, channel, command.Field{
			Prompt:            "Slack channel: ",
			Required:          true,
			NonInteractiveErr: fmt.Errorf("--channel is required for slack recipients in non-interactive mode"),
			EmptyErr:          fmt.Errorf("slack channel is required"),
		})
		if err != nil {
			return nil, err
		}
		details["slack_channel"] = channel

	case "pagerduty":
		integrationKey, err = command.Resolve(opts.IOStreams, integrationKey, command.Field{
			Prompt:            "Integration key: ",
			Required:          true,
			NonInteractiveErr: fmt.Errorf("--integration-key is required for pagerduty recipients in non-interactive mode"),
			EmptyErr:          fmt.Errorf("integration key is required"),
		})
		if err != nil {
			return nil, err
		}
		details["pagerduty_integration_key"] = integrationKey
		name, err = command.Resolve(opts.IOStreams, name, command.Field{
			Prompt: "Integration name (optional): ",
		})
		if err != nil {
			return nil, err
		}
		if name != "" {
			details["pagerduty_integration_name"] = name
		}

	case "msteams", "msteams_workflow":
		url, err = command.Resolve(opts.IOStreams, url, command.Field{
			Prompt:            "Webhook URL: ",
			Required:          true,
			NonInteractiveErr: fmt.Errorf("--url is required for %s recipients in non-interactive mode", recipientType),
			EmptyErr:          fmt.Errorf("webhook URL is required"),
		})
		if err != nil {
			return nil, err
		}
		details["webhook_url"] = url
		name, err = command.Resolve(opts.IOStreams, name, command.Field{
			Prompt: "Webhook name (optional): ",
		})
		if err != nil {
			return nil, err
		}
		if name != "" {
			details["webhook_name"] = name
		}

	case "webhook":
		url, err = command.Resolve(opts.IOStreams, url, command.Field{
			Prompt:            "Webhook URL: ",
			Required:          true,
			NonInteractiveErr: fmt.Errorf("--url is required for webhook recipients in non-interactive mode"),
			EmptyErr:          fmt.Errorf("webhook URL is required"),
		})
		if err != nil {
			return nil, err
		}
		details["webhook_url"] = url
		name, err = command.Resolve(opts.IOStreams, name, command.Field{
			Prompt:            "Webhook name: ",
			Required:          true,
			NonInteractiveErr: fmt.Errorf("--name is required for webhook recipients in non-interactive mode"),
			EmptyErr:          fmt.Errorf("webhook name is required"),
		})
		if err != nil {
			return nil, err
		}
		details["webhook_name"] = name

	default:
		return nil, fmt.Errorf("unsupported recipient type: %s", recipientType)
	}

	return map[string]any{
		"type":    recipientType,
		"details": details,
	}, nil
}

func createFromFile(ctx context.Context, opts *options.RootOptions, file string) error {
	data, err := command.ReadDefinitionFile(opts.IOStreams, file)
	if err != nil {
		return err
	}

	return sendCreate(ctx, opts, data)
}

func sendCreate(ctx context.Context, opts *options.RootOptions, data []byte) error {
	client, err := opts.ClientFor(nil, options.AuthConfig)
	if err != nil {
		return err
	}

	resp, err := client.CreateRecipientWithBodyWithResponse(ctx, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating recipient: %w", err)
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
