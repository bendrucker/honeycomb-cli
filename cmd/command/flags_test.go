package command

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAnyChanged(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		names []string
		want  bool
	}{
		{
			name:  "no flags changed",
			args:  nil,
			names: []string{"name", "description"},
			want:  false,
		},
		{
			name:  "one named flag changed",
			args:  []string{"--name", "foo"},
			names: []string{"name", "description"},
			want:  true,
		},
		{
			name:  "another named flag changed",
			args:  []string{"--description", "bar"},
			names: []string{"name", "description"},
			want:  true,
		},
		{
			name:  "changed flag not in names",
			args:  []string{"--other", "baz"},
			names: []string{"name", "description"},
			want:  false,
		},
		{
			name:  "no names given",
			args:  []string{"--name", "foo"},
			names: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error { return nil }}
			var name, description, other string
			cmd.Flags().StringVar(&name, "name", "", "")
			cmd.Flags().StringVar(&description, "description", "", "")
			cmd.Flags().StringVar(&other, "other", "", "")

			cmd.SetArgs(tt.args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute: %v", err)
			}

			if got := AnyChanged(cmd, tt.names...); got != tt.want {
				t.Errorf("AnyChanged(%v) = %v, want %v", tt.names, got, tt.want)
			}
		})
	}
}
