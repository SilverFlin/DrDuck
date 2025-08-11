package ai

import (
	"fmt"

	"github.com/SilverFlin/DrDuck/internal/config"
	"github.com/SilverFlin/DrDuck/pkg/claude"
	"github.com/SilverFlin/DrDuck/pkg/cursor"
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

// Provider defines the interface for AI integrations
type Provider interface {
	IsAvailable() bool
	GetChangedFiles() ([]string, error)
	SuggestADRContent(adrName string) (map[string]string, error)
	ExtractContext() (string, error)
	AnalyzeChanges(prompt string) (string, error)
	AnalyzeChangesWithTokens(prompt string) (AnalyzeResult, error)
}

// Manager handles AI provider integration
type Manager struct {
	config   *config.Config
	provider Provider
}

// NewManager creates a new AI provider manager
func NewManager(cfg *config.Config) *Manager {
	var provider Provider

	switch cfg.AIProvider {
	case "claude-code":
		provider = &ClaudeProvider{integration: claude.NewIntegration()}
	case "cursor":
		provider = &CursorProvider{integration: cursor.NewIntegration()}
	default:
		// Default to Claude
		provider = &ClaudeProvider{integration: claude.NewIntegration()}
	}

	return &Manager{
		config:   cfg,
		provider: provider,
	}
}

// IsAvailable checks if the configured AI provider is available
func (m *Manager) IsAvailable() bool {
	return m.provider.IsAvailable()
}

// GetProviderName returns the name of the configured AI provider
func (m *Manager) GetProviderName() string {
	return m.config.AIProvider
}

// GetChangedFiles returns files modified in the current AI session
func (m *Manager) GetChangedFiles() ([]string, error) {
	return m.provider.GetChangedFiles()
}

// SuggestADRContent generates content suggestions for an ADR
func (m *Manager) SuggestADRContent(adrName string) (map[string]string, error) {
	return m.provider.SuggestADRContent(adrName)
}

// ExtractContext extracts relevant context from the AI session
func (m *Manager) ExtractContext() (string, error) {
	return m.provider.ExtractContext()
}

// AnalyzeChanges sends a change analysis prompt to the AI provider
func (m *Manager) AnalyzeChanges(prompt string) (string, error) {
	return m.provider.AnalyzeChanges(prompt)
}

// AnalyzeChangesWithTokens sends a change analysis prompt to the AI provider and returns token usage
func (m *Manager) AnalyzeChangesWithTokens(prompt string) (AnalyzeResult, error) {
	return m.provider.AnalyzeChangesWithTokens(prompt)
}

// ClaudeProvider implements Provider for Claude Code CLI
type ClaudeProvider struct {
	integration *claude.Integration
}

func (p *ClaudeProvider) IsAvailable() bool {
	return p.integration.IsAvailable()
}

func (p *ClaudeProvider) GetChangedFiles() ([]string, error) {
	return p.integration.GetChangedFiles()
}

func (p *ClaudeProvider) SuggestADRContent(adrName string) (map[string]string, error) {
	return p.integration.SuggestADRContent(adrName)
}

func (p *ClaudeProvider) ExtractContext() (string, error) {
	session, err := p.integration.GetCurrentSession()
	if err != nil {
		return "", fmt.Errorf("failed to get Claude session: %w", err)
	}

	return p.integration.ExtractContext(session)
}

func (p *ClaudeProvider) AnalyzeChanges(prompt string) (string, error) {
	return p.integration.AnalyzeChanges(prompt)
}

func (p *ClaudeProvider) AnalyzeChangesWithTokens(prompt string) (AnalyzeResult, error) {
	response, tokenUsage, err := p.integration.AnalyzeChangesWithTokens(prompt)
	if err != nil {
		return AnalyzeResult{}, err
	}
	
	// Convert claude.TokenUsage to ai.TokenUsage
	var aiTokenUsage TokenUsage
	if tokenUsage != nil {
		aiTokenUsage = TokenUsage{
			InputTokens:  tokenUsage.InputTokens,
			OutputTokens: tokenUsage.OutputTokens,
			TotalTokens:  tokenUsage.TotalTokens,
		}
	}
	
	return AnalyzeResult{
		Response:   response,
		TokenUsage: aiTokenUsage,
	}, nil
}

// CursorProvider implements Provider for Cursor
type CursorProvider struct {
	integration *cursor.Integration
}

func (p *CursorProvider) IsAvailable() bool {
	return p.integration.IsAvailable()
}

func (p *CursorProvider) GetChangedFiles() ([]string, error) {
	return p.integration.GetChangedFiles()
}

func (p *CursorProvider) SuggestADRContent(adrName string) (map[string]string, error) {
	return p.integration.SuggestADRContent(adrName)
}

func (p *CursorProvider) ExtractContext() (string, error) {
	session, err := p.integration.GetCurrentSession()
	if err != nil {
		return "", fmt.Errorf("failed to get Cursor session: %w", err)
	}

	return p.integration.ExtractContext(session)
}

func (p *CursorProvider) AnalyzeChanges(prompt string) (string, error) {
	return p.integration.AnalyzeChanges(prompt)
}

func (p *CursorProvider) AnalyzeChangesWithTokens(prompt string) (AnalyzeResult, error) {
	response, tokenUsage, err := p.integration.AnalyzeChangesWithTokens(prompt)
	if err != nil {
		return AnalyzeResult{}, err
	}
	
	// Convert cursor.TokenUsage to ai.TokenUsage
	var aiTokenUsage TokenUsage
	if tokenUsage != nil {
		aiTokenUsage = TokenUsage{
			InputTokens:  tokenUsage.InputTokens,
			OutputTokens: tokenUsage.OutputTokens,
			TotalTokens:  tokenUsage.TotalTokens,
		}
	}
	
	return AnalyzeResult{
		Response:   response,
		TokenUsage: aiTokenUsage,
	}, nil
}