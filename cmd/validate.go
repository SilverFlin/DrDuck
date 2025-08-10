package cmd

import (
	"fmt"

	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/SilverFlin/DrDuck/internal/hooks"
	"github.com/spf13/cobra"
)

var (
	preCommitFlag bool
	prePushFlag   bool
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate ADR requirements for current changes",
	Long: `Validate ADR requirements by checking for draft ADRs and analyzing current changes.
This command previews what the git hooks will check without making any commits or pushes.

Examples:
  drduck validate                # General validation
  drduck validate --pre-commit   # Preview pre-commit validation
  drduck validate --pre-push     # Preview pre-push validation`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVar(&preCommitFlag, "pre-commit", false, "Run pre-commit validation")
	validateCmd.Flags().BoolVar(&prePushFlag, "pre-push", false, "Run pre-push validation")
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Check if project is initialized
	initialized, err := config.IsInitialized()
	if err != nil {
		return fmt.Errorf("failed to check initialization status: %w", err)
	}

	if !initialized {
		return fmt.Errorf("❌ DrDuck is not initialized in this project. Run 'drduck init' first")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create validator
	validator := hooks.NewValidator(cfg)

	// Determine which validation to run
	switch {
	case preCommitFlag:
		return runPreCommitValidation(validator)
	case prePushFlag:
		return runPrePushValidation(validator)
	default:
		return runGeneralValidation(validator)
	}
}

func runPreCommitValidation(validator *hooks.Validator) error {
	fmt.Println("🦆 DrDuck: Running pre-commit validation preview...")
	fmt.Println()

	result := validator.ValidatePreCommit()
	fmt.Println(result.Message)

	// Pre-commit validation never fails
	return nil
}

func runPrePushValidation(validator *hooks.Validator) error {
	fmt.Println("🦆 DrDuck: Running pre-push validation preview...")
	fmt.Println()

	result := validator.ValidatePrePush()
	fmt.Println(result.Message)

	if result.ShouldBlock {
		fmt.Println()
		fmt.Println("🚫 This validation would block a push. Use --no-verify to bypass.")
		return fmt.Errorf("validation failed")
	}

	return nil
}

func runGeneralValidation(validator *hooks.Validator) error {
	fmt.Println("🦆 DrDuck: Running comprehensive validation...")
	fmt.Println()

	// Run both validations and show results
	fmt.Println("## Pre-commit Check (Warning Only)")
	preCommitResult := validator.ValidatePreCommit()
	fmt.Println(preCommitResult.Message)
	
	fmt.Println()
	fmt.Println("## Pre-push Check (May Block)")
	prePushResult := validator.ValidatePrePush()
	fmt.Println(prePushResult.Message)

	fmt.Println()
	fmt.Println("## Summary")
	if len(preCommitResult.DraftADRs) > 0 {
		fmt.Printf("📝 Found %d draft ADR(s)\n", len(preCommitResult.DraftADRs))
	}

	if prePushResult.NeedsADR {
		fmt.Println("🤖 AI analysis suggests creating an ADR for current changes")
	}

	if prePushResult.ShouldBlock {
		fmt.Println("🚫 Current state would block git push")
		return fmt.Errorf("validation issues found")
	} else {
		fmt.Println("✅ All checks would pass")
	}

	return nil
}