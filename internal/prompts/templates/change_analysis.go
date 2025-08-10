package templates

import (
	"fmt"
	"strings"

	"github.com/SilverFlin/DrDuck/internal/prompts/personas"
)

// ChangeAnalysisPrompt generates a prompt for analyzing git changes to determine if an ADR is needed
func ChangeAnalysisPrompt(projectName, changes, recentCommits string) string {
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString(personas.DrDuckPersona)
	promptBuilder.WriteString("\n\n")
	
	promptBuilder.WriteString("# CHANGE ANALYSIS REQUEST\n\n")
	
	if projectName != "" {
		promptBuilder.WriteString(fmt.Sprintf("**Project**: %s\n", projectName))
	}
	
	promptBuilder.WriteString("**Task**: Analyze the following code changes and determine if they require an Architectural Decision Record (ADR).\n\n")
	
	if recentCommits != "" {
		promptBuilder.WriteString("## Recent Commit Context\n")
		promptBuilder.WriteString("```\n")
		promptBuilder.WriteString(recentCommits)
		promptBuilder.WriteString("\n```\n\n")
	}
	
	promptBuilder.WriteString("## Code Changes to Analyze\n")
	promptBuilder.WriteString("```diff\n")
	promptBuilder.WriteString(changes)
	promptBuilder.WriteString("\n```\n\n")
	
	promptBuilder.WriteString("## Analysis Required\n")
	promptBuilder.WriteString("Please analyze these changes and provide your assessment in EXACTLY this format:\n\n")
	promptBuilder.WriteString("**Decision**: Yes OR No\n")
	promptBuilder.WriteString("**Reasoning**: Brief explanation of why/why not\n")
	promptBuilder.WriteString("**Suggested ADR Title**: If yes, propose a specific title in kebab-case\n")
	promptBuilder.WriteString("**Key Points**: If yes, 2-3 bullet points of what the ADR should cover\n\n")
	promptBuilder.WriteString("IMPORTANT: Start your response with exactly '**Decision**: Yes' or '**Decision**: No'\n\n")
	
	promptBuilder.WriteString("Focus on architectural significance rather than implementation details. ")
	promptBuilder.WriteString("Consider the long-term impact on the codebase, team understanding, and future maintainability.")
	
	return promptBuilder.String()
}

// DraftCompletionPrompt generates a prompt for suggesting how to complete draft ADRs
func DraftCompletionPrompt(adrTitle, currentContent string, daysSinceDraft int) string {
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString(personas.DrDuckPersona)
	promptBuilder.WriteString("\n\n")
	
	promptBuilder.WriteString("# ADR COMPLETION ASSISTANCE\n\n")
	promptBuilder.WriteString(fmt.Sprintf("**ADR Title**: %s\n", adrTitle))
	
	if daysSinceDraft > 0 {
		promptBuilder.WriteString(fmt.Sprintf("**Days since created**: %d\n", daysSinceDraft))
	}
	
	promptBuilder.WriteString("**Status**: Currently in Draft\n\n")
	
	promptBuilder.WriteString("## Current ADR Content\n")
	promptBuilder.WriteString("```markdown\n")
	promptBuilder.WriteString(currentContent)
	promptBuilder.WriteString("\n```\n\n")
	
	promptBuilder.WriteString("## Assistance Needed\n")
	promptBuilder.WriteString("This ADR has been in draft status and needs completion. Please provide:\n\n")
	promptBuilder.WriteString("1. **Missing Sections**: Which sections need content?\n")
	promptBuilder.WriteString("2. **Content Suggestions**: Brief suggestions for each missing section\n")
	promptBuilder.WriteString("3. **Questions to Consider**: Key questions the team should answer\n")
	promptBuilder.WriteString("4. **Next Steps**: Specific actions to move this ADR forward\n\n")
	
	promptBuilder.WriteString("Keep suggestions practical and actionable. Focus on helping the team ")
	promptBuilder.WriteString("document their decision-making process effectively.")
	
	return promptBuilder.String()
}