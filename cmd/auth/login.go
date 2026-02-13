package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/api"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/bendrucker/honeycomb-cli/internal/prompt"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

type loginResult struct {
	Type        string `json:"type"`
	Team        string `json:"team,omitempty"`
	Environment string `json:"environment,omitempty"`
	KeyID       string `json:"key_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Verified    bool   `json:"verified"`
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
		var fields []huh.Field
		if keyType == "" {
			fields = append(fields, huh.NewSelect[string]().
				Title("Key type").
				Options(
					huh.NewOption("config", "config"),
					huh.NewOption("ingest", "ingest"),
					huh.NewOption("management", "management"),
				).
				Value(&keyType))
		}
		if keyID == "" {
			fields = append(fields, huh.NewInput().
				Title("Key ID").
				Value(&keyID))
		}
		if keySecret == "" {
			fields = append(fields, huh.NewInput().
				Title("Key secret").
				EchoMode(huh.EchoModePassword).
				Value(&keySecret))
		}
		if len(fields) > 0 {
			err := huh.NewForm(huh.NewGroup(fields...)).Run()
			if err == huh.ErrUserAborted {
				return nil
			}
			if err != nil {
				return fmt.Errorf("prompting for credentials: %w", err)
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
			line, err := prompt.ReadLine(ios.In)
			if err != nil {
				return fmt.Errorf("reading key secret from stdin: %w", err)
			}
			keySecret = line
		}
	}

	kt, err := config.ParseKeyType(keyType)
	if err != nil {
		return err
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
	case output.FormatJSON:
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	case output.FormatTable:
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
