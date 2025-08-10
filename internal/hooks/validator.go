package hooks

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/SilverFlin/DrDuck/internal/adr"
	"github.com/SilverFlin/DrDuck/internal/ai"
	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/SilverFlin/DrDuck/internal/prompts/templates"
	"github.com/charmbracelet/huh"
)

// ValidationResult represents the result of hook validation
type ValidationResult struct {
	ShouldBlock     bool
	DraftADRs       []*adr.ADR
	NeedsADR        bool
	SuggestedTitle  string
	Message         string
	AIResponse      string
}

// Validator handles git hook validation logic
type Validator struct {
	config     *config.Config
	adrManager *adr.Manager
	aiManager  *ai.Manager
}

// NewValidator creates a new hook validator
func NewValidator(cfg *config.Config) *Validator {
	return &Validator{
		config:     cfg,
		adrManager: adr.NewManager(cfg),
		aiManager:  ai.NewManager(cfg),
	}
}

// ValidatePreCommit performs pre-commit validation (warns but never blocks)
func (v *Validator) ValidatePreCommit() *ValidationResult {
	result := &ValidationResult{
		ShouldBlock: false, // Pre-commit never blocks
	}

	// Check for draft ADRs
	drafts, err := v.getDraftADRs()
	if err != nil {
		result.Message = fmt.Sprintf("⚠️  Could not check ADR status: %v", err)
		return result
	}

	result.DraftADRs = drafts

	if len(drafts) == 0 {
		result.Message = "🦆 DrDuck: All ADRs are up to date! ✨"
		return result
	}

	// Build encouraging message for draft ADRs
	var messageBuilder strings.Builder
	messageBuilder.WriteString("🦆 DrDuck: Found draft ADRs that could use some attention:\n")

	for _, draft := range drafts {
		daysSince := int(time.Since(draft.Date).Hours() / 24)
		daysText := "today"
		if daysSince == 1 {
			daysText = "1 day ago"
		} else if daysSince > 1 {
			daysText = fmt.Sprintf("%d days ago", daysSince)
		}

		messageBuilder.WriteString(fmt.Sprintf("   📝 ADR-%04d: %s (created %s)\n", 
			draft.ID, draft.Title, daysText))
	}

	messageBuilder.WriteString("\n💡 Consider completing these before your next push:\n")
	for _, draft := range drafts {
		messageBuilder.WriteString(fmt.Sprintf("   drduck edit %04d  # Complete %s\n", 
			draft.ID, draft.Title))
	}

	messageBuilder.WriteString("\n✨ Commit proceeding as normal...")
	result.Message = messageBuilder.String()

	return result
}

