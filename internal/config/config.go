// Package config provides configuration management for the coverage system
package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mrz1836/go-coverage/internal/envfile"
)

// Static error definitions
var (
	ErrInvalidCoverageThreshold = errors.New("coverage threshold must be between 0 and 100")
	ErrEmptyCoverageInput       = errors.New("coverage input file cannot be empty")
	ErrMissingGitHubToken       = errors.New("GitHub token is required for GitHub integration")
	ErrMissingGitHubOwner       = errors.New("GitHub repository owner is required")
	ErrMissingGitHubRepo        = errors.New("GitHub repository name is required")
	ErrInvalidBadgeStyle        = errors.New("invalid badge style")
	ErrInvalidReportTheme       = errors.New("invalid report theme")
	ErrInvalidRetentionDays     = errors.New("history retention days must be positive")
	ErrInvalidMaxEntries        = errors.New("history max entries must be positive")
	ErrEnvFileNotFound          = errors.New("environment configuration file not found")
)

// isMainBranch checks if a branch name is one of the configured main branches
func isMainBranch(branchName string) bool {
	mainBranches := os.Getenv("MAIN_BRANCHES")
	if mainBranches == "" {
		mainBranches = "master,main"
	}

	branches := strings.Split(mainBranches, ",")
	for _, branch := range branches {
		if strings.TrimSpace(branch) == branchName {
			return true
		}
	}

	return false
}

// Config holds the main configuration for the coverage system
type Config struct {
	// Coverage settings
	Coverage CoverageConfig `json:"coverage"`
	// GitHub integration settings
	GitHub GitHubConfig `json:"github"`
	// Badge generation settings
	Badge BadgeConfig `json:"badge"`
	// Report generation settings
	Report ReportConfig `json:"report"`
	// History tracking settings
	History HistoryConfig `json:"history"`
	// Storage settings
	Storage StorageConfig `json:"storage"`
	// Logging settings
	Log LogConfig `json:"log"`
	// Analytics settings
	Analytics AnalyticsConfig `json:"analytics"`
}

// CoverageConfig holds coverage analysis settings
type CoverageConfig struct {
	// Input coverage file path
	InputFile string `json:"input_file"`
	// Output directory for generated files
	OutputDir string `json:"output_dir"`
	// Minimum coverage threshold
	Threshold float64 `json:"threshold"`
	// Allow threshold override via PR labels
	AllowLabelOverride bool `json:"allow_label_override"`
	// Paths to exclude from coverage
	ExcludePaths []string `json:"exclude_paths"`
	// File patterns to exclude
	ExcludeFiles []string `json:"exclude_files"`
	// Whether to exclude test files
	ExcludeTests bool `json:"exclude_tests"`
	// Whether to exclude generated files
	ExcludeGenerated bool `json:"exclude_generated"`
}

// GitHubConfig holds GitHub integration settings
type GitHubConfig struct {
	// GitHub API token
	Token string `json:"token"`
	// Repository owner
	Owner string `json:"owner"`
	// Repository name
	Repository string `json:"repository"`
	// Pull request number (0 if not in PR context)
	PullRequest int `json:"pull_request"`
	// Commit SHA
	CommitSHA string `json:"commit_sha"`
	// Whether to post PR comments
	PostComments bool `json:"post_comments"`
	// Whether to create commit statuses
	CreateStatuses bool `json:"create_statuses"`
	// API timeout
	Timeout time.Duration `json:"timeout"`
}

// BadgeConfig holds badge generation settings
type BadgeConfig struct {
	// Badge style (flat, flat-square, for-the-badge)
	Style string `json:"style"`
	// Label text
	Label string `json:"label"`
	// Logo URL
	Logo string `json:"logo"`
	// Logo color
	LogoColor string `json:"logo_color"`
	// Output file path
	OutputFile string `json:"output_file"`
	// Whether to generate trend badge
	IncludeTrend bool `json:"include_trend"`
	// Max time for all logo fetch attempts
	LogoTimeout time.Duration `json:"logo_timeout"`
	// Per-request timeout for logo fetching
	LogoHTTPTimeout time.Duration `json:"logo_http_timeout"`
	// Number of retries for logo fetching
	LogoRetries int `json:"logo_retries"`
	// Enable GitHub fallback for logo fetching
	LogoGitHubFallback bool `json:"logo_github_fallback"`
}

