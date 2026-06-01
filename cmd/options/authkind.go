package options

import (
	"fmt"

	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
)

// AuthKind names the credential a command authenticates with. It is the single
// command-level declaration of "this command authenticates this way", replacing
// the config.KeyType literals that were spread across every call site and the
// RequireTeam guards duplicated in each management command.
type AuthKind int

const (
	// AuthConfig uses a Configuration API key (X-Honeycomb-Team) for the v1
	// Configuration API: boards, columns, SLOs, triggers, queries, etc.
	AuthConfig AuthKind = iota
	// AuthIngest uses an Ingest API key (X-Honeycomb-Team) for sending events.
	AuthIngest
	// AuthManagement uses a Management v2 key (id:secret bearer) for the
	// Management API: environments and keys. Management commands operate on a
	// team, so this kind requires a team slug.
	AuthManagement
	// AuthMCPOAuth names the OAuth path the mcp commands use to reach the
	// OAuth-protected Honeycomb MCP server through the mcp package's OAuth
	// transport, not ClientFor. It has no keyring API key and cannot build a REST
	// client, so ClientFor rejects it and AuthKinds excludes it from the
	// login/status machinery. It exists to make the auth taxonomy complete; no
	// command passes it to ClientFor.
	AuthMCPOAuth
)

// authKindKeyType maps an AuthKind to the config.KeyType whose credential and
// transport it uses. AuthMCPOAuth has no entry: the MCP server is an OAuth
// protected resource reached through the mcp package's OAuth transport, not a
// keyring API key, so it cannot build a REST client here.
func authKindKeyType(kind AuthKind) (config.KeyType, bool) {
	switch kind {
	case AuthConfig:
		return config.KeyConfig, true
	case AuthIngest:
		return config.KeyIngest, true
	case AuthManagement:
		return config.KeyManagement, true
	default:
		return "", false
	}
}

// requiresTeam reports whether a kind operates on a single team and therefore
// needs a team slug resolved before a request can be issued.
func (k AuthKind) requiresTeam() bool {
	return k == AuthManagement
}

// AuthKinds returns every AuthKind that authenticates with a stored API key, in
// the order auth status reports them. AuthMCPOAuth is excluded: its credential
// is an OAuth token set, not a verifiable API key, and stays out of the
// login/status machinery.
func AuthKinds() []AuthKind {
	return []AuthKind{AuthConfig, AuthIngest, AuthManagement}
}

// KeyType returns the config.KeyType backing a key-based AuthKind. It is only
// valid for kinds with a keyring-stored credential (everything except
// AuthMCPOAuth) and is used by auth status to enumerate the credentials it
// verifies.
func (k AuthKind) KeyType() config.KeyType {
	kt, _ := authKindKeyType(k)
	return kt
}

// ClientFor builds an API client for the declared auth kind. For
// team-scoped kinds (AuthManagement) it first resolves the team slug into
// *team, folding the per-command RequireTeam guard into one place, then bakes
// the matching credential's request editor into the client.
//
// Pass the command's --team flag pointer for team-scoped kinds; pass nil for
// kinds that do not operate on a team.
func (o *RootOptions) ClientFor(team *string, kind AuthKind) (*api.ClientWithResponses, error) {
	kt, ok := authKindKeyType(kind)
	if !ok {
		return nil, fmt.Errorf("auth kind does not use an API key")
	}

	if kind.requiresTeam() {
		if err := o.RequireTeam(team); err != nil {
			return nil, err
		}
	}

	return o.Client(kt)
}
