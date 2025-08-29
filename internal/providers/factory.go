package providers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mrz1836/go-coverage/internal/config"
)

// Static error definitions for provider factory operations
var (
	ErrUnsupportedProvider   = errors.New("unsupported provider type")
	ErrProviderNotConfigured = errors.New("provider not properly configured")
	ErrAutoDetectionFailed   = errors.New("automatic provider detection failed")
	ErrMissingConfiguration  = errors.New("missing required configuration")
	ErrInvalidGitHubContext  = errors.New("invalid GitHub context")
)

// Factory creates and configures coverage providers
type Factory struct {
	logger Logger
}

// Logger interface for factory logging
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// NewFactory creates a new provider factory
func NewFactory(logger Logger) *Factory {
	return &Factory{
		logger: logger,
	}
}

// CreateProvider creates a provider based on the configuration
func (f *Factory) CreateProvider(ctx context.Context, cfg *Config) (Provider, error) {
	if cfg == nil {
		return nil, ErrMissingConfiguration
	}

	// Auto-detect provider if needed
	providerType := cfg.Provider
	if providerType == ProviderTypeAuto {
		providerType = f.detectProvider(cfg)
		f.logger.Info("Auto-detected provider: %s", providerType)
	}

	// Create the appropriate provider
	var provider Provider
	var err error

	switch providerType {
	case ProviderTypeInternal:
		provider, err = f.createInternalProvider(cfg)
	case ProviderTypeCodecov:
		provider, err = f.createCodecovProvider(cfg)
	case ProviderTypeAuto:
		// This case should not be reached since auto-detection happens above,
		// but we include it for exhaustiveness
		return nil, fmt.Errorf("%w: auto-detection should have resolved to a specific provider", ErrAutoDetectionFailed)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedProvider, providerType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create provider %s: %w", providerType, err)
	}

	// Initialize the provider
	if err := provider.Initialize(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize provider %s: %w", providerType, err)
	}

	// Validate the provider configuration
	if err := provider.Validate(); err != nil {
		return nil, fmt.Errorf("provider validation failed for %s: %w", providerType, err)
	}

	f.logger.Info("Successfully created and initialized provider: %s", provider.Name())
	return provider, nil
}

// detectProvider automatically detects the appropriate provider
func (f *Factory) detectProvider(cfg *Config) ProviderType {
	f.logger.Debug("Starting automatic provider detection")

	// Check for explicit provider environment variable
	if providerEnv := os.Getenv("GO_COVERAGE_PROVIDER"); providerEnv != "" {
		switch strings.ToLower(providerEnv) {
		case "internal", "github", "pages":
			f.logger.Debug("Provider detected from GO_COVERAGE_PROVIDER: internal")
			return ProviderTypeInternal
		case "codecov":
			f.logger.Debug("Provider detected from GO_COVERAGE_PROVIDER: codecov")
			return ProviderTypeCodecov
		}
	}

	// Check for Codecov token presence
	codecovToken := os.Getenv("CODECOV_TOKEN")
	if codecovToken != "" {
		f.logger.Debug("Codecov token detected, preferring Codecov provider")
		return ProviderTypeCodecov
	}

	// Check if we're in GitHub Actions environment
	if cfg.GitHubContext != nil && cfg.GitHubContext.IsGitHubActions {
		// Check for GitHub Pages deployment capabilities
		if f.canDeployToGitHubPages(cfg.GitHubContext) {
			f.logger.Debug("GitHub Actions environment with Pages capabilities detected")
			return ProviderTypeInternal
		}
	}

	// Default fallback to internal provider
	f.logger.Debug("No specific provider detected, defaulting to internal provider")
	return ProviderTypeInternal
}

// canDeployToGitHubPages checks if GitHub Pages deployment is possible
func (f *Factory) canDeployToGitHubPages(ctx *GitHubContext) bool {
	// Check for required GitHub context
	if ctx.Repository == "" || ctx.Token == "" {
		return false
	}

	// Pages deployment is generally available for public repositories
	// and private repositories with the right permissions
	return true
}

// createInternalProvider creates an internal GitHub Pages provider
func (f *Factory) createInternalProvider(cfg *Config) (Provider, error) {
	// Ensure we have GitHub context for internal provider
	if cfg.GitHubContext == nil {
		return nil, fmt.Errorf("%w: GitHub context required for internal provider", ErrInvalidGitHubContext)
	}

	// Use default config if not provided
	internalConfig := cfg.InternalConfig
	if internalConfig == nil {
		internalConfig = DefaultInternalProviderConfig()
		f.logger.Debug("Using default internal provider configuration")
	}

	// Create the internal provider
	return NewInternalProvider(internalConfig), nil
}

