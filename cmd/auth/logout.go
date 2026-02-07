package auth

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

type logoutResult struct {
	Type string `json:"type" yaml:"type"`
}

func NewLogoutCmd(opts *options.RootOptions) *cobra.Command {
	var keyType string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored authentication keys",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runAuthLogout(opts, keyType)
		},
	}

	cmd.Flags().StringVar(&keyType, "key-type", "", "Key type to remove (config, ingest, management)")

	return cmd
}

func runAuthLogout(opts *options.RootOptions, keyType string) error {
	profile := opts.ActiveProfile()

	var targets []config.KeyType
	if keyType != "" {
		kt, err := config.ParseKeyType(keyType)
		if err != nil {
			return err
		}
		targets = []config.KeyType{kt}
	} else {
		targets = keyTypes
	}

	var deleted []logoutResult
	for _, kt := range targets {
		err := config.DeleteKey(profile, kt)
		if err == keyring.ErrNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("deleting %s key: %w", kt, err)
		}
		deleted = append(deleted, logoutResult{Type: string(kt)})
	}

	if len(deleted) == 0 {
		return fmt.Errorf("no keys configured for profile %q", profile)
	}

	return opts.WriteFormatted(deleted, func(out io.Writer) error {
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "TYPE")
		for _, d := range deleted {
			_, _ = fmt.Fprintln(w, d.Type)
		}
		return w.Flush()
	})
}
