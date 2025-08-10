package templates

import (
	"fmt"
	"strings"

	"github.com/SilverFlin/DrDuck/internal/prompts/personas"
)

// ADRContentSuggestionPrompt generates a prompt for suggesting ADR content based on context
func ADRContentSuggestionPrompt(adrTitle, projectContext, changeContext string) string {
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString(personas.DrDuckPersona)
	promptBuilder.WriteString("\n\n")
	
	promptBuilder.WriteString("# ADR CONTENT SUGGESTION REQUEST\n\n")
	
	promptBuilder.WriteString(fmt.Sprintf("**ADR Title**: %s\n\n", adrTitle))
	
	if projectContext != "" {
		promptBuilder.WriteString("## Project Context\n")
		promptBuilder.WriteString(projectContext)
		promptBuilder.WriteString("\n\n")
	}
	
	if changeContext != "" {
		promptBuilder.WriteString("## Change Context\n")
		promptBuilder.WriteString(changeContext)
		promptBuilder.WriteString("\n\n")
	}
	
	promptBuilder.WriteString("## Content Generation Request\n")
	promptBuilder.WriteString("Based on the provided context, please suggest content for the following ADR sections:\n\n")
	
	promptBuilder.WriteString("1. **Context**: What problem or situation motivated this decision?\n")
	promptBuilder.WriteString("2. **Decision**: What solution or approach was chosen?\n")
	promptBuilder.WriteString("3. **Rationale**: Why was this particular solution selected?\n")
	promptBuilder.WriteString("4. **Consequences**: What are the positive, negative, and neutral implications?\n")
	promptBuilder.WriteString("5. **Alternatives**: What other options were considered?\n\n")
	
	promptBuilder.WriteString("**Guidelines**:\n")
	promptBuilder.WriteString("- Keep suggestions concise but informative\n")
	promptBuilder.WriteString("- Focus on architectural and long-term considerations\n")
	promptBuilder.WriteString("- Include placeholder text where specific details need team input\n")
	promptBuilder.WriteString("- Highlight areas that need further investigation or discussion\n\n")
	
	promptBuilder.WriteString("Provide practical, ready-to-use content that teams can build upon.")
	
	return promptBuilder.String()
}

// ADRReviewPrompt generates a prompt for reviewing completed ADRs
func ADRReviewPrompt(adrTitle, adrContent string) string {
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString(personas.DrDuckPersona)
	promptBuilder.WriteString("\n\n")
	
	promptBuilder.WriteString("# ADR REVIEW REQUEST\n\n")
	
	promptBuilder.WriteString(fmt.Sprintf("**ADR Title**: %s\n\n", adrTitle))
	
	promptBuilder.WriteString("## ADR Content to Review\n")
	promptBuilder.WriteString("```markdown\n")
	promptBuilder.WriteString(adrContent)
	promptBuilder.WriteString("\n```\n\n")
	
	promptBuilder.WriteString("## Review Areas\n")
	promptBuilder.WriteString("Please review this ADR and provide feedback on:\n\n")
	
	promptBuilder.WriteString("1. **Clarity**: Is the decision and reasoning clearly explained?\n")
	promptBuilder.WriteString("2. **Completeness**: Are all necessary sections adequately filled?\n")
	promptBuilder.WriteString("3. **Consequences**: Are the implications thoroughly considered?\n")
	promptBuilder.WriteString("4. **Alternatives**: Are alternative approaches adequately covered?\n")
	promptBuilder.WriteString("5. **Future Maintainability**: Will this ADR be useful for future team members?\n\n")
	
	promptBuilder.WriteString("**Provide**:\n")
	promptBuilder.WriteString("- **Overall Assessment**: Is this ADR ready for acceptance?\n")
	promptBuilder.WriteString("- **Suggestions**: Specific improvements or additions needed\n")
	promptBuilder.WriteString("- **Strengths**: What this ADR does well\n")
	promptBuilder.WriteString("- **Questions**: Any unclear areas that need elaboration\n\n")
	
	promptBuilder.WriteString("Focus on helping the team create documentation that will be valuable ")
	promptBuilder.WriteString("for current and future developers.")
	
	return promptBuilder.String()
}