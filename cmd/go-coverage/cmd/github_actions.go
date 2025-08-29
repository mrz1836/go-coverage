package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mrz1836/go-coverage/internal/analysis"
	"github.com/mrz1836/go-coverage/internal/artifacts"
	"github.com/mrz1836/go-coverage/internal/config"
	"github.com/mrz1836/go-coverage/internal/diagnostics"
	"github.com/mrz1836/go-coverage/internal/github"
	"github.com/mrz1836/go-coverage/internal/health"
	"github.com/mrz1836/go-coverage/internal/history"
	"github.com/mrz1836/go-coverage/internal/parser"
	"github.com/mrz1836/go-coverage/internal/providers"
	"github.com/mrz1836/go-coverage/internal/validation"
)

// Static error definitions
var (
	ErrNotInGitHubActions          = errors.New("not running in GitHub Actions environment (use --force to override)")
	ErrCoverageInputNotFound       = errors.New("coverage input file not found")
	ErrCoverageUploadFailed        = errors.New("coverage upload failed")
	ErrCriticalHealthCheckFailures = errors.New("critical health check failures detected")
	ErrValidationFailed            = errors.New("validation failed")
)

// GitHubActionsConfig holds configuration for the github-actions command
type GitHubActionsConfig struct {
	InputFile      string
	Provider       string
	DryRun         bool
	Debug          bool
	AutoDetect     bool
	Diagnostics    bool
	SkipHealth     bool
	SkipValidation bool
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

			// Get additional flags
			diagnostics, _ := cmd.Flags().GetBool("diagnostics")
			skipHealth, _ := cmd.Flags().GetBool("skip-health")
			skipValidation, _ := cmd.Flags().GetBool("skip-validation")

			// Create configuration
			cfg := &GitHubActionsConfig{
				InputFile:      inputFile,
				Provider:       provider,
				DryRun:         dryRun,
				Debug:          debug,
				AutoDetect:     autoDetect,
				Diagnostics:    diagnostics,
				SkipHealth:     skipHealth,
				SkipValidation: skipValidation,
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
	cmd.Flags().Bool("diagnostics", false, "Enable diagnostic output for troubleshooting")
	cmd.Flags().Bool("skip-health", false, "Skip health checks before execution")
	cmd.Flags().Bool("skip-validation", false, "Skip input validation checks")

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
	if envErr := loadEnvFiles(); envErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to load .env files: %v\n", envErr)
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
	if validateErr := mainCfg.Validate(); validateErr != nil {
		return fmt.Errorf("configuration validation failed: %w", validateErr)
	}

	if cfg.Debug {
		_, _ = fmt.Fprintf(os.Stderr, "Input file: %s\n", mainCfg.Coverage.InputFile)
		_, _ = fmt.Fprintf(os.Stderr, "Provider: %s\n", cfg.Provider)
		_, _ = fmt.Fprintf(os.Stderr, "Output directory: %s\n", mainCfg.Coverage.OutputDir)
		_, _ = fmt.Fprintln(os.Stderr, "::endgroup::")
	}

	// Step 2.5: Health checks and validation (Phase 6 integration)
	ctx := context.Background()
	diag := diagnostics.NewErrorReporter(ctx)

	// Run health checks unless skipped
	if !cfg.SkipHealth {
		if cfg.Debug {
			_, _ = fmt.Fprintln(os.Stderr, "::group::Health Checks")
		}

		_, _ = fmt.Fprintln(os.Stdout, "🏥 Running system health checks...")

		if healthErr := runHealthChecks(ctx, githubCtx, cfg.Debug); healthErr != nil {
			if cfg.Diagnostics {
				diagErr := diagnostics.NewDiagnosticError(diagnostics.ErrorTypeSystem, "health_checks", healthErr.Error())
				diag.Report(diagErr.Build())
			}
			_, _ = fmt.Fprintf(os.Stderr, "⚠️  Health check warnings: %v\n", healthErr)
		}

		if cfg.Debug {
			_, _ = fmt.Fprintln(os.Stderr, "::endgroup::")
		}
	}

	// Run input validation unless skipped
	if !cfg.SkipValidation {
		if cfg.Debug {
			_, _ = fmt.Fprintln(os.Stderr, "::group::Input Validation")
		}

		_, _ = fmt.Fprintln(os.Stdout, "🔍 Validating inputs...")

		if validationErr := runInputValidation(ctx, githubCtx, mainCfg, cfg); validationErr != nil {
			if cfg.Diagnostics {
				diagErr := diagnostics.NewDiagnosticError(diagnostics.ErrorTypeValidation, "input_validation", validationErr.Error())
				diag.Report(diagErr.Build())
			}
			return fmt.Errorf("input validation failed: %w", validationErr)
		}

		if cfg.Debug {
			_, _ = fmt.Fprintln(os.Stderr, "::endgroup::")
		}
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

	// Execute the complete workflow orchestration with fallback mechanisms
	err = executeWorkflowWithFallback(ctx, githubCtx, mainCfg, cfg, diag)

	// Output diagnostics if enabled and there were errors
	if cfg.Diagnostics && (err != nil || diag.HasErrors()) {
		_, _ = fmt.Fprintln(os.Stderr, "\n🔧 Diagnostic Information:")
		if diagnosticOutput := diag.GenerateReport(); diagnosticOutput != "" {
			_, _ = fmt.Fprintln(os.Stderr, diagnosticOutput)
		}
	}

	return err
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

// executeWorkflow orchestrates the complete GitHub Actions coverage workflow using provider abstraction
func executeWorkflow(ctx context.Context, githubCtx *github.GitHubContext, mainCfg *config.Config, cfg *GitHubActionsConfig) error {
	// Step 1: Parse coverage data
	_, _ = fmt.Fprintln(os.Stdout, "📊 Parsing coverage data...")

	// Check if input file exists
	if _, err := os.Stat(mainCfg.Coverage.InputFile); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrCoverageInputNotFound, mainCfg.Coverage.InputFile)
	}

	// Parse coverage data using the existing parser
	coverageParser := parser.New()
	coverageResult, err := coverageParser.ParseFile(ctx, mainCfg.Coverage.InputFile)
	if err != nil {
		return fmt.Errorf("failed to parse coverage data: %w", err)
	}

	// Step 2: Create provider configuration
	_, _ = fmt.Fprintln(os.Stdout, "⚙️  Configuring coverage provider...")

	providerType := convertProviderType(cfg.Provider)
	providerConfig, err := providers.CreateConfigFromEnvironment(mainCfg, providerType, cfg.DryRun, cfg.Debug, false)
	if err != nil {
		return fmt.Errorf("failed to create provider configuration: %w", err)
	}

	// Update GitHub context from the existing context
	providerConfig.GitHubContext = &providers.GitHubContext{
		IsGitHubActions: githubCtx.IsGitHubActions,
		Repository:      githubCtx.Repository,
		Owner:           extractOwner(githubCtx.Repository),
		Repo:            extractRepo(githubCtx.Repository),
		Branch:          githubCtx.Branch,
		CommitSHA:       githubCtx.CommitSHA,
		PRNumber:        githubCtx.PRNumber,
		EventName:       githubCtx.EventName,
		RunID:           githubCtx.RunID,
		Token:           githubCtx.Token,
	}

	// Step 3: Create and initialize provider
	logger := providers.NewDefaultLogger(cfg.Debug, cfg.Debug)
	factory := providers.NewFactory(logger)

	provider, err := factory.CreateProvider(ctx, providerConfig)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	_, _ = fmt.Fprintf(os.Stdout, "✅ Using %s provider\n", provider.Name())

	// Step 4: Convert coverage data to provider format
	providerCoverageData := convertCoverageData(coverageResult, githubCtx)

	// Step 5: Process coverage data through provider
	_, _ = fmt.Fprintln(os.Stdout, "🎨 Processing coverage data...")

	err = provider.Process(ctx, providerCoverageData)
	if err != nil {
		return fmt.Errorf("failed to process coverage data: %w", err)
	}

	// Step 6: Upload/deploy coverage data
	_, _ = fmt.Fprintf(os.Stdout, "🚀 Uploading coverage via %s provider...\n", provider.Name())

	uploadResult, err := provider.Upload(ctx)
	if err != nil {
		return fmt.Errorf("failed to upload coverage: %w", err)
	}

	if !uploadResult.Success {
		return fmt.Errorf("%w: %s", ErrCoverageUploadFailed, uploadResult.Message)
	}

	// Step 7: Generate additional reports
	if err := provider.GenerateReports(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to generate additional reports: %v\n", err)
	}

	// Step 8: Post PR comment (if this is a PR)
	if githubCtx.PRNumber != "" {
		_, _ = fmt.Fprintln(os.Stdout, "💬 Posting PR comment...")

		postPRComment(ctx, githubCtx, mainCfg, cfg)
	}

	// Step 9: Output results
	if cfg.Debug {
		_, _ = fmt.Fprintf(os.Stdout, "📍 Report URL: %s\n", uploadResult.ReportURL)
		if len(uploadResult.AdditionalURLs) > 0 {
			_, _ = fmt.Fprintln(os.Stdout, "🔗 Additional URLs:")
			for _, url := range uploadResult.AdditionalURLs {
				_, _ = fmt.Fprintf(os.Stdout, "   - %s\n", url)
			}
		}
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "🌐 Coverage report: %s\n", uploadResult.ReportURL)
	}

	// Step 10: Cleanup
	if err := provider.Cleanup(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Provider cleanup failed: %v\n", err)
	}

	_, _ = fmt.Fprintln(os.Stdout, "✅ GitHub Actions coverage workflow completed successfully!")
	return nil
}

