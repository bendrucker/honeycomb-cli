package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "honeycomb-cli"
	keyringTimeout = 3 * time.Second
)

type KeyType string

const (
	KeyConfig     KeyType = "config"
	KeyIngest     KeyType = "ingest"
	KeyManagement KeyType = "management"
)

// KeyTypes returns every key type, in storage order. Callers that act on all
// key types (logout, status, profile listing) range over this rather than
// hand-listing the constants, so adding a key type updates them all.
func KeyTypes() []KeyType {
	return []KeyType{KeyConfig, KeyIngest, KeyManagement}
}

func ParseKeyType(s string) (KeyType, error) {
	switch s {
	case "config":
		return KeyConfig, nil
	case "ingest":
		return KeyIngest, nil
	case "management":
		return KeyManagement, nil
	default:
		return "", fmt.Errorf("invalid key type %q (must be config, ingest, or management)", s)
	}
}

func ApplyAuth(req *http.Request, kt KeyType, key string) {
	switch kt {
	case KeyManagement:
		req.Header.Set("Authorization", "Bearer "+key)
	case KeyIngest, KeyConfig:
		req.Header.Set("X-Honeycomb-Team", key)
	}
}

func keyringKey(profile string, kt KeyType) string {
	return fmt.Sprintf("%s:%s", profile, kt)
}

// mcpTokenSuffix is the keyring sub-key under which the MCP OAuth token set is
// stored. It is deliberately not a KeyType: the token is a JSON document, not a
// raw credential, so it stays out of the login/status/KeyTypes machinery.
const mcpTokenSuffix = "mcp"

func mcpTokenKey(profile string) string {
	return fmt.Sprintf("%s:%s", profile, mcpTokenSuffix)
}

// SetMCPToken stores the OAuth token set for a profile, JSON-encoded under the
// {profile}:mcp keyring entry.
func SetMCPToken(profile string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encoding MCP token: %w", err)
	}
	return withTimeout(func() error {
		return keyring.Set(keyringService, mcpTokenKey(profile), string(data))
	})
}

// GetMCPToken decodes the stored OAuth token set for a profile into v. It
// returns keyring.ErrNotFound when no token has been stored.
func GetMCPToken(profile string, v any) error {
	var raw string
	err := withTimeout(func() error {
		var e error
		raw, e = keyring.Get(keyringService, mcpTokenKey(profile))
		return e
	})
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(raw), v); err != nil {
		return fmt.Errorf("decoding MCP token: %w", err)
	}
	return nil
}

// DeleteMCPToken removes the stored OAuth token set for a profile.
func DeleteMCPToken(profile string) error {
	return withTimeout(func() error {
		return keyring.Delete(keyringService, mcpTokenKey(profile))
	})
}

func SetKey(profile string, kt KeyType, value string) error {
	return withTimeout(func() error {
		return keyring.Set(keyringService, keyringKey(profile, kt), value)
	})
}

// ManagementKey encodes a management key's ID and secret into the single
// credential string the keyring stores and ApplyAuth sends as the bearer
// token. The id:secret wire format lives only here.
func ManagementKey(id, secret string) string {
	return id + ":" + secret
}

// SetManagementKey stores a management key, encoding its ID and secret so
// callers never assemble the id:secret wire format themselves.
func SetManagementKey(profile, id, secret string) error {
	return SetKey(profile, KeyManagement, ManagementKey(id, secret))
}

func GetKey(profile string, kt KeyType) (string, error) {
	var val string
	err := withTimeout(func() error {
		var e error
		val, e = keyring.Get(keyringService, keyringKey(profile, kt))
		return e
	})
	return val, err
}

func DeleteKey(profile string, kt KeyType) error {
	return withTimeout(func() error {
		return keyring.Delete(keyringService, keyringKey(profile, kt))
	})
}

func withTimeout(fn func() error) error {
	ch := make(chan error, 1)
	go func() {
		ch <- fn()
	}()

	select {
	case err := <-ch:
		return err
	case <-time.After(keyringTimeout):
		return fmt.Errorf("keyring operation timed out after %s", keyringTimeout)
	}
}
