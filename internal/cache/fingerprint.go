package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Fingerprinter generates content-based fingerprints for git changes
type Fingerprinter struct {
	config CacheConfig
}

// NewFingerprinter creates a new change fingerprinter
func NewFingerprinter(config CacheConfig) *Fingerprinter {
	return &Fingerprinter{
		config: config,
	}
}

// GenerateFingerprint creates a content hash for the current git changes,
// excluding ADR files and other configured exclusions
func (f *Fingerprinter) GenerateFingerprint() (string, *ChangeFingerprint, error) {
	// Get current branch
	branch, err := f.getCurrentBranch()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Get commit range being analyzed
	commitRange, err := f.getCommitRange()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get commit range: %w", err)
	}

	// Get filtered git changes (excluding ADR files and other exclusions)
	filteredDiff, err := f.getFilteredChanges(commitRange)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get filtered changes: %w", err)
	}

	// Create change fingerprint
	fingerprint := &ChangeFingerprint{
		Files:       make(map[string]string),
		CommitRange: commitRange,
		Branch:      branch,
		Timestamp:   time.Now(),
	}

	// Generate file-level hashes for the filtered changes
	if err := f.generateFileHashes(filteredDiff, fingerprint); err != nil {
		return "", nil, fmt.Errorf("failed to generate file hashes: %w", err)
	}

	// Generate overall content hash from the fingerprint
	contentHash := f.hashFingerprint(fingerprint)

	return contentHash, fingerprint, nil
}

// getCurrentBranch gets the name of the current git branch
func (f *Fingerprinter) getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getCommitRange determines the appropriate commit range for analysis
func (f *Fingerprinter) getCommitRange() (string, error) {
	// Try different strategies to get a meaningful commit range
	
	// Strategy 1: Changes since last push (origin/branch)
	branch, err := f.getCurrentBranch()
	if err == nil {
		remoteBranch := fmt.Sprintf("origin/%s", branch)
		if f.remoteExists(remoteBranch) {
			return fmt.Sprintf("%s..HEAD", remoteBranch), nil
		}
	}

	// Strategy 2: Last few commits if no remote
	if f.hasCommits("HEAD~3") {
		return "HEAD~3..HEAD", nil
	}

	// Strategy 3: All uncommitted changes
	return "HEAD", nil
}

// getFilteredChanges gets git diff but filters out ADR files and other exclusions
func (f *Fingerprinter) getFilteredChanges(commitRange string) (string, error) {
	// Get the full diff first
	var cmd *exec.Cmd
	if commitRange == "HEAD" {
		// Get staged and unstaged changes
		cmd = exec.Command("git", "diff", "HEAD")
	} else {
		cmd = exec.Command("git", "diff", commitRange)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	diff := string(output)
	if diff == "" {
		return "", nil
	}

	// Filter the diff to exclude ADR files and other configured exclusions
	filteredDiff := f.filterDiff(diff)
	return filteredDiff, nil
}

// filterDiff removes changes to files that should be excluded from analysis
func (f *Fingerprinter) filterDiff(diff string) string {
	if diff == "" {
		return ""
	}

	lines := strings.Split(diff, "\n")
	var filteredLines []string
	var currentFile string
	var includeCurrentFile bool

	for _, line := range lines {
		// Detect file headers in diff
		if strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			// Extract filename from diff headers
			if strings.Contains(line, " b/") {
				parts := strings.Split(line, " b/")
				if len(parts) > 1 {
					currentFile = parts[1]
				}
			}
			
			// Determine if we should include this file
			includeCurrentFile = f.shouldIncludeFile(currentFile)
		}

		// Include line if current file should be included
		if includeCurrentFile || currentFile == "" {
			filteredLines = append(filteredLines, line)
		}
	}

	return strings.Join(filteredLines, "\n")
}

// shouldIncludeFile determines if a file should be included in the fingerprint
func (f *Fingerprinter) shouldIncludeFile(filePath string) bool {
	if filePath == "" {
		return true
	}

	// Check if file is in an ADR directory
	for _, adrDir := range f.config.ADRDirs {
		if f.matchesPattern(filePath, adrDir+"/*") || strings.HasPrefix(filePath, adrDir+"/") {
			return false
		}
	}

	// Check explicit exclusions
	for _, pattern := range f.config.ExcludeFiles {
		if f.matchesPattern(filePath, pattern) {
			return false
		}
	}

	// If include patterns are specified, file must match one of them
	if len(f.config.IncludeFiles) > 0 {
		for _, pattern := range f.config.IncludeFiles {
			if f.matchesPattern(filePath, pattern) {
				return true
			}
		}
		return false
	}

	// Default: include file
	return true
}

// matchesPattern checks if a file path matches a glob-like pattern
func (f *Fingerprinter) matchesPattern(filePath, pattern string) bool {
	// Simple pattern matching - could be enhanced with proper glob library
	if pattern == "*" {
		return true
	}
	
	// Handle ** wildcards
	if strings.Contains(pattern, "**") {
		pattern = strings.ReplaceAll(pattern, "**", "*")
	}

	// Handle extension patterns like "*.md"
	if strings.HasPrefix(pattern, "*.") {
		ext := pattern[1:] // Remove the *
		return strings.HasSuffix(filePath, ext)
	}

	// Handle directory patterns like "**/docs/**"
	if strings.Contains(pattern, "*") {
		// Simple wildcard matching
		matched, _ := filepath.Match(pattern, filePath)
		return matched
	}

	// Exact match or prefix match for directories
	return filePath == pattern || strings.HasPrefix(filePath, pattern)
}

// generateFileHashes creates individual hashes for changed files
func (f *Fingerprinter) generateFileHashes(diff string, fingerprint *ChangeFingerprint) error {
	if diff == "" {
		return nil
	}

	// For simplicity, we'll hash the entire filtered diff
	// In a more sophisticated implementation, we could hash individual file changes
	hasher := sha256.New()
	hasher.Write([]byte(diff))
	diffHash := hex.EncodeToString(hasher.Sum(nil))

	fingerprint.Files["filtered_diff"] = diffHash
	return nil
}

// hashFingerprint generates a single hash from the entire fingerprint
func (f *Fingerprinter) hashFingerprint(fingerprint *ChangeFingerprint) string {
	hasher := sha256.New()

	// Include commit range and branch
	hasher.Write([]byte(fingerprint.CommitRange))
	hasher.Write([]byte(fingerprint.Branch))

	// Include all file hashes in a consistent order
	for filePath, fileHash := range fingerprint.Files {
		hasher.Write([]byte(filePath))
		hasher.Write([]byte(fileHash))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// remoteExists checks if a remote branch exists
func (f *Fingerprinter) remoteExists(remoteBranch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", remoteBranch)
	return cmd.Run() == nil
}

// hasCommits checks if a commit reference exists
func (f *Fingerprinter) hasCommits(ref string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	return cmd.Run() == nil
}

// GetCurrentChanges returns a human-readable summary of current changes
// This can be used for logging and debugging
func (f *Fingerprinter) GetCurrentChanges() (string, error) {
	commitRange, err := f.getCommitRange()
	if err != nil {
		return "", err
	}

	return f.getFilteredChanges(commitRange)
}