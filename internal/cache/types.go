package cache

import (
	"time"
)

// AnalysisResult represents a cached AI analysis result
type AnalysisResult struct {
	Decision        string            `json:"decision"`         // "yes" or "no"
	Suggestion      string            `json:"suggestion"`       // AI's reasoning
	Title           string            `json:"title"`            // Suggested ADR title
	Timestamp       time.Time         `json:"timestamp"`        // When analysis was performed
	ChangesAnalyzed []string          `json:"changes_analyzed"` // File paths that were analyzed
	CommitRange     string            `json:"commit_range"`     // Git commit range analyzed
	Resolved        bool              `json:"resolved"`         // Whether an ADR was created for this
	ResolvedADRID   int               `json:"resolved_adr_id"`  // ID of ADR that resolved this
}

// CacheEntry represents a single cache entry keyed by content fingerprint
type CacheEntry struct {
	ContentHash string          `json:"content_hash"` // Hash of analyzed code changes (excluding ADRs)
	Analysis    *AnalysisResult `json:"analysis"`     // The analysis result
}

// Cache represents the entire cache structure
type Cache struct {
	Version  string                `json:"version"`  // Cache format version
	Entries  map[string]CacheEntry `json:"entries"`  // Content hash -> analysis result
	Metadata CacheMetadata         `json:"metadata"` // Cache metadata
}

// CacheMetadata contains cache management information
type CacheMetadata struct {
	LastCleanup time.Time `json:"last_cleanup"` // When cache was last cleaned up
	TotalSize   int       `json:"total_size"`   // Number of entries
	MaxAge      int       `json:"max_age_days"` // Maximum age in days before cleanup
}

// ChangeFingerprint represents the components used to generate a content hash
type ChangeFingerprint struct {
	Files       map[string]string `json:"files"`        // filepath -> file content hash
	CommitRange string            `json:"commit_range"` // git commit range
	Branch      string            `json:"branch"`       // current branch
	Timestamp   time.Time         `json:"timestamp"`    // when fingerprint was created
}

// CacheConfig represents configuration for the cache system
type CacheConfig struct {
	MaxAge        int      `json:"max_age_days"`    // Days to keep cache entries
	MaxEntries    int      `json:"max_entries"`     // Maximum number of cache entries
	ExcludeFiles  []string `json:"exclude_files"`   // File patterns to exclude from fingerprinting
	IncludeFiles  []string `json:"include_files"`   // File patterns to include (empty = all)
	ADRDirs       []string `json:"adr_dirs"`        // ADR directories to exclude from analysis
	CleanupAfter  int      `json:"cleanup_after"`   // Days between automatic cleanup
}

// DefaultCacheConfig returns sensible defaults for cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxAge:       7,  // Keep cache entries for 7 days
		MaxEntries:   100, // Keep max 100 entries
		ExcludeFiles: []string{
			"*.md",
			"*.txt", 
			"*.rst",
			"**/docs/**",
			"**/documentation/**",
			"**/.drduck/**",
			"**/.git/**",
		},
		IncludeFiles: []string{}, // Include all files not explicitly excluded
		ADRDirs: []string{
			"docs/adr",
			"docs/adrs", 
			"architecture/decisions",
			"doc/adr",
			"adr",
		},
		CleanupAfter: 1, // Clean up daily
	}
}