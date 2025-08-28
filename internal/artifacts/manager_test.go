package artifacts

import (
	"testing"
	"time"

	"github.com/mrz1836/go-coverage/internal/history"
)

func TestDefaultOptions(t *testing.T) {
	t.Run("DefaultDownloadOptions", func(t *testing.T) {
		opts := DefaultDownloadOptions()
		if opts.MaxRuns != 8 {
			t.Errorf("Expected MaxRuns to be 8, got %d", opts.MaxRuns)
		}
		if opts.FallbackToBranch != "main" {
			t.Errorf("Expected FallbackToBranch to be 'main', got %s", opts.FallbackToBranch)
		}
		if opts.MaxAge != 24*7*time.Hour {
			t.Errorf("Expected MaxAge to be 1 week, got %v", opts.MaxAge)
		}
	})

	t.Run("DefaultUploadOptions", func(t *testing.T) {
		opts := DefaultUploadOptions()
		if opts.RetentionDays != 30 {
			t.Errorf("Expected RetentionDays to be 30, got %d", opts.RetentionDays)
		}
	})

	t.Run("DefaultListOptions", func(t *testing.T) {
		opts := DefaultListOptions()
		if opts.Limit != 50 {
			t.Errorf("Expected Limit to be 50, got %d", opts.Limit)
		}
		if opts.IncludeExpired != false {
			t.Errorf("Expected IncludeExpired to be false, got %t", opts.IncludeExpired)
		}
	})
}

func TestGenerateArtifactName(t *testing.T) {
	tests := []struct {
		name     string
		opts     *UploadOptions
		expected string
	}{
		{
			name: "Custom name",
			opts: &UploadOptions{
				Name: "custom-artifact",
			},
			expected: "custom-artifact",
		},
		{
			name: "PR artifact",
			opts: &UploadOptions{
				PRNumber:  "123",
				Branch:    "feature",
				CommitSHA: "abc123def456",
			},
			expected: "coverage-history-pr-123",
		},
		{
			name: "Main branch with SHA",
			opts: &UploadOptions{
				Branch:    "main",
				CommitSHA: "abc123def456",
			},
			expected: "coverage-history-main-abc123d-", // timestamp suffix will be added
		},
		{
			name: "Feature branch with SHA",
			opts: &UploadOptions{
				Branch:    "feature-xyz",
				CommitSHA: "abc123def456",
			},
			expected: "coverage-history-feature-xyz-abc123d-", // timestamp suffix will be added
		},
		{
			name: "Branch without SHA",
			opts: &UploadOptions{
				Branch: "develop",
			},
			expected: "coverage-history-develop-", // timestamp suffix will be added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateArtifactName(tt.opts)

			if tt.opts.Name != "" {
				// Exact match for custom names
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			} else if tt.opts.PRNumber != "" {
				// Exact match for PR artifacts
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			} else {
				// Partial match for timestamp-suffixed names
				if !containsPrefix(result, tt.expected) {
					t.Errorf("Expected result to start with %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestMergeHistoryLogic(t *testing.T) {
	// Create a mock manager for testing merge logic
	manager := &Manager{
		maxSize: 10 * 1024 * 1024, // 10MB
	}

	t.Run("MergeNilHistories", func(t *testing.T) {
		result, err := manager.MergeHistory(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != nil {
			t.Error("Expected nil result when both histories are nil")
		}
	})

	t.Run("MergeWithNilCurrent", func(t *testing.T) {
		previous := &History{
			Records: []history.CoverageRecord{
				{
					Timestamp:  time.Now().Add(-time.Hour),
					CommitSHA:  "abc123",
					Branch:     "main",
					Percentage: 85.0,
				},
			},
		}

		result, err := manager.MergeHistory(nil, previous)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != previous {
			t.Error("Expected result to be the previous history")
		}
	})

	t.Run("MergeWithNilPrevious", func(t *testing.T) {
		current := &History{
			Records: []history.CoverageRecord{
				{
					Timestamp:  time.Now(),
					CommitSHA:  "def456",
					Branch:     "main",
					Percentage: 90.0,
				},
			},
			Metadata: &HistoryMetadata{
				CreatedAt: time.Now(),
			},
		}

		result, err := manager.MergeHistory(current, nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != current {
			t.Error("Expected result to be the current history")
		}
	})

	t.Run("MergeTwoHistories", func(t *testing.T) {
		now := time.Now()

		previous := &History{
			Records: []history.CoverageRecord{
				{
					Timestamp:  now.Add(-2 * time.Hour),
					CommitSHA:  "abc123",
					Branch:     "main",
					Percentage: 80.0,
				},
				{
					Timestamp:  now.Add(-time.Hour),
					CommitSHA:  "def456",
					Branch:     "main",
					Percentage: 85.0,
				},
			},
		}

		current := &History{
			Records: []history.CoverageRecord{
				{
					Timestamp:  now,
					CommitSHA:  "ghi789",
					Branch:     "main",
					Percentage: 90.0,
				},
			},
			Metadata: &HistoryMetadata{
				CreatedAt: now,
			},
		}

		result, err := manager.MergeHistory(current, previous)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Records) != 3 {
			t.Errorf("Expected 3 records, got %d", len(result.Records))
		}

		// Check that records are sorted by timestamp (oldest first)
		if len(result.Records) >= 2 {
			for i := 1; i < len(result.Records); i++ {
				if result.Records[i].Timestamp.Before(result.Records[i-1].Timestamp) {
					t.Error("Records are not sorted by timestamp")
					break
				}
			}
		}

		// Check metadata
		if result.Metadata.RecordCount != 3 {
			t.Errorf("Expected record count 3, got %d", result.Metadata.RecordCount)
		}
	})

	t.Run("MergeWithDuplicates", func(t *testing.T) {
		now := time.Now()

		previous := &History{
			Records: []history.CoverageRecord{
				{
					Timestamp:  now,
					CommitSHA:  "abc123",
					Branch:     "main",
					Percentage: 80.0,
				},
			},
		}

		current := &History{
			Records: []history.CoverageRecord{
				{
					Timestamp:  now,
					CommitSHA:  "abc123", // Same commit SHA and timestamp
					Branch:     "main",
					Percentage: 85.0, // Different percentage (updated)
				},
			},
			Metadata: &HistoryMetadata{
				CreatedAt: now,
			},
		}

		result, err := manager.MergeHistory(current, previous)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Records) != 1 {
			t.Errorf("Expected 1 record after deduplication, got %d", len(result.Records))
		}

		// The current record should overwrite the previous one
		if result.Records[0].Percentage != 85.0 {
			t.Errorf("Expected percentage 85.0, got %f", result.Records[0].Percentage)
		}
	})
}

func TestCreateEmptyHistory(t *testing.T) {
	manager := &Manager{}

	history := manager.createEmptyHistory()

	if history == nil {
		t.Fatal("Expected non-nil history")
	}

	if len(history.Records) != 0 {
		t.Errorf("Expected empty records, got %d", len(history.Records))
	}

	if history.Metadata == nil {
		t.Fatal("Expected non-nil metadata")
	}

	if history.Metadata.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", history.Metadata.Version)
	}

	if history.Metadata.RecordCount != 0 {
		t.Errorf("Expected record count 0, got %d", history.Metadata.RecordCount)
	}
}

// Helper function to check if a string contains a prefix
func containsPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
