package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SilverFlin/DrDuck/internal/adr"
	"github.com/SilverFlin/DrDuck/internal/ai"
	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/SilverFlin/DrDuck/internal/prompts/templates"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var createNewADR bool

var completeADRCmd = &cobra.Command{
	Use:   "complete-adr [adr-id]",
	Short: "AI-assisted interactive ADR completion",
	Long: `Complete an ADR through AI-assisted interactive prompts. This command analyzes
your code changes and asks targeted questions to generate comprehensive ADR content.

Usage scenarios:
  drduck complete-adr 0001      # Complete existing draft ADR
  drduck complete-adr --create  # Create and complete new ADR

The command will:
1. Analyze your git changes using AI
2. Ask targeted questions based on change type
3. Generate complete ADR content from your responses
4. Preview the content and allow confirmation
5. Save the completed ADR`,
	Args: func(cmd *cobra.Command, args []string) error {
		if createNewADR && len(args) > 0 {
			return fmt.Errorf("cannot specify ADR ID when using --create flag")
		}
		if !createNewADR && len(args) != 1 {
			return fmt.Errorf("must specify ADR ID or use --create flag")
		}
		return nil
	},
	RunE: runCompleteADR,
}

func init() {
	rootCmd.AddCommand(completeADRCmd)
	completeADRCmd.Flags().BoolVar(&createNewADR, "create", false, "Create a new ADR instead of completing existing one")
}

func runCompleteADR(cmd *cobra.Command, args []string) error {
	// Check if project is initialized
	initialized, err := config.IsInitialized()
	if err != nil {
		return fmt.Errorf("failed to check initialization status: %w", err)
	}

	if !initialized {
		return fmt.Errorf("❌ DrDuck is not initialized in this project. Run 'drduck init' first")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create managers
	adrManager := adr.NewManager(cfg)
	aiManager := ai.NewManager(cfg)

	fmt.Println("🦆 Welcome to AI-Assisted ADR Completion!")
	fmt.Println("=====================================")
	fmt.Println()

	var targetADR *adr.ADR
	var isNewADR bool

	if createNewADR {
		// Create new ADR workflow
		fmt.Println("📝 Creating a new ADR based on your recent changes...")
		targetADR, err = createNewADRFromChanges(adrManager, aiManager)
		if err != nil {
			return fmt.Errorf("failed to create new ADR: %w", err)
		}
		isNewADR = true
	} else {
		// Complete existing ADR workflow
		adrIDStr := args[0]
		adrID, err := strconv.Atoi(strings.TrimLeft(adrIDStr, "0"))
		if err != nil {
			return fmt.Errorf("invalid ADR ID: %s", adrIDStr)
		}

		targetADR, err = adrManager.GetADRByID(adrID)
		if err != nil {
			return fmt.Errorf("ADR not found: %w", err)
		}

		fmt.Printf("📝 Completing ADR-%04d: %s\n", targetADR.ID, targetADR.Title)
		fmt.Printf("📊 Current Status: %s\n", targetADR.Status)
	}

	fmt.Println()

	// Step 1: Analyze current changes
	fmt.Println("🔍 Step 1: Analyzing your code changes...")
	changes, changeAnalysis, err := analyzeRecentChanges(aiManager)
	if err != nil {
		fmt.Printf("⚠️  Could not analyze changes: %v\n", err)
		fmt.Println("Continuing with manual input...")
		changes = ""
		changeAnalysis = ""
	} else {
		fmt.Println("✅ Change analysis completed")
		if changeAnalysis != "" {
			fmt.Println("\n🤖 AI Analysis of Changes:")
			fmt.Println("---")
			fmt.Println(changeAnalysis)
			fmt.Println("---")
			
			// Check if AI recommends skipping ADR creation
			if createNewADR && shouldSkipADRCreation(changeAnalysis) {
				shouldSkip, err := askUserToSkipADR(changeAnalysis)
				if err != nil {
					return fmt.Errorf("failed to get user decision: %w", err)
				}
				if shouldSkip {
					return nil // Exit early, ADR creation skipped
				}
			}
		}
	}

	// Step 2: Interactive questionnaire
	fmt.Println("\n💬 Step 2: Let's gather context about your decision...")
	responses, err := conductInteractiveQuestionnaire(targetADR.Title, changeAnalysis)
	if err != nil {
		return fmt.Errorf("questionnaire failed: %w", err)
	}

	// Step 3: Generate ADR content using AI
	fmt.Println("\n🤖 Step 3: Generating ADR content with AI...")
	generatedContent, err := generateADRContent(aiManager, targetADR, changes, changeAnalysis, responses)
	if err != nil {
		return fmt.Errorf("failed to generate ADR content: %w", err)
	}

	// Step 4: Preview and confirm
	fmt.Println("\n👀 Step 4: Review generated content...")
	confirmed, finalContent, err := previewAndConfirm(generatedContent)
	if err != nil {
		return fmt.Errorf("preview failed: %w", err)
	}

	if !confirmed {
		fmt.Println("❌ ADR completion cancelled by user")
		return nil
	}

	// Step 5: Save the completed ADR
	fmt.Println("\n💾 Step 5: Saving completed ADR...")
	if err := saveCompletedADR(targetADR, finalContent); err != nil {
		return fmt.Errorf("failed to save ADR: %w", err)
	}

	// Step 6: Ask about status
	if err := handleADRStatusUpdate(adrManager, targetADR, isNewADR); err != nil {
		return fmt.Errorf("status update failed: %w", err)
	}

	// Step 7: Ask about title confirmation for new ADRs
	if isNewADR {
		if err := handleTitleConfirmation(adrManager, targetADR); err != nil {
			return fmt.Errorf("title confirmation failed: %w", err)
		}
	}

	fmt.Println("\n🎉 ADR completion successful!")
	fmt.Printf("📄 File: %s\n", targetADR.FilePath)
	fmt.Println("💡 Your changes should now pass the pre-push hook")

	return nil
}

// createNewADRFromChanges creates a new ADR with AI-suggested title
func createNewADRFromChanges(adrManager *adr.Manager, aiManager *ai.Manager) (*adr.ADR, error) {
	// Get git changes to suggest title
	changes, err := getGitChangesSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get git changes: %w", err)
	}

	// Use AI to suggest title
	suggestedTitle := "recent-architectural-changes"
	if aiManager.IsAvailable() && changes != "" {
		prompt := fmt.Sprintf("Based on these git changes, suggest a concise ADR title (2-4 words, kebab-case):\n\n%s", changes)
		response, err := aiManager.AnalyzeChanges(prompt)
		if err == nil && strings.Contains(response, "title") {
			// Extract title from response (simple parsing)
			lines := strings.Split(response, "\n")
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), "title") && strings.Contains(line, ":") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						title := strings.TrimSpace(parts[1])
						title = strings.Trim(title, "\"'`")
						if title != "" {
							suggestedTitle = title
							break
						}
					}
				}
			}
		}
	}

	// Create the ADR
	return adrManager.Create(suggestedTitle)
}

