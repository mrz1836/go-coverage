package artifacts

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mrz1836/go-coverage/internal/github"
	"github.com/mrz1836/go-coverage/internal/history"
)

// Manager is the concrete implementation of ArtifactManager
type Manager struct {
	client  *GitHubCLI
	tempDir string
	maxSize int64 // Maximum history size in bytes (10MB)
	maxAge  time.Duration
}

// NewManager creates a new artifact manager
func NewManager() (*Manager, error) {
	client, err := NewGitHubCLI()
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub CLI client: %w", err)
	}

	tempDir := os.Getenv("RUNNER_TEMP")
	if tempDir == "" {
		tempDir = "/tmp"
	}
	tempDir = filepath.Join(tempDir, "coverage-artifacts")

	return &Manager{
		client:  client,
		tempDir: tempDir,
		maxSize: 10 * 1024 * 1024,    // 10MB
		maxAge:  30 * 24 * time.Hour, // 30 days
	}, nil
}

// DownloadHistory downloads the most recent coverage history from GitHub artifacts
func (m *Manager) DownloadHistory(ctx context.Context, opts *DownloadOptions) (*History, error) {
	if opts == nil {
		opts = DefaultDownloadOptions()
	}

	// Get current GitHub context for branch info
	githubCtx, err := github.DetectEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub context: %w", err)
	}

	// Set branch from context if not specified
	if opts.Branch == "" {
		opts.Branch = githubCtx.Branch
	}

	// List artifacts with current branch preference
	listOpts := &ListOptions{
		Branch: opts.Branch,
		Limit:  opts.MaxRuns * 2, // Get more to have options
	}

	artifacts, err := m.client.ListArtifacts(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w", err)
	}

	// Sort artifacts by creation time (newest first)
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].CreatedAt.After(artifacts[j].CreatedAt)
	})

	// Try to find suitable artifact
	var selectedArtifact *ArtifactInfo

	// First, try current branch
	for _, artifact := range artifacts {
		if artifact.Branch == opts.Branch {
			if opts.MaxAge > 0 && time.Since(artifact.CreatedAt) > opts.MaxAge {
				continue
			}
			selectedArtifact = artifact
			break
		}
	}

	// Fallback to main/master branch if no current branch history found
	if selectedArtifact == nil && opts.FallbackToBranch != "" {
		for _, artifact := range artifacts {
			if artifact.Branch == opts.FallbackToBranch {
				if opts.MaxAge > 0 && time.Since(artifact.CreatedAt) > opts.MaxAge {
					continue
				}
				selectedArtifact = artifact
				break
			}
		}
	}

	// If no artifact found, return empty history
	if selectedArtifact == nil {
		return m.createEmptyHistory(), nil
	}

	// Download and parse the artifact
	return m.downloadAndParseArtifact(ctx, selectedArtifact)
}

// MergeHistory merges current coverage data with previous history
func (m *Manager) MergeHistory(current, previous *History) (*History, error) {
	if current == nil {
		return previous, nil
	}
	if previous == nil {
		return current, nil
	}

	// Create merged history
	merged := &History{
		Records: make([]history.CoverageRecord, 0, len(current.Records)+len(previous.Records)),
		Metadata: &HistoryMetadata{
			Version:   "1.0",
			CreatedAt: current.Metadata.CreatedAt,
			UpdatedAt: time.Now(),
		},
	}

	// Use a map to track records by commit SHA to avoid duplicates
	recordMap := make(map[string]history.CoverageRecord)

	// Add previous records first
	for _, record := range previous.Records {
		key := fmt.Sprintf("%s-%s", record.CommitSHA, record.Timestamp.Format(time.RFC3339))
		recordMap[key] = record
	}

	// Add current records, potentially overwriting old ones with same SHA
	for _, record := range current.Records {
		key := fmt.Sprintf("%s-%s", record.CommitSHA, record.Timestamp.Format(time.RFC3339))
		recordMap[key] = record
	}

	// Convert back to slice and sort by timestamp
	for _, record := range recordMap {
		merged.Records = append(merged.Records, record)
	}

	// Sort by timestamp (oldest first)
	sort.Slice(merged.Records, func(i, j int) bool {
		return merged.Records[i].Timestamp.Before(merged.Records[j].Timestamp)
	})

	// Enforce size limits
	merged = m.enforceSizeLimits(merged)

	// Update metadata
	merged.Metadata.RecordCount = len(merged.Records)
	merged.Metadata.UpdatedAt = time.Now()

	return merged, nil
}