// ValidatePrePush performs pre-push validation (can block)
func (v *Validator) ValidatePrePush() *ValidationResult {
	result := &ValidationResult{}

	// First, check for draft ADRs (blocking)
	drafts, err := v.getDraftADRs()
	if err != nil {
		result.ShouldBlock = true
		result.Message = fmt.Sprintf("❌ Could not check ADR status: %v", err)
		return result
	}

	result.DraftADRs = drafts

	if len(drafts) > 0 {
		result.ShouldBlock = true
		var messageBuilder strings.Builder
		messageBuilder.WriteString("🚫 DrDuck: Cannot push with draft ADRs!\n\n")
		
		for _, draft := range drafts {
			daysSince := int(time.Since(draft.Date).Hours() / 24)
			messageBuilder.WriteString(fmt.Sprintf("   📝 ADR-%04d: %s (%d days in draft)\n", 
				draft.ID, draft.Title, daysSince))
		}

		messageBuilder.WriteString("\n🔧 To proceed:\n")
		messageBuilder.WriteString("   1. Complete draft ADRs with AI assistance:\n")
		for _, draft := range drafts {
			messageBuilder.WriteString(fmt.Sprintf("      drduck complete-adr %04d  # AI-assisted completion\n", draft.ID))
		}
		messageBuilder.WriteString("   2. Or edit manually:\n")
		for _, draft := range drafts {
			messageBuilder.WriteString(fmt.Sprintf("      drduck edit %04d\n", draft.ID))
		}
		messageBuilder.WriteString("   3. Or use emergency bypass: git push --no-verify\n\n")
		messageBuilder.WriteString("💡 Tip: 'complete-adr' uses AI to fill content based on your changes")

		result.Message = messageBuilder.String()
		return result
	}

	// If no drafts, check if changes need a new ADR using AI
	needsADR, aiResponse, suggestedTitle, err := v.analyzeChangesForADR()
	if err != nil {
		// Don't block on AI errors, just warn
		result.Message = fmt.Sprintf("⚠️  Could not analyze changes with AI: %v\n✅ Push proceeding...", err)
		return result
	}

	result.NeedsADR = needsADR
	result.AIResponse = aiResponse
	result.SuggestedTitle = suggestedTitle

	if needsADR {
		// Ask user if they want to create ADR automatically
		shouldCreate, err := v.askUserToCreateADR(suggestedTitle, aiResponse)
		if err == nil && shouldCreate {
			// Run complete-adr --create automatically
			createResult := v.runCompleteADRCreate()
			if createResult.Success {
				result.Message = fmt.Sprintf("🎉 ADR created successfully!\n%s\n\n✅ Push proceeding...", createResult.Message)
				return result
			} else {
				result.ShouldBlock = true
				result.Message = fmt.Sprintf("❌ ADR creation failed: %s\n\nPlease create ADR manually or use --no-verify", createResult.Message)
				return result
			}
		}

		// User declined or error occurred - show original blocking message
		result.ShouldBlock = true
		var messageBuilder strings.Builder
		messageBuilder.WriteString("🚫 DrDuck: These changes appear to need an ADR!\n\n")
		messageBuilder.WriteString("🤖 Dr Duck's Analysis:\n")
		messageBuilder.WriteString(aiResponse)
		messageBuilder.WriteString("\n\n🔧 To proceed:\n")
		messageBuilder.WriteString("   1. Create ADR with AI assistance:\n")
		messageBuilder.WriteString("      drduck complete-adr --create  # AI-guided ADR creation\n")
		messageBuilder.WriteString("   2. Or create manually:\n")
		if suggestedTitle != "" {
			messageBuilder.WriteString(fmt.Sprintf("      drduck new -n \"%s\"\n", suggestedTitle))
		} else {
			messageBuilder.WriteString("      drduck new -n \"your-decision-name\"\n")
		}
		messageBuilder.WriteString("   3. Or use emergency bypass: git push --no-verify\n\n")
		messageBuilder.WriteString("💡 Recommended: Use 'complete-adr --create' for AI-assisted ADR generation")

		result.Message = messageBuilder.String()
		return result
	}

	// All good!
	result.Message = "🦆 DrDuck: All checks passed! ✨ Push proceeding..."
	return result
}

// getDraftADRs returns all ADRs currently in draft status
func (v *Validator) getDraftADRs() ([]*adr.ADR, error) {
	allADRs, err := v.adrManager.List()
	if err != nil {
		return nil, err
	}

	var drafts []*adr.ADR
	for _, a := range allADRs {
		if a.Status == adr.StatusDraft {
			drafts = append(drafts, a)
		}
	}

	return drafts, nil
}