// createCodecovProvider creates a Codecov provider
func (f *Factory) createCodecovProvider(cfg *Config) (Provider, error) {
	// Use default config if not provided
	codecovConfig := cfg.CodecovConfig
	if codecovConfig == nil {
		codecovConfig = DefaultCodecovProviderConfig()
		f.logger.Debug("Using default Codecov provider configuration")
	}

	// Check for Codecov token
	if codecovConfig.Token == "" {
		codecovConfig.Token = os.Getenv("CODECOV_TOKEN")
	}

	if codecovConfig.Token == "" {
		return nil, fmt.Errorf("%w: CODECOV_TOKEN is required for Codecov provider", ErrProviderNotConfigured)
	}

	// Create the Codecov provider
	return NewCodecovProvider(codecovConfig), nil
}

// GetAvailableProviders returns a list of available provider types
func (f *Factory) GetAvailableProviders() []ProviderType {
	return []ProviderType{
		ProviderTypeInternal,
		ProviderTypeCodecov,
	}
}

// ValidateProviderType checks if a provider type is valid
func (f *Factory) ValidateProviderType(providerType ProviderType) error {
	switch providerType {
	case ProviderTypeAuto, ProviderTypeInternal, ProviderTypeCodecov:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedProvider, providerType)
	}
}

// CreateConfigFromEnvironment creates a provider config from environment variables and main config
func CreateConfigFromEnvironment(mainConfig *config.Config, providerType ProviderType, dryRun, debug, force bool) (*Config, error) {
	// Create GitHub context from environment
	githubCtx := &GitHubContext{
		IsGitHubActions: os.Getenv("GITHUB_ACTIONS") == "true",
		Repository:      os.Getenv("GITHUB_REPOSITORY"),
		Branch:          getBranchFromEnv(),
		CommitSHA:       os.Getenv("GITHUB_SHA"),
		PRNumber:        getPRNumberFromEnv(),
		EventName:       os.Getenv("GITHUB_EVENT_NAME"),
		RunID:           os.Getenv("GITHUB_RUN_ID"),
		Token:           os.Getenv("GITHUB_TOKEN"),
	}

	// Extract owner and repo from repository
	if githubCtx.Repository != "" {
		parts := strings.Split(githubCtx.Repository, "/")
		if len(parts) == 2 {
			githubCtx.Owner = parts[0]
			githubCtx.Repo = parts[1]
		}
	}

	// Create provider-specific configurations
	internalConfig := &InternalProviderConfig{
		EnablePages:         getEnvBool("GO_COVERAGE_ENABLE_PAGES"),
		CleanupPatterns:     getEnvStringSlice("GO_COVERAGE_CLEANUP_PATTERNS", DefaultInternalProviderConfig().CleanupPatterns),
		VerificationTimeout: getEnvDuration("GO_COVERAGE_VERIFICATION_TIMEOUT", 30*time.Second),
		GenerateNavigation:  getEnvBool("GO_COVERAGE_GENERATE_NAVIGATION"),
		EnableTrends:        getEnvBool("GO_COVERAGE_ENABLE_TRENDS"),
	}

	codecovConfig := &CodecovProviderConfig{
		Token:            os.Getenv("CODECOV_TOKEN"),
		APIURL:           getEnvString("CODECOV_URL", "https://codecov.io"),
		Flags:            getEnvStringSlice("CODECOV_FLAGS", []string{}),
		Build:            os.Getenv("CODECOV_BUILD"),
		Timeout:          getEnvDuration("CODECOV_TIMEOUT", 60*time.Second),
		EnablePRComments: getEnvBool("CODECOV_ENABLE_PR_COMMENTS"),
	}

	return &Config{
		Main:           mainConfig,
		Provider:       providerType,
		DryRun:         dryRun,
		Debug:          debug,
		Force:          force,
		GitHubContext:  githubCtx,
		InternalConfig: internalConfig,
		CodecovConfig:  codecovConfig,
	}, nil
}

// Helper functions for environment variable parsing

func getBranchFromEnv() string {
	// For pull requests, prefer GITHUB_HEAD_REF (source branch)
	if branch := os.Getenv("GITHUB_HEAD_REF"); branch != "" {
		return branch
	}
	// For push events, use GITHUB_REF_NAME
	if branch := os.Getenv("GITHUB_REF_NAME"); branch != "" {
		return branch
	}
	// Fallback to extracting from GITHUB_REF
	if ref := os.Getenv("GITHUB_REF"); ref != "" && strings.HasPrefix(ref, "refs/heads/") {
		return strings.TrimPrefix(ref, "refs/heads/")
	}
	return ""
}

func getPRNumberFromEnv() string {
	// Extract PR number from GITHUB_EVENT_NAME and event payload
	if os.Getenv("GITHUB_EVENT_NAME") == "pull_request" {
		// In a real implementation, we'd parse the event payload
		// For now, try to extract from ref
		if ref := os.Getenv("GITHUB_REF"); ref != "" && strings.HasPrefix(ref, "refs/pull/") {
			parts := strings.Split(ref, "/")
			if len(parts) >= 3 {
				return parts[2]
			}
		}
	}
	return ""
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return true // Default to true since all callers use true as default
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