// convertPRDiff converts github.PRDiff to analysis.PRDiff
func convertPRDiff(githubPRDiff *github.PRDiff) *analysis.PRDiff {
	analysisPRDiff := &analysis.PRDiff{
		Files: make([]analysis.PRFile, len(githubPRDiff.Files)),
	}

	for i, file := range githubPRDiff.Files {
		analysisPRDiff.Files[i] = analysis.PRFile{
			Filename:         file.Filename,
			Status:           file.Status,
			Additions:        file.Additions,
			Deletions:        file.Deletions,
			Changes:          file.Changes,
			Patch:            file.Patch,
			BlobURL:          file.BlobURL,
			RawURL:           file.RawURL,
			PreviousFilename: file.PreviousFilename,
		}
	}

	return analysisPRDiff
}

// convertAnalysisToGitHub converts analysis.CoverageComparison to github.CoverageComparison
func convertAnalysisToGitHub(analysisComparison *analysis.CoverageComparison) *github.CoverageComparison {
	return &github.CoverageComparison{
		BaseCoverage: github.CoverageData{
			Percentage:        analysisComparison.BaseCoverage.Percentage,
			TotalStatements:   analysisComparison.BaseCoverage.TotalStatements,
			CoveredStatements: analysisComparison.BaseCoverage.CoveredStatements,
			CommitSHA:         analysisComparison.BaseCoverage.CommitSHA,
			Branch:            analysisComparison.BaseCoverage.Branch,
			Timestamp:         analysisComparison.BaseCoverage.Timestamp,
		},
		PRCoverage: github.CoverageData{
			Percentage:        analysisComparison.PRCoverage.Percentage,
			TotalStatements:   analysisComparison.PRCoverage.TotalStatements,
			CoveredStatements: analysisComparison.PRCoverage.CoveredStatements,
			CommitSHA:         analysisComparison.PRCoverage.CommitSHA,
			Branch:            analysisComparison.PRCoverage.Branch,
			Timestamp:         analysisComparison.PRCoverage.Timestamp,
		},
		Difference: analysisComparison.Difference,
		TrendAnalysis: github.TrendData{
			Direction:        analysisComparison.TrendAnalysis.Direction,
			Magnitude:        analysisComparison.TrendAnalysis.Magnitude,
			PercentageChange: analysisComparison.TrendAnalysis.PercentageChange,
			Momentum:         analysisComparison.TrendAnalysis.Momentum,
		},
		FileChanges:      convertAnalysisFileChanges(analysisComparison.FileChanges),
		SignificantFiles: analysisComparison.SignificantFiles,
		PRFileAnalysis:   convertAnalysisPRFileAnalysis(analysisComparison.PRFileAnalysis),
	}
}

