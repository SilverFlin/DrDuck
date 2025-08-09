package cmd

import (
	"fmt"
	"strings"

	"github.com/SilverFlin/DrDuck/internal/adr"
	"github.com/SilverFlin/DrDuck/internal/ai"
	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new Architectural Decision Record (ADR)",
	Long:  `Create a new Architectural Decision Record (ADR) with the specified name using the configured template.`,
	RunE:  runNew,
}

var (
	adrName string
)

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVarP(&adrName, "name", "n", "", "Name of the ADR (required)")
	newCmd.MarkFlagRequired("name")
}

func runNew(cmd *cobra.Command, args []string) error {
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

	// Validate ADR name
	if strings.TrimSpace(adrName) == "" {
		return fmt.Errorf("ADR name cannot be empty")
	}

	// Clean up the name
	cleanName := strings.TrimSpace(adrName)
	
	// Create ADR manager
	manager := adr.NewManager(cfg)

	// Create AI manager
	aiManager := ai.NewManager(cfg)

	fmt.Printf("🦆 Creating new ADR: %s\n", cleanName)

	// Check AI provider availability
	if aiManager.IsAvailable() {
		fmt.Printf("🤖 %s integration detected - ADR will be enhanced with AI insights\n", aiManager.GetProviderName())
	} else {
		fmt.Printf("ℹ️  %s not available - creating basic ADR template\n", aiManager.GetProviderName())
	}

	// Create the ADR
	newADR, err := manager.Create(cleanName)
	if err != nil {
		return fmt.Errorf("failed to create ADR: %w", err)
	}

	fmt.Printf("✅ ADR-%04d created successfully!\n", newADR.ID)
	fmt.Printf("📝 File: %s\n", newADR.FilePath)
	fmt.Printf("🔧 Status: %s\n", newADR.Status)
	fmt.Printf("📅 Date: %s\n", newADR.Date.Format("2006-01-02"))
	fmt.Println()
	fmt.Printf("💡 Edit the ADR to add context, decision rationale, and consequences.\n")
	fmt.Printf("🔄 The ADR will be automatically updated when you push changes to your branch.\n")

	return nil
}