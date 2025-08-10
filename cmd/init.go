package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize DrDuck in the current project",
	Long:  `Initialize DrDuck in the current project with interactive setup for AI provider, documentation storage, git hooks, and ADR templates.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if already initialized
	initialized, err := config.IsInitialized()
	if err != nil {
		return fmt.Errorf("failed to check initialization status: %w", err)
	}

	if initialized {
		fmt.Println("✨ DrDuck is already initialized in this project!")
		
		// Ask if user wants to reconfigure
		var reconfigure bool
		err := huh.NewConfirm().
			Title("Do you want to reconfigure the project?").
			Value(&reconfigure).
			Run()
		if err != nil {
			return err
		}

		if !reconfigure {
			fmt.Println("👋 Initialization cancelled")
			return nil
		}
	}

	fmt.Println("🦆 Welcome to DrDuck! Let's set up your project for automated documentation.")
	fmt.Println()

	// Load existing config or create default
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Interactive setup
	if err := setupInteractive(cfg); err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// Save configuration
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Create directory structure
	if err := createProjectStructure(cfg); err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	// Setup git hooks if requested
	if err := setupGitHooks(cfg); err != nil {
		return fmt.Errorf("failed to setup git hooks: %w", err)
	}

	fmt.Println()
	fmt.Println("🎉 DrDuck has been successfully initialized!")
	fmt.Printf("📁 Configuration saved to: %s\n", config.ConfigDir+"/"+config.ConfigFile)
	if cfg.DocStorage == "same-repo" {
		fmt.Printf("📝 ADRs will be stored in: %s\n", cfg.DocPath)
	} else {
		fmt.Printf("📝 ADRs will be stored in separate repository: %s\n", cfg.SeparateRepoURL)
	}
	fmt.Println()
	fmt.Println("Ready to create your first ADR with: drduck new -n \"feature-name\"")

	return nil
}

func setupInteractive(cfg *config.Config) error {
	var aiProvider string
	var docStorage string
	var adrTemplate string
	var preCommitHook bool
	var prePushHook bool
	var customDocPath string
	var separateRepoURL string

	form := huh.NewForm(
		// AI Provider selection
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which AI coding assistant do you use?").
				Options(
					huh.NewOption("Claude Code CLI", "claude-code").Selected(cfg.AIProvider == "claude-code"),
					huh.NewOption("Cursor", "cursor").Selected(cfg.AIProvider == "cursor"),
				).
				Value(&aiProvider),
		),

		// Documentation storage
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Where should ADRs be stored?").
				Options(
					huh.NewOption("Same repository (docs/adrs/)", "same-repo").Selected(cfg.DocStorage == "same-repo"),
					huh.NewOption("Separate documentation repository", "separate-repo").Selected(cfg.DocStorage == "separate-repo"),
				).
				Value(&docStorage),
		),

		// ADR Template
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose ADR template format:").
				Options(
					huh.NewOption("Nygard template (Michael Nygard's original ADR format)", "nygard").Selected(cfg.ADRTemplate == "nygard"),
					huh.NewOption("MADR (Markdown Any Decision Records)", "madr").Selected(cfg.ADRTemplate == "madr"),
					huh.NewOption("Simple template", "simple").Selected(cfg.ADRTemplate == "simple"),
					huh.NewOption("Custom template", "custom").Selected(cfg.ADRTemplate == "custom"),
				).
				Value(&adrTemplate),
		),

		// Git Hooks
		huh.NewGroup(
			huh.NewConfirm().
				Title("Install pre-commit hook?").
				Description("Validates ADR completeness for staged changes").
				Value(&preCommitHook).
				Affirmative("Yes").
				Negative("No"),

			huh.NewConfirm().
				Title("Install pre-push hook?").
				Description("Ensures significant changes have associated ADRs").
				Value(&prePushHook).
				Affirmative("Yes").
				Negative("No"),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	// Additional prompts based on selections
	if docStorage == "same-repo" {
		err := huh.NewInput().
			Title("ADR storage path:").
			Placeholder(config.DefaultDocPath).
			Value(&customDocPath).
			Run()
		if err != nil {
			return err
		}
		if customDocPath == "" {
			customDocPath = config.DefaultDocPath
		}
	} else {
		err := huh.NewInput().
			Title("Separate repository URL:").
			Placeholder("github.com/yourorg/docs").
			Value(&separateRepoURL).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("repository URL is required")
				}
				return nil
			}).
			Run()
		if err != nil {
			return err
		}
	}

	// Update configuration
	cfg.AIProvider = aiProvider
	cfg.DocStorage = docStorage
	cfg.ADRTemplate = adrTemplate
	cfg.Hooks.PreCommit = preCommitHook
	cfg.Hooks.PrePush = prePushHook

	if docStorage == "same-repo" {
		cfg.DocPath = customDocPath
		cfg.SeparateRepoURL = ""
	} else {
		cfg.SeparateRepoURL = separateRepoURL
		cfg.DocPath = ""
	}

	return nil
}

func createProjectStructure(cfg *config.Config) error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}

	// Create .drduck directory structure
	dirs := []string{
		configDir,
		filepath.Join(configDir, "templates"),
		filepath.Join(configDir, "hooks"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create ADR directory if same-repo storage
	if cfg.DocStorage == "same-repo" {
		adrDir := cfg.DocPath
		if err := os.MkdirAll(adrDir, 0755); err != nil {
			return fmt.Errorf("failed to create ADR directory %s: %w", adrDir, err)
		}

		// Create initial README for ADR directory
		readmePath := filepath.Join(adrDir, "README.md")
		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			readme := `# Architectural Decision Records (ADRs)

