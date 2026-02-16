package recipient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/jsonutil"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/spf13/cobra"
)

var recipientTypes = []string{"email", "slack", "pagerduty", "msteams", "msteams_workflow", "webhook"}

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

func runCreate(ctx context.Context, opts *options.RootOptions, recipientType, target, channel, integrationKey, name, url string) error {
	var err error
	if recipientType == "" {
		if !opts.IOStreams.CanPrompt() {
			return fmt.Errorf("--type or --file is required in non-interactive mode")
		}
		recipientType, err = prompt.Choice(opts.IOStreams.Err, opts.IOStreams.In, "Type (email, slack, pagerduty, msteams, msteams_workflow, webhook): ", recipientTypes)
		if err != nil {
			return err
		}
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
		if target == "" {
			if !opts.IOStreams.CanPrompt() {
				return nil, fmt.Errorf("--target is required for email recipients in non-interactive mode")
			}
			target, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Email address: ")
			if err != nil {
				return nil, err
			}
		}
		if target == "" {
			return nil, fmt.Errorf("email address is required")
		}
		details["email_address"] = target

	case "slack":
		if channel == "" {
			if !opts.IOStreams.CanPrompt() {
				return nil, fmt.Errorf("--channel is required for slack recipients in non-interactive mode")
			}
			channel, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Slack channel: ")
			if err != nil {
				return nil, err
			}
		}
		if channel == "" {
			return nil, fmt.Errorf("slack channel is required")
		}
		details["slack_channel"] = channel

	case "pagerduty":
		if integrationKey == "" {
			if !opts.IOStreams.CanPrompt() {
				return nil, fmt.Errorf("--integration-key is required for pagerduty recipients in non-interactive mode")
			}
			integrationKey, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Integration key: ")
			if err != nil {
				return nil, err
			}
		}
		if integrationKey == "" {
			return nil, fmt.Errorf("integration key is required")
		}
		details["pagerduty_integration_key"] = integrationKey
		if name == "" && opts.IOStreams.CanPrompt() {
			name, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Integration name (optional): ")
			if err != nil {
				return nil, err
			}
		}
		if name != "" {
			details["pagerduty_integration_name"] = name
		}

	case "msteams", "msteams_workflow":
		if url == "" {
			if !opts.IOStreams.CanPrompt() {
				return nil, fmt.Errorf("--url is required for %s recipients in non-interactive mode", recipientType)
			}
			url, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Webhook URL: ")
			if err != nil {
				return nil, err
			}
		}
		if url == "" {
			return nil, fmt.Errorf("webhook URL is required")
		}
		details["webhook_url"] = url
		if name == "" && opts.IOStreams.CanPrompt() {
			name, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Webhook name (optional): ")
			if err != nil {
				return nil, err
			}
		}
		if name != "" {
			details["webhook_name"] = name
		}

	case "webhook":
		if url == "" {
			if !opts.IOStreams.CanPrompt() {
				return nil, fmt.Errorf("--url is required for webhook recipients in non-interactive mode")
			}
			url, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Webhook URL: ")
			if err != nil {
				return nil, err
			}
		}
		if url == "" {
			return nil, fmt.Errorf("webhook URL is required")
		}
		details["webhook_url"] = url
		if name == "" {
			if !opts.IOStreams.CanPrompt() {
				return nil, fmt.Errorf("--name is required for webhook recipients in non-interactive mode")
			}
			name, err = prompt.Line(opts.IOStreams.Err, opts.IOStreams.In, "Webhook name: ")
			if err != nil {
				return nil, err
			}
		}
		if name == "" {
			return nil, fmt.Errorf("webhook name is required")
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
	data, err := readFile(opts, file)
	if err != nil {
		return err
	}

	return sendCreate(ctx, opts, data)
}

func readFile(opts *options.RootOptions, file string) ([]byte, error) {
	var r io.Reader
	if file == "-" {
		r = opts.IOStreams.In
	} else {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	data, err = jsonutil.Sanitize(data)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return data, nil
}

func sendCreate(ctx context.Context, opts *options.RootOptions, data []byte) error {
	auth, err := opts.KeyEditor(config.KeyConfig)
	if err != nil {
		return err
	}

	client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	resp, err := client.CreateRecipientWithBodyWithResponse(ctx, "application/json", bytes.NewReader(data), auth)
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
