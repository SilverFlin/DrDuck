package cmd

import (
	"fmt"
	"time"

	"github.com/SilverFlin/DrDuck/internal/adr"
	"github.com/SilverFlin/DrDuck/internal/ai"
	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show DrDuck configuration and ADR status overview",
	Long: `Display comprehensive status information including:
- DrDuck configuration settings
- ADR counts by status
- Git hooks status
- AI provider availability
- Recent draft ADRs

This command provides a quick overview of your project's documentation state.`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Check if project is initialized
	initialized, err := config.IsInitialized()
	if err != nil {
		return fmt.Errorf("failed to check initialization status: %w", err)
	}

	if !initialized {
		fmt.Println("📋 DrDuck Status")
		fmt.Println("================")
		fmt.Println()
		fmt.Println("❌ DrDuck is not initialized in this project")
		fmt.Println("💡 Run 'drduck init' to get started")
		return nil
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Println("🦆 DrDuck Status")
	fmt.Println("================")
	fmt.Println()

	// Configuration overview
	fmt.Println("## Configuration")
	fmt.Printf("📁 Storage: %s", cfg.DocStorage)
	if cfg.DocStorage == "same-repo" {
		fmt.Printf(" (%s)", cfg.DocPath)
	} else if cfg.SeparateRepoURL != "" {
		fmt.Printf(" (%s)", cfg.SeparateRepoURL)
	}
	fmt.Println()
	fmt.Printf("🤖 AI Provider: %s", cfg.AIProvider)

	// Check AI availability
	aiManager := ai.NewManager(cfg)
	if aiManager.IsAvailable() {
		fmt.Println(" ✅")
	} else {
		fmt.Println(" ❌")
	}

	fmt.Printf("📝 Template: %s\n", cfg.ADRTemplate)
	fmt.Println()

	// Git hooks status
	fmt.Println("## Git Hooks")
	fmt.Printf("Pre-commit: ")
	if cfg.Hooks.PreCommit {
		fmt.Println("✅ Enabled (warns about drafts)")
	} else {
		fmt.Println("❌ Disabled")
	}

	fmt.Printf("Pre-push: ")
	if cfg.Hooks.PrePush {
		fmt.Println("✅ Enabled (blocks on drafts/missing ADRs)")
	} else {
		fmt.Println("❌ Disabled")
	}
	fmt.Println()

	// ADR status overview
	adrManager := adr.NewManager(cfg)
	counts, err := adrManager.GetStatusCounts()
	if err != nil {
		fmt.Printf("⚠️  Could not get ADR status: %v\n", err)
	} else {
		fmt.Println("## ADR Overview")
		total := 0
		for _, count := range counts {
			total += count
		}

		if total == 0 {
			fmt.Println("📝 No ADRs found")
			fmt.Println("💡 Create your first ADR: drduck new -n \"your-decision-name\"")
		} else {
			fmt.Printf("📊 Total ADRs: %d\n", total)
			
			if counts[adr.StatusDraft] > 0 {
				fmt.Printf("   📝 %d Draft\n", counts[adr.StatusDraft])
			}
			if counts[adr.StatusInProgress] > 0 {
				fmt.Printf("   ⚡ %d In Progress\n", counts[adr.StatusInProgress])
			}
			if counts[adr.StatusAccepted] > 0 {
				fmt.Printf("   ✅ %d Accepted\n", counts[adr.StatusAccepted])
			}
			if counts[adr.StatusSuperseded] > 0 {
				fmt.Printf("   ⏭️  %d Superseded\n", counts[adr.StatusSuperseded])
			}
			if counts[adr.StatusRejected] > 0 {
				fmt.Printf("   ❌ %d Rejected\n", counts[adr.StatusRejected])
			}
		}
		fmt.Println()

		// Show draft details if any
		if counts[adr.StatusDraft] > 0 {
			fmt.Println("## Draft ADRs (Attention Needed)")
			drafts, err := adrManager.GetDraftADRs()
			if err == nil {
				for _, draft := range drafts {
					daysSince := int(time.Since(draft.Date).Hours() / 24)
					daysText := "today"
					if daysSince == 1 {
						daysText = "1 day"
					} else if daysSince > 1 {
						daysText = fmt.Sprintf("%d days", daysSince)
					}

					fmt.Printf("   📝 ADR-%04d: %s (%s old)\n", draft.ID, draft.Title, daysText)
				}
				fmt.Println()
				fmt.Println("💡 Complete drafts before pushing:")
				for _, draft := range drafts {
					fmt.Printf("   drduck edit %04d\n", draft.ID)
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("## Quick Commands")
	fmt.Println("   drduck list           # List all ADRs")
	fmt.Println("   drduck new -n \"name\"   # Create new ADR")
	fmt.Println("   drduck validate       # Check current state")
	if cfg.AIProvider != "" {
		fmt.Printf("   AI Provider: %s", cfg.AIProvider)
		if aiManager.IsAvailable() {
			fmt.Println(" (available)")
		} else {
			fmt.Println(" (not available)")
		}
	}

	return nil
}