// analyzeRecentChanges gets git changes and AI analysis with timeout protection
func analyzeRecentChanges(aiManager *ai.Manager) (changes string, analysis string, err error) {
	// Get basic changes summary first (fast)
	changes, err = getGitChangesSummary()
	if err != nil {
		changes = "Could not detect git changes - proceeding with manual input"
		return changes, "Git analysis unavailable", nil
	}

	if changes == "" || changes == "No git changes detected" {
		return changes, "No significant changes detected", nil
	}

	if !aiManager.IsAvailable() {
		return changes, "AI analysis not available - using change detection only", nil
	}

	// Try to get detailed changes for AI analysis (may be large)
	fmt.Print("Getting detailed changes for AI analysis... ")
	detailedChanges, wasTruncated, err := getDetailedChanges()
	if err != nil {
		fmt.Println("failed, using summary")
		detailedChanges = changes // Fallback to summary
		wasTruncated = false
	} else {
		if wasTruncated {
			fmt.Println("large changeset detected, truncated for analysis")
		} else {
			fmt.Println("done")
		}
	}

	// Prepare AI prompt with size-aware content
	var promptChanges string
	if wasTruncated {
		// Use summary + truncated details for large changes
		promptChanges = fmt.Sprintf("LARGE CHANGESET (truncated):\n%s", detailedChanges)
	} else {
		promptChanges = detailedChanges
	}

	// Run AI analysis with timeout protection
	fmt.Print("Running AI analysis... ")
	analysis, err = analyzeWithTimeout(aiManager, promptChanges, 30*time.Second)
	if err != nil {
		fmt.Printf("failed (%v), using fallback\n", err)
		// Provide intelligent fallback analysis based on change patterns
		analysis = generateFallbackAnalysis(changes, wasTruncated)
		err = nil // Clear error so workflow continues
	} else {
		fmt.Println("completed")
	}
	
	return changes, analysis, err
}

// analyzeWithTimeout runs AI analysis with a timeout
func analyzeWithTimeout(aiManager *ai.Manager, changes string, timeout time.Duration) (string, error) {
	type result struct {
		analysis string
		err      error
	}

	// Create channel for result
	resultChan := make(chan result, 1)

	// Run analysis in goroutine
	go func() {
		prompt := templates.ChangeAnalysisPrompt("", changes, "")
		analysis, err := aiManager.AnalyzeChanges(prompt)
		resultChan <- result{analysis, err}
	}()

	// Wait for result or timeout
	select {
	case res := <-resultChan:
		return res.analysis, res.err
	case <-time.After(timeout):
		return "", fmt.Errorf("AI analysis timed out after %v", timeout)
	}
}

