package artifacts

import (
	"context"
	"testing"
	"time"

	"github.com/mrz1836/go-coverage/internal/history"
)

func TestCreateCoverageRecord(t *testing.T) {
	// This test will only work in a GitHub Actions environment
	// In a normal test environment, it should return an error

	record, err := CreateCoverageRecord(85.5, 1000, 855)

	// In test environment (not GitHub Actions), we expect an error
	if err == nil {
		// If we're somehow in GitHub Actions environment, validate the record
		if record == nil {
			t.Fatal("Expected non-nil record")
		}

		if record.Percentage != 85.5 {
			t.Errorf("Expected percentage 85.5, got %f", record.Percentage)
		}

		if record.TotalLines != 1000 {
			t.Errorf("Expected total lines 1000, got %d", record.TotalLines)
		}

		if record.CoveredLines != 855 {
			t.Errorf("Expected covered lines 855, got %d", record.CoveredLines)
		}

		if record.Timestamp.IsZero() {
			t.Error("Expected non-zero timestamp")
		}
	} else {
		// Expected error in test environment
		if record != nil {
			t.Error("Expected nil record when error occurs")
		}
	}
}

func TestEnsureArtifactsAvailable(t *testing.T) {
	// This test will fail in normal test environment since we're not in GitHub Actions
	err := EnsureArtifactsAvailable()

	// We expect this to fail in test environment
	if err == nil {
		t.Log("Running in GitHub Actions environment - artifacts are available")
	} else {
		t.Logf("Expected error in test environment: %v", err)
	}
}

// Mock implementations for testing

type mockArtifactManager struct {
	histories         map[string]*History
	uploadCalled      bool
	downloadCalled    bool
	listCalled        bool
	cleanupCalled     bool
	lastUploadOptions *UploadOptions
	downloadError     error
	uploadError       error
}

func newMockArtifactManager() *mockArtifactManager {
	return &mockArtifactManager{
		histories: make(map[string]*History),
	}
}

func (m *mockArtifactManager) DownloadHistory(ctx context.Context, opts *DownloadOptions) (*History, error) {
	m.downloadCalled = true
	if m.downloadError != nil {
		return nil, m.downloadError
	}

	// Return empty history if none exists
	branch := "main"
	if opts != nil && opts.Branch != "" {
		branch = opts.Branch
	}

	if hist, exists := m.histories[branch]; exists {
		return hist, nil
	}

	// Return empty history
	return &History{
		Records: []history.CoverageRecord{},
		Metadata: &HistoryMetadata{
			Version:     "1.0",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			RecordCount: 0,
		},
	}, nil
}

func (m *mockArtifactManager) MergeHistory(current, previous *History) (*History, error) {
	if current == nil {
		return previous, nil
	}
	if previous == nil {
		return current, nil
	}

	// Simple merge implementation for testing
	merged := &History{
		Records: make([]history.CoverageRecord, 0, len(current.Records)+len(previous.Records)),
		Metadata: &HistoryMetadata{
			Version:   "1.0",
			CreatedAt: previous.Metadata.CreatedAt,
			UpdatedAt: time.Now(),
		},
	}

	merged.Records = append(merged.Records, previous.Records...)
	merged.Records = append(merged.Records, current.Records...)
	merged.Metadata.RecordCount = len(merged.Records)

	return merged, nil
}

func (m *mockArtifactManager) UploadHistory(ctx context.Context, history *History, opts *UploadOptions) error {
	m.uploadCalled = true
	m.lastUploadOptions = opts

	if m.uploadError != nil {
		return m.uploadError
	}

	// Store the history
	branch := "main"
	if opts != nil && opts.Branch != "" {
		branch = opts.Branch
	}

	m.histories[branch] = history
	return nil
}

func (m *mockArtifactManager) CleanupOldArtifacts(ctx context.Context, retentionDays int) error {
	m.cleanupCalled = true
	return nil
}

func (m *mockArtifactManager) ListArtifacts(ctx context.Context, opts *ListOptions) ([]*ArtifactInfo, error) {
	m.listCalled = true

	// Return mock artifacts
	return []*ArtifactInfo{
		{
			ID:        123,
			Name:      "coverage-history-main-latest",
			Branch:    "main",
			CreatedAt: time.Now(),
			Size:      1024,
		},
	}, nil
}

func TestMockManager(t *testing.T) {
	mock := newMockArtifactManager()

	// Test that the mock implements the interface
	var manager ArtifactManager = mock

	t.Run("DownloadHistory", func(t *testing.T) {
		history, err := manager.DownloadHistory(context.Background(), nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !mock.downloadCalled {
			t.Error("Expected download to be called")
		}

		if history == nil {
			t.Fatal("Expected non-nil history")
		}

		if len(history.Records) != 0 {
			t.Errorf("Expected empty history, got %d records", len(history.Records))
		}
	})

	t.Run("UploadHistory", func(t *testing.T) {
		testHistory := &History{
			Records: []history.CoverageRecord{
				{
					Timestamp:  time.Now(),
					CommitSHA:  "abc123",
					Branch:     "main",
					Percentage: 85.0,
				},
			},
		}

		opts := &UploadOptions{
			Branch: "test-branch",
		}

		err := manager.UploadHistory(context.Background(), testHistory, opts)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !mock.uploadCalled {
			t.Error("Expected upload to be called")
		}

		if mock.lastUploadOptions.Branch != "test-branch" {
			t.Errorf("Expected branch 'test-branch', got %s", mock.lastUploadOptions.Branch)
		}
	})

	t.Run("ListArtifacts", func(t *testing.T) {
		artifacts, err := manager.ListArtifacts(context.Background(), nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !mock.listCalled {
			t.Error("Expected list to be called")
		}

		if len(artifacts) != 1 {
			t.Errorf("Expected 1 artifact, got %d", len(artifacts))
		}

		if artifacts[0].Name != "coverage-history-main-latest" {
			t.Errorf("Expected artifact name 'coverage-history-main-latest', got %s", artifacts[0].Name)
		}
	})

	t.Run("CleanupOldArtifacts", func(t *testing.T) {
		err := manager.CleanupOldArtifacts(context.Background(), 30)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !mock.cleanupCalled {
			t.Error("Expected cleanup to be called")
		}
	})

	t.Run("MergeHistory", func(t *testing.T) {
		previous := &History{
			Records: []history.CoverageRecord{
				{
					Timestamp:  time.Now().Add(-time.Hour),
					CommitSHA:  "old123",
					Percentage: 80.0,
				},
			},
			Metadata: &HistoryMetadata{
				CreatedAt: time.Now().Add(-time.Hour),
			},
		}

		current := &History{
			Records: []history.CoverageRecord{
				{
					Timestamp:  time.Now(),
					CommitSHA:  "new456",
					Percentage: 85.0,
				},
			},
			Metadata: &HistoryMetadata{
				CreatedAt: time.Now(),
			},
		}

		merged, err := manager.MergeHistory(current, previous)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(merged.Records) != 2 {
			t.Errorf("Expected 2 records, got %d", len(merged.Records))
		}

		if merged.Metadata.RecordCount != 2 {
			t.Errorf("Expected record count 2, got %d", merged.Metadata.RecordCount)
		}
	})
}
