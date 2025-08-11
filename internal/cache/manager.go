package cache

import (
	"fmt"
	"time"
)

// Manager provides high-level cache operations for ADR analysis
type Manager struct {
	storage      *Storage
	fingerprinter *Fingerprinter
	config       CacheConfig
}

// NewManager creates a new cache manager
func NewManager(workingDir string, config CacheConfig) *Manager {
	storage := NewStorage(workingDir, config)
	fingerprinter := NewFingerprinter(config)

	return &Manager{
		storage:       storage,
		fingerprinter: fingerprinter,
		config:        config,
	}
}

// NewManagerFromConfig creates a cache manager from a config.CacheConfig
func NewManagerFromConfig(workingDir string, cfg interface{}) *Manager {
	// Convert config interface to CacheConfig
	var cacheConfig CacheConfig
	
	// Handle both config.CacheConfig and cache.CacheConfig types
	switch c := cfg.(type) {
	case CacheConfig:
		cacheConfig = c
	default:
		// Use reflection-like approach or default config
		cacheConfig = DefaultCacheConfig()
	}
	
	return NewManager(workingDir, cacheConfig)
}

// GetAnalysis retrieves cached analysis for current changes, if available
func (m *Manager) GetAnalysis() (*AnalysisResult, bool, error) {
	// Generate fingerprint for current changes
	contentHash, _, err := m.fingerprinter.GenerateFingerprint()
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate fingerprint: %w", err)
	}

	// Look up analysis in cache
	analysis, found := m.storage.Get(contentHash)
	if !found {
		return nil, false, nil
	}

	// Double-check that analysis is not resolved
	if analysis.Resolved {
		return nil, false, nil
	}

	return analysis, true, nil
}

// StoreAnalysis saves an analysis result for the current changes
func (m *Manager) StoreAnalysis(decision, suggestion, title string) error {
	// Generate fingerprint for current changes
	contentHash, fingerprint, err := m.fingerprinter.GenerateFingerprint()
	if err != nil {
		return fmt.Errorf("failed to generate fingerprint: %w", err)
	}

	// Extract changed file paths for metadata
	changedFiles := make([]string, 0, len(fingerprint.Files))
	for filePath := range fingerprint.Files {
		if filePath != "filtered_diff" { // Skip our internal key
			changedFiles = append(changedFiles, filePath)
		}
	}

	// Create analysis result
	analysis := &AnalysisResult{
		Decision:        decision,
		Suggestion:      suggestion,
		Title:           title,
		Timestamp:       time.Now(),
		ChangesAnalyzed: changedFiles,
		CommitRange:     fingerprint.CommitRange,
		Resolved:        false,
		ResolvedADRID:   0,
	}

	// Store in cache
	return m.storage.Put(contentHash, analysis)
}

// MarkResolved marks the current changes as resolved by creating an ADR
func (m *Manager) MarkResolved(adrID int) error {
	// Generate fingerprint for current changes
	contentHash, _, err := m.fingerprinter.GenerateFingerprint()
	if err != nil {
		return fmt.Errorf("failed to generate fingerprint: %w", err)
	}

	return m.storage.MarkResolved(contentHash, adrID)
}

// HasUnresolvedAnalysis checks if there are any unresolved analyses
func (m *Manager) HasUnresolvedAnalysis() (bool, error) {
	unresolved, err := m.storage.GetUnresolvedAnalyses()
	if err != nil {
		return false, err
	}

	return len(unresolved) > 0, nil
}

// GetUnresolvedAnalyses returns all unresolved analyses
func (m *Manager) GetUnresolvedAnalyses() (map[string]*AnalysisResult, error) {
	return m.storage.GetUnresolvedAnalyses()
}

// Cleanup removes expired and resolved entries from cache
func (m *Manager) Cleanup() error {
	return m.storage.Cleanup()
}

// Clear removes all cache entries
func (m *Manager) Clear() error {
	return m.storage.Clear()
}

// GetCurrentChanges returns human-readable current changes for debugging
func (m *Manager) GetCurrentChanges() (string, error) {
	return m.fingerprinter.GetCurrentChanges()
}

// ShouldAnalyze determines if the current changes should be analyzed
// Returns false if changes are already cached or resolved
func (m *Manager) ShouldAnalyze() (bool, string, error) {
	// Check if we already have analysis for current changes
	analysis, found, err := m.GetAnalysis()
	if err != nil {
		return true, "", err // If we can't check cache, proceed with analysis
	}

	if found {
		if analysis.Decision == "yes" {
			return false, "ADR needed (cached result)", nil
		} else {
			return false, "No ADR needed (cached result)", nil
		}
	}

	return true, "", nil
}

// GetCacheStats returns statistics about the cache
func (m *Manager) GetCacheStats() (map[string]interface{}, error) {
	cache, err := m.storage.Load()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["total_entries"] = len(cache.Entries)
	stats["version"] = cache.Version
	stats["last_cleanup"] = cache.Metadata.LastCleanup
	
	// Count resolved vs unresolved
	resolved := 0
	unresolved := 0
	for _, entry := range cache.Entries {
		if entry.Analysis.Resolved {
			resolved++
		} else {
			unresolved++
		}
	}
	
	stats["resolved_entries"] = resolved
	stats["unresolved_entries"] = unresolved
	stats["max_age_days"] = m.config.MaxAge
	stats["max_entries"] = m.config.MaxEntries

	return stats, nil
}