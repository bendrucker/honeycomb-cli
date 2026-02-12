package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/fields"
	"github.com/bendrucker/honeycomb-cli/internal/jq"
	"github.com/spf13/cobra"
)

type callOptions struct {
	root    *options.RootOptions
	factory clientFactory

	fieldFlags []string
	typedFlags []string
	input      string
	jqExpr     string
}

func newCallCmd(opts *options.RootOptions, factory clientFactory) *cobra.Command {
	o := &callOptions{root: opts, factory: factory}

	cmd := &cobra.Command{
		Use:   "call <tool-name>",
		Short: "Call an MCP tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCall(cmd, o, args[0])
		},
	}

	cmd.Flags().StringArrayVarP(&o.fieldFlags, "field", "f", nil, "String field: key=value")
	cmd.Flags().StringArrayVarP(&o.typedFlags, "typed-field", "F", nil, "Typed field: bool/number/null coercion, @file")
	cmd.Flags().StringVar(&o.input, "input", "", "Read arguments from file (- for stdin)")
	cmd.Flags().StringVarP(&o.jqExpr, "jq", "q", "", "Filter output with jq expression")

	return cmd
}

func runCall(cmd *cobra.Command, o *callOptions, toolName string) error {
	hasFields := len(o.fieldFlags) > 0 || len(o.typedFlags) > 0
	if hasFields && o.input != "" {
		return fmt.Errorf("cannot use --field/-f or --typed-field/-F with --input")
	}

	args, err := resolveArgs(o)
	if err != nil {
		return err
	}

	factory := o.factory
	if factory == nil {
		factory = defaultClientFactory
	}

	key, err := o.root.RequireKey(config.KeyConfig)
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	c, err := factory(ctx, o.root.ResolveMCPUrl(), key)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()

	request := mcp.CallToolRequest{}
	request.Params.Name = toolName
	request.Params.Arguments = args

	result, err := c.CallTool(ctx, request)
	if err != nil {
		return fmt.Errorf("calling tool %q: %w", toolName, err)
	}

	if result.IsError {
		var msgs []string
		for _, c := range result.Content {
			if tc, ok := mcp.AsTextContent(c); ok {
				msgs = append(msgs, tc.Text)
			}
		}
		if len(msgs) > 0 {
			return fmt.Errorf("tool %q returned error: %s", toolName, strings.Join(msgs, "\n"))
		}
		return fmt.Errorf("tool %q returned error", toolName)
	}

	return writeCallResult(o, result)
}

func resolveArgs(o *callOptions) (map[string]any, error) {
	if o.input != "" {
		return readInputArgs(o.input, o.root.IOStreams.In)
	}
	if len(o.fieldFlags) > 0 || len(o.typedFlags) > 0 {
		return fields.Parse(o.fieldFlags, o.typedFlags, o.root.IOStreams.In)
	}
	return nil, nil
}

func readInputArgs(path string, stdin io.Reader) (map[string]any, error) {
	var r io.Reader
	if path == "-" {
		r = stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("opening input file: %w", err)
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	var args map[string]any
	if err := json.NewDecoder(r).Decode(&args); err != nil {
		return nil, fmt.Errorf("parsing input JSON: %w", err)
	}
	return args, nil
}

func writeCallResult(o *callOptions, result *mcp.CallToolResult) error {
	ios := o.root.IOStreams
	format := o.root.ResolveFormat()

	if format == "json" || format == "yaml" || o.jqExpr != "" {
		type contentItem struct {
			Type string `json:"type"           yaml:"type"`
			Text string `json:"text,omitempty" yaml:"text,omitempty"`
		}
		type callResult struct {
			Content []contentItem `json:"content" yaml:"content"`
			IsError bool          `json:"isError" yaml:"isError"`
		}

		out := callResult{IsError: result.IsError}
		for _, c := range result.Content {
			if tc, ok := mcp.AsTextContent(c); ok {
				out.Content = append(out.Content, contentItem{Type: "text", Text: tc.Text})
			}
		}

		if o.jqExpr != "" {
			b, err := json.Marshal(out)
			if err != nil {
				return fmt.Errorf("encoding result: %w", err)
			}
			return jq.Filter(bytes.NewReader(b), ios.Out, o.jqExpr)
		}

		return o.root.OutputWriter().WriteValue(out, func(w io.Writer) error {
			return nil
		})
	}

	for _, c := range result.Content {
		if tc, ok := mcp.AsTextContent(c); ok {
			_, _ = fmt.Fprintln(ios.Out, tc.Text)
		}
	}
	return nil
}
