package adr

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SilverFlin/DrDuck/internal/config"
)

type Status string

const (
	StatusDraft      Status = "Draft"
	StatusInProgress Status = "In Progress"
	StatusAccepted   Status = "Accepted"
	StatusSuperseded Status = "Superseded"
	StatusRejected   Status = "Rejected"
)

type ADR struct {
	ID          int       `yaml:"id"`
	Title       string    `yaml:"title"`
	Status      Status    `yaml:"status"`
	Date        time.Time `yaml:"date"`
	Context     string    `yaml:"context"`
	Decision    string    `yaml:"decision"`
	Rationale   string    `yaml:"rationale"`
	Consequences string   `yaml:"consequences"`
	Alternatives string   `yaml:"alternatives,omitempty"`
	FilePath    string    `yaml:"-"`
}

type Manager struct {
	config *config.Config
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{config: cfg}
}

// GetNextID returns the next available ADR ID
func (m *Manager) GetNextID() (int, error) {
	adrs, err := m.List()
	if err != nil {
		return 0, err
	}

	if len(adrs) == 0 {
		return 1, nil
	}

	// Find the highest ID
	maxID := 0
	for _, adr := range adrs {
		if adr.ID > maxID {
			maxID = adr.ID
		}
	}

	return maxID + 1, nil
}

// Create creates a new ADR with the given name and template
func (m *Manager) Create(name string) (*ADR, error) {
	id, err := m.GetNextID()
	if err != nil {
		return nil, fmt.Errorf("failed to get next ID: %w", err)
	}

	// Create ADR struct
	adr := &ADR{
		ID:     id,
		Title:  name,
		Status: StatusDraft,
		Date:   time.Now(),
	}

	// Generate file path
	filename := fmt.Sprintf("%04d-%s.md", id, strings.ReplaceAll(strings.ToLower(name), " ", "-"))
	
	var adrPath string
	if m.config.DocStorage == "same-repo" {
		adrPath = filepath.Join(m.config.DocPath, filename)
	} else {
		// For separate repo, we'll create in a temporary location for now
		// TODO: Implement separate repo handling
		adrPath = filepath.Join("temp_adrs", filename)
	}

	adr.FilePath = adrPath

	// Generate content from template
	content, err := m.generateFromTemplate(adr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate template: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(adrPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write ADR file
	if err := os.WriteFile(adrPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write ADR file: %w", err)
	}

	return adr, nil
}

// List returns all ADRs in the project
func (m *Manager) List() ([]*ADR, error) {
	var adrDir string
	if m.config.DocStorage == "same-repo" {
		adrDir = m.config.DocPath
	} else {
		// TODO: Implement separate repo handling
		adrDir = "temp_adrs"
	}

	if _, err := os.Stat(adrDir); os.IsNotExist(err) {
		return []*ADR{}, nil // No ADRs yet
	}

	files, err := os.ReadDir(adrDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read ADR directory: %w", err)
	}

	var adrs []*ADR
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") || file.Name() == "README.md" {
			continue
		}

		// Parse ADR ID from filename
		parts := strings.SplitN(file.Name(), "-", 2)
		if len(parts) < 2 {
			continue
		}

		id, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		adrPath := filepath.Join(adrDir, file.Name())
		adr, err := m.parseADRFile(adrPath, id)
		if err != nil {
			continue // Skip invalid ADRs
		}

		adrs = append(adrs, adr)
	}

	// Sort by ID
	sort.Slice(adrs, func(i, j int) bool {
		return adrs[i].ID < adrs[j].ID
	})

	return adrs, nil
}

// parseADRFile parses an ADR file and extracts metadata
func (m *Manager) parseADRFile(filePath string, id int) (*ADR, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	adr := &ADR{
		ID:       id,
		FilePath: filePath,
		Status:   StatusDraft, // Default status
		Date:     time.Now(),   // Default to current date if not found in file
	}

	// Simple parsing - look for title and status
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Extract title from first heading
		if strings.HasPrefix(trimmed, "# ") && adr.Title == "" {
			adr.Title = strings.TrimPrefix(trimmed, "# ")
			continue
		}

		// Extract status
		if strings.Contains(strings.ToLower(line), "status:") {
			// Find status after "Status:"
			statusIdx := strings.Index(strings.ToLower(line), "status:")
			if statusIdx != -1 {
				statusText := strings.TrimSpace(line[statusIdx+7:])
				adr.Status = Status(statusText)
			}
			continue
		}

		// Extract date
		if strings.Contains(strings.ToLower(line), "date:") {
			dateIdx := strings.Index(strings.ToLower(line), "date:")
			if dateIdx != -1 {
				dateText := strings.TrimSpace(line[dateIdx+5:])
				if parsedDate, err := time.Parse("2006-01-02", dateText); err == nil {
					adr.Date = parsedDate
				}
			}
			continue
		}

		// Extract context, decision, etc. (basic implementation)
		if i < len(lines)-1 && strings.HasPrefix(trimmed, "## Context") {
			// Find content until next section
			j := i + 1
			var contextLines []string
			for j < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[j]), "##") {
				if strings.TrimSpace(lines[j]) != "" {
					contextLines = append(contextLines, lines[j])
				}
				j++
			}
			adr.Context = strings.Join(contextLines, "\n")
		}
	}

	return adr, nil
}

