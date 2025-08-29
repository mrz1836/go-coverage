package analysis

import (
	"context"
	"testing"
	"time"

	"github.com/mrz1836/go-coverage/internal/artifacts"
	"github.com/mrz1836/go-coverage/internal/history"
	"github.com/mrz1836/go-coverage/internal/parser"
)

// MockArtifactManager implements artifacts.ArtifactManager for testing
type MockArtifactManager struct {
	mockHistory *artifacts.History
	mockError   error
}

func (m *MockArtifactManager) DownloadHistory(ctx context.Context, opts *artifacts.DownloadOptions) (*artifacts.History, error) {
	if m.mockError != nil {
		return nil, m.mockError
	}
	return m.mockHistory, nil
}

func (m *MockArtifactManager) MergeHistory(current, previous *artifacts.History) (*artifacts.History, error) {
	return current, nil
}

func (m *MockArtifactManager) UploadHistory(ctx context.Context, history *artifacts.History, opts *artifacts.UploadOptions) error {
	return nil
}

func (m *MockArtifactManager) CleanupOldArtifacts(ctx context.Context, retentionDays int) error {
	return nil
}

func (m *MockArtifactManager) ListArtifacts(ctx context.Context, opts *artifacts.ListOptions) ([]*artifacts.ArtifactInfo, error) {
	return nil, nil
}

func TestCoverageDiffer_CalculateDiff(t *testing.T) {
	tests := []struct {
		name            string
		currentCoverage *parser.CoverageData
		baseCoverage    *artifacts.History
		expectedDiff    float64
		expectError     bool
	}{
		{
			name: "coverage improvement",
			currentCoverage: &parser.CoverageData{
				Percentage:   85.5,
				TotalLines:   1000,
				CoveredLines: 855,
				Timestamp:    time.Now(),
			},
			baseCoverage: &artifacts.History{
				Records: []history.CoverageRecord{
					{
						Percentage:   80.0,
						TotalLines:   1000,
						CoveredLines: 800,
						CommitSHA:    "abc123",
						Branch:       "main",
						Timestamp:    time.Now().Add(-1 * time.Hour),
					},
				},
			},
			expectedDiff: 5.5,
			expectError:  false,
		},
		{
			name: "coverage decline",
			currentCoverage: &parser.CoverageData{
				Percentage:   75.0,
				TotalLines:   1000,
				CoveredLines: 750,
				Timestamp:    time.Now(),
			},
			baseCoverage: &artifacts.History{
				Records: []history.CoverageRecord{
					{
						Percentage:   80.0,
						TotalLines:   1000,
						CoveredLines: 800,
						CommitSHA:    "abc123",
						Branch:       "main",
						Timestamp:    time.Now().Add(-1 * time.Hour),
					},
				},
			},
			expectedDiff: -5.0,
			expectError:  false,
		},
		{
			name: "no change",
			currentCoverage: &parser.CoverageData{
				Percentage:   80.0,
				TotalLines:   1000,
				CoveredLines: 800,
				Timestamp:    time.Now(),
			},
			baseCoverage: &artifacts.History{
				Records: []history.CoverageRecord{
					{
						Percentage:   80.0,
						TotalLines:   1000,
						CoveredLines: 800,
						CommitSHA:    "abc123",
						Branch:       "main",
						Timestamp:    time.Now().Add(-1 * time.Hour),
					},
				},
			},
			expectedDiff: 0.0,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := &MockArtifactManager{
				mockHistory: tt.baseCoverage,
				mockError:   nil,
			}

			differ := NewCoverageDiffer(mockManager)
			comparison, err := differ.CalculateDiff(
				context.Background(),
				tt.currentCoverage,
				"main",
				"feature-branch",
				"123",
			)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if err == nil && comparison.Difference != tt.expectedDiff {
				t.Errorf("expected diff %f, got %f", tt.expectedDiff, comparison.Difference)
			}
		})
	}
}

