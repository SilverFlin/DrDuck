package cursor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TokenUsage tracks token consumption for AI requests
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// AnalyzeResult contains both the response and token usage information
type AnalyzeResult struct {
	Response    string     `json:"response"`
	TokenUsage  TokenUsage `json:"token_usage"`
}

// CursorSession represents information about a Cursor AI session
type CursorSession struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
	Messages  int      `json:"messages"`
	Files     []string `json:"files,omitempty"`
	Context   string   `json:"context,omitempty"`
}

// Integration handles Cursor integration
type Integration struct {
	// Future: Add configuration options
}

// NewIntegration creates a new Cursor integration instance
func NewIntegration() *Integration {
	return &Integration{}
}

// IsAvailable checks if Cursor is available
func (i *Integration) IsAvailable() bool {
	// Check for Cursor command or installation
	_, err := exec.LookPath("cursor")
	if err == nil {
		return true
	}

	// Check common Cursor installation paths
	possiblePaths := []string{
		"/Applications/Cursor.app/Contents/Resources/app/bin/cursor", // macOS
		"/usr/local/bin/cursor",                                      // Linux
		"C:\\Users\\%USERNAME%\\AppData\\Local\\Programs\\cursor\\cursor.exe", // Windows
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// GetCurrentSession attempts to get information about the current Cursor session
func (i *Integration) GetCurrentSession() (*CursorSession, error) {
	if !i.IsAvailable() {
		return nil, fmt.Errorf("cursor not available")
	}

	// TODO: Implement actual Cursor integration
	// This is a placeholder structure for future implementation
	// We would need to:
	// 1. Access Cursor's AI conversation history
	// 2. Parse recent interactions
	// 3. Extract file changes and context
	// 4. Return session information

	return &CursorSession{
		ID:        "placeholder",
		ProjectID: "current-project",
	}, fmt.Errorf("not implemented: Cursor integration is planned for future releases")
}

// ExtractContext extracts relevant context from Cursor session for ADR generation
func (i *Integration) ExtractContext(session *CursorSession) (string, error) {
	// TODO: Implement context extraction
	// This would analyze:
	// - Recent AI conversations
	// - Code changes made through Cursor AI
	// - Architectural discussions
	// - Decision points in the conversation

	return "", fmt.Errorf("not implemented: context extraction is planned for future releases")
}

// GetChangedFiles returns files that have been modified in Cursor AI sessions
func (i *Integration) GetChangedFiles() ([]string, error) {
	// TODO: Implement file change tracking
	// This could work by:
	// 1. Monitoring Cursor's AI-assisted changes
	// 2. Tracking files modified through AI suggestions
	// 3. Integrating with git to identify AI-generated changes
	// 4. Parsing Cursor's conversation logs for file mentions

	return nil, fmt.Errorf("not implemented: file change tracking is planned for future releases")
}

// SuggestADRContent generates ADR content suggestions based on Cursor AI session
func (i *Integration) SuggestADRContent(adrName string) (map[string]string, error) {
	// TODO: Implement AI-powered ADR content suggestion
	// This would analyze the Cursor AI session and suggest:
	// - Context based on AI conversations
	// - Decision rationale from AI recommendations
	// - Consequences identified during AI interactions
	// - Alternatives discussed with AI

	suggestions := map[string]string{
		"context":      "<!-- Context will be auto-populated based on Cursor AI session analysis -->",
		"decision":     "<!-- Decision will be auto-populated based on Cursor AI session analysis -->",
		"rationale":    "<!-- Rationale will be auto-populated based on Cursor AI session analysis -->",
		"consequences": "<!-- Consequences will be auto-populated based on Cursor AI session analysis -->",
	}

	return suggestions, fmt.Errorf("not implemented: AI-powered content suggestion is planned for future releases")
}

// GetCursorDirectory returns the directory where Cursor stores its data
func (i *Integration) GetCursorDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Common locations for Cursor data
	possibleDirs := []string{
		filepath.Join(homeDir, ".cursor"),
		filepath.Join(homeDir, ".config", "cursor"),
		filepath.Join(homeDir, "Library", "Application Support", "Cursor"), // macOS
		filepath.Join(homeDir, "AppData", "Roaming", "Cursor"),             // Windows
	}

	for _, dir := range possibleDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}

	return "", fmt.Errorf("cursor directory not found")
}

// WatchForChanges sets up monitoring for Cursor AI session changes
func (i *Integration) WatchForChanges() error {
	// TODO: Implement change monitoring
	// This would watch for:
	// - New AI conversations
	// - Code changes through AI assistance
	// - Project modifications via Cursor AI
	return fmt.Errorf("not implemented: change monitoring is planned for future releases")
}

