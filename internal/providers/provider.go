// Package providers implements a flexible abstraction layer for coverage deployment providers.
// This package supports both internal (GitHub Pages) and external (Codecov) coverage providers
// with automatic detection and configuration.
package providers

import (
	"context"
	"time"

	"github.com/mrz1836/go-coverage/internal/config"
)

// Provider defines the interface for coverage deployment providers
type Provider interface {
	// Name returns the provider name for identification and logging
	Name() string

	// Initialize prepares the provider with the given configuration
	Initialize(ctx context.Context, config *Config) error

	// Process handles the coverage data and prepares it for deployment
	Process(ctx context.Context, coverage *CoverageData) error

	// Upload deploys the processed coverage data to the target destination
	Upload(ctx context.Context) (*UploadResult, error)

	// GenerateReports creates any additional reports or artifacts
	GenerateReports(ctx context.Context) error

	// GetReportURL returns the URL where coverage reports can be accessed
	GetReportURL() string

	// Cleanup performs any necessary cleanup operations
	Cleanup(ctx context.Context) error

	// Validate checks if the provider is properly configured
	Validate() error

	// Capabilities returns the capabilities supported by this provider
	Capabilities() ProviderCapabilities
}

// Config holds provider configuration
type Config struct {
	// Main configuration from config package
	Main *config.Config

	// Provider-specific settings
	Provider ProviderType
	DryRun   bool
	Debug    bool
	Force    bool

	// GitHub context information
	GitHubContext *GitHubContext

	// Provider-specific configurations
	InternalConfig *InternalProviderConfig
	CodecovConfig  *CodecovProviderConfig
}

// ProviderType represents the type of provider
type ProviderType string

const (
	// ProviderTypeAuto automatically detects the appropriate provider
	ProviderTypeAuto ProviderType = "auto"

	// ProviderTypeInternal uses internal GitHub Pages deployment
	ProviderTypeInternal ProviderType = "internal"

	// ProviderTypeCodecov uses external Codecov service
	ProviderTypeCodecov ProviderType = "codecov"
)

// GitHubContext contains GitHub Actions environment information
type GitHubContext struct {
	IsGitHubActions bool
	Repository      string
	Owner           string
	Repo            string
	Branch          string
	CommitSHA       string
	PRNumber        string
	EventName       string
	RunID           string
	Token           string
}

// CoverageData represents parsed coverage information
type CoverageData struct {
	// Overall coverage percentage
	Percentage float64

	// Total and covered lines
	TotalLines   int64
	CoveredLines int64

	// Package-level coverage
	Packages []PackageCoverage

	// File-level coverage
	Files []FileCoverage

	// Metadata
	Timestamp time.Time
	CommitSHA string
	Branch    string
}

// PackageCoverage represents coverage data for a package
type PackageCoverage struct {
	Name         string
	Coverage     float64
	TotalLines   int64
	CoveredLines int64
	Files        []string
}

// FileCoverage represents coverage data for a file
type FileCoverage struct {
	Filename     string
	Coverage     float64
	TotalLines   int64
	CoveredLines int64
	MissedLines  []int
}

// UploadResult contains the result of a coverage upload
type UploadResult struct {
	// Provider that performed the upload
	Provider string

	// Success indicates if the upload was successful
	Success bool

	// URLs where coverage reports can be accessed
	ReportURL      string
	AdditionalURLs []string

	// Upload metadata
	UploadTime time.Time
	CommitSHA  string
	Branch     string

	// Error information (if upload failed)
	Error   error
	Message string

	// Provider-specific metadata
	Metadata map[string]interface{}
}

// ProviderCapabilities defines what a provider supports
type ProviderCapabilities struct {
	// SupportsHistory indicates if the provider supports coverage history
	SupportsHistory bool

	// SupportsPRComments indicates if the provider supports PR comments
	SupportsPRComments bool

	// SupportsBadges indicates if the provider supports badge generation
	SupportsBadges bool

	// SupportsReports indicates if the provider supports HTML reports
	SupportsReports bool

	// SupportsDeployment indicates if the provider handles deployment
	SupportsDeployment bool

	// RequiresToken indicates if the provider requires authentication
	RequiresToken bool
}

// InternalProviderConfig holds configuration for the internal GitHub Pages provider
type InternalProviderConfig struct {
	// Deploy to GitHub Pages
	EnablePages bool

	// Cleanup patterns for deployment
	CleanupPatterns []string

	// Deployment verification timeout
	VerificationTimeout time.Duration

	// Enable HTML navigation generation
	GenerateNavigation bool

	// Enable trend visualization
	EnableTrends bool
}

// CodecovProviderConfig holds configuration for the Codecov provider
type CodecovProviderConfig struct {
	// Codecov upload token
	Token string

	// Codecov API URL (defaults to official API)
	APIURL string

	// Upload flags for categorization
	Flags []string

	// Build identifier
	Build string

	// Upload timeout
	Timeout time.Duration

	// Enable PR comments via Codecov
	EnablePRComments bool
}

// DefaultInternalProviderConfig returns default configuration for internal provider
func DefaultInternalProviderConfig() *InternalProviderConfig {
	return &InternalProviderConfig{
		EnablePages:         true,
		CleanupPatterns:     []string{"*.go", "*.mod", "*.sum", "*.yml", "*.yaml", "*.md", "LICENSE", "README*"},
		VerificationTimeout: 30 * time.Second,
		GenerateNavigation:  true,
		EnableTrends:        true,
	}
}

// DefaultCodecovProviderConfig returns default configuration for Codecov provider
func DefaultCodecovProviderConfig() *CodecovProviderConfig {
	return &CodecovProviderConfig{
		APIURL:           "https://codecov.io",
		Flags:            []string{},
		Timeout:          60 * time.Second,
		EnablePRComments: true,
	}
}