func convertAnalysisFileChanges(analysisChanges []analysis.FileChange) []github.FileChange {
	githubChanges := make([]github.FileChange, len(analysisChanges))
	for i, change := range analysisChanges {
		githubChanges[i] = github.FileChange{
			Filename:      change.Filename,
			BaseCoverage:  change.BaseCoverage,
			PRCoverage:    change.PRCoverage,
			Difference:    change.Difference,
			LinesAdded:    change.LinesAdded,
			LinesRemoved:  change.LinesRemoved,
			IsSignificant: change.IsSignificant,
		}
	}
	return githubChanges
}

func convertAnalysisPRFileAnalysis(analysisFileAnalysis *analysis.PRFileAnalysis) *github.PRFileAnalysis {
	if analysisFileAnalysis == nil {
		return nil
	}
	return &github.PRFileAnalysis{
		Summary: github.PRFileSummary{
			TotalFiles:          analysisFileAnalysis.Summary.TotalFiles,
			GoFilesCount:        analysisFileAnalysis.Summary.GoFilesCount,
			TestFilesCount:      analysisFileAnalysis.Summary.TestFilesCount,
			ConfigFilesCount:    analysisFileAnalysis.Summary.ConfigFilesCount,
			DocumentationCount:  analysisFileAnalysis.Summary.DocumentationCount,
			GeneratedFilesCount: analysisFileAnalysis.Summary.GeneratedFilesCount,
			OtherFilesCount:     analysisFileAnalysis.Summary.OtherFilesCount,
			HasGoChanges:        analysisFileAnalysis.Summary.HasGoChanges,
			HasTestChanges:      analysisFileAnalysis.Summary.HasTestChanges,
			HasConfigChanges:    analysisFileAnalysis.Summary.HasConfigChanges,
			TotalAdditions:      analysisFileAnalysis.Summary.TotalAdditions,
			TotalDeletions:      analysisFileAnalysis.Summary.TotalDeletions,
			GoAdditions:         analysisFileAnalysis.Summary.GoAdditions,
			GoDeletions:         analysisFileAnalysis.Summary.GoDeletions,
		},
	}
}

