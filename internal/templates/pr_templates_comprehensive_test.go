package templates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFormatCommitSHA(t *testing.T) {
	engine := NewPRTemplateEngine(nil)

	tests := []struct {
		name     string
		sha      string
		expected string
	}{
		{
			name:     "long SHA",
			sha:      "abc123def456789",
			expected: "abc123d",
		},
		{
			name:     "short SHA",
			sha:      "abc123",
			expected: "abc123",
		},
		{
			name:     "empty SHA",
			sha:      "",
			expected: "",
		},
		{
			name:     "exactly 7 chars",
			sha:      "abc1234",
			expected: "abc1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.formatCommitSHA(tt.sha)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTimestamp(t *testing.T) {
	engine := NewPRTemplateEngine(&TemplateConfig{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	testTime := time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC)
	result := engine.formatTimestamp(testTime)
	require.Equal(t, "2023-12-25 14:30:45", result)
}

func TestStatusEmoji(t *testing.T) {
	tests := []struct {
		name          string
		includeEmojis bool
		status        string
		expected      string
	}{
		{
			name:          "emojis disabled",
			includeEmojis: false,
			status:        "excellent",
			expected:      "",
		},
		{
			name:          "excellent status",
			includeEmojis: true,
			status:        "excellent",
			expected:      "🟢",
		},
		{
			name:          "good status",
			includeEmojis: true,
			status:        "good",
			expected:      "🟡",
		},
		{
			name:          "warning status",
			includeEmojis: true,
			status:        "warning",
			expected:      "🟠",
		},
		{
			name:          "critical status",
			includeEmojis: true,
			status:        "critical",
			expected:      "🔴",
		},
		{
			name:          "default status",
			includeEmojis: true,
			status:        "unknown",
			expected:      "⚪",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPRTemplateEngine(&TemplateConfig{
				IncludeEmojis: tt.includeEmojis,
			})
			result := engine.statusEmoji(tt.status)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTrendEmoji(t *testing.T) {
	tests := []struct {
		name          string
		includeEmojis bool
		direction     string
		expected      string
	}{
		{
			name:          "emojis disabled",
			includeEmojis: false,
			direction:     "improved",
			expected:      "",
		},
		{
			name:          "improved direction",
			includeEmojis: true,
			direction:     "improved",
			expected:      "📈",
		},
		{
			name:          "up direction",
			includeEmojis: true,
			direction:     "up",
			expected:      "📈",
		},
		{
			name:          "upward direction",
			includeEmojis: true,
			direction:     "upward",
			expected:      "📈",
		},
		{
			name:          "degraded direction",
			includeEmojis: true,
			direction:     "degraded",
			expected:      "📉",
		},
		{
			name:          "down direction",
			includeEmojis: true,
			direction:     "down",
			expected:      "📉",
		},
		{
			name:          "downward direction",
			includeEmojis: true,
			direction:     "downward",
			expected:      "📉",
		},
		{
			name:          "stable direction",
			includeEmojis: true,
			direction:     "stable",
			expected:      "📊",
		},
		{
			name:          "volatile direction",
			includeEmojis: true,
			direction:     "volatile",
			expected:      "📊",
		},
		{
			name:          "unknown direction",
			includeEmojis: true,
			direction:     "unknown",
			expected:      "📊",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPRTemplateEngine(&TemplateConfig{
				IncludeEmojis: tt.includeEmojis,
			})
			result := engine.trendEmoji(tt.direction)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestRiskEmoji(t *testing.T) {
	tests := []struct {
		name          string
		includeEmojis bool
		risk          string
		expected      string
	}{
		{
			name:          "emojis disabled",
			includeEmojis: false,
			risk:          "high",
			expected:      "",
		},
		{
			name:          "high risk",
			includeEmojis: true,
			risk:          "high",
			expected:      "🚨",
		},
		{
			name:          "critical risk",
			includeEmojis: true,
			risk:          "critical",
			expected:      "🚨",
		},
		{
			name:          "medium risk",
			includeEmojis: true,
			risk:          "medium",
			expected:      "⚠️",
		},
		{
			name:          "low risk",
			includeEmojis: true,
			risk:          "low",
			expected:      "✅",
		},
		{
			name:          "unknown risk",
			includeEmojis: true,
			risk:          "unknown",
			expected:      "ℹ️",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPRTemplateEngine(&TemplateConfig{
				IncludeEmojis: tt.includeEmojis,
			})
			result := engine.riskEmoji(tt.risk)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGradeEmoji(t *testing.T) {
	tests := []struct {
		name          string
		includeEmojis bool
		grade         string
		expected      string
	}{
		{
			name:          "emojis disabled",
			includeEmojis: false,
			grade:         "A+",
			expected:      "",
		},
		{
			name:          "A+ grade",
			includeEmojis: true,
			grade:         "A+",
			expected:      "🏆",
		},
		{
			name:          "A grade",
			includeEmojis: true,
			grade:         "A",
			expected:      "🥇",
		},
		{
			name:          "B+ grade",
			includeEmojis: true,
			grade:         "B+",
			expected:      "🥈",
		},
		{
			name:          "B grade",
			includeEmojis: true,
			grade:         "B",
			expected:      "🥈",
		},
		{
			name:          "C grade",
			includeEmojis: true,
			grade:         "C",
			expected:      "🥉",
		},
		{
			name:          "D grade",
			includeEmojis: true,
			grade:         "D",
			expected:      "⚠️",
		},
		{
			name:          "F grade",
			includeEmojis: true,
			grade:         "F",
			expected:      "🚨",
		},
		{
			name:          "unknown grade",
			includeEmojis: true,
			grade:         "X",
			expected:      "📊",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPRTemplateEngine(&TemplateConfig{
				IncludeEmojis: tt.includeEmojis,
			})
			result := engine.gradeEmoji(tt.grade)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestPriorityEmoji(t *testing.T) {
	tests := []struct {
		name          string
		includeEmojis bool
		priority      string
		expected      string
	}{
		{
			name:          "emojis disabled",
			includeEmojis: false,
			priority:      "high",
			expected:      "",
		},
		{
			name:          "high priority",
			includeEmojis: true,
			priority:      "high",
			expected:      "🔥",
		},
		{
			name:          "medium priority",
			includeEmojis: true,
			priority:      "medium",
			expected:      "📌",
		},
		{
			name:          "low priority",
			includeEmojis: true,
			priority:      "low",
			expected:      "💡",
		},
		{
			name:          "unknown priority",
			includeEmojis: true,
			priority:      "unknown",
			expected:      "ℹ️",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPRTemplateEngine(&TemplateConfig{
				IncludeEmojis: tt.includeEmojis,
			})
			result := engine.priorityEmoji(tt.priority)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestProgressBarComprehensive(t *testing.T) {
	tests := []struct {
		name                string
		includeProgressBars bool
		value               float64
		maxValue            float64
		width               int
		expectedContains    []string
	}{
		{
			name:                "progress bars disabled",
			includeProgressBars: false,
			value:               50,
			maxValue:            100,
			width:               10,
			expectedContains:    []string{""},
		},
		{
			name:                "full progress",
			includeProgressBars: true,
			value:               100,
			maxValue:            100,
			width:               10,
			expectedContains:    []string{"██████████", "100.0%"},
		},
		{
			name:                "half progress",
			includeProgressBars: true,
			value:               50,
			maxValue:            100,
			width:               10,
			expectedContains:    []string{"█████░░░░░", "50.0%"},
		},
		{
			name:                "zero progress",
			includeProgressBars: true,
			value:               0,
			maxValue:            100,
			width:               10,
			expectedContains:    []string{"░░░░░░░░░░", "0.0%"},
		},
		{
			name:                "over max value",
			includeProgressBars: true,
			value:               150,
			maxValue:            100,
			width:               10,
			expectedContains:    []string{"██████████", "150.0%"},
		},
		{
			name:                "negative value",
			includeProgressBars: true,
			value:               -10,
			maxValue:            100,
			width:               10,
			expectedContains:    []string{"░░░░░░░░░░", "-10.0%"},
		},
		{
			name:                "zero width defaults to 20",
			includeProgressBars: true,
			value:               50,
			maxValue:            100,
			width:               0,
			expectedContains:    []string{"50.0%"},
		},
		{
			name:                "negative width defaults to 20",
			includeProgressBars: true,
			value:               50,
			maxValue:            100,
			width:               -5,
			expectedContains:    []string{"50.0%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPRTemplateEngine(&TemplateConfig{
				IncludeProgressBars: tt.includeProgressBars,
			})
			result := engine.progressBar(tt.value, tt.maxValue, tt.width)

			if !tt.includeProgressBars {
				require.Empty(t, result)
				return
			}

			for _, expected := range tt.expectedContains {
				require.Contains(t, result, expected)
			}
		})
	}
}

func TestTrendChart(t *testing.T) {
	tests := []struct {
		name          string
		includeCharts bool
		value         interface{}
		expected      string
	}{
		{
			name:          "charts disabled",
			includeCharts: false,
			value:         95.0,
			expected:      "",
		},
		{
			name:          "single value high",
			includeCharts: true,
			value:         95.0,
			expected:      "📈",
		},
		{
			name:          "single value medium",
			includeCharts: true,
			value:         75.0,
			expected:      "📊",
		},
		{
			name:          "single value low",
			includeCharts: true,
			value:         65.0,
			expected:      "📉",
		},
		{
			name:          "empty slice",
			includeCharts: true,
			value:         []float64{},
			expected:      "",
		},
		{
			name:          "invalid type",
			includeCharts: true,
			value:         "invalid",
			expected:      "",
		},
		{
			name:          "slice of values",
			includeCharts: true,
			value:         []float64{80.0, 85.0, 90.0},
			expected:      "", // Should contain ASCII chart, but we just check it's not empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPRTemplateEngine(&TemplateConfig{
				IncludeCharts: tt.includeCharts,
			})
			result := engine.trendChart(tt.value)

			if tt.name == "slice of values" && tt.includeCharts {
				require.NotEmpty(t, result)
			} else {
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFilterFiles(t *testing.T) {
	tests := []struct {
		name            string
		hideStableFiles bool
		maxFileChanges  int
		files           []FileCoverageData
		expectedLen     int
		expectedFiles   []string
	}{
		{
			name:            "no filtering",
			hideStableFiles: false,
			maxFileChanges:  10,
			files: []FileCoverageData{
				{Filename: "file1.go", Status: "changed", Change: 5.0},
				{Filename: "file2.go", Status: "stable", Change: 0.5},
				{Filename: "file3.go", Status: "new", Change: 10.0},
			},
			expectedLen:   3,
			expectedFiles: []string{"file1.go", "file2.go", "file3.go"},
		},
		{
			name:            "hide stable files",
			hideStableFiles: true,
			maxFileChanges:  10,
			files: []FileCoverageData{
				{Filename: "file1.go", Status: "changed", Change: 5.0},
				{Filename: "file2.go", Status: "stable", Change: 0.5},
				{Filename: "file3.go", Status: "stable", Change: 2.0},
				{Filename: "file4.go", Status: "new", Change: 10.0},
			},
			expectedLen:   3,
			expectedFiles: []string{"file1.go", "file3.go", "file4.go"},
		},
		{
			name:            "limit max files",
			hideStableFiles: false,
			maxFileChanges:  2,
			files: []FileCoverageData{
				{Filename: "file1.go", Status: "changed", Change: 5.0},
				{Filename: "file2.go", Status: "stable", Change: 0.5},
				{Filename: "file3.go", Status: "new", Change: 10.0},
			},
			expectedLen:   2,
			expectedFiles: []string{"file1.go", "file2.go"},
		},
		{
			name:            "hide stable and limit",
			hideStableFiles: true,
			maxFileChanges:  1,
			files: []FileCoverageData{
				{Filename: "file1.go", Status: "changed", Change: 5.0},
				{Filename: "file2.go", Status: "stable", Change: 0.5},
				{Filename: "file3.go", Status: "new", Change: 10.0},
			},
			expectedLen:   1,
			expectedFiles: []string{"file1.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPRTemplateEngine(&TemplateConfig{
				HideStableFiles: tt.hideStableFiles,
				MaxFileChanges:  tt.maxFileChanges,
			})

			result := engine.filterFiles(tt.files)
			require.Len(t, result, tt.expectedLen)

			for i, expectedFile := range tt.expectedFiles {
				if i < len(result) {
					require.Equal(t, expectedFile, result[i].Filename)
				}
			}
		})
	}
}

func TestFilterPackages(t *testing.T) {
	tests := []struct {
		name              string
		hideStableFiles   bool
		maxPackageChanges int
		packages          []PackageCoverageData
		expectedLen       int
		expectedPackages  []string
	}{
		{
			name:              "no filtering",
			hideStableFiles:   false,
			maxPackageChanges: 10,
			packages: []PackageCoverageData{
				{Package: "pkg1", Status: "changed", Change: 5.0},
				{Package: "pkg2", Status: "stable", Change: 0.5},
				{Package: "pkg3", Status: "new", Change: 10.0},
			},
			expectedLen:      3,
			expectedPackages: []string{"pkg1", "pkg2", "pkg3"},
		},
		{
			name:              "hide stable packages",
			hideStableFiles:   true,
			maxPackageChanges: 10,
			packages: []PackageCoverageData{
				{Package: "pkg1", Status: "changed", Change: 5.0},
				{Package: "pkg2", Status: "stable", Change: 0.5},
				{Package: "pkg3", Status: "stable", Change: 2.0},
				{Package: "pkg4", Status: "new", Change: 10.0},
			},
			expectedLen:      3,
			expectedPackages: []string{"pkg1", "pkg3", "pkg4"},
		},
		{
			name:              "limit max packages",
			hideStableFiles:   false,
			maxPackageChanges: 2,
			packages: []PackageCoverageData{
				{Package: "pkg1", Status: "changed", Change: 5.0},
				{Package: "pkg2", Status: "stable", Change: 0.5},
				{Package: "pkg3", Status: "new", Change: 10.0},
			},
			expectedLen:      2,
			expectedPackages: []string{"pkg1", "pkg2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPRTemplateEngine(&TemplateConfig{
				HideStableFiles:   tt.hideStableFiles,
				MaxPackageChanges: tt.maxPackageChanges,
			})

			result := engine.filterPackages(tt.packages)
			require.Len(t, result, tt.expectedLen)

			for i, expectedPkg := range tt.expectedPackages {
				if i < len(result) {
					require.Equal(t, expectedPkg, result[i].Package)
				}
			}
		})
	}
}

// Test missing PR template functions
func TestFilterRecommendations(t *testing.T) {
	tests := []struct {
		name               string
		maxRecommendations int
		recommendations    []RecommendationData
		expectedLen        int
		expectedPriorities []string
	}{
		{
			name:               "sort by priority",
			maxRecommendations: 10,
			recommendations: []RecommendationData{
				{Title: "Low Priority", Priority: "low"},
				{Title: "High Priority", Priority: "high"},
				{Title: "Medium Priority", Priority: "medium"},
			},
			expectedLen:        3,
			expectedPriorities: []string{"high", "medium", "low"},
		},
		{
			name:               "limit recommendations",
			maxRecommendations: 2,
			recommendations: []RecommendationData{
				{Title: "Low Priority", Priority: "low"},
				{Title: "High Priority", Priority: "high"},
				{Title: "Medium Priority", Priority: "medium"},
			},
			expectedLen:        2,
			expectedPriorities: []string{"high", "medium"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPRTemplateEngine(&TemplateConfig{
				MaxRecommendations: tt.maxRecommendations,
			})

			result := engine.filterRecommendations(tt.recommendations)
			require.Len(t, result, tt.expectedLen)

			for i, expectedPriority := range tt.expectedPriorities {
				if i < len(result) {
					require.Equal(t, expectedPriority, result[i].Priority)
				}
			}
		})
	}
}

func TestSortFilesByRisk(t *testing.T) {
	files := []FileCoverageData{
		{Filename: "low.go", Risk: "low", Change: 1.0},
		{Filename: "critical.go", Risk: "critical", Change: 5.0},
		{Filename: "medium.go", Risk: "medium", Change: 3.0},
		{Filename: "high.go", Risk: "high", Change: 2.0},
		{Filename: "critical2.go", Risk: "critical", Change: 10.0}, // Higher change
	}

	engine := NewPRTemplateEngine(nil)
	result := engine.sortFilesByRisk(files)

	require.Len(t, result, 5)
	// Should be sorted by risk first (critical > high > medium > low), then by change magnitude
	require.Equal(t, "critical2.go", result[0].Filename) // critical with highest change
	require.Equal(t, "critical.go", result[1].Filename)  // critical with lower change
	require.Equal(t, "high.go", result[2].Filename)      // high risk
	require.Equal(t, "medium.go", result[3].Filename)    // medium risk
	require.Equal(t, "low.go", result[4].Filename)       // low risk
}

func TestSortByChange(t *testing.T) {
	files := []FileCoverageData{
		{Filename: "small.go", Change: 1.0},
		{Filename: "large.go", Change: 10.0},
		{Filename: "medium.go", Change: 5.0},
		{Filename: "negative.go", Change: -8.0}, // Should be sorted by absolute value
	}

	engine := NewPRTemplateEngine(nil)
	result := engine.sortByChange(files)

	require.Len(t, result, 4)
	// Should be sorted by absolute change value (descending)
	require.Equal(t, "large.go", result[0].Filename)    // |10.0|
	require.Equal(t, "negative.go", result[1].Filename) // |-8.0|
	require.Equal(t, "medium.go", result[2].Filename)   // |5.0|
	require.Equal(t, "small.go", result[3].Filename)    // |1.0|
}

func TestConditionalLogicFunctions(t *testing.T) {
	engine := NewPRTemplateEngine(nil)

	t.Run("IsSignificant", func(t *testing.T) {
		tests := []struct {
			change   float64
			expected bool
		}{
			{1.5, true},
			{-2.0, true},
			{0.5, false},
			{-0.8, false},
			{1.0, true},
			{-1.0, true},
		}

		for _, tt := range tests {
			result := engine.isSignificant(tt.change)
			require.Equal(t, tt.expected, result, "Change: %f", tt.change)
		}
	})

	t.Run("IsDegraded", func(t *testing.T) {
		tests := []struct {
			direction string
			expected  bool
		}{
			{"degraded", true},
			{"down", true},
			{"downward", true},
			{"improved", false},
			{"up", false},
			{"stable", false},
		}

		for _, tt := range tests {
			result := engine.isDegraded(tt.direction)
			require.Equal(t, tt.expected, result, "Direction: %s", tt.direction)
		}
	})

	t.Run("IsStable", func(t *testing.T) {
		tests := []struct {
			direction string
			expected  bool
		}{
			{"stable", true},
			{"improved", false},
			{"degraded", false},
			{"up", false},
			{"down", false},
		}

		for _, tt := range tests {
			result := engine.isStable(tt.direction)
			require.Equal(t, tt.expected, result, "Direction: %s", tt.direction)
		}
	})

	t.Run("NeedsAttention", func(t *testing.T) {
		engineWithThreshold := NewPRTemplateEngine(&TemplateConfig{
			WarningThreshold: 80.0,
		})

		tests := []struct {
			percentage float64
			expected   bool
		}{
			{85.0, false}, // Above threshold
			{75.0, true},  // Below threshold
			{80.0, false}, // At threshold
			{79.9, true},  // Just below threshold
		}

		for _, tt := range tests {
			result := engineWithThreshold.needsAttention(tt.percentage)
			require.Equal(t, tt.expected, result, "Percentage: %f", tt.percentage)
		}
	})
}

func TestTextUtilityFunctions(t *testing.T) {
	engine := NewPRTemplateEngine(nil)

	t.Run("Truncate", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			length   int
			expected string
		}{
			{"ShortString", "hello", 10, "hello"},
			{"ExactLength", "hello", 5, "hello"},
			{"LongString", "hello world", 8, "hello..."},
			{"VeryShort", "hello world", 3, "..."},
			{"EmptyString", "", 5, ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := engine.truncate(tt.input, tt.length)
				require.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("Pluralize", func(t *testing.T) {
		tests := []struct {
			count    int
			singular string
			plural   string
			expected string
		}{
			{0, "file", "files", "files"},
			{1, "file", "files", "file"},
			{2, "file", "files", "files"},
			{5, "test", "tests", "tests"},
		}

		for _, tt := range tests {
			result := engine.pluralize(tt.count, tt.singular, tt.plural)
			require.Equal(t, tt.expected, result, "Count: %d", tt.count)
		}
	})

	t.Run("Capitalize", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"hello", "Hello"},
			{"HELLO", "HELLO"},
			{"h", "H"},
			{"", ""},
			{"hello world", "Hello world"},
		}

		for _, tt := range tests {
			result := engine.capitalize(tt.input)
			require.Equal(t, tt.expected, result, "Input: %s", tt.input)
		}
	})
}

func TestSliceFunction(t *testing.T) {
	engine := NewPRTemplateEngine(nil)

	t.Run("FileCoverageData", func(t *testing.T) {
		files := []FileCoverageData{
			{Filename: "file1.go"},
			{Filename: "file2.go"},
			{Filename: "file3.go"},
			{Filename: "file4.go"},
		}

		tests := []struct {
			name     string
			start    int
			end      int
			expected []string
		}{
			{"Normal", 1, 3, []string{"file2.go", "file3.go"}},
			{"StartFromBeginning", 0, 2, []string{"file1.go", "file2.go"}},
			{"EndBeyondLength", 2, 10, []string{"file3.go", "file4.go"}},
			{"NegativeStart", -1, 2, []string{"file1.go", "file2.go"}},
			{"StartAfterEnd", 3, 1, []string{}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := engine.slice(files, tt.start, tt.end)
				slicedFiles, ok := result.([]FileCoverageData)
				require.True(t, ok)

				require.Len(t, slicedFiles, len(tt.expected))
				for i, expectedFile := range tt.expected {
					require.Equal(t, expectedFile, slicedFiles[i].Filename)
				}
			})
		}
	})

	t.Run("PackageCoverageData", func(t *testing.T) {
		pkgs := []PackageCoverageData{
			{Package: "pkg1"},
			{Package: "pkg2"},
		}

		result := engine.slice(pkgs, 0, 1)
		slicedPkgs, ok := result.([]PackageCoverageData)
		require.True(t, ok)
		require.Len(t, slicedPkgs, 1)
		require.Equal(t, "pkg1", slicedPkgs[0].Package)
	})

	t.Run("RecommendationData", func(t *testing.T) {
		recs := []RecommendationData{
			{Title: "rec1"},
			{Title: "rec2"},
		}

		result := engine.slice(recs, 1, 2)
		slicedRecs, ok := result.([]RecommendationData)
		require.True(t, ok)
		require.Len(t, slicedRecs, 1)
		require.Equal(t, "rec2", slicedRecs[0].Title)
	})

	t.Run("StringSlice", func(t *testing.T) {
		strs := []string{"a", "b", "c"}

		result := engine.slice(strs, 0, 2)
		slicedStrs, ok := result.([]string)
		require.True(t, ok)
		require.Len(t, slicedStrs, 2)
		require.Equal(t, "a", slicedStrs[0])
		require.Equal(t, "b", slicedStrs[1])
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		result := engine.slice(123, 0, 1)
		require.Equal(t, 123, result) // Should return input unchanged
	})
}

func TestLengthFunction(t *testing.T) {
	engine := NewPRTemplateEngine(nil)

	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"FileCoverageData", []FileCoverageData{{}, {}}, 2},
		{"PackageCoverageData", []PackageCoverageData{{}, {}, {}}, 3},
		{"RecommendationData", []RecommendationData{{}}, 1},
		{"StringSlice", []string{"a", "b"}, 2},
		{"String", "hello", 5},
		{"EmptyString", "", 0},
		{"UnsupportedType", 123, 0},
		{"Nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.length(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestAddCustomTemplate(t *testing.T) {
	engine := NewPRTemplateEngine(nil)

	t.Run("ValidTemplate", func(t *testing.T) {
		templateContent := `Hello {{ .Repository.Name }}`
		err := engine.AddCustomTemplate("custom", templateContent)
		require.NoError(t, err)

		// Verify template was added
		_, exists := engine.templates["custom"]
		require.True(t, exists)
	})

	t.Run("InvalidTemplate", func(t *testing.T) {
		templateContent := `Hello {{ .InvalidSyntax`
		err := engine.AddCustomTemplate("invalid", templateContent)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse custom template")
	})
}

func TestFormatGradeMissingCases(t *testing.T) {
	tests := []struct {
		name          string
		includeEmojis bool
		grade         string
		expected      string
	}{
		{"EmojiDisabled", false, "A+", "A+"},
		{"UnknownGrade", true, "Z", "Z"},
		{"CaseInsensitive", true, "a", "a"}, // Should not match since it's case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewPRTemplateEngine(&TemplateConfig{
				IncludeEmojis: tt.includeEmojis,
			})
			result := engine.formatGrade(tt.grade)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestTrendChartMissingCases(t *testing.T) {
	engine := NewPRTemplateEngine(&TemplateConfig{
		IncludeCharts: true,
	})

	t.Run("AllSameValues", func(t *testing.T) {
		values := []float64{80.0, 80.0, 80.0}
		result := engine.trendChart(values)
		// When all values are the same, should return dashes
		require.Equal(t, "───", result)
	})

	t.Run("SingleValueInSlice", func(t *testing.T) {
		values := []float64{85.0}
		result := engine.trendChart(values)
		// Single value should produce one character
		require.Equal(t, "─", result)
	})

	t.Run("VariedValues", func(t *testing.T) {
		values := []float64{10.0, 50.0, 90.0}
		result := engine.trendChart(values)
		// Should produce a chart with different heights
		require.NotEmpty(t, result)
		require.Len(t, []rune(result), 3)
	})
}
