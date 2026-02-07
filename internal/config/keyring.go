package config

import (
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

func SetKey(profile string, kt KeyType, value string) error {
	return withTimeout(func() error {
		return keyring.Set(keyringService, keyringKey(profile, kt), value)
	})
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
