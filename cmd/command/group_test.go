package command

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTestGroup(ran *bool) *cobra.Command {
	parent := Group(&cobra.Command{Use: "parent", Short: "Parent group"})
	parent.AddCommand(&cobra.Command{
		Use: "child",
		RunE: func(_ *cobra.Command, _ []string) error {
			*ran = true
			return nil
		},
	})
	return parent
}

func TestGroup(t *testing.T) {
	for _, tc := range []struct {
		name      string
		args      []string
		wantErr   bool
		wantChild bool
		wantHelp  bool
	}{
		{
			name:     "bare parent prints help and errors",
			args:     []string{},
			wantErr:  true,
			wantHelp: true,
		},
		{
			name:    "unknown subcommand errors",
			args:    []string{"bogus"},
			wantErr: true,
		},
		{
			name:      "valid subcommand runs",
			args:      []string{"child"},
			wantErr:   false,
			wantChild: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var ran bool
			parent := newTestGroup(&ran)

			var out bytes.Buffer
			parent.SetOut(&out)
			parent.SetErr(&out)
			parent.SetArgs(tc.args)

			err := parent.Execute()
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ran != tc.wantChild {
				t.Errorf("child ran = %v, want %v", ran, tc.wantChild)
			}
			if tc.wantHelp && !strings.Contains(out.String(), "Parent group") {
				t.Errorf("expected help output, got %q", out.String())
			}
		})
	}
}