// generateFromTemplate generates ADR content from the configured template
func (m *Manager) generateFromTemplate(adr *ADR) (string, error) {
	switch m.config.ADRTemplate {
	case "madr":
		return m.generateMADRTemplate(adr), nil
	case "simple":
		return m.generateSimpleTemplate(adr), nil
	default:
		return m.generateMADRTemplate(adr), nil // Default to MADR
	}
}

// generateMADRTemplate generates content using MADR template
func (m *Manager) generateMADRTemplate(adr *ADR) string {
	return fmt.Sprintf(`# %s

* **Status**: %s
* **Date**: %s

## Context

<!-- What is the issue that we're seeing that is motivating this decision or change? -->

## Decision

<!-- What is the change that we're proposing and/or doing? -->

## Rationale

<!-- Why are we making this decision? What are the driving forces? -->

## Consequences

### Positive

<!-- What becomes easier or more straightforward? -->

### Negative

<!-- What becomes more difficult or complex? -->

### Neutral

<!-- What are the other implications that are neither positive nor negative? -->

## Alternatives Considered

<!-- What other options were considered? Why were they not chosen? -->

## Links

<!-- Related ADRs, issues, or documentation -->

---
*ADR-%04d created by DrDuck on %s*
`, adr.Title, adr.Status, adr.Date.Format("2006-01-02"), adr.ID, adr.Date.Format("2006-01-02"))
}

// generateSimpleTemplate generates content using simple template
func (m *Manager) generateSimpleTemplate(adr *ADR) string {
	return fmt.Sprintf(`# %s

**Status**: %s  
**Date**: %s

## Problem

<!-- What problem are we trying to solve? -->

## Solution

<!-- What is our solution? -->

## Why This Solution?

<!-- Why did we choose this solution over alternatives? -->

## Impact

<!-- What are the consequences of this decision? -->

---
*ADR-%04d created by DrDuck on %s*
`, adr.Title, adr.Status, adr.Date.Format("2006-01-02"), adr.ID, adr.Date.Format("2006-01-02"))
}

// GetDraftADRs returns all ADRs currently in draft status
func (m *Manager) GetDraftADRs() ([]*ADR, error) {
	allADRs, err := m.List()
	if err != nil {
		return nil, err
	}

	var drafts []*ADR
	for _, adr := range allADRs {
		if adr.Status == StatusDraft {
			drafts = append(drafts, adr)
		}
	}

	return drafts, nil
}

// HasDraftADRs checks if there are any ADRs in draft status
func (m *Manager) HasDraftADRs() (bool, error) {
	drafts, err := m.GetDraftADRs()
	if err != nil {
		return false, err
	}
	return len(drafts) > 0, nil
}

// GetADRByID retrieves an ADR by its ID
func (m *Manager) GetADRByID(id int) (*ADR, error) {
	adrs, err := m.List()
	if err != nil {
		return nil, err
	}

	for _, adr := range adrs {
		if adr.ID == id {
			return adr, nil
		}
	}

	return nil, fmt.Errorf("ADR with ID %d not found", id)
}

// UpdateADRStatus updates the status of an ADR
func (m *Manager) UpdateADRStatus(id int, newStatus Status) error {
	adr, err := m.GetADRByID(id)
	if err != nil {
		return err
	}

	// Read the current file content
	content, err := os.ReadFile(adr.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read ADR file: %w", err)
	}

	// Update the status in the content
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), "status:") {
			// Replace the status line
			if strings.Contains(line, "**Status**:") {
				lines[i] = fmt.Sprintf("**Status**: %s", newStatus)
			} else if strings.Contains(line, "* **Status**:") {
				lines[i] = fmt.Sprintf("* **Status**: %s", newStatus)
			}
			break
		}
	}

	// Write back the updated content
	updatedContent := strings.Join(lines, "\n")
	if err := os.WriteFile(adr.FilePath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to update ADR file: %w", err)
	}

	return nil
}

// GetStatusCounts returns a count of ADRs by status
func (m *Manager) GetStatusCounts() (map[Status]int, error) {
	adrs, err := m.List()
	if err != nil {
		return nil, err
	}

	counts := make(map[Status]int)
	for _, adr := range adrs {
		counts[adr.Status]++
	}

	return counts, nil
}