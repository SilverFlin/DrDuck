package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/SilverFlin/DrDuck/internal/adr"
	"github.com/SilverFlin/DrDuck/internal/ai"
	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/SilverFlin/DrDuck/internal/prompts/templates"
	"github.com/spf13/cobra"
)

var suggestCmd = &cobra.Command{
	Use:   "suggest [adr-id]",
	Short: "Get AI suggestions for completing an ADR",
	Long: `Get AI-powered suggestions for completing ADR content based on the current draft.
This command analyzes the current ADR content and provides suggestions for missing sections.

Examples:
  drduck suggest 0001        # Get suggestions for ADR-0001
  drduck suggest 1           # Get suggestions for ADR-0001 (leading zeros optional)`,
	Args: cobra.ExactArgs(1),
	RunE: runSuggest,
}

func init() {
	rootCmd.AddCommand(suggestCmd)
}

func runSuggest(cmd *cobra.Command, args []string) error {
	// Check if project is initialized
	initialized, err := config.IsInitialized()
	if err != nil {
		return fmt.Errorf("failed to check initialization status: %w", err)
	}

	if !initialized {
		return fmt.Errorf("❌ DrDuck is not initialized in this project. Run 'drduck init' first")
	}

	// Parse ADR ID
	adrIDStr := args[0]
	adrID, err := strconv.Atoi(strings.TrimLeft(adrIDStr, "0"))
	if err != nil {
		return fmt.Errorf("invalid ADR ID: %s", adrIDStr)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create managers
	manager := adr.NewManager(cfg)
	aiManager := ai.NewManager(cfg)

	// Get the ADR
	targetADR, err := manager.GetADRByID(adrID)
	if err != nil {
		return fmt.Errorf("ADR not found: %w", err)
	}

	fmt.Printf("🤖 Getting AI suggestions for ADR-%04d: %s\n", targetADR.ID, targetADR.Title)
	fmt.Printf("📊 Current Status: %s\n", targetADR.Status)
	fmt.Println()

	// Read current content
	content, err := os.ReadFile(targetADR.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read ADR file: %w", err)
	}

	// Check AI availability
	if !aiManager.IsAvailable() {
		fmt.Printf("⚠️  AI provider (%s) not available. Providing basic suggestions...\n\n", cfg.AIProvider)
		return provideFallbackSuggestions(targetADR, string(content))
	}

	fmt.Printf("🔍 Analyzing content with %s...\n", cfg.AIProvider)

	// Calculate days since creation for context
	daysSinceDraft := int(targetADR.Date.Sub(targetADR.Date).Hours() / 24)

	// Generate AI prompt for draft completion
	prompt := templates.DraftCompletionPrompt(targetADR.Title, string(content), daysSinceDraft)

	// Get AI analysis
	response, err := aiManager.AnalyzeChanges(prompt)
	if err != nil {
		fmt.Printf("⚠️  AI analysis failed: %v\nProviding basic suggestions...\n\n", err)
		return provideFallbackSuggestions(targetADR, string(content))
	}

	fmt.Println("🤖 Dr Duck's Suggestions:")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println(response)
	fmt.Println("=" + strings.Repeat("=", 50))

	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Printf("   drduck edit %04d        # Edit the ADR with these suggestions\n", adrID)
	fmt.Printf("   drduck accept %04d      # Accept when ready\n", adrID)

	return nil
}

// provideFallbackSuggestions provides basic content analysis when AI is unavailable
func provideFallbackSuggestions(targetADR *adr.ADR, content string) error {
	fmt.Println("📋 Basic Content Analysis:")
	fmt.Println()

	contentLower := strings.ToLower(content)

	// Check for placeholder comments
	placeholders := []string{
		"<!-- what is the issue",
		"<!-- what is the change", 
		"<!-- why are we making",
		"<!-- what becomes easier",
		"<!-- what becomes more difficult",
		"<!-- what other options",
		"<!-- what problem are we trying",
		"<!-- what is our solution",
		"<!-- why did we choose",
		"<!-- what are the consequences",
	}

	placeholderCount := 0
	for _, placeholder := range placeholders {
		if strings.Contains(contentLower, placeholder) {
			placeholderCount++
		}
	}

	if placeholderCount > 0 {
		fmt.Printf("📝 **Missing Content**: Found %d unfilled sections\n", placeholderCount)
		fmt.Println("   • Replace placeholder comments with actual content")
		fmt.Println("   • Each section should have at least a few sentences")
		fmt.Println()
	}

	// Check key sections
	sections := map[string][]string{
		"Context": {"## context", "## problem"},
		"Decision": {"## decision", "## solution"}, 
		"Rationale": {"## rationale", "## why"},
		"Consequences": {"## consequences", "## impact"},
	}

	missingSections := []string{}
	for sectionName, headers := range sections {
		hasContent := false
		for _, header := range headers {
			if strings.Contains(contentLower, header) {
				hasContent = true
				break
			}
		}
		if !hasContent {
			missingSections = append(missingSections, sectionName)
		}
	}

	if len(missingSections) > 0 {
		fmt.Printf("📋 **Missing Sections**: %s\n", strings.Join(missingSections, ", "))
		fmt.Println("   • Add these sections to complete the ADR")
		fmt.Println()
	}

	// Generic suggestions based on ADR title
	title := strings.ToLower(targetADR.Title)
	fmt.Println("💡 **General Suggestions**:")

	if strings.Contains(title, "database") || strings.Contains(title, "db") {
		fmt.Println("   • Consider data migration strategy")
		fmt.Println("   • Document performance implications")
		fmt.Println("   • Address backup and recovery concerns")
	} else if strings.Contains(title, "api") {
		fmt.Println("   • Define API contract and versioning strategy")
		fmt.Println("   • Consider backward compatibility")
		fmt.Println("   • Document authentication and security")
	} else if strings.Contains(title, "architecture") || strings.Contains(title, "design") {
		fmt.Println("   • Explain the architectural decision clearly")
		fmt.Println("   • Compare with alternative approaches")
		fmt.Println("   • Consider long-term maintainability")
	} else {
		fmt.Println("   • Clearly state the problem being solved")
		fmt.Println("   • Explain why this solution was chosen")
		fmt.Println("   • Consider future implications and risks")
	}

	fmt.Println()
	fmt.Println("🔧 **Action Items**:")
	fmt.Printf("   1. Edit the ADR: drduck edit %04d\n", targetADR.ID)
	fmt.Println("   2. Fill in missing sections with meaningful content")
	fmt.Println("   3. Remove placeholder comments")
	fmt.Printf("   4. Accept when complete: drduck accept %04d\n", targetADR.ID)

	return nil
}