func TestCoverageDiffer_GenerateTrendAnalysis(t *testing.T) {
	differ := NewCoverageDiffer(&MockArtifactManager{})

	tests := []struct {
		name              string
		difference        float64
		baseCoverage      float64
		currentCoverage   float64
		expectedDirection string
		expectedMagnitude string
	}{
		{
			name:              "significant improvement",
			difference:        6.0,
			baseCoverage:      70.0,
			currentCoverage:   76.0,
			expectedDirection: "up",
			expectedMagnitude: "significant",
		},
		{
			name:              "moderate decline",
			difference:        -2.5,
			baseCoverage:      80.0,
			currentCoverage:   77.5,
			expectedDirection: "down",
			expectedMagnitude: "moderate",
		},
		{
			name:              "minor change",
			difference:        0.5,
			baseCoverage:      80.0,
			currentCoverage:   80.5,
			expectedDirection: "up",
			expectedMagnitude: "minor",
		},
		{
			name:              "stable",
			difference:        0.05,
			baseCoverage:      80.0,
			currentCoverage:   80.05,
			expectedDirection: "stable",
			expectedMagnitude: "minor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trend := differ.generateTrendAnalysis(tt.difference, tt.baseCoverage)

			if trend.Direction != tt.expectedDirection {
				t.Errorf("expected direction %s, got %s", tt.expectedDirection, trend.Direction)
			}
			if trend.Magnitude != tt.expectedMagnitude {
				t.Errorf("expected magnitude %s, got %s", tt.expectedMagnitude, trend.Magnitude)
			}
			if trend.PercentageChange != tt.difference {
				t.Errorf("expected percentage change %f, got %f", tt.difference, trend.PercentageChange)
			}
		})
	}
}

func TestCoverageDiffer_IdentifySignificantFiles(t *testing.T) {
	differ := NewCoverageDiffer(&MockArtifactManager{})

	fileChanges := []FileChange{
		{
			Filename:      "main.go",
			BaseCoverage:  80.0,
			PRCoverage:    85.0,
			Difference:    5.0,
			IsSignificant: true,
		},
		{
			Filename:      "utils.go",
			BaseCoverage:  90.0,
			PRCoverage:    89.0,
			Difference:    -1.0,
			IsSignificant: false,
		},
		{
			Filename:      "handler.go",
			BaseCoverage:  70.0,
			PRCoverage:    75.0,
			Difference:    5.0,
			IsSignificant: true,
		},
	}

	significantFiles := differ.identifySignificantFiles(fileChanges)

	expectedCount := 2
	if len(significantFiles) != expectedCount {
		t.Errorf("expected %d significant files, got %d", expectedCount, len(significantFiles))
	}

	expectedFiles := map[string]bool{
		"main.go":    true,
		"handler.go": true,
	}

	for _, file := range significantFiles {
		if !expectedFiles[file] {
			t.Errorf("unexpected significant file: %s", file)
		}
	}
}

func TestCoverageDiffer_NoBaseCoverageAvailable(t *testing.T) {
	// Test behavior when no base coverage is available
	mockManager := &MockArtifactManager{
		mockError: context.DeadlineExceeded, // Simulate download failure
	}

	differ := NewCoverageDiffer(mockManager)

	currentCoverage := &parser.CoverageData{
		Percentage:   85.0,
		TotalLines:   1000,
		CoveredLines: 850,
		Timestamp:    time.Now(),
	}

	comparison, err := differ.CalculateDiff(
		context.Background(),
		currentCoverage,
		"main",
		"feature-branch",
		"123",
	)
	// Should not error when base coverage is unavailable
	if err != nil {
		t.Errorf("unexpected error when base coverage unavailable: %v", err)
	}

	// Should use default base coverage (0%)
	if comparison.BaseCoverage.Percentage != 0.0 {
		t.Errorf("expected base coverage 0.0, got %f", comparison.BaseCoverage.Percentage)
	}

	// Difference should equal current coverage
	if comparison.Difference != 85.0 {
		t.Errorf("expected difference 85.0, got %f", comparison.Difference)
	}
}
