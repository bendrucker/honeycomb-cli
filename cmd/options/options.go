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

// RequireTeam resolves the team slug a management command operates on, writing
// the result back into *flag. Precedence:
//
//  1. An explicit --team flag is used as-is.
//  2. Otherwise the active profile's stored team is inferred, but only when
//     exactly one team is known. A management key can span multiple teams, so
//     inference is deliberately limited to the single-team case the profile
//     records; with zero (or, in the future, more than one) known teams the
//     required-flag error stands.
func (o *RootOptions) RequireTeam(flag *string) error {
	if *flag != "" {
		return nil
	}
	if team, ok := o.inferTeam(); ok {
		*flag = team
		return nil
	}
	return fmt.Errorf("--team is required (or set a single team via honeycomb auth login --team)")
}

// inferTeam returns the team slug to use when --team is unset, reporting false
// when no single team is known. The profile stores at most one team, so a
// recorded team is unambiguous; a missing team leaves the choice to the caller.
func (o *RootOptions) inferTeam() (string, bool) {
	if o.Config == nil {
		return "", false
	}
	profile := o.ActiveProfile()
	if p, ok := o.Config.Profiles[profile]; ok && p.Team != "" {
		return p.Team, true
	}
	return "", false
}

// outputKind selects how an unset --format flag resolves. Detail output follows
// the terminal: a table when interactive, JSON when piped, so scripts and agents
// get structured output by default. List output defaults to a table in both
// modes, since a collection reads best as a table.
type outputKind int

const (
	detailOutput outputKind = iota
	listOutput
)

// resolveFormat picks the concrete output format for kind. An explicit --format
// flag always wins; otherwise the per-kind default applies.
func (o *RootOptions) resolveFormat(kind outputKind) string {
	if o.Format != "" {
		return o.Format
	}
	if kind == detailOutput && !o.IOStreams.IsStdoutTTY() {
		return output.FormatJSON
	}
	return output.FormatTable
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

// Client builds an API client with auth for the given key type baked in via a
// request editor, so call sites issue requests without threading an editor
// argument through every WithResponse call.
func (o *RootOptions) Client(kt config.KeyType) (*api.ClientWithResponses, error) {
	editor, err := o.KeyEditor(kt)
	if err != nil {
		return nil, err
	}
	client, err := api.NewClientWithResponses(o.ResolveAPIUrl(), api.WithRequestEditorFn(editor))
	if err != nil {
		return nil, fmt.Errorf("creating API client: %w", err)
	}
	return client, nil
}

func (o *RootOptions) OutputWriter() *output.Writer {
	return output.New(o.IOStreams.Out, o.resolveFormat(detailOutput))
}

func (o *RootOptions) OutputWriterList() *output.Writer {
	return output.New(o.IOStreams.Out, o.resolveFormat(listOutput))
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
