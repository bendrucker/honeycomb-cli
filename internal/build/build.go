// Package build formats the version metadata that GoReleaser stamps into the
// binary through its default -X main.{version,commit,date} ldflags.
package build

import "runtime/debug"

// Version resolves the version to report. A release build passes a real version
// string; otherwise it falls back to the module version from the build info
// (set for `go install module@version`), then to the supplied default.
func Version(version string) string {
	if version != "dev" {
		return version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}

	return version
}

// String renders the version, appending the commit and date when a release
// build set them. The `go build`/`go run` defaults ("none", "unknown") are
// omitted.
func String(version, commit, date string) string {
	s := Version(version)
	if commit != "" && commit != "none" {
		s += " (" + commit + ")"
	}
	if date != "" && date != "unknown" {
		s += " " + date
	}
	return s
}
