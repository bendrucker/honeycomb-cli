package cmd

import (
	"strings"
	"testing"

	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/spf13/cobra"
)

// coveredResources are the top-level resource commands whose leaf subcommands
// must carry usage examples. Resources outside this set may add examples later
// but are not asserted here.
var coveredResources = map[string]bool{
	"auth":    true,
	"query":   true,
	"dataset": true,
	"board":   true,
	"trigger": true,
	"key":     true,
}

// TestLeafCommandsHaveExamples walks the command tree and asserts that every
// leaf command (one with RunE set and no subcommands) under a covered resource
// has a non-empty Example.
func TestLeafCommandsHaveExamples(t *testing.T) {
	root := NewRootCmd(iostreams.Test(t).IOStreams)

	for _, resource := range root.Commands() {
		if !coveredResources[resource.Name()] {
			continue
		}

		walkLeaves(resource, func(leaf *cobra.Command) {
			t.Run(commandPath(leaf), func(t *testing.T) {
				if strings.TrimSpace(leaf.Example) == "" {
					t.Errorf("leaf command %q has no Example", commandPath(leaf))
				}
			})
		})
	}
}

func walkLeaves(cmd *cobra.Command, fn func(*cobra.Command)) {
	subcommands := cmd.Commands()
	if cmd.RunE != nil && len(subcommands) == 0 {
		fn(cmd)
		return
	}
	for _, sub := range subcommands {
		walkLeaves(sub, fn)
	}
}

func commandPath(cmd *cobra.Command) string {
	var parts []string
	for c := cmd; c != nil; c = c.Parent() {
		parts = append([]string{c.Name()}, parts...)
	}
	return strings.Join(parts, " ")
}
