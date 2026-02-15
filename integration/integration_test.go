//go:build integration

package integration

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var (
	binary          string
	runID           string
	team            string
	apiURL          string
	dataset         string
	configKeyID     string
	configKeySecret string
	hasEnterprise   bool
	hasPro          bool
)

func TestMain(m *testing.M) {
	mgmtKeyID := requireEnv("HONEYCOMB_MANAGEMENT_KEY_ID")
	mgmtKeySecret := requireEnv("HONEYCOMB_MANAGEMENT_KEY_SECRET")
	team = requireEnv("HONEYCOMB_TEAM")
	apiURL = os.Getenv("HONEYCOMB_API_URL")

	runID = generateRunID()
	log.Printf("run ID: %s", runID)

	bin, err := buildBinary()
	if err != nil {
		log.Fatalf("building binary: %v", err)
	}
	binary = bin

	if err := storeManagementKey(mgmtKeyID, mgmtKeySecret); err != nil {
		log.Fatalf("storing management key: %v", err)
	}

	keyID, secret, err := createConfigKey()
	if err != nil {
		log.Fatalf("creating config key: %v", err)
	}
	configKeyID = keyID
	configKeySecret = secret

	if err := storeConfigKey(secret); err != nil {
		log.Fatalf("storing config key: %v", err)
	}

	ds, err := createTestDataset()
	if err != nil {
		log.Fatalf("creating test dataset: %v", err)
	}
	dataset = ds

	probeFeatures()

	code := m.Run()

	cleanup()
	os.Exit(code)
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}

func generateRunID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("generating run ID: %v", err)
	}
	return "it-" + hex.EncodeToString(b)
}

func buildBinary() (string, error) {
	dir, err := os.MkdirTemp("", "honeycomb-integration-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}
	bin := filepath.Join(dir, "honeycomb")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/honeycomb")
	cmd.Dir = findModuleRoot()
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("go build: %w", err)
	}
	return bin, nil
}

func findModuleRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("getting working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Fatal("could not find go.mod")
		}
		dir = parent
	}
}

func storeManagementKey(keyID, keySecret string) error {
	_, err := runErr(nil,
		"auth", "login",
		"--key-type", "management",
		"--key-id", keyID,
		"--key-secret", keySecret,
		"--no-verify",
	)
	return err
}

func createConfigKey() (string, string, error) {
	body := fmt.Sprintf(`{"data":{"type":"keys","attributes":{"name":"%s","key_type":"configuration"}}}`, runID)
	r, err := runErr([]byte(body), "key", "create", "--team", team, "-f", "-")
	if err != nil {
		return "", "", fmt.Errorf("key create: %w\nstderr: %s", err, r.stderr)
	}

	var detail struct {
		ID     string `json:"id"`
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(r.stdout, &detail); err != nil {
		return "", "", fmt.Errorf("parsing key create response: %w\nstdout: %s", err, r.stdout)
	}
	return detail.ID, detail.Secret, nil
}

func storeConfigKey(secret string) error {
	_, err := runErr(nil,
		"auth", "login",
		"--key-type", "config",
		"--key-secret", secret,
		"--no-verify",
	)
	return err
}

func createTestDataset() (string, error) {
	r, err := runErr(nil, "dataset", "create", "--name", runID)
	if err != nil {
		return "", fmt.Errorf("dataset create: %w\nstderr: %s", err, r.stderr)
	}

	var ds struct {
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(r.stdout, &ds); err != nil {
		return "", fmt.Errorf("parsing dataset create response: %w\nstdout: %s", err, r.stdout)
	}
	return ds.Slug, nil
}

func probeFeatures() {
	hasEnterprise = probeEnterprise()
	hasPro = probePro()
	log.Printf("enterprise=%t pro=%t", hasEnterprise, hasPro)
}

func probeEnterprise() bool {
	queryJSON := `{"calculations":[{"op":"COUNT"}],"time_range":60}`
	_, err := runErr([]byte(queryJSON), "query", "run", "--dataset", dataset, "-f", "-")
	return err == nil
}

func probePro() bool {
	_, err := runErr(nil, "slo", "list", "--dataset", dataset)
	return err == nil
}

func cleanup() {
	if dataset != "" {
		r, err := runErr(nil, "dataset", "delete", dataset, "--yes")
		if err != nil {
			log.Printf("cleanup: deleting dataset: %v\nstderr: %s", err, r.stderr)
		}
	}

	if configKeyID != "" {
		r, err := runErr(nil, "key", "delete", configKeyID, "--team", team, "--yes")
		if err != nil {
			log.Printf("cleanup: deleting config key: %v\nstderr: %s", err, r.stderr)
		}
	}

	r, err := runErr(nil, "auth", "logout", "--profile", "integration-test")
	if err != nil {
		log.Printf("cleanup: auth logout: %v\nstderr: %s", err, r.stderr)
	}
}

func commonFlags() []string {
	flags := []string{"--no-interactive", "--profile", "integration-test", "--format", "json"}
	if apiURL != "" {
		flags = append(flags, "--api-url", apiURL)
	}
	return flags
}

func execBinary(stdin []byte, args ...string) (result, error) {
	allArgs := append(args, commonFlags()...)
	cmd := exec.Command(binary, allArgs...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return result{stdout: stdout.Bytes(), stderr: stderr.Bytes()}, err
}
