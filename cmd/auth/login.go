package auth

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

type loginResult struct {
	Type        string `json:"type" yaml:"type"`
	Team        string `json:"team,omitempty" yaml:"team,omitempty"`
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`
	KeyID       string `json:"key_id,omitempty" yaml:"key_id,omitempty"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Verified    bool   `json:"verified" yaml:"verified"`
}

func NewLoginCmd(opts *options.RootOptions) *cobra.Command {
	var (
		keyType   string
		keyID     string
		keySecret string
		verify    bool
		noVerify  bool
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Honeycomb",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Flags().Changed("no-verify") {
				verify = false
			}
			return runAuthLogin(cmd.Context(), opts, keyType, keyID, keySecret, verify)
		},
	}

	cmd.Flags().StringVar(&keyType, "key-type", "", "Key type: config, ingest, management")
	cmd.Flags().StringVar(&keyID, "key-id", "", "Key ID")
	cmd.Flags().StringVar(&keySecret, "key-secret", "", "Key secret (alternative to stdin)")
	cmd.Flags().BoolVar(&verify, "verify", true, "Verify key against the API before storing")
	cmd.Flags().BoolVar(&noVerify, "no-verify", false, "Skip API verification")
	cmd.Flags().Lookup("no-verify").Hidden = true

	return cmd
}

func runAuthLogin(ctx context.Context, opts *options.RootOptions, keyType, keyID, keySecret string, verify bool) error {
	ios := opts.IOStreams

	if ios.CanPrompt() {
		var err error
		if keyType == "" {
			keyType, err = promptChoice(ios.Out, ios.In, "Key type (config, ingest, management): ", []string{"config", "ingest", "management"})
			if err != nil {
				return fmt.Errorf("reading key type: %w", err)
			}
		}
		if keyID == "" {
			keyID, err = promptLine(ios.Out, ios.In, "Key ID: ")
			if err != nil {
				return fmt.Errorf("reading key ID: %w", err)
			}
		}
		if keySecret == "" {
			keySecret, err = promptSecret(ios.Out, ios.In, ios.StdinFd(), "Key secret: ")
			if err != nil {
				return fmt.Errorf("reading key secret: %w", err)
			}
		}
	} else {
		if keyType == "" {
			return fmt.Errorf("--key-type is required in non-interactive mode")
		}
		if keyID == "" {
			return fmt.Errorf("--key-id is required in non-interactive mode")
		}
		if keySecret == "" {
			line, err := readLine(ios.In)
			if err != nil {
				return fmt.Errorf("reading key secret from stdin: %w", err)
			}
			keySecret = line
		}
	}

	kt := config.KeyType(keyType)
	switch kt {
	case config.KeyConfig, config.KeyIngest, config.KeyManagement:
	default:
		return fmt.Errorf("invalid key type: %q (must be config, ingest, or management)", keyType)
	}

	combinedKey := keyID + ":" + keySecret

	result := loginResult{
		Type: keyType,
	}

	if verify {
		client, err := api.NewClientWithResponses(opts.ResolveAPIUrl())
		if err != nil {
			return fmt.Errorf("creating API client: %w", err)
		}

		ks, err := verifyKey(ctx, client, kt, combinedKey)
		if err != nil {
			return err
		}
		if ks.Status == "invalid" {
			return fmt.Errorf("invalid key")
		}
		if ks.Status == "error" {
			return fmt.Errorf("verifying key: %s", ks.Error)
		}

		result.KeyID = ks.KeyID
		result.Team = ks.Team
		result.Environment = ks.Environment
		result.Name = ks.Name
		result.Verified = true
	}

	profile := opts.ActiveProfile()
	if err := config.SetKey(profile, kt, combinedKey); err != nil {
		return fmt.Errorf("storing key: %w", err)
	}

	return writeLoginResult(opts, result)
}

func writeLoginResult(opts *options.RootOptions, result loginResult) error {
	out := opts.IOStreams.Out

	switch opts.ResolveFormat() {
	case "json":
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	case "yaml":
		return yaml.NewEncoder(out).Encode(result)
	case "table":
		if result.Verified {
			if result.Team != "" {
				_, _ = fmt.Fprintf(out, "Authenticated as %s", result.Team)
				if result.Environment != "" {
					_, _ = fmt.Fprintf(out, " (%s)", result.Environment)
				}
				_, _ = fmt.Fprintln(out)
			} else if result.Name != "" {
				_, _ = fmt.Fprintf(out, "Authenticated with key %q\n", result.Name)
			} else {
				_, _ = fmt.Fprintln(out, "Key verified and stored.")
			}
		} else {
			_, _ = fmt.Fprintln(out, "Key stored (unverified).")
		}
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", opts.ResolveFormat())
	}
}

func promptLine(out io.Writer, in io.Reader, prompt string) (string, error) {
	_, _ = fmt.Fprint(out, prompt)
	return readLine(in)
}

func promptChoice(out io.Writer, in io.Reader, prompt string, choices []string) (string, error) {
	for {
		line, err := promptLine(out, in, prompt)
		if err != nil {
			return "", err
		}
		for _, c := range choices {
			if strings.EqualFold(line, c) {
				return c, nil
			}
		}
		_, _ = fmt.Fprintf(out, "Invalid choice. Options: %s\n", strings.Join(choices, ", "))
	}
}

func promptSecret(out io.Writer, in io.Reader, fd uintptr, prompt string) (string, error) {
	_, _ = fmt.Fprint(out, prompt)
	if fd != 0 {
		b, err := term.ReadPassword(int(fd))
		_, _ = fmt.Fprintln(out)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return readLine(in)
}

func readLine(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		return strings.TrimRight(scanner.Text(), "\r\n"), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("unexpected end of input")
}
