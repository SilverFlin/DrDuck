package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information set by main.go
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "drduck",
	Short: "DrDuck - DocOps CLI tool for automated documentation workflows",
	Long: `DrDuck is a CLI tool that integrates with AI coding assistants (Claude Code CLI, Cursor) 
to automate the creation and management of Architectural Decision Records (ADRs) and other 
documentation following DocOps principles.`,
	Version: buildVersion,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func SetVersionInfo(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date
	rootCmd.Version = version
}

func init() {
	rootCmd.SetVersionTemplate(`DrDuck version {{.Version}}
Commit: ` + buildCommit + `
Built: ` + buildDate + `
`)
}