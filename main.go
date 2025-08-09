package main

import (
	"github.com/SilverFlin/DrDuck/cmd"
)

// These variables are set by GoReleaser at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Set version information for the CLI
	cmd.SetVersionInfo(version, commit, date)
	cmd.Execute()
}