// generateFallbackAnalysis creates intelligent fallback when AI fails
func generateFallbackAnalysis(changes string, wasTruncated bool) string {
	var analysis strings.Builder
	
	if wasTruncated {
		analysis.WriteString("🔍 Large changeset detected - providing summary analysis:\n\n")
	} else {
		analysis.WriteString("🔍 AI unavailable - providing heuristic analysis:\n\n")
	}

	// Parse changes for patterns
	changesLower := strings.ToLower(changes)
	
	// Detect change types
	var changeTypes []string
	if strings.Contains(changesLower, "database") || strings.Contains(changesLower, "migration") ||
	   strings.Contains(changesLower, "sql") || strings.Contains(changesLower, "schema") {
		changeTypes = append(changeTypes, "Database/Schema changes")
	}
	if strings.Contains(changesLower, "api") || strings.Contains(changesLower, "endpoint") ||
	   strings.Contains(changesLower, "route") || strings.Contains(changesLower, "controller") {
		changeTypes = append(changeTypes, "API/Interface changes")
	}
	if strings.Contains(changesLower, "config") || strings.Contains(changesLower, "env") ||
	   strings.Contains(changesLower, "settings") {
		changeTypes = append(changeTypes, "Configuration changes")
	}
	if strings.Contains(changesLower, "security") || strings.Contains(changesLower, "auth") ||
	   strings.Contains(changesLower, "permission") {
		changeTypes = append(changeTypes, "Security-related changes")
	}

	if len(changeTypes) > 0 {
		analysis.WriteString("**Detected Change Types:**\n")
		for _, changeType := range changeTypes {
			analysis.WriteString(fmt.Sprintf("- %s\n", changeType))
		}
		analysis.WriteString("\n")
	}

	// Count modified files
	lines := strings.Split(changes, "\n")
	fileCount := 0
	for _, line := range lines {
		if strings.Contains(line, " |") && (strings.Contains(line, "+") || strings.Contains(line, "-")) {
			fileCount++
		}
	}

	if fileCount > 0 {
		analysis.WriteString(fmt.Sprintf("**Scale:** %d files modified\n", fileCount))
		if fileCount > 10 {
			analysis.WriteString("**Impact:** Likely significant - multiple files affected\n")
		} else if fileCount > 3 {
			analysis.WriteString("**Impact:** Moderate - several files affected\n") 
		} else {
			analysis.WriteString("**Impact:** Focused - few files affected\n")
		}
		analysis.WriteString("\n")
	}

	analysis.WriteString("**Recommendation:** Given the scope of changes, documenting the architectural decisions would help team understanding.")
	
	return analysis.String()
}

// getGitChangesSummary gets a summary of git changes using the same logic as pre-push validation
func getGitChangesSummary() (string, error) {
	// First try to get changes since last push (same as pre-push hook)
	changes, err := getChangesSinceLastPush()
	if err == nil && len(strings.TrimSpace(changes)) > 0 {
		return changes, nil
	}

	// Fallback to working directory changes
	var output []byte

	// Try unstaged changes
	cmd := exec.Command("git", "diff", "--stat")
	output, err = cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		return string(output), nil
	}

	// Try staged changes  
	cmd = exec.Command("git", "diff", "--cached", "--stat")
	output, err = cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		return string(output), nil
	}

	// Try recent commit
	cmd = exec.Command("git", "diff", "HEAD~1", "--stat")
	output, err = cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		return string(output), nil
	}

	// Git status as final fallback
	cmd = exec.Command("git", "status", "--porcelain")
	output, err = cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		return "Modified files detected:\n" + string(output), nil
	}

	return "No git changes detected", nil
}

// getChangesSinceLastPush gets changes since last push (same logic as pre-push hook)
func getChangesSinceLastPush() (string, error) {
	// Get current branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Try to get diff against origin (multiple commits)
	remoteBranch := fmt.Sprintf("origin/%s", branch)
	diffCmd := exec.Command("git", "diff", remoteBranch+"..HEAD", "--stat")
	output, err := diffCmd.Output()
	
	// If no remote, try to get last 3 commits
	if err != nil {
		diffCmd = exec.Command("git", "diff", "HEAD~3..HEAD", "--stat")
		output, err = diffCmd.Output()
		
		// If still no luck, get all staged and unstaged changes
		if err != nil {
			diffCmd = exec.Command("git", "diff", "HEAD", "--stat")
			output, err = diffCmd.Output()
		}
	}

	return string(output), err
}

// Constants for diff analysis limits
const (
	MaxDiffLines    = 1000  // Maximum lines of diff to analyze
	MaxDiffSize     = 50000 // Maximum characters in diff
	MaxFilesChanged = 50    // Maximum files to analyze in detail
)

