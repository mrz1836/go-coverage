// Package deployment provides GitHub Pages deployment capabilities for coverage reports.
// This package implements intelligent incremental deployment that preserves existing
// content while aggressively cleaning unwanted files from the gh-pages branch.
package deployment

import (
	"context"
	"time"
)

// DeploymentManager defines the interface for managing GitHub Pages deployments
type DeploymentManager interface {
	// Deploy performs a complete deployment of coverage reports to GitHub Pages
	Deploy(ctx context.Context, opts *DeploymentOptions) (*DeploymentResult, error)

	// Verify checks that the deployment was successful and URLs are accessible
	Verify(ctx context.Context, result *DeploymentResult) error

	// Rollback reverts the deployment to the previous state
	Rollback(ctx context.Context, backupRef string) error

	// GetDeploymentURL returns the URL for the deployed coverage report
	GetDeploymentURL(opts *DeploymentOptions) string

	// ListDeployments returns information about recent deployments
	ListDeployments(ctx context.Context, limit int) ([]*DeploymentInfo, error)
}

// DeploymentOptions configures deployment behavior
type DeploymentOptions struct {
	// CoverageFiles contains the coverage report files to deploy
	CoverageFiles map[string][]byte

	// Repository is the GitHub repository (owner/repo)
	Repository string

	// Branch is the source branch name
	Branch string

	// CommitSHA is the source commit SHA
	CommitSHA string

	// PRNumber is the pull request number (if applicable)
	PRNumber string

	// EventName is the GitHub event name (push, pull_request, etc.)
	EventName string

	// TargetPath specifies the deployment target path structure
	TargetPath DeploymentPath

	// CleanupPatterns specifies file patterns to remove during deployment
	CleanupPatterns []string

	// DryRun performs all operations except the final push
	DryRun bool

	// Force enables force push for conflict resolution
	Force bool

	// VerificationTimeout specifies how long to wait for URL verification
	VerificationTimeout time.Duration
}

// DeploymentPath defines the target path structure for coverage files
type DeploymentPath struct {
	// Type specifies the deployment type (main, branch, pr)
	Type PathType

	// Root is the base path (empty for root deployment)
	Root string

	// Identifier is the branch name or PR number
	Identifier string
}

// PathType represents the type of deployment path
type PathType string

const (
	// PathTypeMain represents the main branch deployment
	PathTypeMain PathType = "main"

	// PathTypeBranch represents a feature branch deployment
	PathTypeBranch PathType = "branch"

	// PathTypePR represents a pull request deployment
	PathTypePR PathType = "pr"

	// PathTypeRoot represents a root-level deployment
	PathTypeRoot PathType = "root"
)

// DeploymentResult contains information about a completed deployment
type DeploymentResult struct {
	// CommitSHA is the SHA of the deployment commit
	CommitSHA string

	// DeploymentURL is the primary URL for the deployed content
	DeploymentURL string

	// AdditionalURLs contains other URLs created during deployment
	AdditionalURLs []string

	// FilesDeployed is the number of files deployed
	FilesDeployed int

	// FilesRemoved is the number of files removed during cleanup
	FilesRemoved int

	// DeploymentTime is when the deployment completed
	DeploymentTime time.Time

	// BackupRef is the git reference for rollback purposes
	BackupRef string

	// Warnings contains any warnings encountered during deployment
	Warnings []string
}

// DeploymentInfo contains information about a deployment
type DeploymentInfo struct {
	// CommitSHA is the deployment commit SHA
	CommitSHA string

	// Branch is the source branch
	Branch string

	// PRNumber is the PR number (if applicable)
	PRNumber string

	// DeploymentTime is when the deployment occurred
	DeploymentTime time.Time

	// URL is the deployment URL
	URL string

	// Status is the deployment status
	Status DeploymentStatus

	// Message is the commit message
	Message string
}

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	// StatusPending represents a deployment in progress
	StatusPending DeploymentStatus = "pending"

	// StatusSuccess represents a successful deployment
	StatusSuccess DeploymentStatus = "success"

	// StatusFailed represents a failed deployment
	StatusFailed DeploymentStatus = "failed"

	// StatusRolledBack represents a rolled back deployment
	StatusRolledBack DeploymentStatus = "rolled_back"
)

// DefaultCleanupPatterns returns the default file patterns to remove during deployment
func DefaultCleanupPatterns() []string {
	return []string{
		"*.go", "*.mod", "*.sum", // Go files
		"*.yml", "*.yaml", // Workflow files
		"*.md", "LICENSE", "README*", // Documentation
		"cmd/", "internal/", "pkg/", // Source directories
		"test/", "testdata/", // Test files
		".github/", ".git/", // Git directories
		"docs/", "examples/", // Documentation directories
		"scripts/", "tools/", // Utility directories
		"*.txt", "*.log", // Text and log files
		"Makefile", "mage*", // Build files
		"go.work*", // Go workspace files
	}
}

// DefaultDeploymentOptions returns default deployment configuration
func DefaultDeploymentOptions() *DeploymentOptions {
	return &DeploymentOptions{
		CleanupPatterns:     DefaultCleanupPatterns(),
		DryRun:              false,
		Force:               false,
		VerificationTimeout: 30 * time.Second,
	}
}

// BuildDeploymentPath creates a deployment path based on the deployment context
func BuildDeploymentPath(eventName, branch, prNumber string) DeploymentPath {
	// Handle pull request deployments
	if eventName == "pull_request" && prNumber != "" {
		return DeploymentPath{
			Type:       PathTypePR,
			Root:       "pr",
			Identifier: prNumber,
		}
	}

	// Handle main branch deployments
	if branch == "main" || branch == "master" {
		return DeploymentPath{
			Type:       PathTypeMain,
			Root:       "main",
			Identifier: branch,
		}
	}

	// Handle feature branch deployments
	return DeploymentPath{
		Type:       PathTypeBranch,
		Root:       "branch",
		Identifier: sanitizeBranchName(branch),
	}
}

// String returns the full path for deployment
func (dp DeploymentPath) String() string {
	if dp.Type == PathTypeRoot {
		return ""
	}
	return dp.Root + "/" + dp.Identifier
}

// sanitizeBranchName sanitizes branch names for filesystem usage
func sanitizeBranchName(branch string) string {
	// Replace common problematic characters
	replacements := map[string]string{
		"/":  "-",
		"\\": "-",
		":":  "-",
		"*":  "-",
		"?":  "-",
		"\"": "-",
		"<":  "-",
		">":  "-",
		"|":  "-",
		" ":  "-",
	}

	result := branch
	for old, new := range replacements {
		result = replaceAll(result, old, new)
	}

	return result
}

// replaceAll is a simple string replacement function
func replaceAll(s, old, newStr string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if i <= len(s)-len(old) && s[i:i+len(old)] == old {
			result += newStr
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}
