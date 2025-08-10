package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/SilverFlin/DrDuck/internal/adr"
	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [adr-id]",
	Short: "Edit an existing ADR",
	Long: `Open an existing ADR for editing in your default editor.
After editing, you can use 'drduck accept [adr-id]' to mark it as accepted.

Examples:
  drduck edit 0001           # Edit ADR-0001
  drduck edit 0001 --status accepted  # Edit and set status
  drduck edit 1              # Edit ADR-0001 (leading zeros optional)`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

var editStatus string

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.Flags().StringVar(&editStatus, "status", "", "Set status after editing (draft, in-progress, accepted, rejected, superseded)")
}

func runEdit(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("📝 Editing ADR-%04d: %s\n", targetADR.ID, targetADR.Title)
	fmt.Printf("📄 File: %s\n", targetADR.FilePath)
	fmt.Printf("📊 Current Status: %s\n", targetADR.Status)
	fmt.Println()

	// Determine editor
	editor := getEditor()
	
	// Open ADR in editor
	fmt.Printf("🚀 Opening in %s...\n", editor)
	editorCmd := exec.Command(editor, targetADR.FilePath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	fmt.Println("✅ Editor closed")

	// Update status if requested
	if editStatus != "" {
		newStatus, err := parseStatus(editStatus)
		if err != nil {
			return fmt.Errorf("invalid status: %w", err)
		}

		fmt.Printf("📊 Updating status from %s to %s...\n", targetADR.Status, newStatus)
		if err := manager.UpdateADRStatus(adrID, newStatus); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}
		fmt.Printf("✅ ADR-%04d status updated to %s\n", adrID, newStatus)
	} else if targetADR.Status == adr.StatusDraft {
		// Offer AI assistance for draft completion
		fmt.Println()
		fmt.Println("🤖 AI assistance available:")
		fmt.Printf("   drduck suggest %04d    # Get AI content suggestions\n", adrID)
		fmt.Println("💡 Status management:")
		fmt.Printf("   drduck accept %04d     # Mark as accepted (with validation)\n", adrID)
		fmt.Printf("   drduck set-status %04d in-progress  # Mark as in-progress\n", adrID)
	}

	return nil
}

func getEditor() string {
	// Check environment variables in order of preference
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}

	// Check for common editors
	editors := []string{"code", "cursor", "vim", "vi", "nano", "notepad"}
	for _, editor := range editors {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	// Fallback
	return "vi"
}

func parseStatus(statusStr string) (adr.Status, error) {
	statusStr = strings.ToLower(strings.TrimSpace(statusStr))
	
	switch statusStr {
	case "draft":
		return adr.StatusDraft, nil
	case "in-progress", "inprogress", "progress":
		return adr.StatusInProgress, nil
	case "accepted", "accept":
		return adr.StatusAccepted, nil
	case "rejected", "reject":
		return adr.StatusRejected, nil
	case "superseded", "supersede":
		return adr.StatusSuperseded, nil
	default:
		return "", fmt.Errorf("unknown status '%s'. Valid statuses: draft, in-progress, accepted, rejected, superseded", statusStr)
	}
}