package main

import (
	"os"

	"github.com/bendrucker/honeycomb-cli/cmd"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
)

func main() {
	ios := iostreams.System()
	rootCmd := cmd.NewRootCmd(ios)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
