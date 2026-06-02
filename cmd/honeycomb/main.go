package main

import (
	"fmt"
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd"
	"github.com/bendrucker/honeycomb-cli/internal/build"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
)

// Build metadata, stamped by GoReleaser's default -X main.* ldflags. They keep
// these defaults for `go build`/`go run`; build.Version fills in the module
// version for `go install`.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	ios := iostreams.System()
	rootCmd := cmd.NewRootCmd(ios)
	rootCmd.Version = build.String(version, commit, date)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
