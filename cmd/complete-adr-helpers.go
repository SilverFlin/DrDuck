package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// shouldSkipADRCreation determines if AI analysis suggests skipping ADR creation
func shouldSkipADRCreation(aiAnalysis string) bool {
	if aiAnalysis == "" {
		return false
	}

	analysisLower := strings.ToLower(aiAnalysis)
	
	// Look for clear "No" decision in AI analysis
	return strings.Contains(analysisLower, "decision**: no") ||
		   strings.Contains(analysisLower, "**decision**: no") ||
		   strings.Contains(analysisLower, "decision: no")
}

// askUserToSkipADR prompts user when AI recommends skipping ADR
func askUserToSkipADR(aiAnalysis string) (bool, error) {
	fmt.Println("💡 Dr Duck's recommendation: These changes may not require an ADR.")
	
	var proceedChoice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Would you like to proceed with ADR creation anyway?").
				Options(
					huh.NewOption("No, skip ADR creation", "skip"),
					huh.NewOption("Yes, create ADR anyway", "proceed"),
					huh.NewOption("Let me review the analysis first", "review"),
				).
				Value(&proceedChoice),
		),
	)

	if err := form.Run(); err != nil {
		return false, err
	}

	switch proceedChoice {
	case "skip":
		fmt.Println("\n✅ ADR creation skipped based on AI recommendation.")
		fmt.Println("💡 You can always create an ADR later if needed with:")
		fmt.Printf("   drduck new -n \"your-decision-name\"\n")
		return true, nil // Skip ADR creation
		
	case "review":
		fmt.Println("\n🤖 Full AI Analysis:")
		fmt.Println("=" + strings.Repeat("=", 50))
		fmt.Println(aiAnalysis)
		fmt.Println("=" + strings.Repeat("=", 50))
		
		var reviewChoice string
		reviewForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("After reviewing the analysis, would you like to proceed?").
					Options(
						huh.NewOption("No, skip ADR creation", "skip"),
						huh.NewOption("Yes, create ADR anyway", "proceed"),
					).
					Value(&reviewChoice),
			),
		)
		
		if err := reviewForm.Run(); err != nil {
			return false, err
		}
		
		if reviewChoice == "skip" {
			fmt.Println("\n✅ ADR creation skipped after review.")
			return true, nil // Skip ADR creation
		}
		
		return false, nil // Proceed with ADR creation
		
	default: // "proceed"
		return false, nil // Proceed with ADR creation
	}
}