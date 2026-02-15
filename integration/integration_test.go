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
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
)

var (
	runID           string
	team            string
	apiURL          string
	environment     string
	dataset         string
	configKeyID     string
	configKeySecret string
	hasEnterprise   bool
	hasPro          bool
)

func TestMain(m *testing.M) {
	team = requireEnv("HONEYCOMB_TEAM")
	apiURL = os.Getenv("HONEYCOMB_API_URL")

	runID = generateRunID()
	log.Printf("run ID: %s", runID)

	if err := setupManagementKey(); err != nil {
		log.Fatalf("setting up management key: %v", err)
	}

	envID, err := createTestEnvironment()
	if err != nil {
		log.Fatalf("creating test environment: %v", err)
	}
	environment = envID

	keyID, secret, err := createConfigKey(environment)
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

func setupManagementKey() error {
	mgmtKeyID := os.Getenv("HONEYCOMB_MANAGEMENT_KEY_ID")
	mgmtKeySecret := os.Getenv("HONEYCOMB_MANAGEMENT_KEY_SECRET")

	if mgmtKeyID != "" && mgmtKeySecret != "" {
		log.Print("using management key from environment")
		_, err := runErr(nil,
			"auth", "login",
			"--key-type", "management",
			"--key-id", mgmtKeyID,
			"--key-secret", mgmtKeySecret,
			"--no-verify",
		)
		return err
	}

	log.Print("using management key from default profile keyring")
	key, err := config.GetKey("default", config.KeyManagement)
	if err != nil {
		return fmt.Errorf("reading management key from default profile: %w (set HONEYCOMB_MANAGEMENT_KEY_ID and HONEYCOMB_MANAGEMENT_KEY_SECRET as an alternative)", err)
	}

	return config.SetKey("integration-test", config.KeyManagement, key)
}

func createTestEnvironment() (string, error) {
	r, err := runErr(nil, "environment", "create", "--team", team, "--name", runID)
	if err != nil {
		return "", fmt.Errorf("environment create: %w\nstderr: %s", err, r.stderr)
	}

	var env struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(r.stdout, &env); err != nil {
		return "", fmt.Errorf("parsing environment create: %w\nstdout: %s", err, r.stdout)
	}
	return env.ID, nil
}

func createConfigKey(envID string) (string, string, error) {
	body, err := json.Marshal(map[string]any{
		"data": map[string]any{
			"type": "api-keys",
			"attributes": map[string]any{
				"name":     runID,
				"key_type": "configuration",
				"permissions": map[string]any{
					"create_datasets":   true,
					"manage_columns":    true,
					"manage_boards":     true,
					"manage_markers":    true,
					"manage_slos":       true,
					"manage_triggers":   true,
					"manage_recipients": true,
					"run_queries":       true,
					"send_events":       true,
				},
			},
			"relationships": map[string]any{
				"environment": map[string]any{
					"data": map[string]any{"id": envID, "type": "environments"},
				},
			},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("marshaling key create body: %w", err)
	}
	r, err := runErr(body, "key", "create", "--team", team, "-f", "-")
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
		runErr(nil, "dataset", "update", dataset, "--delete-protected=false")
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

	if environment != "" {
		runErr(nil, "environment", "update", environment, "--team", team, "--delete-protected=false")
		r, err := runErr(nil, "environment", "delete", environment, "--team", team, "--yes")
		if err != nil {
			log.Printf("cleanup: deleting environment: %v\nstderr: %s", err, r.stderr)
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

func execCmd(stdin []byte, args ...string) (result, error) {
	allArgs := append(args, commonFlags()...)

	ts := iostreams.Test()
	if stdin != nil {
		ts.InBuf.Write(stdin)
	}

	rootCmd := cmd.NewRootCmd(ts.IOStreams)
	rootCmd.SetArgs(allArgs)

	var errBuf bytes.Buffer
	rootCmd.SetErr(&errBuf)

	err := rootCmd.Execute()
	return result{stdout: ts.OutBuf.Bytes(), stderr: errBuf.Bytes()}, err
}