This directory contains the Architectural Decision Records for this project.

## About ADRs

An Architectural Decision Record (ADR) is a document that captures an important architectural decision made along with its context and consequences.

## Index

This section will be automatically updated by DrDuck as new ADRs are created.

<!-- ADR_INDEX_START -->
<!-- ADR_INDEX_END -->

---
*This documentation is managed by [DrDuck](https://github.com/SilverFlin/DrDuck)*
`
			if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
				return fmt.Errorf("failed to create ADR README: %w", err)
			}
		}
	}

	return nil
}

func setupGitHooks(cfg *config.Config) error {
	if !cfg.Hooks.PreCommit && !cfg.Hooks.PrePush {
		return nil // No hooks to install
	}

	// Check if we're in a git repository
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		fmt.Println("⚠️  Warning: Not in a git repository. Git hooks will be created but won't be active until you initialize git.")
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}

	hooksDir := filepath.Join(configDir, "hooks")

	if cfg.Hooks.PreCommit {
		preCommitHook := `#!/bin/sh
# DrDuck pre-commit hook
# Encourages ADR completion but never blocks commits

echo "🦆 DrDuck: Checking for draft ADRs..."

# Check if DrDuck is available
if ! command -v drduck >/dev/null 2>&1; then
    echo "⚠️  DrDuck command not found, but continuing with commit..."
    echo "   Install DrDuck globally: npm install -g drduck"
    exit 0
fi

# Run pre-commit validation (warns but never blocks)
if ! drduck validate --pre-commit 2>/dev/null; then
    echo "⚠️  DrDuck validation had issues, but commit proceeding..."
fi

# Always exit successfully - pre-commit never blocks
exit 0
`
		hookPath := filepath.Join(hooksDir, "pre-commit")
		if err := os.WriteFile(hookPath, []byte(preCommitHook), 0755); err != nil {
			return fmt.Errorf("failed to create pre-commit hook: %w", err)
		}

		// Link to git hooks directory if it exists
		gitHooksDir := ".git/hooks"
		if _, err := os.Stat(gitHooksDir); err == nil {
			gitHookPath := filepath.Join(gitHooksDir, "pre-commit")
			if err := os.Remove(gitHookPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove existing pre-commit hook: %w", err)
			}
			if err := os.Symlink(hookPath, gitHookPath); err != nil {
				// Fallback to copying if symlink fails
				hookContent, err := os.ReadFile(hookPath)
				if err != nil {
					return err
				}
				if err := os.WriteFile(gitHookPath, hookContent, 0755); err != nil {
					return fmt.Errorf("failed to install pre-commit hook: %w", err)
				}
			}
		}
	}

	if cfg.Hooks.PrePush {
		prePushHook := `#!/bin/sh
# DrDuck pre-push hook
# Blocks push if draft ADRs exist or if changes need ADRs

echo "🦆 DrDuck: Validating ADR requirements before push..."

# Check if DrDuck is available
if ! command -v drduck >/dev/null 2>&1; then
    echo "⚠️  DrDuck command not found. Install DrDuck or use --no-verify to skip."
    echo "   Install: npm install -g drduck"
    exit 1
fi

# Run pre-push validation (may block)
if ! drduck validate --pre-push; then
    echo ""
    echo "💡 To bypass this check: git push --no-verify"
    echo "🔗 For help: drduck --help"
    exit 1
fi

# Validation passed
echo "✅ All ADR requirements satisfied!"
exit 0
`
		hookPath := filepath.Join(hooksDir, "pre-push")
		if err := os.WriteFile(hookPath, []byte(prePushHook), 0755); err != nil {
			return fmt.Errorf("failed to create pre-push hook: %w", err)
		}

		// Link to git hooks directory if it exists
		gitHooksDir := ".git/hooks"
		if _, err := os.Stat(gitHooksDir); err == nil {
			gitHookPath := filepath.Join(gitHooksDir, "pre-push")
			if err := os.Remove(gitHookPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove existing pre-push hook: %w", err)
			}
			if err := os.Symlink(hookPath, gitHookPath); err != nil {
				// Fallback to copying if symlink fails
				hookContent, err := os.ReadFile(hookPath)
				if err != nil {
					return err
				}
				if err := os.WriteFile(gitHookPath, hookContent, 0755); err != nil {
					return fmt.Errorf("failed to install pre-push hook: %w", err)
				}
			}
		}
	}

	return nil
}