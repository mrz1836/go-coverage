// Package github provides PR comment template generation for coverage reporting
package github

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrz1836/go-coverage/internal/history"
)

// CommentTemplateGenerator generates formatted PR comment bodies
type CommentTemplateGenerator struct {
	config     *CommentTemplateConfig
	visualizer *TrendVisualizer
}

// CommentTemplateConfig configures comment template generation
type CommentTemplateConfig struct {
	// Repository information for links
	Repository string

	// Display options
	ShowTrendChart     bool
	ShowPackageDetails bool
	ShowFileDetails    bool
	ShowFullHistory    bool

	// Thresholds
	CoverageTarget   float64
	WarningThreshold float64

	// Links
	ReportBaseURL string // GitHub Pages base URL for reports
	BadgeURL      string // Badge URL

	// Branding
	ShowBranding bool
	CompactMode  bool // Compact mode for smaller comments
}

// DefaultCommentTemplateConfig returns default template configuration
func DefaultCommentTemplateConfig() *CommentTemplateConfig {
	return &CommentTemplateConfig{
		ShowTrendChart:     true,
		ShowPackageDetails: true,
		ShowFileDetails:    false,
		ShowFullHistory:    false,
		CoverageTarget:     65.0,
		WarningThreshold:   50.0,
		ShowBranding:       true,
		CompactMode:        false,
	}
}

// NewCommentTemplateGenerator creates a new comment template generator
func NewCommentTemplateGenerator(config *CommentTemplateConfig) *CommentTemplateGenerator {
	if config == nil {
		config = DefaultCommentTemplateConfig()
	}

	return &CommentTemplateGenerator{
		config: config,
		visualizer: NewTrendVisualizer(&TrendConfig{
			Width:      40,
			Height:     6,
			ShowValues: true,
			ChartStyle: "line",
			TimeFormat: "Jan 2",
		}),
	}
}

// GenerateComment creates a formatted PR comment from coverage comparison
func (ctg *CommentTemplateGenerator) GenerateComment(comparison *CoverageComparison, trendHistory []history.CoverageRecord, deploymentURL string) string {
	var sections []string

	// Add comment signature for identification
	sections = append(sections, "[//]: # (go-coverage-v1)")

	// Header section
	sections = append(sections, ctg.generateHeader(comparison))

	// Coverage changes section
	sections = append(sections, ctg.generateCoverageChanges(comparison))

	// Package details section (collapsible)
	if ctg.config.ShowPackageDetails {
		sections = append(sections, ctg.generatePackageDetails(comparison))
	}

	// Trend section (collapsible)
	if ctg.config.ShowTrendChart && len(trendHistory) > 1 {
		sections = append(sections, ctg.generateTrendSection(trendHistory))
	}

	// File details section (collapsible)
	if ctg.config.ShowFileDetails && len(comparison.FileChanges) > 0 {
		sections = append(sections, ctg.generateFileDetails(comparison))
	}

	// Links section
	sections = append(sections, ctg.generateLinks(deploymentURL, comparison))

	// Footer section
	if ctg.config.ShowBranding {
		sections = append(sections, ctg.generateFooter())
	}

	return strings.Join(sections, "\n\n")
}

// generateHeader creates the main header with coverage status
func (ctg *CommentTemplateGenerator) generateHeader(comparison *CoverageComparison) string {
	current := comparison.PRCoverage.Percentage
	target := ctg.config.CoverageTarget
	diff := comparison.Difference

	// Determine status emoji and text
	var statusEmoji, statusText string
	if current >= target {
		statusEmoji = "✅"
		statusText = "PASSING"
	} else if current >= ctg.config.WarningThreshold {
		statusEmoji = "⚠️"
		statusText = "WARNING"
	} else {
		statusEmoji = "❌"
		statusText = "FAILING"
	}

	// Determine change emoji
	var changeEmoji, changeText string
	if diff > 0.5 {
		changeEmoji = "📈"
		changeText = fmt.Sprintf("(+%.2f%%)", diff)
	} else if diff < -0.5 {
		changeEmoji = "📉"
		changeText = fmt.Sprintf("(%.2f%%)", diff)
	} else {
		changeEmoji = "📊"
		changeText = "(no change)"
	}

	header := fmt.Sprintf("## 📊 Coverage Report %s\n\n", changeEmoji)

	if ctg.config.CompactMode {
		header += fmt.Sprintf("**%.2f%%** %s | Target: %.1f%% | Status: **%s** %s",
			current, changeText, target, statusText, statusEmoji)
	} else {
		header += fmt.Sprintf("**Current Coverage:** %.2f%% %s %s\n", current, changeText, changeEmoji)
		header += fmt.Sprintf("**Target:** %.1f%% | **Status:** %s **%s**", target, statusEmoji, statusText)
	}

	return header
}

