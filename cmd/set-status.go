package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/SilverFlin/DrDuck/internal/adr"
	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/spf13/cobra"
)

var setStatusCmd = &cobra.Command{
	Use:   "set-status [adr-id] [status]",
	Short: "Set the status of an ADR",
	Long: `Set the status of an ADR with validation of status transitions.

Valid statuses:
  draft         - Initial status when created
  in-progress   - Work is ongoing  
  accepted      - Decision is finalized (use 'drduck accept' for validation)
  rejected      - Decision was rejected
  superseded    - Replaced by another ADR

Examples:
  drduck set-status 0001 in-progress    # Mark as in progress
  drduck set-status 1 rejected          # Mark as rejected
  drduck set-status 5 superseded        # Mark as superseded`,
	Args: cobra.ExactArgs(2),
	RunE: runSetStatus,
}

func init() {
	rootCmd.AddCommand(setStatusCmd)
}

func runSetStatus(cmd *cobra.Command, args []string) error {
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

	// Parse status
	newStatus, err := parseStatus(args[1])
	if err != nil {
		return fmt.Errorf("invalid status: %w", err)
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

	fmt.Printf("📝 ADR-%04d: %s\n", targetADR.ID, targetADR.Title)
	fmt.Printf("📊 Current Status: %s\n", targetADR.Status)

	// Check if status is already set
	if targetADR.Status == newStatus {
		fmt.Printf("✅ ADR-%04d is already %s\n", adrID, newStatus)
		return nil
	}

	// Validate status transition
	if err := validateStatusTransition(targetADR.Status, newStatus); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Special case: redirect to accept command for accepted status
	if newStatus == adr.StatusAccepted {
		fmt.Println("💡 For accepted status, use 'drduck accept' command which includes content validation")
		fmt.Printf("   drduck accept %04d\n", adrID)
		return nil
	}

	// Update status
	fmt.Printf("📊 Updating status to %s...\n", newStatus)
	if err := manager.UpdateADRStatus(adrID, newStatus); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	fmt.Printf("✅ ADR-%04d status updated from %s to %s\n", adrID, targetADR.Status, newStatus)

	// Show appropriate next steps based on new status
	showNextSteps(newStatus, adrID)

	return nil
}

// validateStatusTransition checks if a status transition is valid
func validateStatusTransition(currentStatus, newStatus adr.Status) error {
	// Define valid transitions
	validTransitions := map[adr.Status][]adr.Status{
		adr.StatusDraft: {
			adr.StatusInProgress,
			adr.StatusAccepted,
			adr.StatusRejected,
		},
		adr.StatusInProgress: {
			adr.StatusDraft,      // Back to draft for more work
			adr.StatusAccepted,
			adr.StatusRejected,
		},
		adr.StatusAccepted: {
			adr.StatusSuperseded, // Can only be superseded once accepted
		},
		adr.StatusRejected: {
			adr.StatusDraft,      // Can reconsider
			adr.StatusInProgress,
		},
		adr.StatusSuperseded: {
			// Superseded ADRs generally shouldn't change status
		},
	}

	allowedTransitions, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("unknown current status: %s", currentStatus)
	}

	for _, allowedStatus := range allowedTransitions {
		if newStatus == allowedStatus {
			return nil // Valid transition
		}
	}

	return fmt.Errorf("cannot transition from %s to %s", currentStatus, newStatus)
}

// showNextSteps provides guidance based on the new status
func showNextSteps(status adr.Status, adrID int) {
	fmt.Println()
	fmt.Println("💡 Next steps:")

	switch status {
	case adr.StatusDraft:
		fmt.Printf("   • Continue editing: drduck edit %04d\n", adrID)
		fmt.Printf("   • When ready: drduck accept %04d\n", adrID)
	case adr.StatusInProgress:
		fmt.Printf("   • Continue working on the decision\n")
		fmt.Printf("   • Update content: drduck edit %04d\n", adrID)
		fmt.Printf("   • When ready: drduck accept %04d\n", adrID)
	case adr.StatusRejected:
		fmt.Println("   • Document reasons for rejection in the ADR")
		fmt.Println("   • Consider if alternative approaches need their own ADRs")
	case adr.StatusSuperseded:
		fmt.Println("   • Reference the ADR that supersedes this one")
		fmt.Println("   • Update any documentation that references this ADR")
	}

	fmt.Println("   • Commit changes: git add . && git commit -m \"Update ADR status\"")
}