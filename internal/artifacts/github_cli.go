package artifacts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mrz1836/go-coverage/internal/github"
)

var (
	// ErrNotGitHubActions indicates the code is not running in GitHub Actions
	ErrNotGitHubActions = errors.New("not running in GitHub Actions environment")
	// ErrFileNotExists indicates a file does not exist
	ErrFileNotExists = errors.New("file does not exist")
	// ErrNoHistoryFound indicates no history file was found in artifact
	ErrNoHistoryFound = errors.New("no history JSON file found in artifact")
)

// GitHubCLI implements GitHub CLI operations for artifact management
type GitHubCLI struct {
	repository string
	token      string
}

// NewGitHubCLI creates a new GitHub CLI client
func NewGitHubCLI() (*GitHubCLI, error) {
	ctx, err := github.DetectEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to detect GitHub environment: %w", err)
	}

	if !ctx.IsGitHubActions {
		return nil, ErrNotGitHubActions
	}

	return &GitHubCLI{
		repository: ctx.Repository,
		token:      ctx.Token,
	}, nil
}

// GitHubArtifact represents artifact information from GitHub API
type GitHubArtifact struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	SizeInBytes       int64     `json:"size_in_bytes"`
	ArchiveURL        string    `json:"archive_download_url"`
	Expired           bool      `json:"expired"`
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	WorkflowRunID     int64     `json:"workflow_run_id"`
	WorkflowRunHead   string    `json:"workflow_run_head_sha"`
	WorkflowRunBranch string    `json:"workflow_run_head_branch"`
}

// GitHubArtifactsResponse represents the response from GitHub artifacts API
type GitHubArtifactsResponse struct {
	TotalCount int               `json:"total_count"`
	Artifacts  []*GitHubArtifact `json:"artifacts"`
}

// ListArtifacts lists available coverage history artifacts using GitHub CLI
func (cli *GitHubCLI) ListArtifacts(ctx context.Context, opts *ListOptions) ([]*ArtifactInfo, error) {
	if opts == nil {
		opts = DefaultListOptions()
	}

	// Use gh api to list artifacts
	args := []string{
		"api",
		fmt.Sprintf("repos/%s/actions/artifacts", cli.repository),
		"--paginate",
	}

	// Add query parameters
	query := []string{
		fmt.Sprintf("per_page=%d", opts.Limit),
	}

	if len(query) > 0 {
		args = append(args, "-F", strings.Join(query, "&"))
	}

	cmd := exec.CommandContext(ctx, "gh", args...) //nolint:gosec // gh CLI command with controlled args
	cmd.Env = append(os.Environ(), fmt.Sprintf("GITHUB_TOKEN=%s", cli.token))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w (stderr: %s)", err, stderr.String())
	}

	var response GitHubArtifactsResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse artifacts response: %w", err)
	}

	// Filter and convert to ArtifactInfo
	artifacts := make([]*ArtifactInfo, 0, len(response.Artifacts))
	for _, artifact := range response.Artifacts {
		// Filter coverage history artifacts
		if !strings.HasPrefix(artifact.Name, "coverage-history-") {
			continue
		}

		// Filter by branch if specified
		if opts.Branch != "" && artifact.WorkflowRunBranch != opts.Branch {
			continue
		}

		// Skip expired artifacts unless explicitly requested
		if artifact.Expired && !opts.IncludeExpired {
			continue
		}

		// Parse artifact name for metadata
		branch, commitSHA, prNumber := parseArtifactName(artifact.Name)

		info := &ArtifactInfo{
			ID:          artifact.ID,
			Name:        artifact.Name,
			Branch:      branch,
			CommitSHA:   commitSHA,
			PRNumber:    prNumber,
			CreatedAt:   artifact.CreatedAt,
			Size:        artifact.SizeInBytes,
			DownloadURL: artifact.ArchiveURL,
		}

		artifacts = append(artifacts, info)
	}

	return artifacts, nil
}

// DownloadArtifact downloads a specific artifact to a temporary directory
func (cli *GitHubCLI) DownloadArtifact(ctx context.Context, artifactID int64, destDir string) error {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0o750); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Use gh run download to download artifact
	args := []string{
		"run", "download",
		"--repo", cli.repository,
		fmt.Sprintf("%d", artifactID),
		"--dir", destDir,
	}

	cmd := exec.CommandContext(ctx, "gh", args...) //nolint:gosec // gh CLI command with controlled args
	cmd.Env = append(os.Environ(), fmt.Sprintf("GITHUB_TOKEN=%s", cli.token))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download artifact: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}