// getDetailedChanges gets actual code changes (not just stats) with size limits
func getDetailedChanges() (string, bool, error) {
	// Try to get actual diff content (not just stats) with limits
	changes, err := getChangesSinceLastPush()
	if err != nil || strings.TrimSpace(changes) == "" {
		// Fallback to working directory changes
		cmd := exec.Command("git", "diff")
		output, err := cmd.Output()
		if err == nil && len(strings.TrimSpace(string(output))) > 0 {
			changes = string(output)
		}
	} else {
		// Get full diff content for the range
		remoteBranch, err := getCurrentRemoteBranch()
		if err == nil {
			cmd := exec.Command("git", "diff", remoteBranch+"..HEAD")
			output, err := cmd.Output()
			if err == nil {
				changes = string(output)
			}
		}
	}

	if strings.TrimSpace(changes) == "" {
		return "No changes detected", false, nil
	}

	// Apply size limits and filtering
	filteredChanges, wasTruncated := filterAndLimitChanges(changes)
	return filteredChanges, wasTruncated, nil
}

// getCurrentRemoteBranch gets the remote tracking branch
func getCurrentRemoteBranch() (string, error) {
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(string(branchOutput))
	return fmt.Sprintf("origin/%s", branch), nil
}

// filterAndLimitChanges applies intelligent filtering and size limits to git changes
func filterAndLimitChanges(changes string) (string, bool) {
	lines := strings.Split(changes, "\n")
	var filteredLines []string
	var currentFile string
	var addedLines, removedLines int
	var filesChanged int
	var wasTruncated bool

	// Track file-level changes for summary
	fileChanges := make(map[string]struct {
		added   int
		removed int
	})

	for _, line := range lines {
		// Check if we've hit size limits
		if len(strings.Join(filteredLines, "\n")) > MaxDiffSize {
			wasTruncated = true
			break
		}

		// Skip auto-generated or less important files
		if strings.HasPrefix(line, "diff --git") {
			filesChanged++
			if filesChanged > MaxFilesChanged {
				wasTruncated = true
				break
			}

			// Extract file path
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				currentFile = parts[3][2:] // Remove "b/" prefix
			}

			// Skip auto-generated files
			if shouldSkipFile(currentFile) {
				currentFile = "" // Mark to skip
				continue
			}
		}

		// Skip if we're in a file we want to ignore
		if currentFile == "" {
			continue
		}

		// Track line changes
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addedLines++
			if stats, exists := fileChanges[currentFile]; exists {
				stats.added++
				fileChanges[currentFile] = stats
			} else {
				fileChanges[currentFile] = struct {
					added   int
					removed int
				}{added: 1, removed: 0}
			}
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removedLines++
			if stats, exists := fileChanges[currentFile]; exists {
				stats.removed++
				fileChanges[currentFile] = stats
			} else {
				fileChanges[currentFile] = struct {
					added   int
					removed int
				}{added: 0, removed: 1}
			}
		}

		// Add line to filtered output
		filteredLines = append(filteredLines, line)

		// Check line count limit
		if len(filteredLines) > MaxDiffLines {
			wasTruncated = true
			break
		}
	}

	result := strings.Join(filteredLines, "\n")

	// Add summary if truncated
	if wasTruncated {
		summary := generateChangesSummary(fileChanges, addedLines, removedLines, filesChanged)
		result = summary + "\n\n" + result + "\n\n[... content truncated due to size ...]"
	}

	return result, wasTruncated
}

// shouldSkipFile determines if a file should be skipped during analysis
func shouldSkipFile(filepath string) bool {
	skipPatterns := []string{
		// Auto-generated files
		".generated.", "_generated.", "generated/",
		
		// Build artifacts
		"node_modules/", "vendor/", "target/", "build/", "dist/", ".git/",
		
		// Lock files
		"package-lock.json", "yarn.lock", "Cargo.lock", "go.sum", "composer.lock",
		
		// Documentation (unless it's architectural)
		".md", ".txt", ".rst", ".asciidoc",
		
		// Binary files
		".png", ".jpg", ".jpeg", ".gif", ".pdf", ".zip", ".tar", ".gz",
		
		// IDE files
		".vscode/", ".idea/", "*.iml", ".DS_Store",
		
		// Test data
		"test-data/", "fixtures/", "mocks/",
	}

	lowerPath := strings.ToLower(filepath)
	for _, pattern := range skipPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	return false
}

// generateChangesSummary creates a high-level summary of changes
func generateChangesSummary(fileChanges map[string]struct {
	added   int
	removed int
}, totalAdded, totalRemoved, totalFiles int) string {
	var summary strings.Builder
	
	summary.WriteString("=== CHANGES SUMMARY ===\n")
	summary.WriteString(fmt.Sprintf("Files changed: %d, Lines added: %d, Lines removed: %d\n\n", 
		totalFiles, totalAdded, totalRemoved))
	
	summary.WriteString("Key files modified:\n")
	count := 0
	for file, stats := range fileChanges {
		if count >= 10 { // Show top 10 files
			break
		}
		summary.WriteString(fmt.Sprintf("- %s (+%d -%d)\n", file, stats.added, stats.removed))
		count++
	}
	
	if len(fileChanges) > 10 {
		summary.WriteString(fmt.Sprintf("... and %d more files\n", len(fileChanges)-10))
	}
	
	return summary.String()
}

