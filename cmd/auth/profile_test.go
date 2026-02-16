package auth

import (
	"encoding/json"
	"testing"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/bendrucker/honeycomb-cli/internal/output"
)

func TestProfileList(t *testing.T) {
	for _, tc := range []struct {
		name    string
		config  *config.Config
		profile string
		keys    map[string]map[config.KeyType]string
		want    []profileEntry
	}{
		{
			name:   "default profile with keys",
			config: &config.Config{},
			keys: map[string]map[config.KeyType]string{
				"default": {config.KeyConfig: "cfg-secret"},
			},
			want: []profileEntry{
				{Name: "default", Active: true, Keys: []string{"config"}},
			},
		},
		{
			name: "multiple profiles",
			config: &config.Config{
				Profiles: map[string]*config.Profile{
					"default": {Team: "my-team"},
					"staging": {Team: "staging-team"},
				},
			},
			keys: map[string]map[config.KeyType]string{
				"default": {config.KeyConfig: "cfg", config.KeyManagement: "mgmt"},
				"staging": {config.KeyConfig: "stg"},
			},
			want: []profileEntry{
				{Name: "default", Active: true, Keys: []string{"config", "management"}, Team: "my-team"},
				{Name: "staging", Keys: []string{"config"}, Team: "staging-team"},
			},
		},
		{
			name:    "non-default active profile",
			config:  &config.Config{ActiveProfile: "work"},
			profile: "work",
			keys: map[string]map[config.KeyType]string{
				"work": {config.KeyConfig: "work-key"},
			},
			want: []profileEntry{
				{Name: "work", Active: true, Keys: []string{"config"}},
			},
		},
		{
			name: "active profile from flag",
			config: &config.Config{
				Profiles: map[string]*config.Profile{
					"default": {Team: "default-team"},
					"other":   {},
				},
			},
			profile: "other",
			keys: map[string]map[config.KeyType]string{
				"other":   {config.KeyIngest: "ingest-key"},
				"default": {config.KeyConfig: "cfg"},
			},
			want: []profileEntry{
				{Name: "other", Active: true, Keys: []string{"ingest"}},
				{Name: "default", Keys: []string{"config"}, Team: "default-team"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for profile, keys := range tc.keys {
				for kt, val := range keys {
					if err := config.SetKey(profile, kt, val); err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = config.DeleteKey(profile, kt) })
				}
			}

			ts := iostreams.Test(t)
			opts := &options.RootOptions{
				IOStreams: ts.IOStreams,
				Config:    tc.config,
				Format:    output.FormatJSON,
				Profile:   tc.profile,
			}

			if err := runProfileList(opts); err != nil {
				t.Fatal(err)
			}

			var got []profileEntry
			if err := json.Unmarshal(ts.OutBuf.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal output: %v", err)
			}

			if len(got) != len(tc.want) {
				t.Fatalf("got %d entries, want %d", len(got), len(tc.want))
			}

			for i, want := range tc.want {
				g := got[i]
				if g.Name != want.Name {
					t.Errorf("[%d] name = %q, want %q", i, g.Name, want.Name)
				}
				if g.Active != want.Active {
					t.Errorf("[%d] active = %v, want %v", i, g.Active, want.Active)
				}
				if g.Team != want.Team {
					t.Errorf("[%d] team = %q, want %q", i, g.Team, want.Team)
				}

				gotKeys := g.Keys
				if gotKeys == nil {
					gotKeys = []string{}
				}
				wantKeys := want.Keys
				if wantKeys == nil {
					wantKeys = []string{}
				}
				if len(gotKeys) != len(wantKeys) {
					t.Errorf("[%d] keys = %v, want %v", i, gotKeys, wantKeys)
					continue
				}
				for j := range wantKeys {
					if gotKeys[j] != wantKeys[j] {
						t.Errorf("[%d] keys[%d] = %q, want %q", i, j, gotKeys[j], wantKeys[j])
					}
				}
			}
		})
	}
}

func TestProfileList_NoProfiles(t *testing.T) {
	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		Format:    output.FormatJSON,
	}

	err := runProfileList(opts)
	if err == nil {
		t.Fatal("expected error for no profiles")
	}
	want := "no profiles configured (run honeycomb auth login)"
	if err.Error() != want {
		t.Errorf("got error %q, want %q", err.Error(), want)
	}
}
