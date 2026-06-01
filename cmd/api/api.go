package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	clientapi "github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/fields"
	"github.com/bendrucker/honeycomb-cli/internal/jq"
	"github.com/spf13/cobra"
)

type apiOptions struct {
	root *options.RootOptions

	method   string
	fields   []string
	typed    []string
	headers  []string
	jqExpr   string
	include  bool
	paginate bool
	keyType  string
	input    string
	raw      bool
}

func NewCmd(opts *options.RootOptions) *cobra.Command {
	o := &apiOptions{root: opts}

	cmd := &cobra.Command{
		Use:   "api [method] <path>",
		Short: "Make an authenticated API request",
		Long:  "Make an authenticated API request to Honeycomb and print the response.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := parsePositionalMethod(cmd, o, args)
			if err != nil {
				return err
			}
			return run(cmd, o, path)
		},
	}

	cmd.Flags().StringVarP(&o.method, "method", "X", "", "HTTP method (default: auto-detected)")
	cmd.Flags().StringArrayVarP(&o.fields, "field", "f", nil, "String field: key=value")
	cmd.Flags().StringArrayVarP(&o.typed, "typed-field", "F", nil, "Typed field: bool/number/null coercion, @file")
	cmd.Flags().StringArrayVarP(&o.headers, "header", "H", nil, "Request header: key:value")
	cmd.Flags().StringVarP(&o.jqExpr, "jq", "q", "", "Filter response with jq expression")
	cmd.Flags().BoolVarP(&o.include, "include", "i", false, "Show response status line and headers")
	cmd.Flags().BoolVar(&o.paginate, "paginate", false, "Follow Link rel=\"next\" pagination")
	cmd.Flags().StringVar(&o.keyType, "key-type", "", "Override auth key type: config, ingest, management")
	cmd.Flags().StringVar(&o.input, "input", "", "Read body from file (- for stdin)")
	cmd.Flags().BoolVar(&o.raw, "raw", false, "Output the full JSON:API envelope instead of flattened attributes")

	hideFormatFlag(cmd)

	return cmd
}

// hideFormatFlag hides the inherited persistent --format flag for the api
// command, which streams responses verbatim and never applies output
// formatting. The flag is registered on the root command, so it can only be
// resolved from the inherited flag set, which cobra populates lazily. Hiding it
// in the help function keeps it out of `api --help` without disabling the flag
// for parsing.
func hideFormatFlag(cmd *cobra.Command) {
	help := cmd.HelpFunc()
	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		if c == cmd {
			if flag := c.InheritedFlags().Lookup("format"); flag != nil {
				flag.Hidden = true
			}
		}
		help(c, args)
	})
}

func run(cmd *cobra.Command, o *apiOptions, path string) error {
	f, err := fields.Parse(o.fields, o.typed, o.root.IOStreams.In)
	if err != nil {
		return err
	}

	body, cleanup, err := resolveBody(o)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	method := resolveMethod(o, body != nil)

	if o.paginate && method != http.MethodGet {
		return fmt.Errorf("--paginate is only supported with GET requests")
	}

	kt, err := resolveKeyType(o, path)
	if err != nil {
		return err
	}

	key, err := o.root.RequireKey(kt)
	if err != nil {
		return err
	}

	baseURL := o.root.ResolveAPIUrl()
	client := &http.Client{}
	ios := o.root.IOStreams

	if isV2Path(path) && body == nil && len(f) > 0 {
		switch method {
		case http.MethodPost, http.MethodPatch, http.MethodPut:
			resourceType := inferResourceType(method, path)
			f = wrapJSONAPI(f, resourceType)
		}
	}

	for {
		req, err := buildRequest(method, baseURL, path, f, body, o.headers)
		if err != nil {
			return err
		}
		req = req.WithContext(cmd.Context())
		config.ApplyAuth(req, kt, key)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}

		if o.include {
			writeResponseHeaders(ios.Err, resp)
		}

		respBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}

		if err := clientapi.CheckResponse(resp.StatusCode, respBody); err != nil {
			writeBody(ios.Err, respBody)
			return fmt.Errorf("%s %s: %w", method, path, err)
		}

		if !o.raw && isV2Path(path) {
			respBody, err = unwrapJSONAPI(respBody)
			if err != nil {
				return fmt.Errorf("unwrapping JSON:API response: %w", err)
			}
		}

		if o.jqExpr != "" {
			if err := jq.Filter(bytes.NewReader(respBody), ios.Out, o.jqExpr); err != nil {
				return err
			}
		} else {
			writeBody(ios.Out, respBody)
		}

		if !o.paginate {
			break
		}
		next := nextPageURL(resp)
		if next == "" {
			break
		}

		path = next
		baseURL = ""
		f = nil
		body = nil
	}

	return nil
}

// writeBody writes the response body to w, appending a trailing newline when
// the body lacks one. Re-marshaled JSON:API output has no newline, so this
// keeps it from running into the next shell prompt; bodies that already end in
// a newline (v1 responses, --raw passthrough) are written unchanged.
func writeBody(w io.Writer, body []byte) {
	_, _ = w.Write(body)
	if len(body) > 0 && body[len(body)-1] != '\n' {
		_, _ = w.Write([]byte{'\n'})
	}
}

// parsePositionalMethod supports the gh-style `api <method> <path>` form. With a
// single positional, it is treated as the path and the method comes from -X or
// auto-detection. With two positionals, the first must be a known HTTP method,
// which sets o.method (uppercased), and the second is the path. Passing both a
// positional method and -X is a conflict.
func parsePositionalMethod(cmd *cobra.Command, o *apiOptions, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}

	method := args[0]
	if !knownMethod(method) {
		return "", fmt.Errorf("unknown HTTP method %q: expected GET, POST, PUT, PATCH, DELETE, or HEAD", method)
	}
	if cmd.Flags().Changed("method") {
		return "", fmt.Errorf("cannot set the method via both a positional argument and -X/--method")
	}

	o.method = strings.ToUpper(method)
	return args[1], nil
}

// knownMethod reports whether s is a recognized HTTP method, ignoring case.
func knownMethod(s string) bool {
	switch strings.ToUpper(s) {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead:
		return true
	default:
		return false
	}
}

func resolveMethod(o *apiOptions, hasBody bool) string {
	if o.method != "" {
		return o.method
	}
	if len(o.fields) > 0 || len(o.typed) > 0 || hasBody {
		return http.MethodPost
	}
	return http.MethodGet
}

func resolveBody(o *apiOptions) (io.Reader, func(), error) {
	if o.input == "" {
		return nil, nil, nil
	}
	if o.input == "-" {
		return o.root.IOStreams.In, nil, nil
	}
	f, err := os.Open(o.input)
	if err != nil {
		return nil, nil, fmt.Errorf("opening input file: %w", err)
	}
	return f, func() { _ = f.Close() }, nil
}

func resolveKeyType(o *apiOptions, path string) (config.KeyType, error) {
	if o.keyType != "" {
		return config.ParseKeyType(o.keyType)
	}
	return inferKeyType(path), nil
}