// ReportConfig holds HTML report generation settings
type ReportConfig struct {
	// Output file path
	OutputFile string `json:"output_file"`
	// Report title
	Title string `json:"title"`
	// Theme (github-dark, light, etc.)
	Theme string `json:"theme"`
	// Whether to show package breakdown
	ShowPackages bool `json:"show_packages"`
	// Whether to show file breakdown
	ShowFiles bool `json:"show_files"`
	// Whether to show missing lines
	ShowMissing bool `json:"show_missing"`
}

// HistoryConfig holds history tracking settings
type HistoryConfig struct {
	// Whether to enable history tracking
	Enabled bool `json:"enabled"`
	// Storage path for history files
	StoragePath string `json:"storage_path"`
	// Number of days to retain history
	RetentionDays int `json:"retention_days"`
	// Maximum number of entries to keep
	MaxEntries int `json:"max_entries"`
	// Whether to enable automatic cleanup
	AutoCleanup bool `json:"auto_cleanup"`
	// Whether to enable detailed metrics
	MetricsEnabled bool `json:"metrics_enabled"`
}

// StorageConfig holds storage settings
type StorageConfig struct {
	// Base directory for all coverage files
	BaseDir string `json:"base_dir"`
	// Whether to create directories automatically
	AutoCreate bool `json:"auto_create"`
	// File permissions for created files
	FileMode os.FileMode `json:"file_mode"`
	// Directory permissions for created directories
	DirMode os.FileMode `json:"dir_mode"`
}

// LogConfig holds logging configuration settings
type LogConfig struct {
	// Log level (DEBUG, INFO, WARN, ERROR)
	Level string `json:"level"`
	// Log format (text, json)
	Format string `json:"format"`
	// Whether to enable logging
	Enabled bool `json:"enabled"`
}

// AnalyticsConfig holds analytics tracking settings
type AnalyticsConfig struct {
	// Google Analytics tracking ID
	GoogleAnalyticsID string `json:"google_analytics_id"`
	// Whether to include branding in reports
	BrandingEnabled bool `json:"branding_enabled"`
}

