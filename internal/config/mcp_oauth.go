package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

// ErrMCPTokenCorrupt indicates a stored MCP token entry exists but could not be
// decoded. Callers treat this like a missing token: the entry is unusable, so
// re-authorizing is the only recovery.
var ErrMCPTokenCorrupt = errors.New("stored MCP token is corrupt")

// MCPStore persists the OAuth credentials for the Honeycomb MCP server in the OS
// keyring, scoped to a profile. The token set and the Dynamic Client
// Registration client ID are OAuth artifacts, not raw Honeycomb credentials, so
// they live behind their own store rather than in the KeyType vault and stay out
// of the login/status/KeyTypes machinery.
type MCPStore struct {
	profile string
}

// NewMCPStore returns the OAuth credential store for a profile.
func NewMCPStore(profile string) MCPStore {
	return MCPStore{profile: profile}
}

// mcpTokenSuffix is the keyring sub-key under which the OAuth token set is
// stored, and mcpClientIDSuffix the sub-key for the DCR client ID. Persisting
// the client ID lets the refresh-token grant send a stable client_id across CLI
// invocations instead of re-registering and forcing a fresh browser flow on
// every token expiry.
const (
	mcpTokenSuffix    = "mcp"
	mcpClientIDSuffix = "mcp-client"
)

func (s MCPStore) tokenKey() string {
	return fmt.Sprintf("%s:%s", s.profile, mcpTokenSuffix)
}

func (s MCPStore) clientIDKey() string {
	return fmt.Sprintf("%s:%s", s.profile, mcpClientIDSuffix)
}

// SetToken stores the OAuth token set, JSON-encoded under the {profile}:mcp
// keyring entry.
func (s MCPStore) SetToken(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encoding MCP token: %w", err)
	}
	return withTimeout(func() error {
		return keyring.Set(keyringService, s.tokenKey(), string(data))
	})
}

// Token decodes the stored OAuth token set into v. It returns
// keyring.ErrNotFound when no token has been stored and ErrMCPTokenCorrupt when
// the stored entry cannot be decoded.
func (s MCPStore) Token(v any) error {
	var raw string
	err := withTimeout(func() error {
		var e error
		raw, e = keyring.Get(keyringService, s.tokenKey())
		return e
	})
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(raw), v); err != nil {
		return fmt.Errorf("%w: %w", ErrMCPTokenCorrupt, err)
	}
	return nil
}

// DeleteToken removes the stored OAuth token set.
func (s MCPStore) DeleteToken() error {
	return withTimeout(func() error {
		return keyring.Delete(keyringService, s.tokenKey())
	})
}

// SetClientID stores the DCR-registered OAuth client ID.
func (s MCPStore) SetClientID(clientID string) error {
	return withTimeout(func() error {
		return keyring.Set(keyringService, s.clientIDKey(), clientID)
	})
}

// ClientID returns the stored OAuth client ID, or keyring.ErrNotFound when none
// has been registered.
func (s MCPStore) ClientID() (string, error) {
	var val string
	err := withTimeout(func() error {
		var e error
		val, e = keyring.Get(keyringService, s.clientIDKey())
		return e
	})
	return val, err
}

// DeleteClientID removes the stored OAuth client ID.
func (s MCPStore) DeleteClientID() error {
	return withTimeout(func() error {
		return keyring.Delete(keyringService, s.clientIDKey())
	})
}