// UploadArtifact uploads a file as a GitHub artifact
func (cli *GitHubCLI) UploadArtifact(ctx context.Context, name, filePath string, retentionDays int) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrFileNotExists, filePath)
	}

	// Use gh run upload to upload artifact (would need to be part of a workflow run)
	// This is a placeholder - actual artifact upload happens within GitHub Actions
	// The real implementation would write to a designated upload directory that
	// the workflow will pick up and upload as an artifact

	uploadDir := os.Getenv("RUNNER_TEMP")
	if uploadDir == "" {
		uploadDir = "/tmp"
	}

	uploadDir = filepath.Join(uploadDir, "coverage-artifacts")
	if err := os.MkdirAll(uploadDir, 0o750); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Copy file to upload directory with the artifact name
	destPath := filepath.Join(uploadDir, name+".json")

	// Read source file
	data, err := os.ReadFile(filePath) //nolint:gosec // controlled file path from artifact processing
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Write to upload directory
	if err := os.WriteFile(destPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write to upload directory: %w", err)
	}

	return nil
}

// DeleteArtifact deletes a specific artifact
func (cli *GitHubCLI) DeleteArtifact(ctx context.Context, artifactID int64) error {
	args := []string{
		"api",
		fmt.Sprintf("repos/%s/actions/artifacts/%d", cli.repository, artifactID),
		"--method", "DELETE",
	}

	cmd := exec.CommandContext(ctx, "gh", args...) //nolint:gosec // gh CLI command with controlled args
	cmd.Env = append(os.Environ(), fmt.Sprintf("GITHUB_TOKEN=%s", cli.token))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete artifact: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}

// parseArtifactName extracts metadata from artifact names
// Expected formats:
// - coverage-history-{branch}-{sha}-{timestamp}
// - coverage-history-main-latest
// - coverage-history-pr-{number}
func parseArtifactName(name string) (branch, commitSHA, prNumber string) {
	// Remove the prefix
	if !strings.HasPrefix(name, "coverage-history-") {
		return "", "", ""
	}

	remaining := name[17:] // Remove "coverage-history-" prefix

	if remaining == "" {
		return "", "", ""
	}

	// Handle special cases first
	if strings.HasPrefix(remaining, "main-latest") {
		return "main", "", ""
	}

	if strings.HasPrefix(remaining, "master-latest") {
		return "master", "", ""
	}

	if strings.HasPrefix(remaining, "pr-") {
		prNumber = strings.TrimPrefix(remaining, "pr-")
		return "", "", prNumber
	}

	// For standard format, we need to work backwards from the end
	// since branch names can contain hyphens but timestamps are always numeric
	parts := strings.Split(remaining, "-")

	if len(parts) >= 3 {
		// Check if last part looks like a timestamp (all digits)
		lastPart := parts[len(parts)-1]
		isTimestamp := true
		for _, r := range lastPart {
			if r < '0' || r > '9' {
				isTimestamp = false
				break
			}
		}

		if isTimestamp && len(lastPart) >= 8 { // Reasonable timestamp length
			// We have branch-...-sha-timestamp format
			commitSHA = parts[len(parts)-2]
			// Branch is everything except the last two parts
			branchParts := parts[:len(parts)-2]
			branch = strings.Join(branchParts, "-")
			return branch, commitSHA, ""
		}
	}

	// Fallback: treat the entire remaining string as branch name
	branch = remaining
	return branch, "", ""
}

// generateArtifactName generates a standardized artifact name
func GenerateArtifactName(opts *UploadOptions) string {
	if opts.Name != "" {
		return opts.Name
	}

	timestamp := time.Now().Unix()

	// Handle PR artifacts
	if opts.PRNumber != "" {
		return fmt.Sprintf("coverage-history-pr-%s", opts.PRNumber)
	}

	// Handle branch artifacts
	if opts.Branch != "" {
		if opts.Branch == "main" || opts.Branch == "master" {
			// Special case for main branch
			if opts.CommitSHA != "" {
				shortSHA := opts.CommitSHA
				if len(shortSHA) > 7 {
					shortSHA = shortSHA[:7]
				}
				return fmt.Sprintf("coverage-history-%s-%s-%d", opts.Branch, shortSHA, timestamp)
			}
			return fmt.Sprintf("coverage-history-%s-latest", opts.Branch)
		}

		// Regular branch
		if opts.CommitSHA != "" {
			shortSHA := opts.CommitSHA
			if len(shortSHA) > 7 {
				shortSHA = shortSHA[:7]
			}
			return fmt.Sprintf("coverage-history-%s-%s-%d", opts.Branch, shortSHA, timestamp)
		}
		return fmt.Sprintf("coverage-history-%s-%d", opts.Branch, timestamp)
	}

	// Fallback
	return fmt.Sprintf("coverage-history-%d", timestamp)
}