// generateCoverageChanges creates the coverage changes table
func (ctg *CommentTemplateGenerator) generateCoverageChanges(comparison *CoverageComparison) string {
	if ctg.config.CompactMode {
		return "" // Skip detailed table in compact mode
	}

	base := comparison.BaseCoverage.Percentage
	current := comparison.PRCoverage.Percentage
	diff := comparison.Difference

	var changeIcon string
	if diff > 0.1 {
		changeIcon = "🟢"
	} else if diff < -0.1 {
		changeIcon = "🔴"
	} else {
		changeIcon = "🔵"
	}

	table := "### Coverage Changes\n\n"
	table += "| Metric | Base | Current | Change | Status |\n"
	table += "|--------|------|---------|--------|---------|\n"
	table += fmt.Sprintf("| Coverage | %.2f%% | %.2f%% | %+.2f%% | %s |\n",
		base, current, diff, changeIcon)
	table += fmt.Sprintf("| Statements | %d | %d | %+d | - |\n",
		comparison.BaseCoverage.CoveredStatements,
		comparison.PRCoverage.CoveredStatements,
		comparison.PRCoverage.CoveredStatements-comparison.BaseCoverage.CoveredStatements)

	return table
}

// generatePackageDetails creates collapsible package coverage breakdown
func (ctg *CommentTemplateGenerator) generatePackageDetails(comparison *CoverageComparison) string {
	if len(comparison.FileChanges) == 0 {
		return ""
	}

	details := "<details>\n<summary>📦 Package Coverage Details</summary>\n\n"
	details += "| Package | Coverage | Change | Files |\n"
	details += "|---------|----------|--------|---------|\n"

	// Show up to 10 most significant packages
	count := 0
	for _, change := range comparison.FileChanges {
		if count >= 10 {
			break
		}

		var changeText string
		if change.Difference > 0.1 {
			changeText = fmt.Sprintf("+%.1f%%", change.Difference)
		} else if change.Difference < -0.1 {
			changeText = fmt.Sprintf("%.1f%%", change.Difference)
		} else {
			changeText = "~"
		}

		// Extract package name from filename
		pkgName := change.Filename
		if strings.Contains(pkgName, "/") {
			parts := strings.Split(pkgName, "/")
			pkgName = parts[len(parts)-1]
		}
		if pkgName == "" {
			pkgName = "main"
		}

		details += fmt.Sprintf("| %s | %.1f%% | %s | - |\n",
			pkgName, change.PRCoverage, changeText)
		count++
	}

	details += "\n</details>"
	return details
}

// generateTrendSection creates collapsible trend analysis with ASCII chart
func (ctg *CommentTemplateGenerator) generateTrendSection(trendHistory []history.CoverageRecord) string {
	if len(trendHistory) < 2 {
		return ""
	}

	trend := "<details>\n<summary>📈 Coverage Trend</summary>\n\n"

	// Generate trend summary
	summary := ctg.visualizer.GenerateTrendSummary(trendHistory)
	trend += summary + "\n\n"

	// Generate ASCII chart
	chart := ctg.visualizer.GenerateASCIIChart(trendHistory)
	trend += "```\n" + chart + "\n```\n\n"

	// Add compact sparkline as alternative
	sparkline := ctg.visualizer.GenerateCompactTrend(trendHistory)
	trend += fmt.Sprintf("**Sparkline:** %s\n", sparkline)

	trend += "\n</details>"
	return trend
}

