package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "drduck",
	Short: "DrDuck - DocOps CLI tool for automated documentation workflows",
	Long: `DrDuck is a CLI tool that integrates with AI coding assistants (Claude Code CLI, Cursor) 
to automate the creation and management of Architectural Decision Records (ADRs) and other 
documentation following DocOps principles.`,
	Version: "0.1.0",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("DrDuck version {{.Version}}\n")
}