package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/fields"
	"github.com/bendrucker/honeycomb-cli/internal/jq"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
)

type callOptions struct {
	root    *options.RootOptions
	token   *string
	factory clientFactory

	fieldFlags []string
	typedFlags []string
	input      string
	jqExpr     string
}

func newCallCmd(opts *options.RootOptions, token *string, factory clientFactory) *cobra.Command {
	o := &callOptions{root: opts, token: token, factory: factory}

	cmd := &cobra.Command{
		Use:   "call <tool-name>",
		Short: "Call an MCP tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCall(cmd, o, args[0])
		},
	}

	cmd.Flags().StringArrayVarP(&o.fieldFlags, "field", "f", nil, "String field: key=value (repeatable; not a file)")
	cmd.Flags().StringArrayVarP(&o.typedFlags, "typed-field", "F", nil, "Typed field: bool/number/null coercion, @file")
	cmd.Flags().StringVar(&o.input, "input", "", "Read full arguments JSON from a file (- for stdin)")
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

	ctx := cmd.Context()
	c, err := connect(ctx, o.root, derefToken(o.token), o.factory)
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

type callResult struct {
	Content []mcp.Content `json:"content"`
	IsError bool          `json:"is_error"`
}

func writeCallResult(o *callOptions, result *mcp.CallToolResult) error {
	ios := o.root.IOStreams

	out := callResult{Content: result.Content, IsError: result.IsError}

	if o.jqExpr != "" {
		b, err := json.Marshal(out)
		if err != nil {
			return fmt.Errorf("encoding result: %w", err)
		}
		return jq.Filter(bytes.NewReader(b), ios.Out, o.jqExpr)
	}

	td := output.DynamicTableDef{
		Headers: []string{"Index", "Type", "Content"},
		Rows:    make([][]string, len(out.Content)),
	}
	for i, c := range out.Content {
		td.Rows[i] = []string{strconv.Itoa(i), contentType(c), contentSummary(c)}
	}

	return o.root.OutputWriter().WriteDynamic(out, td)
}

// contentType returns the MCP content type discriminator for a content item,
// falling back to the Go type name for content that carries no type field.
func contentType(c mcp.Content) string {
	switch v := c.(type) {
	case mcp.TextContent:
		return v.Type
	case mcp.ImageContent:
		return v.Type
	case mcp.AudioContent:
		return v.Type
	case mcp.ResourceLink:
		return v.Type
	case mcp.EmbeddedResource:
		return v.Type
	default:
		return fmt.Sprintf("%T", c)
	}
}

// contentSummary renders a single-line summary of a content item for table
// output. Text content shows its text; binary and resource content show a
// bracketed descriptor (e.g. "[image image/png]") so non-text results are
// visible instead of silently dropped.
func contentSummary(c mcp.Content) string {
	switch v := c.(type) {
	case mcp.TextContent:
		return v.Text
	case mcp.ImageContent:
		return fmt.Sprintf("[image %s]", v.MIMEType)
	case mcp.AudioContent:
		return fmt.Sprintf("[audio %s]", v.MIMEType)
	case mcp.ResourceLink:
		return fmt.Sprintf("[resource_link %s]", v.URI)
	case mcp.EmbeddedResource:
		return fmt.Sprintf("[resource %s]", embeddedResourceURI(v))
	default:
		return fmt.Sprintf("[%s]", contentType(c))
	}
}

func embeddedResourceURI(r mcp.EmbeddedResource) string {
	switch res := r.Resource.(type) {
	case mcp.TextResourceContents:
		return res.URI
	case mcp.BlobResourceContents:
		return res.URI
	default:
		return ""
	}
}
