package deployment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Static error definitions for deployment operations
var (
	ErrNoBackupForRollback = errors.New("no backup reference provided for rollback")
	ErrHTTPStatusError     = errors.New("URL returned non-success status")
)

// Manager is the concrete implementation of DeploymentManager
type Manager struct {
	gitClient     GitOperations
	cleanupEngine CleanupEngine
	htmlGenerator HTMLGenerator
	workDir       string
	dryRun        bool
	verbose       bool
}

// NewManager creates a new deployment manager
func NewManager(repository, token string, dryRun, verbose bool) (*Manager, error) {
	// Create work directory
	workDir := filepath.Join(os.TempDir(), fmt.Sprintf("gh-pages-deploy-%d", time.Now().Unix()))

	// Create git client
	gitClient := NewGitClient(repository, token)

	// Create cleanup engine
	cleanupEngine := NewFileCleanup(dryRun, verbose)

	// Create HTML generator
	baseURL := fmt.Sprintf("https://%s.github.io", getOwnerFromRepository(repository))
	htmlGenerator := NewReportGenerator(repository, baseURL)

	return &Manager{
		gitClient:     gitClient,
		cleanupEngine: cleanupEngine,
		htmlGenerator: htmlGenerator,
		workDir:       workDir,
		dryRun:        dryRun,
		verbose:       verbose,
	}, nil
}

// Deploy performs a complete deployment of coverage reports to GitHub Pages
func (m *Manager) Deploy(ctx context.Context, opts *DeploymentOptions) (*DeploymentResult, error) {
	if opts == nil {
		opts = DefaultDeploymentOptions()
	}

	result := &DeploymentResult{
		DeploymentTime: time.Now(),
		Warnings:       make([]string, 0),
	}

	// Acquire deployment lock
	// Sanitize repository and branch names to prevent path traversal issues
	sanitizedRepo := strings.ReplaceAll(opts.Repository, "/", "-")
	sanitizedBranch := strings.ReplaceAll(opts.Branch, "/", "-")
	lockName := fmt.Sprintf("%s-%s", sanitizedRepo, sanitizedBranch)
	if err := m.gitClient.AcquireLock(ctx, lockName, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("failed to acquire deployment lock: %w", err)
	}
	defer func() {
		if err := m.gitClient.ReleaseLock(ctx, lockName); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to release deployment lock: %v", err))
		}
	}()

	// Create backup before making changes
	if backupRef, err := m.createDeploymentBackup(ctx); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to create backup: %v", err))
	} else {
		result.BackupRef = backupRef
	}

	// Clone or create gh-pages branch
	if err := m.gitClient.CloneOrCreateBranch(ctx, m.workDir); err != nil {
		return nil, fmt.Errorf("failed to setup gh-pages branch: %w", err)
	}
	defer m.cleanup()

	// Perform aggressive cleanup
	cleanupResult, err := m.performCleanup(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup files: %w", err)
	}
	result.FilesRemoved = cleanupResult.FilesRemoved

	// Deploy coverage files
	filesDeployed, err := m.deployCoverageFiles(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy coverage files: %w", err)
	}
	result.FilesDeployed = filesDeployed

	// Discover existing reports and generate navigation
	if navErr := m.generateNavigation(); navErr != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to generate navigation: %v", navErr))
	}

	// Commit changes
	commitMessage := m.buildCommitMessage(opts)
	commitSHA, err := m.gitClient.CommitChanges(ctx, m.workDir, commitMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to commit changes: %w", err)
	}
	result.CommitSHA = commitSHA

	// Push changes (skip if dry run)
	if !m.dryRun {
		if err := m.gitClient.PushChanges(ctx, m.workDir, opts.Force); err != nil {
			return nil, fmt.Errorf("failed to push changes: %w", err)
		}
	}

	// Build URLs
	result.DeploymentURL = m.GetDeploymentURL(opts)
	result.AdditionalURLs = m.buildAdditionalURLs(opts)

	return result, nil
}

// Verify checks that the deployment was successful and URLs are accessible
func (m *Manager) Verify(ctx context.Context, result *DeploymentResult) error {
	if m.dryRun {
		return nil // Skip verification in dry run mode
	}

	// Wait for GitHub Pages to propagate
	time.Sleep(5 * time.Second)

	// Check primary URL
	if err := m.verifyURL(ctx, result.DeploymentURL); err != nil {
		return fmt.Errorf("primary deployment URL not accessible: %w", err)
	}

	// Check additional URLs
	for _, url := range result.AdditionalURLs {
		if err := m.verifyURL(ctx, url); err != nil {
			return fmt.Errorf("additional URL not accessible (%s): %w", url, err)
		}
	}

	return nil
}

// Rollback reverts the deployment to the previous state
func (m *Manager) Rollback(ctx context.Context, backupRef string) error {
	if backupRef == "" {
		return ErrNoBackupForRollback
	}

	// Setup work directory
	if err := m.gitClient.CloneOrCreateBranch(ctx, m.workDir); err != nil {
		return fmt.Errorf("failed to setup work directory for rollback: %w", err)
	}
	defer m.cleanup()

	// Perform rollback
	return m.gitClient.Rollback(ctx, m.workDir, backupRef)
}

// GetDeploymentURL returns the URL for the deployed coverage report
func (m *Manager) GetDeploymentURL(opts *DeploymentOptions) string {
	owner := getOwnerFromRepository(opts.Repository)
	repo := getRepoFromRepository(opts.Repository)
	baseURL := fmt.Sprintf("https://%s.github.io/%s", owner, repo)

	if opts.TargetPath.Type == PathTypeRoot {
		return baseURL + "/coverage.html"
	}

	return baseURL + "/" + opts.TargetPath.String() + "/coverage.html"
}

