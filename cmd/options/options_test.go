package options

import (
	"testing"

	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/bendrucker/honeycomb-cli/internal/output"
)

func TestResolveFormat(t *testing.T) {
	for _, tc := range []struct {
		name   string
		format string
		tty    bool
		kind   outputKind
		want   string
	}{
		{
			name: "detail follows tty to a table",
			tty:  true,
			kind: detailOutput,
			want: output.FormatTable,
		},
		{
			name: "detail defaults to json when piped",
			tty:  false,
			kind: detailOutput,
			want: output.FormatJSON,
		},
		{
			name: "list defaults to a table on a tty",
			tty:  true,
			kind: listOutput,
			want: output.FormatTable,
		},
		{
			name: "list defaults to a table when piped",
			tty:  false,
			kind: listOutput,
			want: output.FormatTable,
		},
		{
			name:   "explicit format wins for detail",
			format: output.FormatJSON,
			tty:    true,
			kind:   detailOutput,
			want:   output.FormatJSON,
		},
		{
			name:   "explicit format wins for list",
			format: output.FormatJSON,
			tty:    false,
			kind:   listOutput,
			want:   output.FormatJSON,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var ts *iostreams.TestStreams
			if tc.tty {
				ts = iostreams.TestPromptable(t)
			} else {
				ts = iostreams.Test(t)
			}
			o := &RootOptions{IOStreams: ts.IOStreams, Format: tc.format}
			if got := o.resolveFormat(tc.kind); got != tc.want {
				t.Errorf("resolveFormat = %q, want %q", got, tc.want)
			}
		})
	}
}