// analyzeChangesForADR uses AI to determine if the current changes require an ADR
func (v *Validator) analyzeChangesForADR() (needsADR bool, aiResponse string, suggestedTitle string, err error) {
	// Check if AI provider is available
	if !v.aiManager.IsAvailable() {
		return false, "", "", fmt.Errorf("AI provider (%s) not available", v.aiManager.GetProviderName())
	}

	// Get git changes since last push
	changes, err := v.getGitChangesSinceLastPush()
	if err != nil {
		return false, "", "", fmt.Errorf("failed to get git changes: %w", err)
	}

	if strings.TrimSpace(changes) == "" {
		return false, "No changes to analyze", "", nil
	}

	// Get recent commit context
	recentCommits, err := v.getRecentCommits()
	if err != nil {
		// Don't fail if we can't get commits, just continue without context
		recentCommits = ""
	}

	// Generate analysis prompt
	prompt := templates.ChangeAnalysisPrompt("", changes, recentCommits)

	// Use AI to analyze changes (this will be implemented in the AI integration)
	response, err := v.analyzeWithAI(prompt)
	if err != nil {
		return false, "", "", err
	}

	// Parse AI response to determine if ADR is needed
	// Look for the structured response format from our prompt
	responseLower := strings.ToLower(response)
	needsADR = strings.Contains(responseLower, "**decision**: yes") ||
		strings.Contains(responseLower, "decision**: yes") ||
		strings.Contains(responseLower, "decision: yes")

	// Try to extract suggested title (simple regex parsing)
	if needsADR {
		lines := strings.Split(response, "\n")
		for _, line := range lines {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "suggested adr title") || strings.Contains(lower, "title:") {
				// Extract title from line - this is a simple implementation
				if idx := strings.Index(lower, ":"); idx != -1 && idx+1 < len(line) {
					title := strings.TrimSpace(line[idx+1:])
					title = strings.Trim(title, "\"'`")
					if title != "" {
						suggestedTitle = title
						break
					}
				}
			}
		}
	}

	return needsADR, response, suggestedTitle, nil
}

// getGitChangesSinceLastPush gets the diff since the last push to current branch
func (v *Validator) getGitChangesSinceLastPush() (string, error) {
	// Get current branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get diff since last push (using origin/branch or fallback to last few commits)
	var diffCmd *exec.Cmd
	
	// Try to get diff against origin
	remoteBranch := fmt.Sprintf("origin/%s", branch)
	diffCmd = exec.Command("git", "diff", remoteBranch+"..HEAD")
	output, err := diffCmd.Output()
	
	// If no remote, try to get last 3 commits
	if err != nil {
		diffCmd = exec.Command("git", "diff", "HEAD~3..HEAD")
		output, err = diffCmd.Output()
		
		// If still no luck, get all staged and unstaged changes
		if err != nil {
			diffCmd = exec.Command("git", "diff", "HEAD")
			output, err = diffCmd.Output()
			if err != nil {
				return "", fmt.Errorf("failed to get git changes: %w", err)
			}
		}
	}

	return string(output), nil
}

// getRecentCommits gets recent commit messages for context
func (v *Validator) getRecentCommits() (string, error) {
	cmd := exec.Command("git", "log", "--oneline", "-5")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// analyzeWithAI sends the prompt to the configured AI provider
func (v *Validator) analyzeWithAI(prompt string) (string, error) {
	return v.aiManager.AnalyzeChanges(prompt)
}

// ADRCreateResult represents the result of automatic ADR creation
type ADRCreateResult struct {
	Success bool
	Message string
}

// askUserToCreateADR prompts the user to create an ADR automatically
func (v *Validator) askUserToCreateADR(suggestedTitle, aiAnalysis string) (bool, error) {
	var shouldCreate bool
	
	title := "DrDuck detected changes that need an ADR. Create one now?"
	description := "This will run 'drduck complete-adr --create' automatically using AI assistance."
	
	if suggestedTitle != "" {
		description += fmt.Sprintf("\nSuggested title: %s", suggestedTitle)
	}
	
	form := huh.NewConfirm().
		Title(title).
		Description(description).
		Value(&shouldCreate).
		Affirmative("Yes, create ADR now").
		Negative("No, I'll handle it manually")

	err := form.Run()
	if err != nil {
		return false, err
	}

	return shouldCreate, nil
}

// runCompleteADRCreate executes the complete-adr --create command
func (v *Validator) runCompleteADRCreate() ADRCreateResult {
	// Import the complete-adr functionality
	// We'll use os/exec to call the drduck command to avoid circular imports
	cmd := exec.Command("drduck", "complete-adr", "--create")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ADRCreateResult{
			Success: false,
			Message: fmt.Sprintf("Command failed: %v\nOutput: %s", err, string(output)),
		}
	}

	return ADRCreateResult{
		Success: true,
		Message: string(output),
	}
}