package options

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/zalando/go-keyring"
)

type RootOptions struct {
	IOStreams *iostreams.IOStreams
	Config    *config.Config

	NoInteractive bool
	Format        string
	APIUrl        string
	MCPUrl        string
	Profile       string
	ConfigPath    string
}

const defaultAPIUrl = "https://api.honeycomb.io"
const defaultMCPUrl = "https://mcp.honeycomb.io/mcp"

func (o *RootOptions) ActiveProfile() string {
	if o.Profile != "" {
		return o.Profile
	}
	if o.Config != nil && o.Config.ActiveProfile != "" {
		return o.Config.ActiveProfile
	}
	return "default"
}

func (o *RootOptions) ResolveConfigPath() string {
	if o.ConfigPath != "" {
		return o.ConfigPath
	}
	return config.DefaultPath()
}

func (o *RootOptions) RequireTeam(flag *string) error {
	if *flag != "" {
		return nil
	}
	if o.Config != nil {
		profile := o.ActiveProfile()
		if p, ok := o.Config.Profiles[profile]; ok && p.Team != "" {
			*flag = p.Team
			return nil
		}
	}
	return fmt.Errorf("--team is required (or set via honeycomb auth login --team)")
}

func (o *RootOptions) ResolveFormat() string {
	if o.Format != "" {
		return o.Format
	}
	if o.IOStreams.IsStdoutTTY() {
		return output.FormatTable
	}
	return output.FormatJSON
}

func (o *RootOptions) RequireKey(kt config.KeyType) (string, error) {
	profile := o.ActiveProfile()
	key, err := config.GetKey(profile, kt)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", fmt.Errorf("no %s key configured for profile %q (run honeycomb auth login --key-type %s)", kt, profile, kt)
	}
	if err != nil {
		return "", fmt.Errorf("reading %s key: %w", kt, err)
	}
	return key, nil
}

func (o *RootOptions) KeyEditor(kt config.KeyType) (api.RequestEditorFn, error) {
	key, err := o.RequireKey(kt)
	if err != nil {
		return nil, err
	}
	return func(_ context.Context, req *http.Request) error {
		config.ApplyAuth(req, kt, key)
		return nil
	}, nil
}

func (o *RootOptions) OutputWriter() *output.Writer {
	return output.New(o.IOStreams.Out, o.ResolveFormat())
}

func (o *RootOptions) OutputWriterList() *output.Writer {
	format := o.Format
	if format == "" {
		format = output.FormatTable
	}
	return output.New(o.IOStreams.Out, format)
}

func (o *RootOptions) ResolveMCPUrl() string {
	if o.MCPUrl != "" {
		return o.MCPUrl
	}
	if o.Config != nil {
		profile := o.ActiveProfile()
		if p, ok := o.Config.Profiles[profile]; ok && p.MCPUrl != "" {
			return p.MCPUrl
		}
		if o.Config.MCPUrl != "" {
			return o.Config.MCPUrl
		}
	}
	return defaultMCPUrl
}

func (o *RootOptions) ResolveAPIUrl() string {
	if o.APIUrl != "" {
		return o.APIUrl
	}
	if o.Config != nil {
		profile := o.ActiveProfile()
		if p, ok := o.Config.Profiles[profile]; ok && p.APIUrl != "" {
			return p.APIUrl
		}
		if o.Config.APIUrl != "" {
			return o.Config.APIUrl
		}
	}
	return defaultAPIUrl
}
