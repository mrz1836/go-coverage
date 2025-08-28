package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mrz1836/go-coverage/internal/config"
	"github.com/mrz1836/go-coverage/internal/github"
)

// Static error definitions
var (
	ErrNotInGitHubActions = errors.New("not running in GitHub Actions environment (use --force to override)")
)

// GitHubActionsConfig holds configuration for the github-actions command
type GitHubActionsConfig struct {
	InputFile  string
	Provider   string
	DryRun     bool
	Debug      bool
	AutoDetect bool
}

// newGitHubActionsCmd creates the github-actions command
func (c *Commands) newGitHubActionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "github-actions",
		Short: "Automated GitHub Actions integration",
		Long: `Automated GitHub Actions integration that consolidates complex workflows into a single intelligent command.

This command automatically detects the GitHub Actions environment, loads configuration from
environment variables and .env files, and orchestrates the complete coverage pipeline including:
- Coverage parsing and analysis
- Badge and report generation
- History tracking and management
- PR comment generation
- GitHub Pages deployment

Supports both internal (GitHub Pages) and external (Codecov) providers with automatic detection.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get flags
			inputFile, _ := cmd.Flags().GetString("input")
			provider, _ := cmd.Flags().GetString("provider")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			debug, _ := cmd.Flags().GetBool("debug")
			autoDetect, _ := cmd.Flags().GetBool("auto-detect")
			force, _ := cmd.Flags().GetBool("force")

			// Create configuration
			cfg := &GitHubActionsConfig{
				InputFile:  inputFile,
				Provider:   provider,
				DryRun:     dryRun,
				Debug:      debug,
				AutoDetect: autoDetect,
			}

			// Execute the github-actions workflow
			return runGitHubActions(cfg, force)
		},
	}

	// Add command flags
	cmd.Flags().StringP("input", "i", "", "Coverage input file (defaults to coverage.txt)")
	cmd.Flags().StringP("provider", "p", "auto", "Coverage provider (auto|internal|codecov)")
	cmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	cmd.Flags().Bool("debug", false, "Enable verbose debug output")
	cmd.Flags().Bool("auto-detect", true, "Automatically detect GitHub Actions environment")
	cmd.Flags().Bool("force", false, "Force execution even when not in GitHub Actions")

	return cmd
}

// runGitHubActions executes the main GitHub Actions workflow
func runGitHubActions(cfg *GitHubActionsConfig, force bool) error {
	// Step 1: Detect GitHub environment
	if cfg.Debug {
		_, _ = fmt.Fprintln(os.Stderr, "::group::Environment Detection")
	}

	githubCtx, err := github.DetectEnvironment()
	if err != nil {
		return fmt.Errorf("failed to detect GitHub environment: %w", err)
	}

	if !githubCtx.IsGitHubActions && !force {
		return ErrNotInGitHubActions
	}

	if cfg.Debug {
		_, _ = fmt.Fprintf(os.Stderr, "GitHub Actions: %v\n", githubCtx.IsGitHubActions)
		_, _ = fmt.Fprintf(os.Stderr, "Repository: %s\n", githubCtx.Repository)
		_, _ = fmt.Fprintf(os.Stderr, "Branch: %s\n", githubCtx.Branch)
		_, _ = fmt.Fprintf(os.Stderr, "Event: %s\n", githubCtx.EventName)
		if githubCtx.PRNumber != "" {
			_, _ = fmt.Fprintf(os.Stderr, "PR Number: %s\n", githubCtx.PRNumber)
		}
	}

	// Step 2: Load configuration
	if cfg.Debug {
		_, _ = fmt.Fprintln(os.Stderr, "::group::Configuration Loading")
	}

	// Load environment files if they exist
	if err := loadEnvFiles(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to load .env files: %v\n", err)
	}

	// Load main configuration
	mainCfg := config.Load()

	// Apply command-line overrides
	if cfg.InputFile != "" {
		mainCfg.Coverage.InputFile = cfg.InputFile
	}

	// Set defaults if not specified
	if mainCfg.Coverage.InputFile == "" {
		mainCfg.Coverage.InputFile = "coverage.txt"
	}

	// Validate configuration
	if err := mainCfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	if cfg.Debug {
		_, _ = fmt.Fprintf(os.Stderr, "Input file: %s\n", mainCfg.Coverage.InputFile)
		_, _ = fmt.Fprintf(os.Stderr, "Provider: %s\n", cfg.Provider)
		_, _ = fmt.Fprintf(os.Stderr, "Output directory: %s\n", mainCfg.Coverage.OutputDir)
		_, _ = fmt.Fprintln(os.Stderr, "::endgroup::")
	}

	// Step 3: Execute workflow
	if cfg.DryRun {
		_, _ = fmt.Fprintln(os.Stdout, "::group::Dry Run - Execution Plan")
		_, _ = fmt.Fprintln(os.Stdout, "✓ Parse coverage data")
		_, _ = fmt.Fprintln(os.Stdout, "✓ Generate badge")
		_, _ = fmt.Fprintln(os.Stdout, "✓ Generate HTML report")
		_, _ = fmt.Fprintln(os.Stdout, "✓ Update coverage history")
		if githubCtx.PRNumber != "" {
			_, _ = fmt.Fprintln(os.Stdout, "✓ Post PR comment")
		}
		if cfg.Provider == "auto" || cfg.Provider == "internal" {
			_, _ = fmt.Fprintln(os.Stdout, "✓ Deploy to GitHub Pages")
		}
		_, _ = fmt.Fprintln(os.Stdout, "::endgroup::")
		_, _ = fmt.Fprintln(os.Stdout, "✅ Dry run completed successfully")
		return nil
	}

	_, _ = fmt.Fprintln(os.Stdout, "🚀 Starting GitHub Actions coverage workflow...")

	// TODO: Implement actual workflow orchestration
	// This will be expanded to call the existing commands in proper sequence
	_, _ = fmt.Fprintln(os.Stdout, "⚠️  Full implementation coming in subsequent phases")

	return nil
}

// loadEnvFiles loads environment variables from .github/.env.* files
func loadEnvFiles() error {
	envFiles := []string{".github/.env.base", ".github/.env.custom"}

	for _, file := range envFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue // Skip non-existent files
		}

		// #nosec G304 - file path is controlled and limited to .github/.env.* files
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		parseEnvContent(string(content))
	}

	return nil
}

// parseEnvContent parses environment file content and sets environment variables
func parseEnvContent(content string) {
	lines := splitLines(content)

	for _, line := range lines {
		line = trimSpace(line)
		if line == "" || line[0] == '#' {
			continue // Skip empty lines and comments
		}

		parts := splitOnFirst(line, '=')
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := trimSpace(parts[0])
		value := trimSpace(parts[1])

		if key != "" {
			_ = os.Setenv(key, value)
		}
	}
}

// Helper functions for string manipulation
func splitLines(s string) []string {
	if s == "" {
		return nil
	}

	var lines []string
	start := 0

	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}

	if start < len(s) {
		lines = append(lines, s[start:])
	}

	return lines
}

func splitOnFirst(s string, sep byte) []string {
	idx := -1
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			idx = i
			break
		}
	}

	if idx == -1 {
		return []string{s}
	}

	return []string{s[:idx], s[idx+1:]}
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && isSpace(s[start]) {
		start++
	}

	for end > start && isSpace(s[end-1]) {
		end--
	}

	return s[start:end]
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r'
}
