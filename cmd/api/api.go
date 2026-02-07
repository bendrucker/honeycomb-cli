package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
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
}

func NewCmd(opts *options.RootOptions) *cobra.Command {
	o := &apiOptions{root: opts}

	cmd := &cobra.Command{
		Use:   "api <path>",
		Short: "Make an authenticated API request",
		Long:  "Make an authenticated API request to Honeycomb and print the response.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, o, args[0])
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

	return cmd
}

func run(cmd *cobra.Command, o *apiOptions, path string) error {
	fields, err := parseFields(o.fields, o.typed, o.root.IOStreams.In)
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

	profile := o.root.ActiveProfile()
	key, err := config.GetKey(profile, kt)
	if err == keyring.ErrNotFound {
		return fmt.Errorf("no %s key configured for profile %q (run honeycomb auth login --key-type %s)", kt, profile, kt)
	}
	if err != nil {
		return fmt.Errorf("reading %s key: %w", kt, err)
	}

	baseURL := o.root.ResolveAPIUrl()
	client := &http.Client{}
	ios := o.root.IOStreams

	for {
		req, err := buildRequest(method, baseURL, path, fields, body, o.headers)
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

		if o.jqExpr != "" {
			if err := filterJQ(bytes.NewReader(respBody), ios.Out, o.jqExpr); err != nil {
				return err
			}
		} else {
			_, _ = ios.Out.Write(respBody)
		}

		if resp.StatusCode >= 400 {
			return fmt.Errorf("%s %s: HTTP %d", method, path, resp.StatusCode)
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
		fields = nil
		body = nil
	}

	return nil
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