// QuestionnairResponse holds user responses to ADR questions
type QuestionnaireResponse struct {
	ProblemContext    string
	DecisionMade      string
	WhyThisSolution   string
	AlternativesConsidered string
	TradeOffs         string
	FutureImplications string
	AdditionalContext string
}

// conductInteractiveQuestionnaire asks targeted questions based on the ADR context
func conductInteractiveQuestionnaire(adrTitle, changeAnalysis string) (*QuestionnaireResponse, error) {
	responses := &QuestionnaireResponse{}

	fmt.Println("I'll ask you some questions to help generate comprehensive ADR content.")
	fmt.Println("You can skip questions by leaving them blank if not applicable.")
	fmt.Println()

	// Determine question style based on title/changes
	questions := getContextualQuestions(adrTitle, changeAnalysis)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("What problem or need motivated this decision?").
				Description(questions.ProblemPrompt).
				Value(&responses.ProblemContext),

			huh.NewText().
				Title("What solution or approach did you choose?").
				Description(questions.DecisionPrompt).
				Value(&responses.DecisionMade),
		),

		huh.NewGroup(
			huh.NewText().
				Title("Why did you choose this particular approach?").
				Description(questions.RationalePrompt).
				Value(&responses.WhyThisSolution),

			huh.NewText().
				Title("What other options did you consider?").
				Description(questions.AlternativesPrompt).
				Value(&responses.AlternativesConsidered),
		),

		huh.NewGroup(
			huh.NewText().
				Title("What are the main trade-offs or implications?").
				Description(questions.ConsequencesPrompt).
				Value(&responses.TradeOffs),

			huh.NewText().
				Title("Any concerns about future maintenance or scalability?").
				Description(questions.FuturePrompt).
				Value(&responses.FutureImplications),
		),

		huh.NewGroup(
			huh.NewText().
				Title("Any additional context or details?").
				Description("Anything else that would help others understand this decision?").
				Value(&responses.AdditionalContext),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	return responses, nil
}

// ContextualQuestions holds prompts tailored to the specific change type
type ContextualQuestions struct {
	ProblemPrompt     string
	DecisionPrompt    string
	RationalePrompt   string
	AlternativesPrompt string
	ConsequencesPrompt string
	FuturePrompt      string
}

// getContextualQuestions returns targeted questions based on the ADR context
func getContextualQuestions(title, analysis string) ContextualQuestions {
	titleLower := strings.ToLower(title)
	analysisLower := strings.ToLower(analysis)

	// Database-related questions
	if strings.Contains(titleLower, "database") || strings.Contains(titleLower, "db") || 
		strings.Contains(analysisLower, "database") {
		return ContextualQuestions{
			ProblemPrompt:     "What data storage or performance issue needed addressing?",
			DecisionPrompt:    "Which database technology/approach did you choose?",
			RationalePrompt:   "Why this database over alternatives (performance, consistency, cost, etc.)?",
			AlternativesPrompt: "What other databases or storage approaches did you evaluate?",
			ConsequencesPrompt: "Impact on performance, data consistency, operational complexity?",
			FuturePrompt:      "Migration strategy, backup plans, scaling considerations?",
		}
	}

	// API-related questions  
	if strings.Contains(titleLower, "api") || strings.Contains(analysisLower, "api") {
		return ContextualQuestions{
			ProblemPrompt:     "What API or integration requirement drove this decision?",
			DecisionPrompt:    "What API design or integration approach did you implement?",
			RationalePrompt:   "Why this API pattern (REST, GraphQL, gRPC, etc.)?",
			AlternativesPrompt: "What other API approaches did you consider?",
			ConsequencesPrompt: "Impact on client integration, versioning, performance?",
			FuturePrompt:      "Backward compatibility, versioning strategy, rate limiting needs?",
		}
	}

	// Architecture/Design questions
	if strings.Contains(titleLower, "architecture") || strings.Contains(titleLower, "design") ||
		strings.Contains(analysisLower, "architecture") {
		return ContextualQuestions{
			ProblemPrompt:     "What architectural challenge or requirement needed addressing?",
			DecisionPrompt:    "What architectural pattern or structure did you implement?",
			RationalePrompt:   "Why this architectural approach over others?",
			AlternativesPrompt: "What other architectural patterns did you evaluate?",
			ConsequencesPrompt: "Impact on maintainability, testability, performance?",
			FuturePrompt:      "How will this scale? What are the long-term implications?",
		}
	}

	// Security-related questions
	if strings.Contains(titleLower, "security") || strings.Contains(titleLower, "auth") ||
		strings.Contains(analysisLower, "security") {
		return ContextualQuestions{
			ProblemPrompt:     "What security requirement or vulnerability needed addressing?",
			DecisionPrompt:    "What security approach or mechanism did you implement?", 
			RationalePrompt:   "Why this security solution over alternatives?",
			AlternativesPrompt: "What other security approaches did you consider?",
			ConsequencesPrompt: "Impact on user experience, performance, compliance?",
			FuturePrompt:      "Ongoing security maintenance, audit requirements, updates needed?",
		}
	}

	// Generic questions for other changes
	return ContextualQuestions{
		ProblemPrompt:     "What problem, requirement, or opportunity motivated this change?",
		DecisionPrompt:    "What solution, approach, or change did you implement?",
		RationalePrompt:   "What were the key factors that led you to choose this approach?",
		AlternativesPrompt: "What other solutions or approaches did you consider?",
		ConsequencesPrompt: "What are the positive and negative impacts of this decision?",
		FuturePrompt:      "What should future developers know about maintaining or extending this?",
	}
}