// postPRComment posts a comment on the PR with coverage information
func postPRComment(ctx context.Context, githubCtx *github.GitHubContext, mainCfg *config.Config, cfg *GitHubActionsConfig) {
	if cfg.Debug {
		_, _ = fmt.Fprintf(os.Stdout, "📝 Processing PR comment for PR #%s\n", githubCtx.PRNumber)
	}

	// Step 1: Parse current coverage data
	parser := parser.New()
	coverageData, err := parser.ParseFile(ctx, mainCfg.Coverage.InputFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to parse coverage data: %v\n", err)
		return
	}

	// Step 2: Set up GitHub client
	client := github.NewWithConfig(&github.Config{
		Token:      githubCtx.Token,
		UserAgent:  "go-coverage/1.0",
		Timeout:    30 * time.Second,
		BaseURL:    "https://api.github.com",
		RetryCount: 3,
	})

	// Step 3: Set up artifacts manager for coverage diff
	artifactManager, err := artifacts.NewManager()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to create artifact manager: %v\n", err)
		return
	}

	// Step 4: Calculate coverage diff
	differ := analysis.NewCoverageDiffer(artifactManager)

	// Extract base branch from PR (default to main if not available)
	baseBranch := "main"
	// In a full implementation, we'd extract base branch from PR event data
	// For now, always use main as default for all events

	comparison, err := differ.CalculateDiff(ctx, coverageData, baseBranch, githubCtx.Branch, githubCtx.PRNumber)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to calculate coverage diff: %v\n", err)
		// Continue with basic comparison
		comparison = &analysis.CoverageComparison{
			PRCoverage: analysis.CoverageData{
				Percentage:        coverageData.Percentage,
				TotalStatements:   coverageData.TotalLines,
				CoveredStatements: coverageData.CoveredLines,
				CommitSHA:         githubCtx.CommitSHA,
				Branch:            githubCtx.Branch,
				Timestamp:         time.Now(),
			},
			BaseCoverage: analysis.CoverageData{
				Percentage: 0.0,
			},
			Difference: coverageData.Percentage,
			TrendAnalysis: analysis.TrendData{
				Direction: "up",
				Magnitude: "minor",
			},
		}
	}

	// Step 5: Get trend history for visualization
	trendHistory, err := differ.GetTrendHistory(ctx, githubCtx.Branch, 30)
	if err != nil {
		if cfg.Debug {
			_, _ = fmt.Fprintf(os.Stderr, "Debug: Failed to get trend history: %v\n", err)
		}
		trendHistory = []history.CoverageRecord{}
	}

	// Step 6: Enhance with PR diff information
	prNumber, parseErr := strconv.Atoi(githubCtx.PRNumber)
	if parseErr == nil {
		// Extract repository owner and name
		repoParts := strings.Split(githubCtx.Repository, "/")
		if len(repoParts) == 2 {
			owner, repo := repoParts[0], repoParts[1]
			prDiff, diffErr := client.GetPRDiff(ctx, owner, repo, prNumber)
			if diffErr == nil {
				// Convert github.PRDiff to analysis.PRDiff
				analysisPRDiff := convertPRDiff(prDiff)
				differ.EnhanceWithPRDiff(comparison, analysisPRDiff)
			}
		}
	}

	// Step 7: Generate comment body
	templateConfig := github.DefaultCommentTemplateConfig()
	templateConfig.Repository = githubCtx.Repository
	templateConfig.CoverageTarget = mainCfg.Coverage.Threshold

	// Set deployment URL if using internal provider
	deploymentURL := ""
	if cfg.Provider == "auto" || cfg.Provider == "internal" {
		repoParts := strings.Split(githubCtx.Repository, "/")
		if len(repoParts) == 2 {
			owner := repoParts[0]
			deploymentURL = fmt.Sprintf("https://%s.github.io/%s/coverage/", owner, repoParts[1])
		}
	}

	templateGenerator := github.NewCommentTemplateGenerator(templateConfig)
	// Convert analysis types to github types for the template generator
	githubComparison := convertAnalysisToGitHub(comparison)
	commentBody := templateGenerator.GenerateComment(githubComparison, trendHistory, deploymentURL)

	// Step 8: Post or update PR comment
	commentConfig := &github.PRCommentConfig{
		MinUpdateIntervalMinutes: 5,
		MaxCommentsPerPR:         1,
		CommentSignature:         "go-coverage-v1",
		IncludeTrend:             true,
		IncludeCoverageDetails:   true,
		IncludeFileAnalysis:      false,
		ShowCoverageHistory:      true,
		BadgeStyle:               "flat",
		EnableStatusChecks:       true,
		FailBelowThreshold:       true,
		BlockMergeOnFailure:      false,
	}
	commentManager := github.NewPRCommentManager(client, commentConfig)

	repoParts := strings.Split(githubCtx.Repository, "/")
	if len(repoParts) != 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Invalid repository format: %s\n", githubCtx.Repository)
		return
	}

	owner, repo := repoParts[0], repoParts[1]
	response, err := commentManager.CreateOrUpdatePRComment(ctx, owner, repo, prNumber, commentBody, githubComparison)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: Failed to post PR comment: %v\n", err)
		return
	}

	// Step 9: Output result
	if cfg.Debug {
		_, _ = fmt.Fprintf(os.Stdout, "✅ PR comment %s (ID: %d)\n", response.Action, response.CommentID)
		_, _ = fmt.Fprintf(os.Stdout, "   Reason: %s\n", response.Reason)
		if response.StatusCheckURL != "" {
			_, _ = fmt.Fprintf(os.Stdout, "   Status check: %s\n", response.StatusCheckURL)
		}
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "💬 PR comment %s successfully\n", response.Action)
	}
}

