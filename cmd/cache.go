package cmd

import (
	"fmt"

	"github.com/SilverFlin/DrDuck/internal/cache"
	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage analysis cache",
	Long: `Commands for managing the analysis cache system.

The cache stores AI analysis results to avoid re-analyzing the same code changes.
This speeds up git hooks and complete-adr operations.`,
}

var cacheStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cache status and statistics",
	RunE:  runCacheStatus,
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all cache entries",
	RunE:  runCacheClear,
}

var cacheCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove expired and resolved cache entries",
	RunE:  runCacheCleanup,
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheStatusCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheCleanupCmd)
}

func runCacheStatus(cmd *cobra.Command, args []string) error {
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

	// Create cache manager
	cacheManager := cache.NewManagerFromMainConfig(cfg.Cache)

	// Get cache statistics
	stats, err := cacheManager.GetCacheStats()
	if err != nil {
		return fmt.Errorf("failed to get cache statistics: %w", err)
	}

	// Display cache status
	fmt.Println("🦆 DrDuck Analysis Cache Status")
	fmt.Println("===============================")
	fmt.Printf("Total entries: %v\n", stats["total_entries"])
	fmt.Printf("Resolved entries: %v\n", stats["resolved_entries"])
	fmt.Printf("Unresolved entries: %v\n", stats["unresolved_entries"])
	fmt.Printf("Max age (days): %v\n", stats["max_age_days"])
	fmt.Printf("Max entries: %v\n", stats["max_entries"])
	fmt.Printf("Cache version: %v\n", stats["version"])
	
	if lastCleanup, ok := stats["last_cleanup"]; ok {
		fmt.Printf("Last cleanup: %v\n", lastCleanup)
	}

	// Show current changes fingerprint for debugging
	changes, err := cacheManager.GetCurrentChanges()
	if err == nil && changes != "" {
		fmt.Println("\n🔍 Current Changes Detection:")
		if len(changes) > 200 {
			fmt.Printf("Changes detected: %d characters (truncated)\n", len(changes))
			fmt.Printf("Sample: %s...\n", changes[:200])
		} else if changes != "" {
			fmt.Printf("Changes detected: %s\n", changes)
		} else {
			fmt.Println("No changes detected")
		}
	}

	return nil
}

func runCacheClear(cmd *cobra.Command, args []string) error {
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

	// Create cache manager
	cacheManager := cache.NewManagerFromMainConfig(cfg.Cache)

	// Clear cache
	if err := cacheManager.Clear(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	fmt.Println("✅ Analysis cache cleared successfully")
	return nil
}

func runCacheCleanup(cmd *cobra.Command, args []string) error {
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

	// Create cache manager
	cacheManager := cache.NewManagerFromMainConfig(cfg.Cache)

	// Get stats before cleanup
	statsBefore, err := cacheManager.GetCacheStats()
	if err != nil {
		return fmt.Errorf("failed to get cache statistics: %w", err)
	}

	// Run cleanup
	if err := cacheManager.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup cache: %w", err)
	}

	// Get stats after cleanup
	statsAfter, err := cacheManager.GetCacheStats()
	if err != nil {
		return fmt.Errorf("failed to get cache statistics: %w", err)
	}

	entriesBefore := statsBefore["total_entries"].(int)
	entriesAfter := statsAfter["total_entries"].(int)
	removed := entriesBefore - entriesAfter

	fmt.Printf("✅ Cache cleanup completed\n")
	fmt.Printf("   Entries before: %d\n", entriesBefore)
	fmt.Printf("   Entries after: %d\n", entriesAfter)
	fmt.Printf("   Entries removed: %d\n", removed)

	return nil
}