// findEnvDir looks for the modular .github/env/ directory by walking up from the
// current working directory. Returns empty string if not found.
// For testing, the GO_COVERAGE_TEST_CONFIG_DIR environment variable overrides detection.
func findEnvDir() string {
	// Test override
	if testDir := os.Getenv("GO_COVERAGE_TEST_CONFIG_DIR"); testDir != "" {
		envDir := filepath.Join(testDir, ".github", "env")
		if hasEnvFiles(envDir) {
			return envDir
		}
		return ""
	}

	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		candidate := filepath.Join(dir, ".github", "env")
		if hasEnvFiles(candidate) {
			return candidate
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

// hasEnvFiles checks if a directory exists and contains at least one .env file.
func hasEnvFiles(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if err != nil || !info.IsDir() {
		return false
	}

	matches, err := filepath.Glob(filepath.Join(dirPath, "*.env"))
	if err != nil {
		return false
	}

	return len(matches) > 0
}

// findBaseEnvFile looks for the legacy .github/.env.base file by walking up from the
// current working directory. Returns the path and nil error if found.
// For testing, the GO_COVERAGE_TEST_CONFIG_DIR environment variable overrides detection.
func findBaseEnvFile() (string, error) {
	var startDir string

	if testDir := os.Getenv("GO_COVERAGE_TEST_CONFIG_DIR"); testDir != "" {
		startDir = testDir
	} else {
		var err error
		startDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	dir := startDir
	for {
		candidate := filepath.Join(dir, ".github", ".env.base")
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", ErrEnvFileNotFound
}

// findCustomEnvFile looks for .env.custom in the same directory as the base env file.
func findCustomEnvFile(basePath string) string {
	customPath := filepath.Join(filepath.Dir(basePath), ".env.custom")
	if _, err := os.Stat(customPath); err == nil {
		return customPath
	}
	return ""
}

// isCI returns true when running in a CI environment.
func isCI() bool {
	return os.Getenv("CI") == "true"
}

// Load loads configuration from environment variables with defaults.
// It first attempts to load modular .github/env/*.env files (preferred),
// then falls back to legacy .github/.env.base + .env.custom.
// If no env files are found, it proceeds silently with os.Getenv() defaults.
func Load() (*Config, error) {
	// Try modular mode first (preferred)
	if envDir := findEnvDir(); envDir != "" {
		if err := envfile.LoadDir(envDir, isCI()); err != nil {
			return nil, fmt.Errorf("failed to load modular configuration from %s: %w", envDir, err)
		}
	} else {
		// Fall back to legacy mode
		basePath, err := findBaseEnvFile()
		if err == nil {
			if loadErr := envfile.Load(basePath); loadErr != nil {
				return nil, fmt.Errorf("failed to load %s: %w", basePath, loadErr)
			}
			if customPath := findCustomEnvFile(basePath); customPath != "" {
				if overloadErr := envfile.Overload(customPath); overloadErr != nil {
					return nil, fmt.Errorf("failed to load %s: %w", customPath, overloadErr)
				}
			}
		}
		// If no env files found at all, continue silently (backward compatible)
	}

	config := &Config{
		Coverage: CoverageConfig{
			InputFile:          getEnvString("GO_COVERAGE_INPUT_FILE", "coverage.txt"),
			OutputDir:          getEnvString("GO_COVERAGE_OUTPUT_DIR", "coverage"),
			Threshold:          getEnvFloat("GO_COVERAGE_THRESHOLD", 80.0),
			AllowLabelOverride: getEnvBool("GO_COVERAGE_ALLOW_LABEL_OVERRIDE", false),
			ExcludePaths:       getEnvStringSlice("GO_COVERAGE_EXCLUDE_PATHS", []string{"vendor/", "test/", "testdata/"}),
			ExcludeFiles:       getEnvStringSlice("GO_COVERAGE_EXCLUDE_FILES", []string{"*_test.go", "*.pb.go"}),
			ExcludeTests:       getEnvBool("GO_COVERAGE_EXCLUDE_TESTS", true),
			ExcludeGenerated:   getEnvBool("GO_COVERAGE_EXCLUDE_GENERATED", true),
		},
		GitHub: GitHubConfig{
			Token:          getEnvString("GITHUB_TOKEN", ""),
			Owner:          getEnvString("GITHUB_REPOSITORY_OWNER", ""),
			Repository:     getRepositoryFromEnv(),
			PullRequest:    getEnvInt("GITHUB_PR_NUMBER", 0),
			CommitSHA:      getEnvString("GITHUB_SHA", ""),
			PostComments:   getEnvBool("GO_COVERAGE_POST_COMMENTS", true),
			CreateStatuses: getEnvBool("GO_COVERAGE_CREATE_STATUSES", true),
			Timeout:        getEnvDuration("GITHUB_TIMEOUT", 30*time.Second),
		},
		Badge: BadgeConfig{
			Style:              getEnvString("GO_COVERAGE_BADGE_STYLE", "flat"),
			Label:              getEnvString("GO_COVERAGE_BADGE_LABEL", "coverage"),
			Logo:               getEnvString("GO_COVERAGE_BADGE_LOGO", ""),
			LogoColor:          getEnvString("GO_COVERAGE_BADGE_LOGO_COLOR", "white"),
			OutputFile:         getEnvString("GO_COVERAGE_BADGE_OUTPUT", "coverage.svg"),
			IncludeTrend:       getEnvBool("GO_COVERAGE_BADGE_TREND", false),
			LogoTimeout:        getEnvDuration("GO_COVERAGE_LOGO_TIMEOUT", 8*time.Second),
			LogoHTTPTimeout:    getEnvDuration("GO_COVERAGE_LOGO_HTTP_TIMEOUT", 3*time.Second),
			LogoRetries:        getEnvInt("GO_COVERAGE_LOGO_RETRIES", 2),
			LogoGitHubFallback: getEnvBool("GO_COVERAGE_LOGO_GITHUB_FALLBACK", false),
		},
		Report: ReportConfig{
			OutputFile:   getEnvString("GO_COVERAGE_REPORT_OUTPUT", "coverage.html"),
			Title:        getEnvString("GO_COVERAGE_REPORT_TITLE", "Coverage Report"),
			Theme:        getEnvString("GO_COVERAGE_REPORT_THEME", "github-dark"),
			ShowPackages: getEnvBool("GO_COVERAGE_REPORT_PACKAGES", true),
			ShowFiles:    getEnvBool("GO_COVERAGE_REPORT_FILES", true),
			ShowMissing:  getEnvBool("GO_COVERAGE_REPORT_MISSING", true),
		},
		History: HistoryConfig{
			Enabled:        getEnvBool("GO_COVERAGE_HISTORY_ENABLED", true),
			StoragePath:    getEnvString("GO_COVERAGE_HISTORY_PATH", "coverage/history"),
			RetentionDays:  getEnvInt("GO_COVERAGE_HISTORY_RETENTION", 90),
			MaxEntries:     getEnvInt("GO_COVERAGE_HISTORY_MAX_ENTRIES", 1000),
			AutoCleanup:    getEnvBool("GO_COVERAGE_HISTORY_CLEANUP", true),
			MetricsEnabled: getEnvBool("GO_COVERAGE_HISTORY_METRICS", true),
		},
		Storage: StorageConfig{
			BaseDir:    getEnvString("GO_COVERAGE_BASE_DIR", "coverage"),
			AutoCreate: getEnvBool("GO_COVERAGE_AUTO_CREATE_DIRS", true),
			FileMode:   os.FileMode(getEnvIntBounded("GO_COVERAGE_FILE_MODE", 0o644, 0, 0o777)),
			DirMode:    os.FileMode(getEnvIntBounded("GO_COVERAGE_DIR_MODE", 0o755, 0, 0o777)),
		},
		Log: LogConfig{
			Level:   getEnvString("GO_COVERAGE_LOG_LEVEL", "INFO"),
			Format:  getEnvString("GO_COVERAGE_LOG_FORMAT", "text"),
			Enabled: getEnvBool("GO_COVERAGE_LOG_ENABLED", true),
		},
		Analytics: AnalyticsConfig{
			GoogleAnalyticsID: getEnvString("GOOGLE_ANALYTICS_ID", ""),
			BrandingEnabled:   getEnvBool("GO_COVERAGE_BRANDING_ENABLED", true),
		},
	}

	return config, nil
}

// Validate validates the configuration and returns an error if invalid
func (c *Config) Validate() error {
	// Validate coverage settings
	if c.Coverage.Threshold < 0 || c.Coverage.Threshold > 100 {
		return fmt.Errorf("%w, got: %.1f", ErrInvalidCoverageThreshold, c.Coverage.Threshold)
	}

	// No additional validation needed for AllowLabelOverride - it's just a boolean

	if c.Coverage.InputFile == "" {
		return ErrEmptyCoverageInput
	}

	// Validate GitHub settings if GitHub integration is enabled
	if c.GitHub.PostComments || c.GitHub.CreateStatuses {
		if c.GitHub.Token == "" {
			return ErrMissingGitHubToken
		}
		if c.GitHub.Owner == "" {
			return ErrMissingGitHubOwner
		}
		if c.GitHub.Repository == "" {
			return ErrMissingGitHubRepo
		}
	}

	// Validate badge settings
	validStyles := []string{"flat", "flat-square", "for-the-badge"}
	if !contains(validStyles, c.Badge.Style) {
		return fmt.Errorf("%w: %s, must be one of: %v", ErrInvalidBadgeStyle, c.Badge.Style, validStyles)
	}

	// Validate report settings
	validThemes := []string{"github-dark", "light", "github-light"}
	if !contains(validThemes, c.Report.Theme) {
		return fmt.Errorf("%w: %s, must be one of: %v", ErrInvalidReportTheme, c.Report.Theme, validThemes)
	}

	// Validate history settings
	if c.History.Enabled {
		if c.History.RetentionDays <= 0 {
			return fmt.Errorf("%w: got %d", ErrInvalidRetentionDays, c.History.RetentionDays)
		}
		if c.History.MaxEntries <= 0 {
			return fmt.Errorf("%w: got %d", ErrInvalidMaxEntries, c.History.MaxEntries)
		}
	}

	return nil
}

// IsGitHubContext returns true if running in a GitHub Actions context
func (c *Config) IsGitHubContext() bool {
	return c.GitHub.Owner != "" && c.GitHub.Repository != "" && c.GitHub.CommitSHA != ""
}

// IsPullRequestContext returns true if running in a pull request context
func (c *Config) IsPullRequestContext() bool {
	return c.IsGitHubContext() && c.GitHub.PullRequest > 0
}

// GetBadgeURL returns the URL for the coverage badge
func (c *Config) GetBadgeURL() string {
	if c.GitHub.Owner == "" || c.GitHub.Repository == "" {
		return ""
	}

	// Use GitHub Pages URL structure
	baseURL := fmt.Sprintf("https://%s.github.io/%s", c.GitHub.Owner, c.GitHub.Repository)

	// If in PR context, return PR-specific badge URL
	if c.IsPullRequestContext() {
		return fmt.Sprintf("%s/badges/pr/%d/coverage.svg", baseURL, c.GitHub.PullRequest)
	}

	// For branch-specific badges, get current branch (default to master)
	branch := c.getCurrentBranch()
	if isMainBranch(branch) {
		// Main branch badge deployed at root
		return fmt.Sprintf("%s/coverage.svg", baseURL)
	}

	// Branch-specific badge (still uses subdirectory structure for branches)
	return fmt.Sprintf("%s/badges/%s/coverage.svg", baseURL, branch)
}

// GetReportURL returns the URL for the coverage report
func (c *Config) GetReportURL() string {
	if c.GitHub.Owner == "" || c.GitHub.Repository == "" {
		return ""
	}

	// Use GitHub Pages URL structure
	baseURL := fmt.Sprintf("https://%s.github.io/%s", c.GitHub.Owner, c.GitHub.Repository)

	// If in PR context, return PR-specific report URL
	if c.IsPullRequestContext() {
		return fmt.Sprintf("%s/reports/pr/%d/coverage.html", baseURL, c.GitHub.PullRequest)
	}

	// For branch-specific reports, get current branch (default to master)
	branch := c.getCurrentBranch()
	if isMainBranch(branch) {
		// Main branch report deployed at root (dashboard at root, detailed report as coverage.html)
		return fmt.Sprintf("%s/", baseURL)
	}

	// Branch-specific report (still uses subdirectory structure for branches)
	return fmt.Sprintf("%s/reports/branch/%s/coverage.html", baseURL, branch)
}

// getCurrentBranch returns the current branch name, with intelligent fallback detection
func (c *Config) getCurrentBranch() string {
	// Try to get branch from environment variables (GitHub Actions context)
	// For pull requests, prefer GITHUB_HEAD_REF (source branch)
	if branch := os.Getenv("GITHUB_HEAD_REF"); branch != "" {
		return branch
	}
	if branch := os.Getenv("GITHUB_REF_NAME"); branch != "" {
		return branch
	}
	if ref := os.Getenv("GITHUB_REF"); ref != "" {
		// Extract branch name from refs/heads/branch-name
		if strings.HasPrefix(ref, "refs/heads/") {
			return strings.TrimPrefix(ref, "refs/heads/")
		}
	}

	// Try to get branch from Git command as fallback
	if branch := c.getBranchFromGit(); branch != "" {
		return branch
	}

	// Final fallback - default to master (this repo's default branch)
	// We avoid using commit SHA as branch name since it creates invalid GitHub URLs
	return "master"
}

// GetCurrentBranch returns the current branch name with intelligent fallback detection (public method)
func (c *Config) GetCurrentBranch() string {
	return c.getCurrentBranch()
}

// GetRepositoryRoot attempts to find the repository root directory
func (c *Config) GetRepositoryRoot() (string, error) {
	// Try to get working directory first
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check if we're already in repo root (has .git directory)
	if _, err := os.Stat(filepath.Join(workingDir, ".git")); err == nil {
		return workingDir, nil
	}

	// Try to use git command to find repo root
	if repoRoot := c.getRepoRootFromGit(); repoRoot != "" {
		return repoRoot, nil
	}

	// Fallback: handle different coverage directory structures
	var repoRoot string
	if strings.Contains(workingDir, "coverage/cmd/go-coverage") ||
		strings.Contains(workingDir, "coverage"+string(filepath.Separator)+"cmd"+string(filepath.Separator)+"go-coverage") {
		// New structure: coverage/cmd/go-coverage - go up 3 levels
		repoRoot = filepath.Join(workingDir, "../../../")
	} else if strings.Contains(workingDir, ".github/coverage/cmd/go-coverage") ||
		strings.Contains(workingDir, ".github"+string(filepath.Separator)+"coverage"+string(filepath.Separator)+"cmd"+string(filepath.Separator)+"go-coverage") {
		// Old structure: .github/coverage/cmd/go-coverage - go up 4 levels
		repoRoot = filepath.Join(workingDir, "../../../../")
	}

	if repoRoot != "" {
		if absPath, err := filepath.Abs(repoRoot); err == nil {
			// Verify this looks like a repo root
			if _, err := os.Stat(filepath.Join(absPath, ".git")); err == nil {
				return absPath, nil
			}
			// Return it anyway as best guess
			return absPath, nil
		}
	}

	// Final fallback - return working directory
	return workingDir, nil
}

// getRepoRootFromGit attempts to get repository root using git commands
func (c *Config) getRepoRootFromGit() string {
	ctx := context.Background()

	// Try git rev-parse --show-toplevel
	if output, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output(); err == nil {
		repoRoot := strings.TrimSpace(string(output))
		if repoRoot != "" {
			return repoRoot
		}
	}

	return ""
}

// ResolveHistoryStoragePath resolves the history storage path to an absolute path
func (c *Config) ResolveHistoryStoragePath() (string, error) {
	historyPath := c.History.StoragePath

	// If already absolute, return as-is
	if filepath.IsAbs(historyPath) {
		return historyPath, nil
	}

	// Get repository root
	repoRoot, err := c.GetRepositoryRoot()
	if err != nil {
		return "", fmt.Errorf("failed to get repository root: %w", err)
	}

	// Join with repo root and make absolute
	resolvedPath := filepath.Join(repoRoot, historyPath)
	absPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return absPath, nil
}

// getBranchFromGit attempts to get the current branch name using Git commands
func (c *Config) getBranchFromGit() string {
	ctx := context.Background()

	// Try git rev-parse --abbrev-ref HEAD first (most reliable)
	if output, err := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		branch := strings.TrimSpace(string(output))
		// Ignore if we're in detached HEAD state
		if branch != "" && branch != "HEAD" {
			return branch
		}
	}

	// Try git branch --show-current as alternative (Git 2.22+)
	if output, err := exec.CommandContext(ctx, "git", "branch", "--show-current").Output(); err == nil {
		branch := strings.TrimSpace(string(output))
		if branch != "" {
			return branch
		}
	}

	// No branch detected
	return ""
}

// Helper functions for environment variable parsing

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvIntBounded parses an integer from environment with min/max bounds
func getEnvIntBounded(key string, defaultValue, minValue, maxValue int) uint32 {
	// For file permissions, we know the valid range is 0-0777 (0-511 decimal)
	// Ensure bounds are reasonable for uint32
	const maxFileMode = 0o777

	// Validate and adjust bounds
	if minValue < 0 {
		minValue = 0
	}
	if maxValue < 0 || maxValue > maxFileMode {
		maxValue = maxFileMode
	}
	if minValue > maxValue {
		minValue = 0
		maxValue = maxFileMode
	}

	// Start with default value
	value := defaultValue

	// Parse environment variable if present
	if envValue := os.Getenv(key); envValue != "" {
		// Parse as uint to ensure non-negative
		if parsed, err := strconv.ParseUint(envValue, 0, 32); err == nil && parsed <= uint64(maxFileMode) {
			value = int(parsed)
		}
	}

	// Apply bounds checking
	if value < minValue {
		value = minValue
	} else if value > maxValue {
		value = maxValue
	}

	// At this point, value is guaranteed to be between 0 and maxFileMode (0o777 = 511)
	// Ensure the value is non-negative before converting to uint32
	if value < 0 {
		value = 0
	}
	// #nosec G115 -- value is bounded to 0-511 range which safely fits in uint32
	return uint32(value)
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
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

func getRepositoryFromEnv() string {
	// GitHub Actions provides GITHUB_REPOSITORY in "owner/repo" format
	if repo := os.Getenv("GITHUB_REPOSITORY"); repo != "" {
		parts := strings.Split(repo, "/")
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return ""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