// AnalyzeChanges sends a prompt to Cursor for change analysis
func (i *Integration) AnalyzeChanges(prompt string) (string, error) {
	if !i.IsAvailable() {
		return "", fmt.Errorf("cursor not available")
	}

	// Try to use Cursor's AI capabilities for analysis
	// This is a basic implementation - Cursor doesn't have a direct CLI for prompts like Claude
	// We'll provide a fallback analysis similar to Claude's implementation
	
	return i.fallbackAnalysis(prompt)
}

// AnalyzeChangesWithTokens sends a prompt to Cursor for change analysis and returns token usage
func (i *Integration) AnalyzeChangesWithTokens(prompt string) (string, *TokenUsage, error) {
	response, err := i.fallbackAnalysis(prompt)
	if err != nil {
		return "", nil, err
	}
	
	tokenUsage := &TokenUsage{
		InputTokens:  estimateTokens(prompt),
		OutputTokens: estimateTokens(response),
		TotalTokens:  estimateTokens(prompt) + estimateTokens(response),
	}
	return response, tokenUsage, nil
}

// estimateTokens provides a rough estimate of token count for a given text
func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	cleanText := strings.ReplaceAll(text, "\n", " ")
	return len(strings.TrimSpace(cleanText)) / 4
}

// fallbackAnalysis provides basic heuristic analysis when AI is unavailable
func (i *Integration) fallbackAnalysis(prompt string) (string, error) {
	// Extract changes from the prompt for basic analysis
	changes := i.extractChangesFromPrompt(prompt)
	
	// Basic heuristics for architectural decisions (similar to Claude implementation)
	architecturalKeywords := []string{
		"database", "api", "framework", "architecture", "design pattern",
		"authentication", "authorization", "security", "performance",
		"integration", "microservice", "monolith", "deployment",
		"technology stack", "library", "dependency", "configuration",
		"schema", "migration", "service", "interface",
	}
	
	uiKeywords := []string{
		"css", "style", "ui", "frontend", "button", "color", "theme",
		"layout", "responsive", "animation", "visual", "design system",
		"component", "react", "vue", "angular",
	}
	
	bugfixKeywords := []string{
		"fix", "bug", "error", "issue", "typo", "hotfix",
		"patch", "correction", "debug", "resolve",
	}
	
	testKeywords := []string{
		"test", "spec", "unittest", "integration test", "e2e",
		"coverage", "mock", "stub", "assert",
	}
	
	changes = strings.ToLower(changes)
	
	// Check for bug fixes first (lowest priority)
	bugScore := 0
	for _, keyword := range bugfixKeywords {
		if strings.Contains(changes, keyword) {
			bugScore++
		}
	}
	
	if bugScore > 0 {
		return `**Decision**: No
**Reasoning**: Changes appear to be bug fixes or patches, which typically don't require architectural documentation
**Suggested ADR Title**: N/A
**Key Points**: N/A`, nil
	}
	
	// Check for test-only changes
	testScore := 0
	for _, keyword := range testKeywords {
		if strings.Contains(changes, keyword) {
			testScore++
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
	if archScore >= 2 {
		return `**Decision**: Yes
**Reasoning**: Changes contain multiple architectural indicators suggesting significant system decisions that should be documented
**Suggested ADR Title**: document-recent-architectural-changes
**Key Points**: 
- Document the architectural decision and its rationale
- Consider long-term implications and alternatives
- Ensure team alignment on the approach`, nil
	}
	
	if testScore > 1 && archScore == 0 && uiScore == 0 {
		return `**Decision**: No
**Reasoning**: Changes appear to be primarily test-related improvements without architectural implications
**Suggested ADR Title**: N/A
**Key Points**: N/A`, nil
	}
	
	if uiScore > 2 && archScore == 0 {
		return `**Decision**: No
**Reasoning**: Changes appear to be primarily UI/styling updates without architectural implications
**Suggested ADR Title**: N/A
**Key Points**: N/A`, nil
	}
	
	// If we detect one architectural keyword, be cautious but recommend ADR
	if archScore == 1 {
		return `**Decision**: Yes
**Reasoning**: Changes touch architectural components - recommending ADR to ensure proper documentation and team alignment
**Suggested ADR Title**: document-architectural-change
**Key Points**: 
- Clarify the architectural decision being made
- Document reasoning and alternatives considered
- Ensure team understanding of the impact`, nil
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