// generateADRContent uses AI to create complete ADR content
func generateADRContent(aiManager *ai.Manager, targetADR *adr.ADR, changes, changeAnalysis string, responses *QuestionnaireResponse) (string, error) {
	if !aiManager.IsAvailable() {
		return generateFallbackContent(targetADR, responses), nil
	}

	// Create comprehensive prompt combining all information
	prompt := createComprehensiveADRPrompt(targetADR, changes, changeAnalysis, responses)

	// Get AI-generated content
	content, err := aiManager.AnalyzeChanges(prompt)
	if err != nil {
		fmt.Printf("⚠️  AI generation failed, using fallback: %v\n", err)
		return generateFallbackContent(targetADR, responses), nil
	}

	// Clean up and format the AI response
	return formatGeneratedContent(content, targetADR), nil
}

// createComprehensiveADRPrompt builds a detailed prompt for AI content generation
func createComprehensiveADRPrompt(targetADR *adr.ADR, changes, changeAnalysis string, responses *QuestionnaireResponse) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString("You are Dr Duck, an expert software architect. Your task is to write a complete ADR (Architectural Decision Record) in proper MADR format.\n\n")
	
	promptBuilder.WriteString("IMPORTANT: Do NOT provide analysis or recommendations. Write the actual ADR content as if the decision has already been made.\n\n")

	promptBuilder.WriteString("Information to use:\n")
	promptBuilder.WriteString(fmt.Sprintf("- ADR Title: %s\n", targetADR.Title))
	promptBuilder.WriteString(fmt.Sprintf("- Date: %s\n", targetADR.Date.Format("2006-01-02")))

	if changes != "" {
		promptBuilder.WriteString(fmt.Sprintf("- Code Changes: %s\n", changes))
	}

	promptBuilder.WriteString(fmt.Sprintf("- Problem/Context: %s\n", responses.ProblemContext))
	promptBuilder.WriteString(fmt.Sprintf("- Decision Made: %s\n", responses.DecisionMade))
	promptBuilder.WriteString(fmt.Sprintf("- Rationale: %s\n", responses.WhyThisSolution))
	promptBuilder.WriteString(fmt.Sprintf("- Alternatives: %s\n", responses.AlternativesConsidered))
	promptBuilder.WriteString(fmt.Sprintf("- Trade-offs: %s\n", responses.TradeOffs))
	promptBuilder.WriteString(fmt.Sprintf("- Future Considerations: %s\n", responses.FutureImplications))
	if responses.AdditionalContext != "" {
		promptBuilder.WriteString(fmt.Sprintf("- Additional Context: %s\n", responses.AdditionalContext))
	}

	promptBuilder.WriteString("\nGenerate ONLY the ADR content in this exact format:\n\n")
	promptBuilder.WriteString("# [title]\n\n")
	promptBuilder.WriteString("* **Status**: Draft\n")
	promptBuilder.WriteString("* **Date**: [date]\n\n")
	promptBuilder.WriteString("## Context\n\n")
	promptBuilder.WriteString("[Describe the problem/situation]\n\n")
	promptBuilder.WriteString("## Decision\n\n")
	promptBuilder.WriteString("[Describe what was decided]\n\n")
	promptBuilder.WriteString("## Rationale\n\n")
	promptBuilder.WriteString("[Explain why this decision was made]\n\n")
	promptBuilder.WriteString("## Consequences\n\n")
	promptBuilder.WriteString("### Positive\n\n")
	promptBuilder.WriteString("[What becomes easier or better]\n\n")
	promptBuilder.WriteString("### Negative\n\n")
	promptBuilder.WriteString("[What becomes more difficult]\n\n")
	promptBuilder.WriteString("### Neutral\n\n")
	promptBuilder.WriteString("[Other implications]\n\n")
	promptBuilder.WriteString("## Alternatives Considered\n\n")
	promptBuilder.WriteString("[Other options that were evaluated]\n\n")
	
	promptBuilder.WriteString("Write in clear, professional language. Replace [brackets] with actual content. Do NOT include analysis comments or suggestions - just write the ADR content.")

	return promptBuilder.String()
}

