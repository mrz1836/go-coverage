// Package analysis provides coverage comparison and diff calculation
package analysis

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/mrz1836/go-coverage/internal/artifacts"
	"github.com/mrz1836/go-coverage/internal/history"
	"github.com/mrz1836/go-coverage/internal/parser"
)

// Static error definitions
var (
	ErrNoCoverageRecords = errors.New("no coverage records found in base branch history")
)

// CoverageDiffer calculates coverage differences between current and base branches
type CoverageDiffer struct {
	artifactManager artifacts.ArtifactManager
}

// CoverageComparison represents coverage comparison between base and PR branches
type CoverageComparison struct {
	BaseCoverage     CoverageData    `json:"base_coverage"`
	PRCoverage       CoverageData    `json:"pr_coverage"`
	Difference       float64         `json:"difference"`
	TrendAnalysis    TrendData       `json:"trend_analysis"`
	FileChanges      []FileChange    `json:"file_changes"`
	SignificantFiles []string        `json:"significant_files"`
	PRFileAnalysis   *PRFileAnalysis `json:"pr_file_analysis,omitempty"`
}

// CoverageData represents coverage information for a specific commit
type CoverageData struct {
	Percentage        float64   `json:"percentage"`
	TotalStatements   int       `json:"total_statements"`
	CoveredStatements int       `json:"covered_statements"`
	CommitSHA         string    `json:"commit_sha"`
	Branch            string    `json:"branch"`
	Timestamp         time.Time `json:"timestamp"`
}

// TrendData represents trend analysis information
type TrendData struct {
	Direction        string  `json:"direction"` // "up", "down", "stable"
	Magnitude        string  `json:"magnitude"` // "significant", "moderate", "minor"
	PercentageChange float64 `json:"percentage_change"`
	Momentum         string  `json:"momentum"` // "accelerating", "steady", "decelerating"
}

// FileChange represents coverage change for a specific file
type FileChange struct {
	Filename      string  `json:"filename"`
	BaseCoverage  float64 `json:"base_coverage"`
	PRCoverage    float64 `json:"pr_coverage"`
	Difference    float64 `json:"difference"`
	LinesAdded    int     `json:"lines_added"`
	LinesRemoved  int     `json:"lines_removed"`
	IsSignificant bool    `json:"is_significant"`
}

// PRFile represents a file in a PR diff
type PRFile struct {
	Filename         string `json:"filename"`
	Status           string `json:"status"` // "added", "removed", "modified", "renamed"
	Additions        int    `json:"additions"`
	Deletions        int    `json:"deletions"`
	Changes          int    `json:"changes"`
	Patch            string `json:"patch,omitempty"` // Diff content
	BlobURL          string `json:"blob_url"`
	RawURL           string `json:"raw_url"`
	PreviousFilename string `json:"previous_filename,omitempty"` // For renamed files
}

// PRDiff represents the diff of a pull request
type PRDiff struct {
	Files []PRFile `json:"files"`
}

// PRFileAnalysis contains analyzed information about PR files
type PRFileAnalysis struct {
	GoFiles            []PRFile
	TestFiles          []PRFile
	ConfigFiles        []PRFile
	DocumentationFiles []PRFile
	GeneratedFiles     []PRFile
	OtherFiles         []PRFile
	Summary            PRFileSummary
}

// PRFileSummary provides summary statistics about the PR files
type PRFileSummary struct {
	TotalFiles          int
	GoFilesCount        int
	TestFilesCount      int
	ConfigFilesCount    int
	DocumentationCount  int
	GeneratedFilesCount int
	OtherFilesCount     int
	HasGoChanges        bool
	HasTestChanges      bool
	HasConfigChanges    bool
	TotalAdditions      int
	TotalDeletions      int
	GoAdditions         int
	GoDeletions         int
}

// NewCoverageDiffer creates a new coverage differ
func NewCoverageDiffer(artifactManager artifacts.ArtifactManager) *CoverageDiffer {
	return &CoverageDiffer{
		artifactManager: artifactManager,
	}
}

