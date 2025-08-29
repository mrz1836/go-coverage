package github

import (
	"strings"
	"testing"
	"time"

	"github.com/mrz1836/go-coverage/internal/history"
)

func TestTrendVisualizer_GenerateASCIIChart(t *testing.T) {
	visualizer := NewTrendVisualizer(DefaultTrendConfig())

	tests := []struct {
		name             string
		records          []history.CoverageRecord
		expectedContains []string
		config           *TrendConfig
	}{
		{
			name:    "empty records",
			records: []history.CoverageRecord{},
			expectedContains: []string{
				"No coverage history data available",
			},
		},
		{
			name: "single record",
			records: []history.CoverageRecord{
				{
					Percentage: 85.5,
					Timestamp:  time.Now(),
				},
			},
			expectedContains: []string{
				"85.5%",
				"●",
			},
		},
		{
			name: "trend data",
			records: []history.CoverageRecord{
				{
					Percentage: 80.0,
					Timestamp:  time.Now().Add(-2 * time.Hour),
				},
				{
					Percentage: 82.5,
					Timestamp:  time.Now().Add(-1 * time.Hour),
				},
				{
					Percentage: 85.0,
					Timestamp:  time.Now(),
				},
			},
			expectedContains: []string{
				"85.0%", // Current value should be shown
				"●",     // Data points
				"┤",     // Y-axis
				"─",     // X-axis
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config != nil {
				visualizer = NewTrendVisualizer(tt.config)
			}

			chart := visualizer.GenerateASCIIChart(tt.records)

			for _, expected := range tt.expectedContains {
				if !strings.Contains(chart, expected) {
					t.Errorf("expected chart to contain '%s', but it didn't.\nChart:\n%s", expected, chart)
				}
			}
		})
	}
}

func TestTrendVisualizer_GenerateSparkline(t *testing.T) {
	visualizer := NewTrendVisualizer(DefaultTrendConfig())

	tests := []struct {
		name             string
		records          []history.CoverageRecord
		expectedContains []string
	}{
		{
			name:    "empty records",
			records: []history.CoverageRecord{},
			expectedContains: []string{
				"",
			},
		},
		{
			name: "single record",
			records: []history.CoverageRecord{
				{Percentage: 85.0},
			},
			expectedContains: []string{
				"85.0%",
			},
		},
		{
			name: "increasing trend",
			records: []history.CoverageRecord{
				{Percentage: 70.0},
				{Percentage: 75.0},
				{Percentage: 80.0},
				{Percentage: 85.0},
			},
			expectedContains: []string{
				"85.0%", // Current value
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sparkline := visualizer.generateSparkline(tt.records)

			for _, expected := range tt.expectedContains {
				if expected != "" && !strings.Contains(sparkline, expected) {
					t.Errorf("expected sparkline to contain '%s', but it didn't.\nSparkline: %s", expected, sparkline)
				}
			}

			// Check that sparkline contains spark characters when there are multiple records
			if len(tt.records) > 1 {
				sparkChars := "▁▂▃▄▅▆▇█"
				containsSparkChar := false
				for _, char := range sparkChars {
					if strings.ContainsRune(sparkline, char) {
						containsSparkChar = true
						break
					}
				}
				if !containsSparkChar {
					t.Errorf("expected sparkline to contain spark characters, but it didn't.\nSparkline: %s", sparkline)
				}
			}
		})
	}
}

func TestTrendVisualizer_GenerateBarChart(t *testing.T) {
	config := DefaultTrendConfig()
	config.ChartStyle = "bar"
	visualizer := NewTrendVisualizer(config)

	records := []history.CoverageRecord{
		{
			Percentage: 75.0,
			Timestamp:  time.Now().Add(-2 * time.Hour),
		},
		{
			Percentage: 80.0,
			Timestamp:  time.Now().Add(-1 * time.Hour),
		},
		{
			Percentage: 85.0,
			Timestamp:  time.Now(),
		},
	}

	chart := visualizer.GenerateASCIIChart(records)

	expectedContains := []string{
		"75.0%",
		"80.0%",
		"85.0%",
		"█", // Bar character
		"│", // Separator
	}

	for _, expected := range expectedContains {
		if !strings.Contains(chart, expected) {
			t.Errorf("expected bar chart to contain '%s', but it didn't.\nChart:\n%s", expected, chart)
		}
	}
}