// generateFileDetails creates collapsible file-level changes
func (ctg *CommentTemplateGenerator) generateFileDetails(comparison *CoverageComparison) string {
	if len(comparison.FileChanges) == 0 {
		return ""
	}

	details := "<details>\n<summary>📁 File Coverage Changes</summary>\n\n"

	// Only show significant files
	significantFiles := 0
	for _, change := range comparison.FileChanges {
		if change.IsSignificant {
			significantFiles++
		}
	}

	if significantFiles == 0 {
		details += "No significant file coverage changes detected.\n"
	} else {
		details += "| File | Base | Current | Change | Lines |\n"
		details += "|------|------|---------|--------|---------|\n"

		for _, change := range comparison.FileChanges {
			if !change.IsSignificant {
				continue
			}

			// Truncate long filenames
			filename := change.Filename
			if len(filename) > 40 {
				filename = "..." + filename[len(filename)-37:]
			}

			var changeIcon string
			if change.Difference > 1.0 {
				changeIcon = "🟢"
			} else if change.Difference < -1.0 {
				changeIcon = "🔴"
			} else {
				changeIcon = "🔵"
			}

			lineChange := ""
			if change.LinesAdded > 0 || change.LinesRemoved > 0 {
				lineChange = fmt.Sprintf("+%d/-%d", change.LinesAdded, change.LinesRemoved)
			}

			details += fmt.Sprintf("| `%s` | %.1f%% | %.1f%% | %+.1f%% %s | %s |\n",
				filename, change.BaseCoverage, change.PRCoverage,
				change.Difference, changeIcon, lineChange)
		}
	}

	details += "\n</details>"
	return details
}

// generateLinks creates links section with report and badge URLs
func (ctg *CommentTemplateGenerator) generateLinks(deploymentURL string, comparison *CoverageComparison) string {
	if deploymentURL == "" && ctg.config.BadgeURL == "" {
		return ""
	}

	links := "### Links\n\n"

	if deploymentURL != "" {
		links += fmt.Sprintf("- 📊 [**Full Coverage Report**](%s)\n", deploymentURL)

		// Add branch-specific link if available
		if comparison.PRCoverage.Branch != "" {
			branchURL := strings.Replace(deploymentURL, "/coverage/", fmt.Sprintf("/branch/%s/", comparison.PRCoverage.Branch), 1)
			links += fmt.Sprintf("- 🌿 [Branch Coverage](%s)\n", branchURL)
		}
	}

	if ctg.config.BadgeURL != "" {
		links += fmt.Sprintf("- 🏷️ [Coverage Badge](%s)\n", ctg.config.BadgeURL)
	}

	if ctg.config.Repository != "" {
		historyURL := fmt.Sprintf("https://%s.github.io/coverage/history/",
			strings.Replace(ctg.config.Repository, "/", ".", 1))
		links += fmt.Sprintf("- 📜 [Coverage History](%s)\n", historyURL)
	}

	return strings.TrimSuffix(links, "\n")
}

// generateFooter creates the comment footer with branding and metadata
func (ctg *CommentTemplateGenerator) generateFooter() string {
	timestamp := time.Now().Format("2006-01-02 15:04:05 UTC")

	footer := "---\n\n"
	footer += fmt.Sprintf("*Generated by Go Coverage on %s*", timestamp)

	return footer
}

// GenerateCompactComment generates a minimal comment for situations with space constraints
func (ctg *CommentTemplateGenerator) GenerateCompactComment(comparison *CoverageComparison) string {
	// Create a compact config
	compactConfig := *ctg.config
	compactConfig.CompactMode = true
	compactConfig.ShowPackageDetails = false
	compactConfig.ShowFileDetails = false
	compactConfig.ShowTrendChart = false
	compactConfig.ShowBranding = false

	compactGenerator := &CommentTemplateGenerator{
		config:     &compactConfig,
		visualizer: ctg.visualizer,
	}

	return compactGenerator.GenerateComment(comparison, nil, "")
}

// GenerateStatusSummary creates a one-line status summary
func (ctg *CommentTemplateGenerator) GenerateStatusSummary(comparison *CoverageComparison) string {
	current := comparison.PRCoverage.Percentage
	diff := comparison.Difference
	target := ctg.config.CoverageTarget

	var status, emoji string
	if current >= target {
		status = "PASS"
		emoji = "✅"
	} else {
		status = "FAIL"
		emoji = "❌"
	}

	var changeText string
	if diff > 0.1 {
		changeText = fmt.Sprintf(" (+%.1f%%)", diff)
	} else if diff < -0.1 {
		changeText = fmt.Sprintf(" (%.1f%%)", diff)
	}

	return fmt.Sprintf("%s Coverage: %.1f%%%s | Target: %.1f%% | %s",
		emoji, current, changeText, target, status)
}