// convertProviderType converts the string provider type to the providers.ProviderType
func convertProviderType(providerStr string) providers.ProviderType {
	switch providerStr {
	case "auto":
		return providers.ProviderTypeAuto
	case "internal":
		return providers.ProviderTypeInternal
	case "codecov":
		return providers.ProviderTypeCodecov
	default:
		return providers.ProviderTypeAuto
	}
}

// extractOwner extracts the owner from a repository string "owner/repo"
func extractOwner(repository string) string {
	parts := strings.Split(repository, "/")
	if len(parts) >= 2 {
		return parts[0]
	}
	return ""
}

// extractRepo extracts the repo name from a repository string "owner/repo"
func extractRepo(repository string) string {
	parts := strings.Split(repository, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// convertCoverageData converts parser result to provider coverage data format
func convertCoverageData(parserResult *parser.CoverageData, githubCtx *github.GitHubContext) *providers.CoverageData {
	if parserResult == nil {
		return nil
	}

	// Convert packages
	packages := make([]providers.PackageCoverage, 0, len(parserResult.Packages))
	var files []providers.FileCoverage

	for _, pkg := range parserResult.Packages {
		// Get list of file names for this package
		var fileNames []string
		for fileName := range pkg.Files {
			fileNames = append(fileNames, fileName)
		}

		packages = append(packages, providers.PackageCoverage{
			Name:         pkg.Name,
			Coverage:     pkg.Percentage,
			TotalLines:   int64(pkg.TotalLines),
			CoveredLines: int64(pkg.CoveredLines),
			Files:        fileNames,
		})

		// Convert files from this package
		for _, file := range pkg.Files {
			// Calculate missed lines from statements
			var missedLines []int
			for _, stmt := range file.Statements {
				if stmt.Count == 0 {
					// Add all lines in the uncovered statement range
					for line := stmt.StartLine; line <= stmt.EndLine; line++ {
						missedLines = append(missedLines, line)
					}
				}
			}

			files = append(files, providers.FileCoverage{
				Filename:     file.Path,
				Coverage:     file.Percentage,
				TotalLines:   int64(file.TotalLines),
				CoveredLines: int64(file.CoveredLines),
				MissedLines:  missedLines,
			})
		}
	}

	return &providers.CoverageData{
		Percentage:   parserResult.Percentage,
		TotalLines:   int64(parserResult.TotalLines),
		CoveredLines: int64(parserResult.CoveredLines),
		Packages:     packages,
		Files:        files,
		Timestamp:    time.Now(),
		CommitSHA:    githubCtx.CommitSHA,
		Branch:       githubCtx.Branch,
	}
}

// runHealthChecks performs system health checks before workflow execution
func runHealthChecks(ctx context.Context, githubCtx *github.GitHubContext, debug bool) error {
	// Create health manager
	manager := health.NewManager()

	// Add specific health checkers
	manager.AddChecker(health.NewGitHubAPIChecker(githubCtx.Token, githubCtx.Repository))
	manager.AddChecker(health.NewDiskSpaceChecker(".", 100*1024*1024, 5.0)) // 100MB min, 5% min
	manager.AddChecker(health.NewNetworkConnectivityChecker([]string{"github.com", "api.github.com"}))

	// Set timeout
	manager.SetTimeout(30 * time.Second)

	// Run health checks
	report := manager.CheckAll(ctx)

	// Log results
	if debug {
		for _, result := range report.Checks {
			status := "✅"
			if result.Status != health.StatusHealthy {
				status = "⚠️"
			}
			_, _ = fmt.Fprintf(os.Stderr, "%s %s: %s\n", status, result.Name, result.Message)
		}
	}

	// Return error only if there are critical failures
	if report.OverallStatus == health.StatusUnhealthy {
		return ErrCriticalHealthCheckFailures
	}

	return nil
}

// runInputValidation validates inputs before workflow execution
func runInputValidation(ctx context.Context, _ *github.GitHubContext, mainCfg *config.Config, cfg *GitHubActionsConfig) error {
	// Create validators
	coverageValidator := validation.NewCoverageFileValidator(mainCfg.Coverage.InputFile)
	githubValidator := validation.NewGitHubEnvironmentValidator()

	// Convert config to map for validator
	configMap := map[string]interface{}{
		"coverage_input":     mainCfg.Coverage.InputFile,
		"coverage_threshold": mainCfg.Coverage.Threshold,
		"output_dir":         mainCfg.Coverage.OutputDir,
	}
	configValidator := validation.NewConfigValidator(configMap)

	// Run all validations
	result := validation.ValidateAll(ctx, coverageValidator, githubValidator, configValidator)

	if cfg.Debug {
		for _, warning := range result.Warnings {
			_, _ = fmt.Fprintf(os.Stderr, "⚠️  %s: %s\n", warning.Field, warning.Message)
		}
		for _, err := range result.Errors {
			_, _ = fmt.Fprintf(os.Stderr, "❌ %s: %s\n", err.Field, err.Message)
		}
	}

	// Return error if validation failed
	if !result.Valid {
		return fmt.Errorf("%w: %d errors found", ErrValidationFailed, len(result.Errors))
	}

	if cfg.Debug {
		_, _ = fmt.Fprintf(os.Stderr, "✅ All input validations passed\n")
	}

	return nil
}

// executeWorkflowWithFallback wraps executeWorkflow with fallback mechanisms
func executeWorkflowWithFallback(ctx context.Context, githubCtx *github.GitHubContext, mainCfg *config.Config, cfg *GitHubActionsConfig, diag *diagnostics.ErrorReporter) error {
	// For now, we'll use a simple wrapper that adds basic error recovery
	// The fallback mechanisms are built into the individual components
	// (like retry logic in network operations, partial upload in artifacts, etc.)

	// Execute the main workflow
	err := executeWorkflow(ctx, githubCtx, mainCfg, cfg)

	// Record execution results in diagnostics
	if cfg.Diagnostics && err != nil {
		diagErr := diagnostics.NewDiagnosticError(diagnostics.ErrorTypeSystem, "workflow_execution", err.Error())
		diag.Report(diagErr.Build())
		_, _ = fmt.Fprintf(os.Stderr, "🔧 Workflow execution failed, diagnostic information available\n")
	}

	return err
}