func TestTrendVisualizer_GenerateTrendSummary(t *testing.T) {
	visualizer := NewTrendVisualizer(DefaultTrendConfig())

	tests := []struct {
		name             string
		records          []history.CoverageRecord
		expectedContains []string
	}{
		{
			name:    "empty records",
			records: []history.CoverageRecord{},
			expectedContains: []string{
				"No trend data available",
			},
		},
		{
			name: "single record",
			records: []history.CoverageRecord{
				{Percentage: 85.0},
			},
			expectedContains: []string{
				"Current coverage: 85.0%",
			},
		},
		{
			name: "improving trend",
			records: []history.CoverageRecord{
				{
					Percentage: 75.0,
					Timestamp:  time.Now().Add(-1 * time.Hour),
				},
				{
					Percentage: 80.0,
					Timestamp:  time.Now(),
				},
			},
			expectedContains: []string{
				"📈",
				"improving",
				"75.0%",
				"80.0%",
				"+5.0%",
			},
		},
		{
			name: "declining trend",
			records: []history.CoverageRecord{
				{
					Percentage: 85.0,
					Timestamp:  time.Now().Add(-1 * time.Hour),
				},
				{
					Percentage: 80.0,
					Timestamp:  time.Now(),
				},
			},
			expectedContains: []string{
				"📉",
				"declining",
				"85.0%",
				"80.0%",
				"-5.0%",
			},
		},
		{
			name: "stable trend",
			records: []history.CoverageRecord{
				{
					Percentage: 80.0,
					Timestamp:  time.Now().Add(-1 * time.Hour),
				},
				{
					Percentage: 80.2,
					Timestamp:  time.Now(),
				},
			},
			expectedContains: []string{
				"📊",
				"stable",
				"80.0%",
				"80.2%",
			},
		},
		{
			name: "long term trend",
			records: []history.CoverageRecord{
				{
					Percentage: 70.0,
					Timestamp:  time.Now().Add(-48 * time.Hour),
				},
				{
					Percentage: 75.0,
					Timestamp:  time.Now().Add(-24 * time.Hour),
				},
				{
					Percentage: 80.0,
					Timestamp:  time.Now(),
				},
			},
			expectedContains: []string{
				"70.0%",
				"80.0%",
				"+10.0%",
				"over 2 days",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := visualizer.GenerateTrendSummary(tt.records)

			for _, expected := range tt.expectedContains {
				if !strings.Contains(summary, expected) {
					t.Errorf("expected summary to contain '%s', but it didn't.\nSummary: %s", expected, summary)
				}
			}
		})
	}
}

func TestTrendVisualizer_GenerateCompactTrend(t *testing.T) {
	visualizer := NewTrendVisualizer(DefaultTrendConfig())

	tests := []struct {
		name             string
		records          []history.CoverageRecord
		expectedContains []string
	}{
		{
			name:    "empty records",
			records: []history.CoverageRecord{},
			expectedContains: []string{
				"No data",
			},
		},
		{
			name: "single record",
			records: []history.CoverageRecord{
				{Percentage: 85.0},
			},
			expectedContains: []string{
				"85.0%",
			},
		},
		{
			name: "trend with improvement",
			records: []history.CoverageRecord{
				{Percentage: 75.0},
				{Percentage: 80.0},
			},
			expectedContains: []string{
				"80.0%",
				"(+5.0%)",
			},
		},
		{
			name: "trend with decline",
			records: []history.CoverageRecord{
				{Percentage: 85.0},
				{Percentage: 80.0},
			},
			expectedContains: []string{
				"80.0%",
				"(-5.0%)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compact := visualizer.GenerateCompactTrend(tt.records)

			for _, expected := range tt.expectedContains {
				if !strings.Contains(compact, expected) {
					t.Errorf("expected compact trend to contain '%s', but it didn't.\nCompact: %s", expected, compact)
				}
			}
		})
	}
}

