package templates

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test NewTemplateManager function
func TestNewTemplateManager(t *testing.T) {
	t.Run("ValidCreation", func(t *testing.T) {
		tm, err := NewTemplateManager()
		require.NoError(t, err)
		require.NotNil(t, tm)
		require.NotNil(t, tm.templates)
		require.NotNil(t, tm.funcs)

		// Check that all expected functions are registered
		expectedFuncs := []string{
			"formatFloat", "formatPercentage", "formatTime",
			"colorForCoverage", "badgeColor", "add", "sub",
			"mul", "div", "githubRepoURL", "githubUserURL",
			"githubBranchURL", "githubCommitURL", "githubFileURL",
			"githubDirURL",
		}
		for _, funcName := range expectedFuncs {
			_, exists := tm.funcs[funcName]
			assert.True(t, exists, "Function %s should be registered", funcName)
		}
	})
}

// Test deprecated RenderDashboard function
func TestRenderDashboard(t *testing.T) {
	tm, err := NewTemplateManager()
	require.NoError(t, err)

	ctx := context.Background()
	data := DashboardData{
		ProjectName:   "test-project",
		TotalCoverage: 85.5,
	}

	result, err := tm.RenderDashboard(ctx, data)
	require.Error(t, err)
	require.Equal(t, ErrDashboardDeprecated, err)
	require.Empty(t, result)
}

// Test deprecated RenderReport function
func TestRenderReport(t *testing.T) {
	tm, err := NewTemplateManager()
	require.NoError(t, err)

	ctx := context.Background()
	data := ReportData{
		Title:           "Test Report",
		ProjectName:     "test-project",
		OverallCoverage: 85.5,
	}

	result, err := tm.RenderReport(ctx, data)
	require.Error(t, err)
	require.Equal(t, ErrReportDeprecated, err)
	require.Empty(t, result)
}

// Test deprecated WriteDashboard function
func TestWriteDashboard(t *testing.T) {
	tm, err := NewTemplateManager()
	require.NoError(t, err)

	ctx := context.Background()
	var buf strings.Builder
	data := DashboardData{ProjectName: "test"}

	err = tm.WriteDashboard(ctx, &buf, data)
	require.Error(t, err)
	require.Equal(t, ErrDashboardDeprecated, err)
}

// Test deprecated WriteReport function
func TestWriteReport(t *testing.T) {
	tm, err := NewTemplateManager()
	require.NoError(t, err)

	ctx := context.Background()
	var buf strings.Builder
	data := ReportData{Title: "test"}

	err = tm.WriteReport(ctx, &buf, data)
	require.Error(t, err)
	require.Equal(t, ErrReportDeprecated, err)
}

// Test deprecated GetEmbeddedFile function
func TestGetEmbeddedFile(t *testing.T) {
	tm, err := NewTemplateManager()
	require.NoError(t, err)

	result, err := tm.GetEmbeddedFile("test.txt")
	require.Error(t, err)
	require.Equal(t, ErrEmbeddedFilesMoved, err)
	require.Nil(t, result)
}

// Test deprecated ListEmbeddedFiles function
func TestListEmbeddedFiles(t *testing.T) {
	tm, err := NewTemplateManager()
	require.NoError(t, err)

	result, err := tm.ListEmbeddedFiles()
	require.Error(t, err)
	require.Equal(t, ErrEmbeddedFilesMoved, err)
	require.Nil(t, result)
}

// Test formatFloat helper function
func TestFormatFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"PositiveFloat", 85.567, "85.6"},
		{"NegativeFloat", -12.34, "-12.3"},
		{"Zero", 0.0, "0.0"},
		{"LargeNumber", 1234.789, "1234.8"},
		{"SmallDecimal", 0.123, "0.1"},
		{"ExactDecimal", 85.5, "85.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFloat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test formatPercentage helper function
func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"PositivePercentage", 85.567, "85.6%"},
		{"NegativePercentage", -12.34, "-12.3%"},
		{"Zero", 0.0, "0.0%"},
		{"FullCoverage", 100.0, "100.0%"},
		{"SmallPercentage", 0.1, "0.1%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPercentage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test formatTime helper function
func TestFormatTime(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC)
	expected := "2023-12-25 14:30:45 UTC"
	result := formatTime(testTime)
	assert.Equal(t, expected, result)
}

