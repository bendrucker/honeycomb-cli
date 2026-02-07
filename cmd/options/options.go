package options

import (
	"fmt"

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
	Profile       string
}

const defaultAPIUrl = "https://api.honeycomb.io"

func (o *RootOptions) ActiveProfile() string {
	if o.Profile != "" {
		return o.Profile
	}
	if o.Config != nil && o.Config.ActiveProfile != "" {
		return o.Config.ActiveProfile
	}
	return "default"
}

func (o *RootOptions) ResolveFormat() string {
	if o.Format != "" {
		return o.Format
	}
	if o.IOStreams.IsStdoutTTY() {
		return "table"
	}
	return "json"
}

func (o *RootOptions) RequireKey(kt config.KeyType) (string, error) {
	profile := o.ActiveProfile()
	key, err := config.GetKey(profile, kt)
	if err == keyring.ErrNotFound {
		return "", fmt.Errorf("no %s key configured for profile %q (run honeycomb auth login --key-type %s)", kt, profile, kt)
	}
	if err != nil {
		return "", fmt.Errorf("reading %s key: %w", kt, err)
	}
	return key, nil
}

func (o *RootOptions) OutputWriter() *output.Writer {
	return output.New(o.IOStreams.Out, o.ResolveFormat())
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
