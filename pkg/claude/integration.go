package claude

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ClaudeSession represents information about a Claude Code CLI session
type ClaudeSession struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Messages    int      `json:"messages"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	Files       []string `json:"files,omitempty"`
	Context     string   `json:"context,omitempty"`
	LastMessage string   `json:"last_message,omitempty"`
}

// Integration handles Claude Code CLI integration
type Integration struct {
	// Future: Add configuration options
}

// NewIntegration creates a new Claude integration instance
func NewIntegration() *Integration {
	return &Integration{}
}

// IsAvailable checks if Claude Code CLI is available
func (i *Integration) IsAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// GetCurrentSession attempts to get information about the current Claude session
func (i *Integration) GetCurrentSession() (*ClaudeSession, error) {
	if !i.IsAvailable() {
		return nil, fmt.Errorf("claude command not available")
	}

	// TODO: Implement actual Claude CLI integration
	// This is a placeholder structure for future implementation
	// We would need to:
	// 1. Check for Claude session files/state
	// 2. Parse Claude's conversation history
	// 3. Extract file changes and context
	// 4. Return session information

	return &ClaudeSession{
		ID:   "placeholder",
		Name: "Current Session",
	}, fmt.Errorf("not implemented: Claude CLI integration is planned for future releases")
}

// ExtractContext extracts relevant context from Claude session for ADR generation
func (i *Integration) ExtractContext(session *ClaudeSession) (string, error) {
	// TODO: Implement context extraction
	// This would analyze:
	// - Recent conversation messages
	// - Code changes discussed
	// - Decision points mentioned
	// - Architectural considerations

	return "", fmt.Errorf("not implemented: context extraction is planned for future releases")
}

// GetChangedFiles returns files that have been modified in the current session
func (i *Integration) GetChangedFiles() ([]string, error) {
	// TODO: Implement file change tracking
	// This could work by:
	// 1. Monitoring Claude's file access
	// 2. Comparing timestamps
	// 3. Integrating with git to track changes
	// 4. Parsing Claude's conversation for file mentions

	return nil, fmt.Errorf("not implemented: file change tracking is planned for future releases")
}

// SuggestADRContent generates ADR content suggestions based on Claude session
func (i *Integration) SuggestADRContent(adrName string) (map[string]string, error) {
	// TODO: Implement AI-powered ADR content suggestion
	// This would analyze the Claude session and suggest:
	// - Context section content
	// - Decision rationale
	// - Consequences
	// - Alternatives considered

	suggestions := map[string]string{
		"context":      "<!-- Context will be auto-populated based on Claude session analysis -->",
		"decision":     "<!-- Decision will be auto-populated based on Claude session analysis -->",
		"rationale":    "<!-- Rationale will be auto-populated based on Claude session analysis -->",
		"consequences": "<!-- Consequences will be auto-populated based on Claude session analysis -->",
	}

	return suggestions, fmt.Errorf("not implemented: AI-powered content suggestion is planned for future releases")
}

// GetClaudeDirectory returns the directory where Claude stores its data
func (i *Integration) GetClaudeDirectory() (string, error) {
	// Common locations for Claude Code CLI data
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	possibleDirs := []string{
		filepath.Join(homeDir, ".claude"),
		filepath.Join(homeDir, ".config", "claude"),
		filepath.Join(homeDir, "Library", "Application Support", "claude"), // macOS
	}

	for _, dir := range possibleDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}

	return "", fmt.Errorf("claude directory not found")
}

// WatchForChanges sets up file system watching for Claude session changes
func (i *Integration) WatchForChanges() error {
	// TODO: Implement file system watching
	// This would monitor Claude's session files and trigger ADR updates
	return fmt.Errorf("not implemented: change watching is planned for future releases")
}

// AnalyzeChanges sends a prompt to Claude for change analysis
func (i *Integration) AnalyzeChanges(prompt string) (string, error) {
	if !i.IsAvailable() {
		return "", fmt.Errorf("claude command not available")
	}

	// Use claude command with --print flag for non-interactive analysis
	cmd := exec.Command("claude", "--print", prompt)
	
	output, err := cmd.Output()
	if err != nil {
		// If direct command fails, provide a fallback analysis
		return i.fallbackAnalysis(prompt)
	}

	response := strings.TrimSpace(string(output))
	if response == "" {
		return i.fallbackAnalysis(prompt)
	}

	return response, nil
}

// fallbackAnalysis provides basic heuristic analysis when AI is unavailable
func (i *Integration) fallbackAnalysis(prompt string) (string, error) {
	// Extract changes from the prompt for basic analysis
	changes := i.extractChangesFromPrompt(prompt)
	
	// Basic heuristics for architectural decisions
	architecturalKeywords := []string{
		"database", "api", "framework", "architecture", "design pattern",
		"authentication", "authorization", "security", "performance",
		"integration", "microservice", "monolith", "deployment",
		"technology stack", "library", "dependency", "configuration",
	}
	
	uiKeywords := []string{
		"css", "style", "ui", "frontend", "button", "color", "theme",
		"layout", "responsive", "animation", "visual", "design system",
	}
	
	bugfixKeywords := []string{
		"fix", "bug", "error", "issue", "typo", "hotfix",
		"patch", "correction", "debug",
	}
	
	changes = strings.ToLower(changes)
	
	// Check for bug fixes first (lowest priority)
	for _, keyword := range bugfixKeywords {
		if strings.Contains(changes, keyword) {
			return `**Decision**: No
**Reasoning**: Changes appear to be bug fixes or patches, which typically don't require architectural documentation
**Suggested ADR Title**: N/A
**Key Points**: N/A`, nil
		}
	}
	
	// Check for UI-only changes
	uiScore := 0
	for _, keyword := range uiKeywords {
		if strings.Contains(changes, keyword) {
			uiScore++
		}
	}
	
	// Check for architectural changes
	archScore := 0
	for _, keyword := range architecturalKeywords {
		if strings.Contains(changes, keyword) {
			archScore++
		}
	}
	
	// Decision logic
	if archScore > 0 {
		return `**Decision**: Yes
**Reasoning**: Changes contain architectural keywords suggesting significant system decisions that should be documented
**Suggested ADR Title**: document-recent-architectural-changes
**Key Points**: 
- Document the architectural decision and its rationale
- Consider long-term implications and alternatives
- Ensure team alignment on the approach`, nil
	}
	
	if uiScore > 2 && archScore == 0 {
		return `**Decision**: No
**Reasoning**: Changes appear to be primarily UI/styling updates without architectural implications
**Suggested ADR Title**: N/A
**Key Points**: N/A`, nil
	}
	
	// Default to requiring ADR for safety when uncertain
	return `**Decision**: Yes
**Reasoning**: Unable to definitively categorize changes - recommending ADR for safety and team communication
**Suggested ADR Title**: document-recent-changes
**Key Points**: 
- Review and document the purpose of these changes
- Consider if they establish new patterns or approaches
- Ensure team understanding and alignment`, nil
}

// extractChangesFromPrompt extracts the actual code changes from the analysis prompt
func (i *Integration) extractChangesFromPrompt(prompt string) string {
	lines := strings.Split(prompt, "\n")
	inChangesSection := false
	var changes []string
	
	for _, line := range lines {
		if strings.Contains(line, "## Code Changes to Analyze") {
			inChangesSection = true
			continue
		}
		if inChangesSection {
			if strings.HasPrefix(line, "##") && !strings.Contains(line, "Code Changes") {
				break
			}
			if !strings.HasPrefix(line, "```") {
				changes = append(changes, line)
			}
		}
	}
	
	return strings.Join(changes, "\n")
}