// UploadHistory uploads the updated history as a GitHub artifact
func (m *Manager) UploadHistory(ctx context.Context, hist *History, opts *UploadOptions) error {
	if opts == nil {
		opts = DefaultUploadOptions()
	}

	// Get current GitHub context if not provided
	if opts.Branch == "" || opts.CommitSHA == "" {
		githubCtx, err := github.DetectEnvironment()
		if err != nil {
			return fmt.Errorf("failed to get GitHub context: %w", err)
		}
		if opts.Branch == "" {
			opts.Branch = githubCtx.Branch
		}
		if opts.CommitSHA == "" {
			opts.CommitSHA = githubCtx.CommitSHA
		}
		if opts.PRNumber == "" {
			opts.PRNumber = githubCtx.PRNumber
		}
	}

	// Update history metadata
	if hist.Metadata == nil {
		hist.Metadata = &HistoryMetadata{}
	}
	hist.Metadata.UpdatedAt = time.Now()
	hist.Metadata.RecordCount = len(hist.Records)

	// Serialize history to JSON
	data, err := json.MarshalIndent(hist, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// Create temporary file
	tempFile := filepath.Join(m.tempDir, "coverage-history-temp.json")
	if mkdirErr := os.MkdirAll(m.tempDir, 0o750); mkdirErr != nil {
		return fmt.Errorf("failed to create temp directory: %w", mkdirErr)
	}

	if writeErr := os.WriteFile(tempFile, data, 0o600); writeErr != nil {
		return fmt.Errorf("failed to write temp file: %w", writeErr)
	}

	// Generate artifact name
	artifactName := GenerateArtifactName(opts)

	// Upload artifact
	err = m.client.UploadArtifact(ctx, artifactName, tempFile, opts.RetentionDays)
	if err != nil {
		// Clean up temp file
		_ = os.Remove(tempFile)
		return fmt.Errorf("failed to upload artifact: %w", err)
	}

	// Clean up temp file
	_ = os.Remove(tempFile)

	return nil
}

// CleanupOldArtifacts removes expired artifacts based on retention policy
func (m *Manager) CleanupOldArtifacts(ctx context.Context, retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = 30 // Default retention
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	// List all coverage history artifacts
	listOpts := &ListOptions{
		Limit:          100,
		IncludeExpired: true,
	}

	artifacts, err := m.client.ListArtifacts(ctx, listOpts)
	if err != nil {
		return fmt.Errorf("failed to list artifacts: %w", err)
	}

	var deletedCount int
	for _, artifact := range artifacts {
		// Skip recent artifacts
		if artifact.CreatedAt.After(cutoffTime) {
			continue
		}

		// Don't delete main branch "latest" artifacts
		if strings.Contains(artifact.Name, "main-latest") {
			continue
		}

		// Delete old artifact
		if err := m.client.DeleteArtifact(ctx, artifact.ID); err != nil {
			// Log error but continue with other artifacts
			fmt.Fprintf(os.Stderr, "Warning: Failed to delete artifact %s: %v\n", artifact.Name, err)
			continue
		}

		deletedCount++
	}

	if deletedCount > 0 {
		fmt.Fprintf(os.Stderr, "Cleaned up %d old coverage history artifacts\n", deletedCount)
	}

	return nil
}

// ListArtifacts lists available coverage history artifacts
func (m *Manager) ListArtifacts(ctx context.Context, opts *ListOptions) ([]*ArtifactInfo, error) {
	return m.client.ListArtifacts(ctx, opts)
}

// downloadAndParseArtifact downloads and parses a specific artifact
func (m *Manager) downloadAndParseArtifact(ctx context.Context, artifact *ArtifactInfo) (*History, error) {
	// Create temporary download directory
	downloadDir := filepath.Join(m.tempDir, "download")
	if err := os.MkdirAll(downloadDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create download directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(downloadDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to clean up download directory: %v\n", err)
		}
	}()

	// Download artifact
	if err := m.client.DownloadArtifact(ctx, artifact.ID, downloadDir); err != nil {
		return nil, fmt.Errorf("failed to download artifact: %w", err)
	}

	// Find the JSON file (artifacts are typically zipped)
	historyFile, err := m.findHistoryFile(downloadDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find history file: %w", err)
	}

	// Read and parse the history file
	data, err := os.ReadFile(historyFile) //nolint:gosec // controlled file path from artifact processing
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var hist History
	if err := json.Unmarshal(data, &hist); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return &hist, nil
}

// findHistoryFile finds the coverage history JSON file in the downloaded artifact
func (m *Manager) findHistoryFile(dir string) (string, error) {
	var historyFile string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".json") && strings.Contains(info.Name(), "history") {
			historyFile = path
			return nil
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	if historyFile == "" {
		return "", ErrNoHistoryFound
	}

	return historyFile, nil
}

// createEmptyHistory creates an empty history structure
func (m *Manager) createEmptyHistory() *History {
	return &History{
		Records: []history.CoverageRecord{},
		Metadata: &HistoryMetadata{
			Version:     "1.0",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			RecordCount: 0,
		},
	}
}

// enforceSizeLimits ensures the history doesn't exceed size limits
func (m *Manager) enforceSizeLimits(hist *History) *History {
	// First, try to estimate size
	data, err := json.Marshal(hist)
	if err != nil || int64(len(data)) <= m.maxSize {
		return hist
	}

	// Remove oldest records until under size limit
	for len(hist.Records) > 1 {
		// Remove 10% of oldest records at a time
		removeCount := len(hist.Records) / 10
		if removeCount < 1 {
			removeCount = 1
		}

		hist.Records = hist.Records[removeCount:]

		// Check size again
		data, err := json.Marshal(hist)
		if err == nil && int64(len(data)) <= m.maxSize {
			break
		}
	}

	// Update metadata
	hist.Metadata.RecordCount = len(hist.Records)
	hist.Metadata.UpdatedAt = time.Now()

	return hist
}
