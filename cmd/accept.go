package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/SilverFlin/DrDuck/internal/adr"
	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/spf13/cobra"
)

var acceptCmd = &cobra.Command{
	Use:   "accept [adr-id]",
	Short: "Mark an ADR as accepted",
	Long: `Mark an ADR as accepted, indicating the decision has been finalized.
This command will validate that the ADR has sufficient content before accepting.

Examples:
  drduck accept 0001         # Accept ADR-0001
  drduck accept 1            # Accept ADR-0001 (leading zeros optional)
  drduck accept 5 --force    # Accept even if validation fails`,
	Args: cobra.ExactArgs(1),
	RunE: runAccept,
}

var forceAccept bool

func init() {
	rootCmd.AddCommand(acceptCmd)
	acceptCmd.Flags().BoolVar(&forceAccept, "force", false, "Accept ADR even if content validation fails")
}

func runAccept(cmd *cobra.Command, args []string) error {
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

	// Create ADR manager
	manager := adr.NewManager(cfg)

	// Get the ADR
	targetADR, err := manager.GetADRByID(adrID)
	if err != nil {
		return fmt.Errorf("ADR not found: %w", err)
	}

	fmt.Printf("🔍 Reviewing ADR-%04d: %s\n", targetADR.ID, targetADR.Title)
	fmt.Printf("📊 Current Status: %s\n", targetADR.Status)

	// Check current status
	if targetADR.Status == adr.StatusAccepted {
		fmt.Printf("✅ ADR-%04d is already accepted!\n", adrID)
		return nil
	}

	// Validate content if not forcing
	if !forceAccept {
		fmt.Println("🔎 Validating ADR content...")
		
		validation, err := validateADRContent(targetADR)
		if err != nil {
			return fmt.Errorf("failed to validate ADR: %w", err)
		}

		if !validation.IsComplete {
			fmt.Println("⚠️  ADR content validation failed:")
			for _, issue := range validation.Issues {
				fmt.Printf("   • %s\n", issue)
			}
			fmt.Println()
			fmt.Println("🔧 To fix these issues:")
			fmt.Printf("   drduck edit %04d       # Edit the ADR\n", adrID)
			fmt.Printf("   drduck accept %04d --force  # Accept anyway\n", adrID)
			return fmt.Errorf("ADR is not ready for acceptance")
		}

		fmt.Println("✅ ADR content validation passed")
	}

	// Update status to accepted
	fmt.Printf("📊 Accepting ADR-%04d...\n", adrID)
	if err := manager.UpdateADRStatus(adrID, adr.StatusAccepted); err != nil {
		return fmt.Errorf("failed to accept ADR: %w", err)
	}

	fmt.Printf("🎉 ADR-%04d (%s) has been accepted!\n", adrID, targetADR.Title)
	fmt.Println("✅ Status updated from", targetADR.Status, "to Accepted")

	// Show next steps
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("   • Commit this change: git add . && git commit -m \"Accept ADR-" + fmt.Sprintf("%04d", adrID) + "\"")
	fmt.Println("   • Share with your team for visibility")
	fmt.Printf("   • View all ADRs: drduck list\n")

	return nil
}

// ADRValidation represents the result of ADR content validation
type ADRValidation struct {
	IsComplete bool
	Issues     []string
}

// validateADRContent checks if an ADR has sufficient content to be accepted
func validateADRContent(targetADR *adr.ADR) (*ADRValidation, error) {
	validation := &ADRValidation{
		IsComplete: true,
		Issues:     []string{},
	}

	// Read the full content to analyze
	content, err := os.ReadFile(targetADR.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ADR file: %w", err)
	}

	contentStr := strings.ToLower(string(content))

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
		if strings.Contains(contentStr, placeholder) {
			placeholderCount++
		}
	}

	if placeholderCount > 2 {
		validation.IsComplete = false
		validation.Issues = append(validation.Issues, fmt.Sprintf("Found %d unfilled placeholder sections", placeholderCount))
	}

	// Check for minimum content in key sections
	sections := map[string][]string{
		"Context": {"## context", "## problem"},
		"Decision": {"## decision", "## solution"},
		"Rationale": {"## rationale", "## why"},
	}

	for sectionName, sectionHeaders := range sections {
		hasContent := false
		for _, header := range sectionHeaders {
			if idx := strings.Index(contentStr, header); idx != -1 {
				// Look for content after the header
				afterHeader := contentStr[idx:]
				nextHeader := strings.Index(afterHeader[len(header):], "##")
				
				var sectionContent string
				if nextHeader == -1 {
					sectionContent = afterHeader[len(header):]
				} else {
					sectionContent = afterHeader[len(header) : len(header)+nextHeader]
				}
				
				// Remove common non-content
				sectionContent = strings.ReplaceAll(sectionContent, "<!--", "")
				sectionContent = strings.ReplaceAll(sectionContent, "-->", "")
				sectionContent = strings.TrimSpace(sectionContent)
				
				if len(sectionContent) > 20 { // Minimum meaningful content
					hasContent = true
					break
				}
			}
		}
		
		if !hasContent {
			validation.IsComplete = false
			validation.Issues = append(validation.Issues, fmt.Sprintf("%s section needs more content", sectionName))
		}
	}

	return validation, nil
}