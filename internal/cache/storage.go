package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	CacheVersion     = "1.0"
	CacheFileName    = "analysis.json"
	DefaultCacheDir  = ".drduck/cache"
)

// Storage handles reading and writing cache data to disk
type Storage struct {
	cacheDir  string
	cacheFile string
	config    CacheConfig
}

// NewStorage creates a new cache storage instance
func NewStorage(workingDir string, config CacheConfig) *Storage {
	cacheDir := filepath.Join(workingDir, DefaultCacheDir)
	cacheFile := filepath.Join(cacheDir, CacheFileName)
	
	return &Storage{
		cacheDir:  cacheDir,
		cacheFile: cacheFile,
		config:    config,
	}
}

// Load reads the cache from disk, creating empty cache if file doesn't exist
func (s *Storage) Load() (*Cache, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(s.cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// If cache file doesn't exist, return empty cache
	if _, err := os.Stat(s.cacheFile); os.IsNotExist(err) {
		return s.createEmptyCache(), nil
	}

	// Read and parse cache file
	data, err := os.ReadFile(s.cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		// If cache is corrupted, start fresh
		return s.createEmptyCache(), nil
	}

	// Migrate cache version if needed
	if cache.Version != CacheVersion {
		return s.migrateCache(&cache)
	}

	return &cache, nil
}

// Save writes the cache to disk
func (s *Storage) Save(cache *Cache) error {
	// Update metadata
	cache.Metadata.TotalSize = len(cache.Entries)
	
	// Ensure cache directory exists
	if err := os.MkdirAll(s.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Marshal cache to JSON
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Write to file
	if err := os.WriteFile(s.cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Get retrieves an analysis result by content hash
func (s *Storage) Get(contentHash string) (*AnalysisResult, bool) {
	cache, err := s.Load()
	if err != nil {
		return nil, false
	}

	entry, exists := cache.Entries[contentHash]
	if !exists {
		return nil, false
	}

	// Check if entry is expired
	if s.isExpired(entry.Analysis) {
		// Remove expired entry
		delete(cache.Entries, contentHash)
		s.Save(cache) // Best effort save
		return nil, false
	}

	return entry.Analysis, true
}

// Put stores an analysis result with the given content hash
func (s *Storage) Put(contentHash string, analysis *AnalysisResult) error {
	cache, err := s.Load()
	if err != nil {
		return err
	}

	// Create cache entry
	entry := CacheEntry{
		ContentHash: contentHash,
		Analysis:    analysis,
	}

	cache.Entries[contentHash] = entry

	// Clean up old entries if needed
	if len(cache.Entries) > s.config.MaxEntries {
		s.cleanupOldEntries(cache)
	}

	return s.Save(cache)
}

// MarkResolved marks an analysis as resolved with the given ADR ID
func (s *Storage) MarkResolved(contentHash string, adrID int) error {
	cache, err := s.Load()
	if err != nil {
		return err
	}

	entry, exists := cache.Entries[contentHash]
	if !exists {
		return fmt.Errorf("cache entry not found for hash: %s", contentHash)
	}

	entry.Analysis.Resolved = true
	entry.Analysis.ResolvedADRID = adrID
	cache.Entries[contentHash] = entry

	return s.Save(cache)
}

// Cleanup removes expired and resolved entries from cache
func (s *Storage) Cleanup() error {
	cache, err := s.Load()
	if err != nil {
		return err
	}

	originalSize := len(cache.Entries)
	now := time.Now()

	// Remove expired or resolved entries
	for hash, entry := range cache.Entries {
		if s.isExpired(entry.Analysis) || entry.Analysis.Resolved {
			delete(cache.Entries, hash)
		}
	}

	// Update cleanup timestamp
	cache.Metadata.LastCleanup = now

	// Only save if something was cleaned up
	if len(cache.Entries) < originalSize {
		return s.Save(cache)
	}

	return nil
}

// GetUnresolvedAnalyses returns all unresolved analyses
func (s *Storage) GetUnresolvedAnalyses() (map[string]*AnalysisResult, error) {
	cache, err := s.Load()
	if err != nil {
		return nil, err
	}

	unresolved := make(map[string]*AnalysisResult)
	for hash, entry := range cache.Entries {
		if !entry.Analysis.Resolved && !s.isExpired(entry.Analysis) {
			unresolved[hash] = entry.Analysis
		}
	}

	return unresolved, nil
}

// createEmptyCache creates a new empty cache with default metadata
func (s *Storage) createEmptyCache() *Cache {
	return &Cache{
		Version: CacheVersion,
		Entries: make(map[string]CacheEntry),
		Metadata: CacheMetadata{
			LastCleanup: time.Now(),
			TotalSize:   0,
			MaxAge:      s.config.MaxAge,
		},
	}
}

// migrateCache handles cache version migrations
func (s *Storage) migrateCache(cache *Cache) (*Cache, error) {
	// For now, just create a new cache if version doesn't match
	// In the future, implement actual migration logic here
	return s.createEmptyCache(), nil
}

// isExpired checks if a cache entry is expired based on config
func (s *Storage) isExpired(analysis *AnalysisResult) bool {
	maxAge := time.Duration(s.config.MaxAge) * 24 * time.Hour
	return time.Since(analysis.Timestamp) > maxAge
}

// cleanupOldEntries removes oldest entries to stay within MaxEntries limit
func (s *Storage) cleanupOldEntries(cache *Cache) {
	if len(cache.Entries) <= s.config.MaxEntries {
		return
	}

	// Convert to slice for sorting by timestamp
	type entryWithHash struct {
		hash  string
		entry CacheEntry
	}

	entries := make([]entryWithHash, 0, len(cache.Entries))
	for hash, entry := range cache.Entries {
		entries = append(entries, entryWithHash{hash: hash, entry: entry})
	}

	// Sort by timestamp (oldest first)
	// Simple bubble sort for now
	for i := 0; i < len(entries)-1; i++ {
		for j := 0; j < len(entries)-i-1; j++ {
			if entries[j].entry.Analysis.Timestamp.After(entries[j+1].entry.Analysis.Timestamp) {
				entries[j], entries[j+1] = entries[j+1], entries[j]
			}
		}
	}

	// Remove oldest entries until we're within the limit
	for len(cache.Entries) > s.config.MaxEntries && len(entries) > 0 {
		delete(cache.Entries, entries[0].hash)
		entries = entries[1:]
	}
}

// Clear removes all cache entries
func (s *Storage) Clear() error {
	cache := s.createEmptyCache()
	return s.Save(cache)
}