// CalculateDiff compares current coverage with base branch coverage from artifacts
func (cd *CoverageDiffer) CalculateDiff(ctx context.Context, currentCoverage *parser.CoverageData, baseBranch, currentBranch, prNumber string) (*CoverageComparison, error) {
	// Create current coverage data
	currentData := CoverageData{
		Percentage:        currentCoverage.Percentage,
		TotalStatements:   currentCoverage.TotalLines,
		CoveredStatements: currentCoverage.CoveredLines,
		CommitSHA:         "", // Will be filled by caller
		Branch:            currentBranch,
		Timestamp:         currentCoverage.Timestamp,
	}

	// Get base branch coverage from artifacts
	baseData, err := cd.getBaseCoverageData(ctx, baseBranch)
	if err != nil {
		// If no base data available, create default with 0% coverage
		baseData = &CoverageData{
			Percentage:        0.0,
			TotalStatements:   0,
			CoveredStatements: 0,
			CommitSHA:         "",
			Branch:            baseBranch,
			Timestamp:         time.Now(),
		}
	}

	// Calculate difference
	difference := currentData.Percentage - baseData.Percentage

	// Generate trend analysis
	trendAnalysis := cd.generateTrendAnalysis(difference, baseData.Percentage)

	// Generate file changes analysis
	fileChanges := cd.generateFileChanges(currentCoverage, baseData)

	// Create comparison
	comparison := &CoverageComparison{
		BaseCoverage:     *baseData,
		PRCoverage:       currentData,
		Difference:       difference,
		TrendAnalysis:    trendAnalysis,
		FileChanges:      fileChanges,
		SignificantFiles: cd.identifySignificantFiles(fileChanges),
	}

	return comparison, nil
}

// getBaseCoverageData retrieves base branch coverage from artifacts
func (cd *CoverageDiffer) getBaseCoverageData(ctx context.Context, baseBranch string) (*CoverageData, error) {
	// Download options for base branch
	opts := &artifacts.DownloadOptions{
		Branch:           baseBranch,
		MaxRuns:          8,
		FallbackToBranch: "main",             // Fallback to main if base branch not found
		MaxAge:           24 * 7 * time.Hour, // Look back 1 week
	}

	// Download history from artifacts
	historyData, err := cd.artifactManager.DownloadHistory(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to download base branch history: %w", err)
	}

	if len(historyData.Records) == 0 {
		return nil, ErrNoCoverageRecords
	}

	// Get the most recent record
	latestRecord := historyData.Records[len(historyData.Records)-1]

	// Convert history record to coverage data
	baseData := &CoverageData{
		Percentage:        latestRecord.Percentage,
		TotalStatements:   latestRecord.TotalLines,
		CoveredStatements: latestRecord.CoveredLines,
		CommitSHA:         latestRecord.CommitSHA,
		Branch:            latestRecord.Branch,
		Timestamp:         latestRecord.Timestamp,
	}

	return baseData, nil
}

// generateTrendAnalysis creates trend analysis based on coverage difference
func (cd *CoverageDiffer) generateTrendAnalysis(difference, _ float64) TrendData {
	var direction string
	var magnitude string
	var momentum string

	// Determine direction
	const minThreshold = 0.1 // 0.1% threshold for detecting change
	if difference > minThreshold {
		direction = "up"
	} else if difference < -minThreshold {
		direction = "down"
	} else {
		direction = "stable"
	}

	// Determine magnitude based on absolute difference
	absChange := math.Abs(difference)
	switch {
	case absChange >= 5.0:
		magnitude = "significant"
	case absChange >= 1.0:
		magnitude = "moderate"
	default:
		magnitude = "minor"
	}

	// Determine momentum (simplified - could be enhanced with more historical data)
	if direction == "stable" {
		momentum = "steady"
	} else if absChange >= 2.0 {
		momentum = "accelerating"
	} else {
		momentum = "steady"
	}

	return TrendData{
		Direction:        direction,
		Magnitude:        magnitude,
		PercentageChange: difference,
		Momentum:         momentum,
	}
}