// ListDeployments returns information about recent deployments
func (m *Manager) ListDeployments(ctx context.Context, limit int) ([]*DeploymentInfo, error) {
	// This would typically query git history or GitHub API
	// For now, return empty list as this is a complex feature
	return []*DeploymentInfo{}, nil
}

// createDeploymentBackup creates a backup for rollback purposes
func (m *Manager) createDeploymentBackup(ctx context.Context) (string, error) {
	// Create temporary directory for backup
	backupDir := filepath.Join(os.TempDir(), fmt.Sprintf("gh-pages-backup-%d", time.Now().Unix()))
	defer func() { _ = os.RemoveAll(backupDir) }()

	// Clone current state
	if err := m.gitClient.CloneOrCreateBranch(ctx, backupDir); err != nil {
		return "", err
	}

	// Create backup reference
	return m.gitClient.CreateBackup(ctx, backupDir)
}

// performCleanup performs aggressive file cleanup
func (m *Manager) performCleanup(opts *DeploymentOptions) (*CleanupResult, error) {
	preservePatterns := DefaultPreservePatterns()
	return m.cleanupEngine.CleanupFiles(m.workDir, opts.CleanupPatterns, preservePatterns)
}

// deployCoverageFiles deploys the coverage files to the appropriate locations
func (m *Manager) deployCoverageFiles(opts *DeploymentOptions) (int, error) {
	filesDeployed := 0

	for filename, data := range opts.CoverageFiles {
		// Determine target path
		var targetPath string
		if opts.TargetPath.Type == PathTypeRoot {
			targetPath = filename
		} else {
			targetPath = filepath.Join(opts.TargetPath.String(), filename)
		}

		// Deploy the file
		if err := m.htmlGenerator.GenerateReportHTML(m.workDir, targetPath, data); err != nil {
			return filesDeployed, fmt.Errorf("failed to deploy file %s: %w", filename, err)
		}

		filesDeployed++
	}

	// Also copy coverage files to root level for latest coverage
	if opts.TargetPath.Type != PathTypeRoot {
		for filename, data := range opts.CoverageFiles {
			if filename == "coverage.html" || filename == "coverage.svg" {
				if err := m.htmlGenerator.GenerateReportHTML(m.workDir, filename, data); err != nil {
					return filesDeployed, fmt.Errorf("failed to deploy root file %s: %w", filename, err)
				}
			}
		}
	}

	return filesDeployed, nil
}

// generateNavigation generates the navigation index.html file
func (m *Manager) generateNavigation() error {
	// Discover existing reports
	reports, err := m.htmlGenerator.DiscoverReports(m.workDir)
	if err != nil {
		return fmt.Errorf("failed to discover reports: %w", err)
	}

	// Generate index.html
	return m.htmlGenerator.GenerateIndexHTML(m.workDir, reports)
}

// buildCommitMessage creates a commit message for the deployment
func (m *Manager) buildCommitMessage(opts *DeploymentOptions) string {
	switch opts.TargetPath.Type {
	case PathTypeMain:
		return fmt.Sprintf("Deploy coverage for %s branch (%s)", opts.Branch, opts.CommitSHA[:7])
	case PathTypeBranch:
		return fmt.Sprintf("Deploy coverage for branch %s (%s)", opts.Branch, opts.CommitSHA[:7])
	case PathTypePR:
		return fmt.Sprintf("Deploy coverage for PR #%s (%s)", opts.PRNumber, opts.CommitSHA[:7])
	case PathTypeRoot:
		return fmt.Sprintf("Deploy coverage (%s)", opts.CommitSHA[:7])
	default:
		return fmt.Sprintf("Deploy coverage (%s)", opts.CommitSHA[:7])
	}
}

// buildAdditionalURLs creates additional URLs for the deployment
func (m *Manager) buildAdditionalURLs(opts *DeploymentOptions) []string {
	var urls []string

	owner := getOwnerFromRepository(opts.Repository)
	repo := getRepoFromRepository(opts.Repository)
	baseURL := fmt.Sprintf("https://%s.github.io/%s", owner, repo)

	// Add badge URL
	if opts.TargetPath.Type == PathTypeRoot {
		urls = append(urls, baseURL+"/coverage.svg")
	} else {
		urls = append(urls, baseURL+"/"+opts.TargetPath.String()+"/coverage.svg")
	}

	// Add navigation URL
	urls = append(urls, baseURL+"/index.html")

	return urls
}

// verifyURL verifies that a URL is accessible
func (m *Manager) verifyURL(ctx context.Context, url string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to access URL: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d", ErrHTTPStatusError, resp.StatusCode)
	}

	return nil
}

// cleanup removes the temporary work directory
func (m *Manager) cleanup() {
	if m.workDir != "" {
		_ = os.RemoveAll(m.workDir)
	}
}

// getOwnerFromRepository extracts the owner from a repository string
func getOwnerFromRepository(repository string) string {
	parts := splitString(repository, "/")
	if len(parts) >= 2 {
		return parts[0]
	}
	return ""
}

// getRepoFromRepository extracts the repo name from a repository string
func getRepoFromRepository(repository string) string {
	parts := splitString(repository, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// splitString splits a string by delimiter (simple implementation)
func splitString(s, delimiter string) []string {
	var parts []string
	current := ""

	for i := 0; i < len(s); i++ {
		if i <= len(s)-len(delimiter) && s[i:i+len(delimiter)] == delimiter {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
			i += len(delimiter) - 1
		} else {
			current += string(s[i])
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
