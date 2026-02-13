package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mrz1836/go-coverage/internal/analysis"
	"github.com/mrz1836/go-coverage/internal/badge"
	"github.com/mrz1836/go-coverage/internal/config"
	"github.com/mrz1836/go-coverage/internal/github"
	"github.com/mrz1836/go-coverage/internal/history"
	"github.com/mrz1836/go-coverage/internal/parser"
	"github.com/mrz1836/go-coverage/internal/templates"
)

var (
	// ErrGitHubTokenRequired indicates GitHub token was not provided
	ErrGitHubTokenRequired = errors.New("GitHub token is required")
	// ErrGitHubOwnerRequired indicates repository owner was not provided
	ErrGitHubOwnerRequired = errors.New("GitHub repository owner is required")
	// ErrGitHubRepoRequired indicates repository name was not provided
	ErrGitHubRepoRequired = errors.New("GitHub repository name is required")
	// ErrPRNumberRequired indicates PR number was not provided
	ErrPRNumberRequired = errors.New("pull request number is required")
)

// newCommentCmd creates the comment command
func (c *Commands) newCommentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Create PR coverage comment with analysis and templates",
		Long: `Create or update pull request comments with coverage information.

Features:
- Intelligent PR comment management with anti-spam features
- Coverage comparison and analysis between base and PR branches
- Dynamic template rendering with multiple template options
- PR-specific badge generation with unique naming
- GitHub status check integration for blocking PR merges
- Smart update logic and lifecycle management`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get flags
			prNumber, _ := cmd.Flags().GetInt("pr")
			inputFile, _ := cmd.Flags().GetString("coverage")
			baseCoverageFile, _ := cmd.Flags().GetString("base-coverage")
			badgeURL, _ := cmd.Flags().GetString("badge-url")
			reportURL, _ := cmd.Flags().GetString("report-url")
			createStatus, _ := cmd.Flags().GetBool("status")
			blockOnFailure, _ := cmd.Flags().GetBool("block-merge")
			generateBadges, _ := cmd.Flags().GetBool("generate-badges")
			enableAnalysis, _ := cmd.Flags().GetBool("enable-analysis")
			antiSpam, _ := cmd.Flags().GetBool("anti-spam")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Validate GitHub configuration
			if cfg.GitHub.Token == "" {
				return ErrGitHubTokenRequired
			}
			if cfg.GitHub.Owner == "" {
				return ErrGitHubOwnerRequired
			}
			if cfg.GitHub.Repository == "" {
				return ErrGitHubRepoRequired
			}

			// Use PR number from config if not provided
			if prNumber == 0 {
				prNumber = cfg.GitHub.PullRequest
			}
			if prNumber == 0 {
				return ErrPRNumberRequired
			}

			// Set defaults
			if inputFile == "" {
				inputFile = cfg.Coverage.InputFile
			}
			if badgeURL == "" {
				badgeURL = cfg.GetBadgeURL()
			}
			if reportURL == "" {
				reportURL = cfg.GetReportURL()
			}
			// URLs will be passed to template data below

			// Parse current coverage data
			p := parser.New()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			coverage, err := p.ParseFile(ctx, inputFile)
			if err != nil {
				return fmt.Errorf("failed to parse coverage file: %w", err)
			}

			// Parse base coverage data for comparison (if provided)
			var baseCoverage *parser.CoverageData
			if baseCoverageFile != "" {
				baseCoverage, err = p.ParseFile(ctx, baseCoverageFile)
				if err != nil {
					cmd.Printf("Warning: failed to parse base coverage file: %v\n", err)
					baseCoverage = nil
				}
			}

			// Get trend information if history is enabled
			trend := "stable"
			if cfg.History.Enabled {
				historyConfig := &history.Config{
					StoragePath:    cfg.History.StoragePath,
					RetentionDays:  cfg.History.RetentionDays,
					MaxEntries:     cfg.History.MaxEntries,
					AutoCleanup:    cfg.History.AutoCleanup,
					MetricsEnabled: cfg.History.MetricsEnabled,
				}
				tracker := history.NewWithConfig(historyConfig)

				// Get latest entry to compare
				branch := cfg.GitHub.CommitSHA
				if branch == "" {
					branch = "master"
				}

				if latest, latestErr := tracker.GetLatestEntry(ctx, branch); latestErr == nil {
					if coverage.Percentage > latest.Coverage.Percentage {
						trend = "up"
					} else if coverage.Percentage < latest.Coverage.Percentage {
						trend = "down"
					}
				}
			}

			// Create GitHub client
			githubConfig := &github.Config{
				Token:      cfg.GitHub.Token,
				BaseURL:    "https://api.github.com",
				Timeout:    cfg.GitHub.Timeout,
				RetryCount: 3,
				UserAgent:  "go-coverage/2.0",
			}
			client := github.NewWithConfig(githubConfig)

			// Analyze PR files to understand the impact
			var prFileAnalysis *github.PRFileAnalysis
			if enableAnalysis {
				prDiff, diffErr := client.GetPRDiff(ctx, cfg.GitHub.Owner, cfg.GitHub.Repository, prNumber)
				if diffErr != nil {
					cmd.Printf("Warning: failed to get PR diff: %v\n", diffErr)
				} else {
					prFileAnalysis = github.AnalyzePRFiles(prDiff)
					cmd.Printf("📋 PR Analysis: %s\n", prFileAnalysis.Summary.GetSummaryText())
				}
			}

			// Initialize PR comment system
			prCommentConfig := &github.PRCommentConfig{
				MinUpdateIntervalMinutes: 5,
				MaxCommentsPerPR:         1,
				CommentSignature:         "go-coverage-v1",
				IncludeTrend:             true,
				IncludeCoverageDetails:   true,
				IncludeFileAnalysis:      enableAnalysis,
				ShowCoverageHistory:      true,
				EnableStatusChecks:       createStatus,
				FailBelowThreshold:       true,
				BlockMergeOnFailure:      blockOnFailure,
			}

			// Adjust settings for anti-spam mode
			if antiSpam {
				prCommentConfig.MinUpdateIntervalMinutes = 15
				prCommentConfig.MaxCommentsPerPR = 1
			}

			prCommentManager := github.NewPRCommentManager(client, prCommentConfig)

			// Perform coverage comparison and analysis if base coverage is available
			var comparison *github.CoverageComparison
			if baseCoverage != nil && enableAnalysis {
				comparisonEngine := analysis.NewComparisonEngine(nil)

				// Convert parser data to comparison snapshots
				baseSnapshot := convertToSnapshot(baseCoverage, "master", "")
				prSnapshot := convertToSnapshot(coverage, "current", cfg.GitHub.CommitSHA)

				comparisonResult, compErr := comparisonEngine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
				if compErr != nil {
					cmd.Printf("Warning: failed to perform coverage comparison: %v\n", compErr)
				} else {
					// Convert comparison result to PR comment format
					comparison = &github.CoverageComparison{
						BaseCoverage: github.CoverageData{
							Percentage:        baseCoverage.Percentage,
							TotalStatements:   baseCoverage.TotalLines,   // Actually statement count, not line count
							CoveredStatements: baseCoverage.CoveredLines, // Actually covered statement count, not line count
							CommitSHA:         "",
							Branch:            "master",
							Timestamp:         time.Now(),
						},
						PRCoverage: github.CoverageData{
							Percentage:        coverage.Percentage,
							TotalStatements:   coverage.TotalLines,   // Actually statement count, not line count
							CoveredStatements: coverage.CoveredLines, // Actually covered statement count, not line count
							CommitSHA:         cfg.GitHub.CommitSHA,
							Branch:            "current",
							Timestamp:         time.Now(),
						},
						Difference:       coverage.Percentage - baseCoverage.Percentage,
						TrendAnalysis:    convertTrendData(comparisonResult.TrendAnalysis),
						FileChanges:      convertFileChanges(comparisonResult.FileChanges),
						SignificantFiles: extractSignificantFiles(comparisonResult.FileChanges),
						PRFileAnalysis:   prFileAnalysis,
					}
				}
			}

			// Fall back to simple comparison if no base coverage or analysis disabled
			if comparison == nil {
				comparison = &github.CoverageComparison{
					// Only set base coverage if we actually have base data
					BaseCoverage: github.CoverageData{
						// Leave base coverage empty when no baseline is available
						// This prevents misleading "0 → current" comparisons
						Percentage:        0,
						TotalStatements:   0,
						CoveredStatements: 0,
						CommitSHA:         "",
						Branch:            "",
						Timestamp:         time.Time{}, // Empty timestamp indicates no baseline
					},
					PRCoverage: github.CoverageData{
						Percentage:        coverage.Percentage,
						TotalStatements:   coverage.TotalLines,   // Actually contains statement count, not line count
						CoveredStatements: coverage.CoveredLines, // Actually contains covered statement count, not line count
						CommitSHA:         cfg.GitHub.CommitSHA,
						Branch:            "current",
						Timestamp:         time.Now(),
					},
					Difference: 0, // No meaningful difference without baseline
					TrendAnalysis: github.TrendData{
						Direction:        trend,
						Magnitude:        "minor",
						PercentageChange: 0,
						Momentum:         "steady",
					},
					PRFileAnalysis: prFileAnalysis, // Include PR file analysis even without baseline
				}
			}

			// Initialize template engine for comment generation
			templateEngine := templates.NewPRTemplateEngine(&templates.TemplateConfig{
				IncludeEmojis:          true,
				IncludeCharts:          true,
				MaxFileChanges:         20,
				MaxRecommendations:     5,
				UseMarkdownTables:      true,
				UseCollapsibleSections: true,
				IncludeProgressBars:    true,
				BrandingEnabled:        true,
			})

			// Build template data
			templateData := buildTemplateData(cfg, prNumber, comparison, coverage, badgeURL, reportURL)

			// Render comment using template engine
			commentBody, renderErr := templateEngine.RenderComment(ctx, "", templateData)
			if renderErr != nil {
				return fmt.Errorf("failed to render comment template: %w", renderErr)
			}

			if dryRun {
				// Display preview for dry run
				cmd.Printf("PR Comment Preview (Dry Run)\n")
				cmd.Printf("=====================================\n")
				cmd.Printf("Template: comprehensive\n")
				cmd.Printf("PR: %d\n", prNumber)
				cmd.Printf("Repository: %s/%s\n", cfg.GitHub.Owner, cfg.GitHub.Repository)
				cmd.Printf("Coverage: %.2f%%\n", coverage.Percentage)
				if comparison.BaseCoverage.Percentage > 0 {
					cmd.Printf("Base Coverage: %.2f%%\n", comparison.BaseCoverage.Percentage)
					cmd.Printf("Difference: %+.2f%%\n", comparison.Difference)
				}
				cmd.Printf("Features enabled:\n")
				cmd.Printf("  - Analysis: %v\n", enableAnalysis)
				cmd.Printf("  - Status Checks: %v\n", createStatus)
				cmd.Printf("  - Badge Generation: %v\n", generateBadges)
				cmd.Printf("  - Merge Blocking: %v\n", blockOnFailure)
				cmd.Printf("  - Anti-spam: %v\n", antiSpam)
				cmd.Printf("=====================================\n")
				cmd.Println(commentBody)
				cmd.Printf("=====================================\n")

				return nil
			}

			// Create or update PR comment
			ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			result, err := prCommentManager.CreateOrUpdatePRComment(ctx, cfg.GitHub.Owner, cfg.GitHub.Repository, prNumber, commentBody, comparison)
			if err != nil {
				return fmt.Errorf("failed to create PR comment: %w", err)
			}

			cmd.Printf("Coverage comment %s successfully!\n", result.Action)
			cmd.Printf("Comment ID: %d\n", result.CommentID)
			cmd.Printf("Coverage: %.2f%%\n", comparison.PRCoverage.Percentage)
			if comparison.BaseCoverage.Percentage > 0 {
				cmd.Printf("Change: %+.2f%% vs base\n", comparison.Difference)
			}
			cmd.Printf("Action taken: %s (%s)\n", result.Action, result.Reason)

			// Create status checks if requested
			if createStatus && cfg.GitHub.CommitSHA != "" {
				statusManager := github.NewStatusCheckManager(client, &github.StatusCheckConfig{
					ContextPrefix:          "go-coverage",
					MainContext:            "coverage/total",
					AdditionalContexts:     []string{"coverage/trend", "coverage/quality"},
					EnableBlocking:         true,
					BlockOnFailure:         true,
					BlockOnError:           false,
					RequireAllPassing:      false,
					CoverageThreshold:      cfg.Coverage.Threshold,
					QualityThreshold:       "C",
					AllowThresholdOverride: cfg.Coverage.AllowLabelOverride,
					AllowLabelOverride:     cfg.Coverage.AllowLabelOverride,
					EnableQualityGates:     true,
					IncludeTargetURLs:      true,
					UpdateStrategy:         github.UpdateAlways,
					StatusTimeout:          30 * time.Second,
					RetrySettings: github.RetrySettings{
						MaxRetries:    3,
						RetryDelay:    1 * time.Second,
						BackoffFactor: 2.0,
					},
				})

				statusRequest := &github.StatusCheckRequest{
					Owner:      cfg.GitHub.Owner,
					Repository: cfg.GitHub.Repository,
					CommitSHA:  cfg.GitHub.CommitSHA,
					PRNumber:   prNumber,
					Branch:     "current",
					BaseBranch: "master",
					Coverage: github.CoverageStatusData{
						Percentage:        coverage.Percentage,
						TotalStatements:   coverage.TotalLines,
						CoveredStatements: coverage.CoveredLines,
						Change:            comparison.Difference,
						Trend:             comparison.TrendAnalysis.Direction,
					},
					Comparison: github.ComparisonStatusData{
						BasePercentage:    comparison.BaseCoverage.Percentage,
						CurrentPercentage: comparison.PRCoverage.Percentage,
						Difference:        comparison.Difference,
						IsSignificant:     comparison.Difference > 1.0 || comparison.Difference < -1.0,
						Direction:         comparison.TrendAnalysis.Direction,
					},
					Quality: github.QualityStatusData{
						Grade:     calculateQualityGrade(coverage.Percentage),
						Score:     coverage.Percentage,
						RiskLevel: calculateRiskLevel(coverage.Percentage),
					},
				}

				statusResult, err := statusManager.CreateStatusChecks(ctx, statusRequest)
				if err != nil {
					cmd.Printf("Warning: failed to create status checks: %v\n", err)
				} else {
					cmd.Printf("Created %d status checks\n", statusResult.TotalChecks)
					cmd.Printf("Passed: %d, Failed: %d, Errors: %d\n",
						statusResult.PassedChecks, statusResult.FailedChecks, statusResult.ErrorChecks)
					if statusResult.BlockingPR {
						cmd.Printf("⚠️ PR merge is blocked due to failed required checks\n")
					}
					if len(statusResult.RequiredFailed) > 0 {
						cmd.Printf("Failed required checks: %v\n", statusResult.RequiredFailed)
					}
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().IntP("pr", "p", 0, "Pull request number")
	cmd.Flags().StringP("coverage", "c", "", "Path to coverage profile file")
	cmd.Flags().String("base-coverage", "", "Path to base branch coverage file for comparison")
	cmd.Flags().String("badge-url", "", "Custom badge URL (optional)")
	cmd.Flags().String("report-url", "", "Custom report URL (optional)")
	cmd.Flags().Bool("status", true, "Create GitHub commit status")
	cmd.Flags().Bool("block-merge", false, "Block PR merge on coverage failure")
	cmd.Flags().Bool("generate-badges", false, "Generate PR-specific badges")
	cmd.Flags().Bool("enable-analysis", true, "Enable code quality analysis")
	cmd.Flags().Bool("anti-spam", true, "Enable anti-spam features")
	cmd.Flags().Bool("dry-run", false, "Show what would be posted without actually posting")

	return cmd
}

// Helper functions for converting data structures

func convertToSnapshot(coverage *parser.CoverageData, branch, commitSHA string) *analysis.CoverageSnapshot {
	return &analysis.CoverageSnapshot{
		Branch:    branch,
		CommitSHA: commitSHA,
		Timestamp: time.Now(),
		OverallCoverage: analysis.CoverageMetrics{
			Percentage: coverage.Percentage,
			// parser.CoverageData fields are confusingly named:
			// TotalLines actually contains total STATEMENTS
			// CoveredLines actually contains covered STATEMENTS
			TotalStatements:   coverage.TotalLines,   // Actually statement count from parser
			CoveredStatements: coverage.CoveredLines, // Actually covered statement count from parser
			TotalLines:        coverage.TotalLines,   // Using statement count as approximation for lines
			CoveredLines:      coverage.CoveredLines, // Using covered statement count as approximation for covered lines
		},
		FileCoverage:    make(map[string]analysis.FileMetrics),
		PackageCoverage: make(map[string]analysis.PackageMetrics),
		TestMetadata: analysis.TestMetadata{
			TestDuration: 0,
			TestCount:    0,
		},
	}
}

func convertTrendData(trend analysis.TrendAnalysis) github.TrendData {
	return github.TrendData{
		Direction:        trend.Direction,
		Magnitude:        "minor", // Simplified
		PercentageChange: 0,       // Would need calculation
		Momentum:         trend.Momentum,
	}
}

func convertFileChanges(changes []analysis.FileChangeAnalysis) []github.FileChange {
	fileChanges := make([]github.FileChange, 0, len(changes))
	for _, change := range changes {
		fileChanges = append(fileChanges, github.FileChange{
			Filename:      change.Filename,
			BaseCoverage:  change.BasePercentage,
			PRCoverage:    change.PRPercentage,
			Difference:    change.PercentageChange,
			LinesAdded:    change.LinesAdded,
			LinesRemoved:  change.LinesRemoved,
			IsSignificant: change.IsSignificant,
		})
	}
	return fileChanges
}

func extractSignificantFiles(changes []analysis.FileChangeAnalysis) []string {
	var significantFiles []string
	for _, change := range changes {
		if change.IsSignificant {
			significantFiles = append(significantFiles, change.Filename)
		}
	}
	return significantFiles
}

func buildTemplateData(cfg *config.Config, prNumber int, comparison *github.CoverageComparison, _ *parser.CoverageData, badgeURL, reportURL string) *templates.TemplateData {
	return &templates.TemplateData{
		Repository: templates.RepositoryInfo{
			Owner:         cfg.GitHub.Owner,
			Name:          cfg.GitHub.Repository,
			DefaultBranch: "master",
			URL:           fmt.Sprintf("https://github.com/%s/%s", cfg.GitHub.Owner, cfg.GitHub.Repository),
		},
		PullRequest: templates.PullRequestInfo{
			Number:     prNumber,
			Title:      "",
			Branch:     "current",
			BaseBranch: "master",
			Author:     "",
			CommitSHA:  cfg.GitHub.CommitSHA,
			URL:        fmt.Sprintf("https://github.com/%s/%s/pull/%d", cfg.GitHub.Owner, cfg.GitHub.Repository, prNumber),
		},
		Timestamp: time.Now(),
		Coverage: templates.CoverageData{
			Overall: templates.CoverageMetrics{
				Percentage:        comparison.PRCoverage.Percentage,
				TotalStatements:   comparison.PRCoverage.TotalStatements,
				CoveredStatements: comparison.PRCoverage.CoveredStatements,
				Grade:             calculateQualityGrade(comparison.PRCoverage.Percentage),
				Status:            calculateCoverageStatus(comparison.PRCoverage.Percentage),
			},
			Summary: templates.CoverageSummary{
				Direction:     comparison.TrendAnalysis.Direction,
				Magnitude:     comparison.TrendAnalysis.Magnitude,
				OverallImpact: determineOverallImpact(comparison.Difference),
			},
		},
		Comparison: templates.ComparisonData{
			BasePercentage:    comparison.BaseCoverage.Percentage,
			CurrentPercentage: comparison.PRCoverage.Percentage,
			Change:            comparison.Difference,
			Direction:         comparison.TrendAnalysis.Direction,
			Magnitude:         comparison.TrendAnalysis.Magnitude,
			IsSignificant:     comparison.Difference > 1.0 || comparison.Difference < -1.0,
		},
		Quality: templates.QualityData{
			OverallGrade:  calculateQualityGrade(comparison.PRCoverage.Percentage),
			CoverageGrade: calculateQualityGrade(comparison.PRCoverage.Percentage),
			TrendGrade:    calculateTrendGrade(comparison.TrendAnalysis.Direction),
			RiskLevel:     calculateRiskLevel(comparison.PRCoverage.Percentage),
			Score:         comparison.PRCoverage.Percentage,
		},
		PRFiles: convertPRFileAnalysis(comparison.PRFileAnalysis),
		Resources: templates.ResourceLinks{
			BadgeURL:      badgeURL,
			ReportURL:     reportURL,
			DashboardURL:  fmt.Sprintf("https://%s.github.io/%s/coverage/", cfg.GitHub.Owner, cfg.GitHub.Repository),
			HistoricalURL: fmt.Sprintf("https://%s.github.io/%s/coverage/trends/", cfg.GitHub.Owner, cfg.GitHub.Repository),
		},
	}
}

func calculateQualityGrade(percentage float64) string {
	switch {
	case percentage >= 95:
		return "A+"
	case percentage >= 90:
		return "A"
	case percentage >= 85:
		return "B+"
	case percentage >= 80:
		return "B"
	case percentage >= 70:
		return "C"
	case percentage >= 60:
		return "D"
	default:
		return "F"
	}
}

func calculateCoverageStatus(percentage float64) string {
	switch {
	case percentage >= 90:
		return "excellent"
	case percentage >= 80:
		return "good"
	case percentage >= 70:
		return "warning"
	default:
		return "critical"
	}
}

func calculateRiskLevel(percentage float64) string {
	switch {
	case percentage >= 80:
		return "low"
	case percentage >= 60:
		return "medium"
	case percentage >= 40:
		return "high"
	default:
		return "critical"
	}
}

func calculateTrendGrade(direction string) string {
	switch direction {
	case "up", "improved":
		return "A"
	case "stable":
		return "B"
	case "down", "degraded":
		return "D"
	default:
		return "C"
	}
}

func determineOverallImpact(difference float64) string {
	if difference > 1.0 {
		return "positive"
	} else if difference < -1.0 {
		return "negative"
	}
	return "neutral"
}

func determineBadgeTrend(direction string) badge.TrendDirection {
	switch strings.ToLower(direction) {
	case "up", "improved":
		return badge.TrendUp
	case "down", "degraded":
		return badge.TrendDown
	default:
		return badge.TrendStable
	}
}

// convertPRFileAnalysis converts GitHub PR file analysis to template data format
func convertPRFileAnalysis(analysis *github.PRFileAnalysis) *templates.PRFileAnalysisData {
	if analysis == nil {
		return nil
	}

	return &templates.PRFileAnalysisData{
		Summary: templates.PRFileSummaryData{
			TotalFiles:          analysis.Summary.TotalFiles,
			GoFilesCount:        analysis.Summary.GoFilesCount,
			TestFilesCount:      analysis.Summary.TestFilesCount,
			ConfigFilesCount:    analysis.Summary.ConfigFilesCount,
			DocumentationCount:  analysis.Summary.DocumentationCount,
			GeneratedFilesCount: analysis.Summary.GeneratedFilesCount,
			OtherFilesCount:     analysis.Summary.OtherFilesCount,
			HasGoChanges:        analysis.Summary.HasGoChanges,
			HasTestChanges:      analysis.Summary.HasTestChanges,
			HasConfigChanges:    analysis.Summary.HasConfigChanges,
			TotalAdditions:      analysis.Summary.TotalAdditions,
			TotalDeletions:      analysis.Summary.TotalDeletions,
			GoAdditions:         analysis.Summary.GoAdditions,
			GoDeletions:         analysis.Summary.GoDeletions,
			SummaryText:         analysis.Summary.GetSummaryText(),
		},
		GoFiles:            convertPRFiles(analysis.GoFiles),
		TestFiles:          convertPRFiles(analysis.TestFiles),
		ConfigFiles:        convertPRFiles(analysis.ConfigFiles),
		DocumentationFiles: convertPRFiles(analysis.DocumentationFiles),
		GeneratedFiles:     convertPRFiles(analysis.GeneratedFiles),
		OtherFiles:         convertPRFiles(analysis.OtherFiles),
	}
}

// convertPRFiles converts PR files to template format
func convertPRFiles(files []github.PRFile) []templates.PRFileData {
	result := make([]templates.PRFileData, len(files))
	for i, file := range files {
		result[i] = templates.PRFileData{
			Filename:  file.Filename,
			Status:    file.Status,
			Additions: file.Additions,
			Deletions: file.Deletions,
			Changes:   file.Changes,
		}
	}
	return result
}