// Test colorForCoverage helper function
func TestColorForCoverage(t *testing.T) {
	tests := []struct {
		name     string
		coverage float64
		expected string
	}{
		{"Excellent_90", 90.0, "#3fb950"},
		{"Excellent_95", 95.0, "#3fb950"},
		{"Good_80", 80.0, "#90c978"},
		{"Good_85", 85.0, "#90c978"},
		{"Acceptable_70", 70.0, "#d29922"},
		{"Acceptable_75", 75.0, "#d29922"},
		{"Low_60", 60.0, "#f85149"},
		{"Low_65", 65.0, "#f85149"},
		{"Poor_50", 50.0, "#da3633"},
		{"Poor_0", 0.0, "#da3633"},
		{"Poor_59", 59.9, "#da3633"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorForCoverage(tt.coverage)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test badgeColor helper function
func TestBadgeColor(t *testing.T) {
	// badgeColor is just a wrapper for colorForCoverage
	result := badgeColor(85.0)
	expected := colorForCoverage(85.0)
	assert.Equal(t, expected, result)
}

// Test math helper functions
func TestMathHelpers(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		tests := []struct {
			a, b     int
			expected int
		}{
			{5, 3, 8},
			{0, 0, 0},
			{-5, 3, -2},
			{100, 200, 300},
		}

		for _, tt := range tests {
			result := add(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("Sub", func(t *testing.T) {
		tests := []struct {
			a, b     int
			expected int
		}{
			{5, 3, 2},
			{0, 0, 0},
			{3, 5, -2},
			{100, 50, 50},
		}

		for _, tt := range tests {
			result := sub(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("Mul", func(t *testing.T) {
		tests := []struct {
			a, b     float64
			expected float64
		}{
			{5.0, 3.0, 15.0},
			{0.0, 10.0, 0.0},
			{-2.0, 4.0, -8.0},
			{0.5, 2.0, 1.0},
		}

		for _, tt := range tests {
			result := mul(tt.a, tt.b)
			if tt.expected == 0.0 {
				assert.Equal(t, tt.expected, result) //nolint:testifylint // comparing with zero
			} else {
				assert.InEpsilon(t, tt.expected, result, 0.001)
			}
		}
	})

	t.Run("Div", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b     int
			expected float64
		}{
			{"NormalDivision", 10, 2, 5.0},
			{"DivideByZero", 10, 0, 0.0},
			{"ZeroDividend", 0, 5, 0.0},
			{"IntegerResult", 15, 3, 5.0},
			{"DecimalResult", 10, 3, 3.333333333333333},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := div(tt.a, tt.b)
				if tt.expected == 0.0 {
					assert.Equal(t, tt.expected, result) //nolint:testifylint // comparing with zero
				} else {
					assert.InEpsilon(t, tt.expected, result, 0.001)
				}
			})
		}
	})
}

// Test GitHub URL building functions
func TestGitHubURLFunctions(t *testing.T) {
	t.Run("GithubRepoURL", func(t *testing.T) {
		tests := []struct {
			name     string
			owner    string
			repo     string
			expected string
		}{
			{"ValidRepo", "owner", "repo", "https://github.com/owner/repo"},
			{"EmptyOwner", "", "repo", ""},
			{"EmptyRepo", "owner", "", ""},
			{"BothEmpty", "", "", ""},
			{"WithDashes", "my-org", "my-repo", "https://github.com/my-org/my-repo"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := githubRepoURL(tt.owner, tt.repo)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("GithubUserURL", func(t *testing.T) {
		tests := []struct {
			name     string
			username string
			expected string
		}{
			{"ValidUser", "testuser", "https://github.com/testuser"},
			{"EmptyUser", "", ""},
			{"UserWithDashes", "test-user", "https://github.com/test-user"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := githubUserURL(tt.username)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("GithubBranchURL", func(t *testing.T) {
		tests := []struct {
			name     string
			owner    string
			repo     string
			branch   string
			expected string
		}{
			{"ValidBranch", "owner", "repo", "main", "https://github.com/owner/repo/tree/main"},
			{"FeatureBranch", "owner", "repo", "feature/test", "https://github.com/owner/repo/tree/feature/test"},
			{"EmptyOwner", "", "repo", "main", ""},
			{"EmptyRepo", "owner", "", "main", ""},
			{"EmptyBranch", "owner", "repo", "", ""},
			{"AllEmpty", "", "", "", ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := githubBranchURL(tt.owner, tt.repo, tt.branch)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("GithubCommitURL", func(t *testing.T) {
		tests := []struct {
			name     string
			owner    string
			repo     string
			sha      string
			expected string
		}{
			{"ValidCommit", "owner", "repo", "abc123", "https://github.com/owner/repo/commit/abc123"},
			{"LongSHA", "owner", "repo", "abc123def456789", "https://github.com/owner/repo/commit/abc123def456789"},
			{"EmptyOwner", "", "repo", "abc123", ""},
			{"EmptyRepo", "owner", "", "abc123", ""},
			{"EmptySHA", "owner", "repo", "", ""},
			{"AllEmpty", "", "", "", ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := githubCommitURL(tt.owner, tt.repo, tt.sha)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("GithubFileURL", func(t *testing.T) {
		tests := []struct {
			name     string
			owner    string
			repo     string
			branch   string
			filepath string
			expected string
		}{
			{"ValidFile", "owner", "repo", "main", "file.go", "https://github.com/owner/repo/blob/main/file.go"},
			{"NestedFile", "owner", "repo", "main", "pkg/file.go", "https://github.com/owner/repo/blob/main/pkg/file.go"},
			{"EmptyOwner", "", "repo", "main", "file.go", ""},
			{"EmptyRepo", "owner", "", "main", "file.go", ""},
			{"EmptyBranch", "owner", "repo", "", "file.go", ""},
			{"EmptyFilepath", "owner", "repo", "main", "", ""},
			{"AllEmpty", "", "", "", "", ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := githubFileURL(tt.owner, tt.repo, tt.branch, tt.filepath)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("GithubDirURL", func(t *testing.T) {
		tests := []struct {
			name     string
			owner    string
			repo     string
			branch   string
			dirpath  string
			expected string
		}{
			{"ValidDir", "owner", "repo", "main", "pkg", "https://github.com/owner/repo/tree/main/pkg"},
			{"NestedDir", "owner", "repo", "main", "internal/pkg", "https://github.com/owner/repo/tree/main/internal/pkg"},
			{"EmptyOwner", "", "repo", "main", "pkg", ""},
			{"EmptyRepo", "owner", "", "main", "pkg", ""},
			{"EmptyBranch", "owner", "repo", "", "pkg", ""},
			{"EmptyDirpath", "owner", "repo", "main", "", ""},
			{"AllEmpty", "", "", "", "", ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := githubDirURL(tt.owner, tt.repo, tt.branch, tt.dirpath)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// Test error constants
func TestErrorConstants(t *testing.T) {
	assert.Contains(t, ErrDashboardDeprecated.Error(), "deprecated")
	assert.Contains(t, ErrReportDeprecated.Error(), "deprecated")
	assert.Contains(t, ErrEmbeddedFilesMoved.Error(), "moved")
}

// Test data structures
func TestDataStructures(t *testing.T) {
	t.Run("DashboardData", func(t *testing.T) {
		data := DashboardData{
			ProjectName:   "test-project",
			TotalCoverage: 85.5,
			LastUpdated:   time.Now(),
		}
		assert.Equal(t, "test-project", data.ProjectName)
		assert.InEpsilon(t, 85.5, data.TotalCoverage, 0.001)
	})

	t.Run("ReportData", func(t *testing.T) {
		data := ReportData{
			Title:           "Test Report",
			ProjectName:     "test-project",
			OverallCoverage: 90.0,
			Generated:       time.Now(),
		}
		assert.Equal(t, "Test Report", data.Title)
		assert.Equal(t, "test-project", data.ProjectName)
		assert.InEpsilon(t, 90.0, data.OverallCoverage, 0.01)
	})

	t.Run("BranchData", func(t *testing.T) {
		data := BranchData{
			Name:         "main",
			Coverage:     85.5,
			CoveredLines: 855,
			TotalLines:   1000,
			Protected:    true,
			LastCommit:   time.Now(),
			Trend:        2.5,
			GitHubURL:    "https://github.com/owner/repo/tree/main",
		}
		assert.Equal(t, "main", data.Name)
		assert.InEpsilon(t, 85.5, data.Coverage, 0.01)
		assert.True(t, data.Protected)
	})
}
