package cmd

import (
	"fmt"
	"strings"

	"github.com/SilverFlin/DrDuck/internal/adr"
	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Architectural Decision Records (ADRs)",
	Long:  `List all Architectural Decision Records (ADRs) in the project with their status and dates.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
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

	// Create ADR manager
	manager := adr.NewManager(cfg)

	// Get all ADRs
	adrs, err := manager.List()
	if err != nil {
		return fmt.Errorf("failed to list ADRs: %w", err)
	}

	if len(adrs) == 0 {
		fmt.Println("📝 No ADRs found in this project.")
		fmt.Println("💡 Create your first ADR with: drduck new -n \"your-decision-name\"")
		return nil
	}

	fmt.Printf("🦆 Found %d ADR(s) in this project:\n\n", len(adrs))

	// Display ADRs in a table-like format
	for _, a := range adrs {
		statusIcon := getStatusIcon(a.Status)
		
		fmt.Printf("ADR-%04d %s %s\n", a.ID, statusIcon, a.Title)
		fmt.Printf("        📅 %s", a.Date.Format("2006-01-02"))
		if a.FilePath != "" {
			fmt.Printf(" • 📄 %s", a.FilePath)
		}
		fmt.Println()
		
		// Show context preview if available
		if a.Context != "" {
			contextPreview := strings.TrimSpace(a.Context)
			if len(contextPreview) > 80 {
				contextPreview = contextPreview[:77] + "..."
			}
			if contextPreview != "" {
				fmt.Printf("        💭 %s\n", contextPreview)
			}
		}
		fmt.Println()
	}

	// Show summary
	statusCounts := make(map[adr.Status]int)
	for _, a := range adrs {
		statusCounts[a.Status]++
	}

	fmt.Println("📊 Summary:")
	if statusCounts[adr.StatusDraft] > 0 {
		fmt.Printf("   📝 %d Draft\n", statusCounts[adr.StatusDraft])
	}
	if statusCounts[adr.StatusInProgress] > 0 {
		fmt.Printf("   ⚡ %d In Progress\n", statusCounts[adr.StatusInProgress])
	}
	if statusCounts[adr.StatusAccepted] > 0 {
		fmt.Printf("   ✅ %d Accepted\n", statusCounts[adr.StatusAccepted])
	}
	if statusCounts[adr.StatusSuperseded] > 0 {
		fmt.Printf("   ⏭️  %d Superseded\n", statusCounts[adr.StatusSuperseded])
	}
	if statusCounts[adr.StatusRejected] > 0 {
		fmt.Printf("   ❌ %d Rejected\n", statusCounts[adr.StatusRejected])
	}

	return nil
}

func getStatusIcon(status adr.Status) string {
	switch status {
	case adr.StatusDraft:
		return "📝"
	case adr.StatusInProgress:
		return "⚡"
	case adr.StatusAccepted:
		return "✅"
	case adr.StatusSuperseded:
		return "⏭️ "
	case adr.StatusRejected:
		return "❌"
	default:
		return "❓"
	}
}