package auth

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

type profileEntry struct {
	Name   string   `json:"name"`
	Active bool     `json:"active"`
	Keys   []string `json:"keys"`
	Team   string   `json:"team,omitempty"`
}

var profileTable = output.TableDef{
	Columns: []output.Column{
		{Header: "Name", Value: func(v any) string { return v.(profileEntry).Name }},
		{Header: "Active", Value: func(v any) string {
			if v.(profileEntry).Active {
				return "*"
			}
			return ""
		}},
		{Header: "Keys", Value: func(v any) string { return strings.Join(v.(profileEntry).Keys, ", ") }},
		{Header: "Team", Value: func(v any) string { return v.(profileEntry).Team }},
	},
}

func NewProfileCmd(opts *options.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage authentication profiles",
	}

	cmd.AddCommand(newProfileListCmd(opts))

	return cmd
}

func newProfileListCmd(opts *options.RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured profiles",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runProfileList(opts)
		},
	}
}

func runProfileList(opts *options.RootOptions) error {
	active := opts.ActiveProfile()
	profiles := discoverProfiles(opts.Config, active)

	var entries []profileEntry
	for _, name := range profiles {
		entry := profileEntry{
			Name:   name,
			Active: name == active,
		}

		for _, kt := range keyTypes {
			_, err := config.GetKey(name, kt)
			if errors.Is(err, keyring.ErrNotFound) {
				continue
			}
			if err != nil {
				return fmt.Errorf("reading %s key for profile %q: %w", kt, name, err)
			}
			entry.Keys = append(entry.Keys, string(kt))
		}

		if opts.Config != nil {
			if p, ok := opts.Config.Profiles[name]; ok {
				entry.Team = p.Team
			}
		}

		if len(entry.Keys) == 0 && entry.Team == "" {
			continue
		}

		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no profiles configured (run honeycomb auth login)")
	}

	return opts.OutputWriterList().Write(entries, profileTable)
}

func discoverProfiles(cfg *config.Config, active string) []string {
	seen := map[string]bool{active: true}
	var rest []string

	if cfg != nil {
		for name := range cfg.Profiles {
			if seen[name] {
				continue
			}
			seen[name] = true
			rest = append(rest, name)
		}
	}

	slices.Sort(rest)
	return append([]string{active}, rest...)
}
