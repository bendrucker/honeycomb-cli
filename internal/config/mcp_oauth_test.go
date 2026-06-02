package config

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestMCPStoreToken(t *testing.T) {
	keyring.MockInit()

	type tokenSet struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	store := NewMCPStore("mcp-test")
	t.Cleanup(func() { _ = store.DeleteToken() })

	var empty tokenSet
	if err := store.Token(&empty); !errors.Is(err, keyring.ErrNotFound) {
		t.Fatalf("Token on empty = %v, want ErrNotFound", err)
	}

	want := tokenSet{AccessToken: "a-tok", RefreshToken: "r-tok"}
	if err := store.SetToken(want); err != nil {
		t.Fatalf("SetToken() error = %v", err)
	}

	var got tokenSet
	if err := store.Token(&got); err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if got != want {
		t.Errorf("Token() = %+v, want %+v", got, want)
	}

	if err := store.DeleteToken(); err != nil {
		t.Fatalf("DeleteToken() error = %v", err)
	}
	if err := store.Token(&got); !errors.Is(err, keyring.ErrNotFound) {
		t.Errorf("Token after delete = %v, want ErrNotFound", err)
	}
}

func TestMCPStoreCorruptToken(t *testing.T) {
	keyring.MockInit()

	store := NewMCPStore("mcp-corrupt")
	t.Cleanup(func() { _ = store.DeleteToken() })

	if err := keyring.Set(keyringService, store.tokenKey(), "not json"); err != nil {
		t.Fatalf("seeding corrupt entry: %v", err)
	}

	var got map[string]any
	if err := store.Token(&got); !errors.Is(err, ErrMCPTokenCorrupt) {
		t.Errorf("Token on corrupt entry = %v, want ErrMCPTokenCorrupt", err)
	}
}

func TestMCPStoreClientID(t *testing.T) {
	keyring.MockInit()

	store := NewMCPStore("mcp-test")
	t.Cleanup(func() { _ = store.DeleteClientID() })

	if _, err := store.ClientID(); !errors.Is(err, keyring.ErrNotFound) {
		t.Fatalf("ClientID on empty = %v, want ErrNotFound", err)
	}

	if err := store.SetClientID("client-123"); err != nil {
		t.Fatalf("SetClientID() error = %v", err)
	}

	got, err := store.ClientID()
	if err != nil {
		t.Fatalf("ClientID() error = %v", err)
	}
	if got != "client-123" {
		t.Errorf("ClientID() = %q, want %q", got, "client-123")
	}
}

func TestMCPStoreKeys(t *testing.T) {
	store := NewMCPStore("alice")
	if got := store.tokenKey(); got != "alice:mcp" {
		t.Errorf("tokenKey() = %q, want %q", got, "alice:mcp")
	}
	if got := store.clientIDKey(); got != "alice:mcp-client" {
		t.Errorf("clientIDKey() = %q, want %q", got, "alice:mcp-client")
	}
}