// generateFileChanges creates file-level change analysis (simplified version)
func (cd *CoverageDiffer) generateFileChanges(currentCoverage *parser.CoverageData, baseCoverage *CoverageData) []FileChange {
	fileChanges := make([]FileChange, 0, len(currentCoverage.Packages))

	// For now, create package-level changes since we don't have detailed base file data
	// In a full implementation, this would compare file-by-file coverage
	for pkgName, pkg := range currentCoverage.Packages {
		change := FileChange{
			Filename:      pkgName,
			BaseCoverage:  baseCoverage.Percentage, // Simplified - using overall base coverage
			PRCoverage:    pkg.Percentage,
			Difference:    pkg.Percentage - baseCoverage.Percentage,
			LinesAdded:    0, // Would need PR diff to calculate
			LinesRemoved:  0, // Would need PR diff to calculate
			IsSignificant: math.Abs(pkg.Percentage-baseCoverage.Percentage) > 2.0,
		}
		fileChanges = append(fileChanges, change)
	}

	// Sort by significance and difference
	sort.Slice(fileChanges, func(i, j int) bool {
		if fileChanges[i].IsSignificant && !fileChanges[j].IsSignificant {
			return true
		}
		if !fileChanges[i].IsSignificant && fileChanges[j].IsSignificant {
			return false
		}
		return math.Abs(fileChanges[i].Difference) > math.Abs(fileChanges[j].Difference)
	})

	return fileChanges
}

// identifySignificantFiles identifies files with significant coverage changes
func (cd *CoverageDiffer) identifySignificantFiles(fileChanges []FileChange) []string {
	var significantFiles []string

	for _, change := range fileChanges {
		if change.IsSignificant {
			significantFiles = append(significantFiles, change.Filename)
		}
	}

	// Limit to top 5 most significant files to avoid overwhelming the comment
	if len(significantFiles) > 5 {
		significantFiles = significantFiles[:5]
	}

	return significantFiles
}

// EnhanceWithPRDiff enhances coverage comparison with PR diff information
func (cd *CoverageDiffer) EnhanceWithPRDiff(comparison *CoverageComparison, prDiff *PRDiff) {
	if prDiff == nil {
		return
	}

	// Analyze PR files and enhance file changes
	prAnalysis := AnalyzePRFiles(prDiff)
	comparison.PRFileAnalysis = prAnalysis

	// Update file changes with actual PR diff data
	for i, change := range comparison.FileChanges {
		for _, prFile := range prDiff.Files {
			if prFile.Filename == change.Filename ||
				fmt.Sprintf("/%s", prFile.Filename) == change.Filename {
				comparison.FileChanges[i].LinesAdded = prFile.Additions
				comparison.FileChanges[i].LinesRemoved = prFile.Deletions
				break
			}
		}
	}
}

// GetTrendHistory gets trend data for the last N records for visualization
func (cd *CoverageDiffer) GetTrendHistory(ctx context.Context, branch string, maxRecords int) ([]history.CoverageRecord, error) {
	opts := &artifacts.DownloadOptions{
		Branch:           branch,
		MaxRuns:          20, // Look through more runs to get enough history
		FallbackToBranch: "main",
		MaxAge:           24 * 30 * time.Hour, // Look back 30 days
	}

	historyData, err := cd.artifactManager.DownloadHistory(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to download trend history: %w", err)
	}

	if len(historyData.Records) == 0 {
		return []history.CoverageRecord{}, nil
	}

	records := historyData.Records
	if len(records) > maxRecords {
		// Return the most recent maxRecords
		records = records[len(records)-maxRecords:]
	}

	return records, nil
}

// AnalyzePRFiles is a simplified version that would normally be in the github package
// This prevents the import cycle by duplicating the functionality here
func AnalyzePRFiles(prDiff *PRDiff) *PRFileAnalysis {
	analysis := &PRFileAnalysis{
		Summary: PRFileSummary{},
	}

	for _, file := range prDiff.Files {
		// Update summary totals
		analysis.Summary.TotalFiles++
		analysis.Summary.TotalAdditions += file.Additions
		analysis.Summary.TotalDeletions += file.Deletions

		// Simple categorization - could be enhanced
		if strings.HasSuffix(file.Filename, ".go") && !strings.HasSuffix(file.Filename, "_test.go") {
			analysis.GoFiles = append(analysis.GoFiles, file)
			analysis.Summary.GoFilesCount++
			analysis.Summary.HasGoChanges = true
			analysis.Summary.GoAdditions += file.Additions
			analysis.Summary.GoDeletions += file.Deletions
		} else if strings.HasSuffix(file.Filename, "_test.go") {
			analysis.TestFiles = append(analysis.TestFiles, file)
			analysis.Summary.TestFilesCount++
			analysis.Summary.HasTestChanges = true
		} else {
			analysis.OtherFiles = append(analysis.OtherFiles, file)
			analysis.Summary.OtherFilesCount++
		}
	}

	return analysis
}