func TestTrendVisualizer_FindMinMax(t *testing.T) {
	visualizer := NewTrendVisualizer(DefaultTrendConfig())

	tests := []struct {
		name        string
		records     []history.CoverageRecord
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "empty records",
			records:     []history.CoverageRecord{},
			expectedMin: 0,
			expectedMax: 100,
		},
		{
			name: "single record",
			records: []history.CoverageRecord{
				{Percentage: 85.0},
			},
			expectedMin: 84.0, // With padding
			expectedMax: 86.0, // With padding
		},
		{
			name: "multiple records",
			records: []history.CoverageRecord{
				{Percentage: 70.0},
				{Percentage: 80.0},
				{Percentage: 90.0},
			},
			expectedMin: 68.0, // 70 - 2 (10% of 20 range)
			expectedMax: 92.0, // 90 + 2 (10% of 20 range)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minVal, maxVal := visualizer.findMinMax(tt.records)

			if minVal != tt.expectedMin {
				t.Errorf("expected min %.1f, got %.1f", tt.expectedMin, minVal)
			}
			if maxVal != tt.expectedMax {
				t.Errorf("expected max %.1f, got %.1f", tt.expectedMax, maxVal)
			}
		})
	}
}

func TestTrendVisualizer_ConfigVariations(t *testing.T) {
	records := []history.CoverageRecord{
		{Percentage: 75.0, Timestamp: time.Now().Add(-2 * time.Hour)},
		{Percentage: 80.0, Timestamp: time.Now().Add(-1 * time.Hour)},
		{Percentage: 85.0, Timestamp: time.Now()},
	}

	tests := []struct {
		name   string
		config *TrendConfig
	}{
		{
			name: "compact chart",
			config: &TrendConfig{
				Width:      20,
				Height:     4,
				ShowValues: false,
				ChartStyle: "line",
			},
		},
		{
			name: "sparkline style",
			config: &TrendConfig{
				ChartStyle: "sparkline",
			},
		},
		{
			name: "bar chart style",
			config: &TrendConfig{
				ChartStyle: "bar",
				Width:      30,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visualizer := NewTrendVisualizer(tt.config)
			chart := visualizer.GenerateASCIIChart(records)

			// Basic validation that chart is generated
			if len(chart) == 0 {
				t.Error("expected non-empty chart")
			}

			// Check if we should expect "85.0%" based on ShowValues setting
			showValues := tt.config.ShowValues != false || tt.name != "compact chart"
			if showValues && !strings.Contains(chart, "85.0%") {
				t.Errorf("expected chart to contain current value 85.0%%.\nChart:\n%s", chart)
			}
		})
	}
}

func TestTrendVisualizer_EdgeCases(t *testing.T) {
	visualizer := NewTrendVisualizer(DefaultTrendConfig())

	t.Run("all same values", func(t *testing.T) {
		records := []history.CoverageRecord{
			{Percentage: 80.0},
			{Percentage: 80.0},
			{Percentage: 80.0},
		}

		chart := visualizer.GenerateASCIIChart(records)
		if len(chart) == 0 {
			t.Error("expected non-empty chart for same values")
		}
	})

	t.Run("extreme values", func(t *testing.T) {
		records := []history.CoverageRecord{
			{Percentage: 0.0},
			{Percentage: 100.0},
		}

		chart := visualizer.GenerateASCIIChart(records)
		if len(chart) == 0 {
			t.Error("expected non-empty chart for extreme values")
		}

		// Should contain both extreme values
		if !strings.Contains(chart, "0.0%") {
			t.Error("expected chart to contain 0.0%")
		}
		if !strings.Contains(chart, "100.0%") {
			t.Error("expected chart to contain 100.0%")
		}
	})

	t.Run("very small differences", func(t *testing.T) {
		records := []history.CoverageRecord{
			{Percentage: 80.001},
			{Percentage: 80.002},
			{Percentage: 80.003},
		}

		chart := visualizer.GenerateASCIIChart(records)
		if len(chart) == 0 {
			t.Error("expected non-empty chart for small differences")
		}
	})
}