// generateFallbackContent creates ADR content without AI
func generateFallbackContent(targetADR *adr.ADR, responses *QuestionnaireResponse) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# %s\n\n", targetADR.Title))
	content.WriteString(fmt.Sprintf("* **Status**: %s\n", targetADR.Status))
	content.WriteString(fmt.Sprintf("* **Date**: %s\n\n", time.Now().Format("2006-01-02")))

	content.WriteString("## Context\n\n")
	if responses.ProblemContext != "" {
		content.WriteString(responses.ProblemContext)
	} else {
		content.WriteString("<!-- Describe the problem or situation that motivated this decision -->")
	}
	content.WriteString("\n\n")

	content.WriteString("## Decision\n\n")
	if responses.DecisionMade != "" {
		content.WriteString(responses.DecisionMade)
	} else {
		content.WriteString("<!-- Describe the solution or approach that was chosen -->")
	}
	content.WriteString("\n\n")

	content.WriteString("## Rationale\n\n")
	if responses.WhyThisSolution != "" {
		content.WriteString(responses.WhyThisSolution)
	} else {
		content.WriteString("<!-- Explain why this particular solution was selected -->")
	}
	content.WriteString("\n\n")

	content.WriteString("## Consequences\n\n")
	content.WriteString("### Positive\n\n")
	if responses.TradeOffs != "" && (strings.Contains(strings.ToLower(responses.TradeOffs), "benefit") || 
		strings.Contains(strings.ToLower(responses.TradeOffs), "positive")) {
		content.WriteString(responses.TradeOffs)
	} else {
		content.WriteString("<!-- What becomes easier or better with this decision -->")
	}
	content.WriteString("\n\n")

	content.WriteString("### Negative\n\n")
	if responses.TradeOffs != "" {
		content.WriteString(responses.TradeOffs)
	} else {
		content.WriteString("<!-- What becomes more difficult or complex -->")
	}
	content.WriteString("\n\n")

	content.WriteString("### Neutral\n\n")
	if responses.FutureImplications != "" {
		content.WriteString(responses.FutureImplications)
	} else {
		content.WriteString("<!-- Other implications that are neither positive nor negative -->")
	}
	content.WriteString("\n\n")

	content.WriteString("## Alternatives Considered\n\n")
	if responses.AlternativesConsidered != "" {
		content.WriteString(responses.AlternativesConsidered)
	} else {
		content.WriteString("<!-- What other options were evaluated and why they weren't chosen -->")
	}
	content.WriteString("\n\n")

	if responses.AdditionalContext != "" {
		content.WriteString("## Additional Notes\n\n")
		content.WriteString(responses.AdditionalContext)
		content.WriteString("\n\n")
	}

	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("*ADR-%04d completed with DrDuck AI assistance on %s*\n", 
		targetADR.ID, time.Now().Format("2006-01-02")))

	return content.String()
}

// formatGeneratedContent cleans up and formats AI-generated content
func formatGeneratedContent(content string, targetADR *adr.ADR) string {
	// If AI returned complete markdown with title, use as-is
	if strings.HasPrefix(content, "#") {
		return content
	}

	// If AI returned sections without title, add title and metadata
	var formatted strings.Builder
	formatted.WriteString(fmt.Sprintf("# %s\n\n", targetADR.Title))
	formatted.WriteString(fmt.Sprintf("* **Status**: %s\n", targetADR.Status))
	formatted.WriteString(fmt.Sprintf("* **Date**: %s\n\n", targetADR.Date.Format("2006-01-02")))
	formatted.WriteString(content)
	formatted.WriteString(fmt.Sprintf("\n\n---\n*ADR-%04d completed with DrDuck AI assistance on %s*\n", 
		targetADR.ID, targetADR.Date.Format("2006-01-02")))

	return formatted.String()
}

// previewAndConfirm shows the generated content and asks for confirmation
func previewAndConfirm(content string) (bool, string, error) {
	fmt.Println("Generated ADR Content:")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println(content)
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	var confirmed bool
	var editChoice string
	
	// Ask what user wants to do
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to do with this content?").
				Options(
					huh.NewOption("Accept as-is", "accept"),
					huh.NewOption("Make manual edits", "edit"),
					huh.NewOption("Regenerate with different approach", "regenerate"),
					huh.NewOption("Cancel", "cancel"),
				).
				Value(&editChoice),
		),
	)

	if err := form.Run(); err != nil {
		return false, "", err
	}

	switch editChoice {
	case "accept":
		confirmed = true
		return confirmed, content, nil
	
	case "edit":
		fmt.Println("💡 Opening content for editing...")
		editedContent, err := editContentInteractively(content)
		if err != nil {
			return false, "", fmt.Errorf("editing failed: %w", err)
		}
		return true, editedContent, nil

	case "regenerate":
		fmt.Println("🔄 To regenerate, run the command again with different responses")
		return false, "", nil

	case "cancel":
		return false, "", nil

	default:
		return false, "", nil
	}
}

