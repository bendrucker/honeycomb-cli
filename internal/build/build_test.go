package build

import "testing"

func TestString(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version string
		commit  string
		date    string
		want    string
	}{
		{
			name:    "explicit version",
			version: "v1.2.3",
			want:    "v1.2.3",
		},
		{
			name:    "version with commit and date",
			version: "v1.2.3",
			commit:  "abc1234",
			date:    "2026-05-31",
			want:    "v1.2.3 (abc1234) 2026-05-31",
		},
		{
			name:    "dev defaults omit commit and date",
			version: "dev",
			commit:  "none",
			date:    "unknown",
			want:    "dev",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := String(tc.version, tc.commit, tc.date); got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
