package artifacts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mrz1836/go-coverage/internal/github"
	"github.com/mrz1836/go-coverage/internal/history"
)

// HistoryIntegration provides a bridge between the artifacts system and the existing history package
type HistoryIntegration struct {
	manager ArtifactManager
}

// NewHistoryIntegration creates a new history integration
func NewHistoryIntegration() (*HistoryIntegration, error) {
	manager, err := NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create artifact manager: %w", err)
	}

	return &HistoryIntegration{
		manager: manager,
	}, nil
}

// SaveRecord saves a coverage record to artifact-based history
func (hi *HistoryIntegration) SaveRecord(ctx context.Context, record *history.CoverageRecord) error {
	// Get current GitHub context
	githubCtx, err := github.DetectEnvironment()
	if err != nil {
		return fmt.Errorf("failed to get GitHub context: %w", err)
	}

	// Download existing history
	downloadOpts := DefaultDownloadOptions()
	downloadOpts.Branch = githubCtx.Branch

	existingHistory, err := hi.manager.DownloadHistory(ctx, downloadOpts)
	if err != nil {
		return fmt.Errorf("failed to download existing history: %w", err)
	}

	// Create new history with the current record
	currentHistory := &History{
		Records: []history.CoverageRecord{*record},
		Metadata: &HistoryMetadata{
			Version:     "1.0",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Repository:  githubCtx.Repository,
			RecordCount: 1,
		},
	}

	// Merge with existing history
	mergedHistory, err := hi.manager.MergeHistory(currentHistory, existingHistory)
	if err != nil {
		return fmt.Errorf("failed to merge history: %w", err)
	}

	// Upload merged history
	uploadOpts := DefaultUploadOptions()
	uploadOpts.Branch = githubCtx.Branch
	uploadOpts.CommitSHA = githubCtx.CommitSHA
	uploadOpts.PRNumber = githubCtx.PRNumber

	err = hi.manager.UploadHistory(ctx, mergedHistory, uploadOpts)
	if err != nil {
		return fmt.Errorf("failed to upload history: %w", err)
	}

	return nil
}

// GetLastRecord returns the most recent coverage record from artifact history
func (hi *HistoryIntegration) GetLastRecord(ctx context.Context) (*history.CoverageRecord, error) {
	// Get current GitHub context
	githubCtx, err := github.DetectEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub context: %w", err)
	}

	// Download history
	downloadOpts := DefaultDownloadOptions()
	downloadOpts.Branch = githubCtx.Branch

	hist, err := hi.manager.DownloadHistory(ctx, downloadOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to download history: %w", err)
	}

	if len(hist.Records) == 0 {
		return nil, history.ErrNoHistory
	}

	// Return the most recent record (history is sorted by timestamp)
	return &hist.Records[len(hist.Records)-1], nil
}

// GetChangeStatus compares current coverage with the last recorded coverage
func (hi *HistoryIntegration) GetChangeStatus(ctx context.Context, currentPercentage float64) (string, float64, error) {
	lastRecord, err := hi.GetLastRecord(ctx)
	if err != nil {
		if errors.Is(err, history.ErrNoHistory) {
			return "stable", 0.0, nil
		}
		return "stable", 0.0, err
	}

	previousPercentage := lastRecord.Percentage
	diff := currentPercentage - previousPercentage

	// Define thresholds for change detection
	const threshold = 0.1 // 0.1% threshold

	if diff > threshold {
		return "improved", previousPercentage, nil
	} else if diff < -threshold {
		return "declined", previousPercentage, nil
	}

	return "stable", previousPercentage, nil
}

// GetHistory returns all coverage records from artifact history
func (hi *HistoryIntegration) GetHistory(ctx context.Context) ([]history.CoverageRecord, error) {
	// Get current GitHub context
	githubCtx, err := github.DetectEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub context: %w", err)
	}

	// Download history
	downloadOpts := DefaultDownloadOptions()
	downloadOpts.Branch = githubCtx.Branch

	hist, err := hi.manager.DownloadHistory(ctx, downloadOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to download history: %w", err)
	}

	return hist.Records, nil
}

// CleanupHistory removes old coverage history artifacts
func (hi *HistoryIntegration) CleanupHistory(ctx context.Context) error {
	return hi.manager.CleanupOldArtifacts(ctx, 30) // 30 days retention
}

// CompatibilityManager provides a drop-in replacement for the existing history.Manager
// This allows existing code to use artifact-based history without modification
type CompatibilityManager struct {
	integration *HistoryIntegration
}

// NewCompatibilityManager creates a compatibility wrapper for the existing history.Manager API
func NewCompatibilityManager() (*CompatibilityManager, error) {
	integration, err := NewHistoryIntegration()
	if err != nil {
		return nil, err
	}

	return &CompatibilityManager{
		integration: integration,
	}, nil
}

// SaveRecord implements the same interface as history.Manager.SaveRecord
func (cm *CompatibilityManager) SaveRecord(record *history.CoverageRecord) error {
	return cm.integration.SaveRecord(context.Background(), record)
}

// GetLastRecord implements the same interface as history.Manager.GetLastRecord
func (cm *CompatibilityManager) GetLastRecord() (*history.CoverageRecord, error) {
	return cm.integration.GetLastRecord(context.Background())
}

// GetChangeStatus implements the same interface as history.Manager.GetChangeStatus
func (cm *CompatibilityManager) GetChangeStatus(currentPercentage float64) (string, float64, error) {
	return cm.integration.GetChangeStatus(context.Background(), currentPercentage)
}

// CreateCoverageRecord creates a coverage record with current GitHub context
func CreateCoverageRecord(percentage float64, totalLines, coveredLines int) (*history.CoverageRecord, error) {
	githubCtx, err := github.DetectEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub context: %w", err)
	}

	return &history.CoverageRecord{
		Timestamp:    time.Now(),
		CommitSHA:    githubCtx.CommitSHA,
		Branch:       githubCtx.Branch,
		Percentage:   percentage,
		TotalLines:   totalLines,
		CoveredLines: coveredLines,
	}, nil
}

// EnsureArtifactsAvailable checks if the GitHub Actions environment is properly set up for artifacts
func EnsureArtifactsAvailable() error {
	// Validate GitHub Actions environment
	if err := github.ValidateEnvironment(); err != nil {
		return fmt.Errorf("GitHub Actions environment validation failed: %w", err)
	}

	// Check if we can create the artifact manager
	_, err := NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize artifact manager: %w", err)
	}

	return nil
}
