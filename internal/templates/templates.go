package templates

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"time"
)

// embeddedFiles removed - assets are now in analytics/assets package

// TemplateManager handles template loading and rendering
type TemplateManager struct {
	templates *template.Template
	funcs     template.FuncMap
}

// DashboardData contains data for dashboard template rendering
type DashboardData struct {
	// Project information
	ProjectName string `json:"project_name"`

	// Repository information
	RepositoryOwner string `json:"repository_owner,omitempty"`
	RepositoryName  string `json:"repository_name,omitempty"`
	RepositoryURL   string `json:"repository_url,omitempty"`
	DefaultBranch   string `json:"default_branch,omitempty"`

	// Overall metrics
	TotalCoverage     float64 `json:"total_coverage"`
	CoverageTrend     float64 `json:"coverage_trend"`
	CoveredFiles      int     `json:"covered_files"`
	TotalFiles        int     `json:"total_files"`
	FilesTrend        int     `json:"files_trend"`
	LinesToCover      int     `json:"lines_to_cover"`
	LinesToCoverTrend int     `json:"lines_to_cover_trend"`
	PackagesTracked   int     `json:"packages_tracked"`

	// Branch information
	Branches []BranchData `json:"branches"`

	// Package information with GitHub URLs
	Packages []Package `json:"packages,omitempty"`

	// Metadata
	LastUpdated time.Time `json:"last_updated"`
	CommitSha   string    `json:"commit_sha"`

	// UI settings
	Theme       string `json:"theme"`
	ShowTrends  bool   `json:"show_trends"`
	ShowDetails bool   `json:"show_details"`

	// Build status information
	BuildStatus *BuildStatus `json:"build_status,omitempty"`
}

// BranchData contains branch-specific coverage information
type BranchData struct {
	Name         string    `json:"name"`
	Coverage     float64   `json:"coverage"`
	CoveredLines int       `json:"covered_lines"`
	TotalLines   int       `json:"total_lines"`
	Protected    bool      `json:"protected"`
	LastCommit   time.Time `json:"last_commit"`
	Trend        float64   `json:"trend"`
	GitHubURL    string    `json:"github_url,omitempty"`
}

// ReportData contains data for coverage report template rendering
type ReportData struct {
	// Report metadata
	Title       string    `json:"title"`
	ProjectName string    `json:"project_name"`
	Generated   time.Time `json:"generated"`
	Branch      string    `json:"branch"`
	CommitSha   string    `json:"commit_sha"`

	// Coverage summary
	OverallCoverage float64        `json:"overall_coverage"`
	PackageStats    []PackageStats `json:"package_stats"`
	FileStats       []FileStats    `json:"file_stats"`

	// Configuration
	Theme       string `json:"theme"`
	ShowDetails bool   `json:"show_details"`

	// GitHub integration
	GitHubOwner      string `json:"github_owner,omitempty"`
	GitHubRepository string `json:"github_repository,omitempty"`
	GitHubBranch     string `json:"github_branch,omitempty"`

	// Repository context (aliases for template compatibility)
	RepositoryOwner string `json:"repository_owner,omitempty"`
	RepositoryName  string `json:"repository_name,omitempty"`
}

// PackageStats contains package-level coverage statistics
type PackageStats struct {
	Name         string  `json:"name"`
	Coverage     float64 `json:"coverage"`
	Files        int     `json:"files"`
	Lines        int     `json:"lines"`
	CoveredLines int     `json:"covered_lines"`
}

// FileStats contains file-level coverage statistics
type FileStats struct {
	Name         string  `json:"name"`
	Path         string  `json:"path"`
	Package      string  `json:"package"`
	Coverage     float64 `json:"coverage"`
	Lines        int     `json:"lines"`
	CoveredLines int     `json:"covered_lines"`
	Functions    int     `json:"functions"`
	CoveredFuncs int     `json:"covered_funcs"`
}

// Package represents a package with coverage data for dashboard display
type Package struct {
	Name         string  `json:"name"`
	Coverage     float64 `json:"coverage"`
	CoveredLines int     `json:"covered_lines"`
	TotalLines   int     `json:"total_lines"`
	Files        []File  `json:"files,omitempty"`
	GitHubURL    string  `json:"github_url,omitempty"`
}

// File represents a file with coverage data for dashboard display
type File struct {
	Name      string  `json:"name"`
	Coverage  float64 `json:"coverage"`
	GitHubURL string  `json:"github_url,omitempty"`
}

// BuildStatus represents the current build status for live updates
type BuildStatus struct {
	State        string    `json:"state"`           // "queued", "in_progress", "completed"
	Conclusion   string    `json:"conclusion"`      // "success", "failure", "canceled", "skipped", "timed_out", "action_required", "neutral"
	WorkflowName string    `json:"workflow_name"`   // Name of the workflow
	RunID        int64     `json:"run_id"`          // GitHub Actions run ID
	RunNumber    int       `json:"run_number"`      // Run number for display
	RunURL       string    `json:"run_url"`         // URL to the GitHub Actions run
	StartedAt    time.Time `json:"started_at"`      // When the run started
	UpdatedAt    time.Time `json:"updated_at"`      // Last update time
	Duration     string    `json:"duration"`        // Human-readable duration
	HeadSHA      string    `json:"head_sha"`        // Commit SHA being built
	HeadBranch   string    `json:"head_branch"`     // Branch being built
	Event        string    `json:"event"`           // Event that triggered the run (push, pull_request, etc.)
	DisplayTitle string    `json:"display_title"`   // Display title for the run
	Available    bool      `json:"available"`       // Whether build status is available
	Error        string    `json:"error,omitempty"` // Error message if status fetch failed
}