// editContentInteractively allows user to edit the generated content
func editContentInteractively(content string) (string, error) {
	// Create temp file for editing
	tempFile := fmt.Sprintf("/tmp/drduck-edit-%d.md", 
		int64(len(content))) // Simple unique identifier

	// Write content to temp file
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile) // Clean up

	// Open in editor
	editor := getEditor() // Use same function from edit.go
	editorCmd := exec.Command(editor, tempFile)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return "", fmt.Errorf("editor failed: %w", err)
	}

	// Read edited content
	editedContent, err := os.ReadFile(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to read edited content: %w", err)
	}

	return string(editedContent), nil
}

// saveCompletedADR writes the completed content to the ADR file
func saveCompletedADR(targetADR *adr.ADR, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(targetADR.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write content to file
	if err := os.WriteFile(targetADR.FilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write ADR file: %w", err)
	}

	return nil
}

// handleADRStatusUpdate asks user about status and updates accordingly
func handleADRStatusUpdate(adrManager *adr.Manager, targetADR *adr.ADR, isNewADR bool) error {
	var statusChoice string
	
	statusPrompt := "What status should this ADR have?"
	if isNewADR {
		statusPrompt = "The ADR has been created with complete content. What status should it have?"
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(statusPrompt).
				Options(
					huh.NewOption("Accepted - Decision is finalized", "accepted"),
					huh.NewOption("In Progress - Still being refined", "in-progress"),
					huh.NewOption("Draft - Needs review", "draft"),
				).
				Value(&statusChoice),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	var newStatus adr.Status
	switch statusChoice {
	case "accepted":
		newStatus = adr.StatusAccepted
	case "in-progress":
		newStatus = adr.StatusInProgress
	case "draft":
		newStatus = adr.StatusDraft
	default:
		return nil // Keep current status
	}

	if newStatus != targetADR.Status {
		fmt.Printf("📊 Updating ADR status to %s...\n", newStatus)
		if err := adrManager.UpdateADRStatus(targetADR.ID, newStatus); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}
		fmt.Printf("✅ ADR-%04d status updated to %s\n", targetADR.ID, newStatus)
	}

	return nil
}

// handleTitleConfirmation prompts user to confirm or change the ADR title
func handleTitleConfirmation(adrManager *adr.Manager, targetADR *adr.ADR) error {
	var keepTitle bool
	
	currentTitle := targetADR.Title
	prompt := fmt.Sprintf("Keep the current title '%s'?", currentTitle)
	
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(prompt).
				Affirmative("Yes, keep it").
				Negative("No, change it").
				Value(&keepTitle),
		),
	)
	
	if err := form.Run(); err != nil {
		return fmt.Errorf("failed to get title confirmation: %w", err)
	}
	
	if !keepTitle {
		var newTitle string
		newTitleForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter new ADR title:").
					Placeholder("e.g., api-versioning-strategy").
					Value(&newTitle).
					Validate(func(str string) error {
						if strings.TrimSpace(str) == "" {
							return fmt.Errorf("title cannot be empty")
						}
						return nil
					}),
			),
		)
		
		if err := newTitleForm.Run(); err != nil {
			return fmt.Errorf("failed to get new title: %w", err)
		}
		
		newTitle = strings.TrimSpace(newTitle)
		if newTitle != currentTitle {
			fmt.Printf("📝 Renaming ADR from '%s' to '%s'...\n", currentTitle, newTitle)
			
			// Update the title and rename the file
			if err := renameADRFile(adrManager, targetADR, newTitle); err != nil {
				return fmt.Errorf("failed to rename ADR: %w", err)
			}
			
			fmt.Printf("✅ ADR-%04d renamed to '%s'\n", targetADR.ID, newTitle)
		}
	}
	
	return nil
}

// renameADRFile renames an ADR file and updates the title in the content
func renameADRFile(adrManager *adr.Manager, targetADR *adr.ADR, newTitle string) error {
	// Read current file content
	currentContent, err := os.ReadFile(targetADR.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read current ADR file: %w", err)
	}
	
	// Generate new filename
	newFilename := fmt.Sprintf("%04d-%s.md", targetADR.ID, strings.ReplaceAll(strings.ToLower(newTitle), " ", "-"))
	newFilePath := filepath.Join(filepath.Dir(targetADR.FilePath), newFilename)
	
	// Update title in content
	content := string(currentContent)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			lines[i] = fmt.Sprintf("# %s", newTitle)
			break
		}
	}
	updatedContent := strings.Join(lines, "\n")
	
	// Write to new file
	if err := os.WriteFile(newFilePath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write new ADR file: %w", err)
	}
	
	// Remove old file if it's different from the new one
	if targetADR.FilePath != newFilePath {
		if err := os.Remove(targetADR.FilePath); err != nil {
			// Log warning but don't fail - new file was created successfully
			fmt.Printf("⚠️  Warning: failed to remove old file %s: %v\n", targetADR.FilePath, err)
		}
	}
	
	// Update the ADR object
	targetADR.Title = newTitle
	targetADR.FilePath = newFilePath
	
	return nil
}