// NewTemplateManager creates a new template manager with embedded templates
func NewTemplateManager() (*TemplateManager, error) {
	tm := &TemplateManager{
		funcs: template.FuncMap{
			"formatFloat":      formatFloat,
			"formatPercentage": formatPercentage,
			"formatTime":       formatTime,
			"colorForCoverage": colorForCoverage,
			"badgeColor":       badgeColor,
			"add":              add,
			"sub":              sub,
			"mul":              mul,
			"div":              div,
			"githubRepoURL":    githubRepoURL,
			"githubUserURL":    githubUserURL,
			"githubBranchURL":  githubBranchURL,
			"githubCommitURL":  githubCommitURL,
			"githubFileURL":    githubFileURL,
			"githubDirURL":     githubDirURL,
		},
	}

	// No templates to parse for now, just set up the funcs
	tm.templates = template.New("").Funcs(tm.funcs)
	return tm, nil
}

var (
	// ErrDashboardDeprecated indicates the dashboard rendering method is deprecated
	ErrDashboardDeprecated = errors.New("RenderDashboard is deprecated - use dashboard package's embedded template")
	// ErrReportDeprecated indicates the report rendering method is deprecated
	ErrReportDeprecated = errors.New("RenderReport is deprecated - use report package's embedded template")
	// ErrEmbeddedFilesMoved indicates embedded files have been moved to analytics/assets package
	ErrEmbeddedFilesMoved = errors.New("embedded files have been moved to analytics/assets package")
)

// RenderDashboard is deprecated - dashboard generator uses embedded template
func (tm *TemplateManager) RenderDashboard(_ context.Context, _ DashboardData) (string, error) {
	return "", ErrDashboardDeprecated
}

// RenderReport is deprecated - report generator uses embedded template
func (tm *TemplateManager) RenderReport(_ context.Context, _ ReportData) (string, error) {
	return "", ErrReportDeprecated
}

// WriteDashboard is deprecated - dashboard generator uses embedded template
func (tm *TemplateManager) WriteDashboard(_ context.Context, _ io.Writer, _ DashboardData) error {
	return ErrDashboardDeprecated
}

// WriteReport is deprecated - report generator uses embedded template
func (tm *TemplateManager) WriteReport(_ context.Context, _ io.Writer, _ ReportData) error {
	return ErrReportDeprecated
}

// GetEmbeddedFile is deprecated - assets are now in analytics/assets package
func (tm *TemplateManager) GetEmbeddedFile(filename string) ([]byte, error) {
	_ = filename // Avoid unused parameter warning
	return nil, ErrEmbeddedFilesMoved
}

// ListEmbeddedFiles is deprecated - assets are now in analytics/assets package
func (tm *TemplateManager) ListEmbeddedFiles() ([]string, error) {
	return nil, ErrEmbeddedFilesMoved
}

// Template helper functions

func formatFloat(f float64) string {
	return fmt.Sprintf("%.1f", f)
}

func formatPercentage(f float64) string {
	return fmt.Sprintf("%.1f%%", f)
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 UTC")
}

func colorForCoverage(coverage float64) string {
	switch {
	case coverage >= 90:
		return "#3fb950" // Excellent - bright green
	case coverage >= 80:
		return "#90c978" // Good - green
	case coverage >= 70:
		return "#d29922" // Acceptable - yellow
	case coverage >= 60:
		return "#f85149" // Low - orange
	default:
		return "#da3633" // Poor - red
	}
}

func badgeColor(coverage float64) string {
	return colorForCoverage(coverage)
}

// Math helper functions for templates
func add(a, b int) int { return a + b }

func sub(a, b int) int { return a - b }

func mul(a, b float64) float64 { return a * b }

func div(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}

// GitHub URL helper functions for templates

// githubRepoURL generates a GitHub repository URL
func githubRepoURL(owner, repo string) string {
	if owner == "" || repo == "" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/%s", owner, repo)
}

// githubUserURL generates a GitHub user profile URL
func githubUserURL(username string) string {
	if username == "" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s", username)
}

// githubBranchURL generates a GitHub branch URL
func githubBranchURL(owner, repo, branch string) string {
	if owner == "" || repo == "" || branch == "" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/%s/tree/%s", owner, repo, branch)
}

// githubCommitURL generates a GitHub commit URL
func githubCommitURL(owner, repo, sha string) string {
	if owner == "" || repo == "" || sha == "" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/%s/commit/%s", owner, repo, sha)
}

// githubFileURL generates a GitHub file URL
func githubFileURL(owner, repo, branch, filepath string) string {
	if owner == "" || repo == "" || branch == "" || filepath == "" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s", owner, repo, branch, filepath)
}

// githubDirURL generates a GitHub directory URL
func githubDirURL(owner, repo, branch, dirpath string) string {
	if owner == "" || repo == "" || branch == "" || dirpath == "" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/%s/tree/%s/%s", owner, repo, branch, dirpath)
}
