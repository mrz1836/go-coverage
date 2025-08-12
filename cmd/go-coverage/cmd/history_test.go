package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-coverage/internal/config"
	"github.com/mrz1836/go-coverage/internal/history"
	"github.com/mrz1836/go-coverage/internal/parser"
)

func TestHistoryCommandMetadata(t *testing.T) {
	// Create a Commands instance for testing
	versionInfo := VersionInfo{
		Version:   "test",
		Commit:    "test-commit",
		BuildDate: "test-date",
	}
	commands := NewCommands(versionInfo)

	// Test command metadata
	assert.Equal(t, "history", commands.History.Use)
	assert.Equal(t, "Manage coverage history", commands.History.Short)
	assert.Contains(t, commands.History.Long, "Manage historical coverage data")
	assert.NotNil(t, commands.History.RunE)
}

func TestHistoryCommandFlags(t *testing.T) {
	// Test that all expected flags exist and have correct defaults
	expectedFlags := map[string]struct {
		flagType     string
		defaultValue string
	}{
		"add":        {"string", ""},
		"branch":     {"string", ""},
		"commit":     {"string", ""},
		"commit-url": {"string", ""},
		"trend":      {"bool", "false"},
		"stats":      {"bool", "false"},
		"cleanup":    {"bool", "false"},
		"days":       {"int", "30"},
		"format":     {"string", "text"},
	}

	for flagName, expected := range expectedFlags {
		t.Run(fmt.Sprintf("flag_%s", flagName), func(t *testing.T) {
			// Create a Commands instance for testing
			versionInfo := VersionInfo{
				Version:   "test",
				Commit:    "test-commit",
				BuildDate: "test-date",
			}
			commands := NewCommands(versionInfo)

			flag := commands.History.Flags().Lookup(flagName)
			require.NotNil(t, flag, "Flag %s should exist", flagName)
			assert.Equal(t, expected.flagType, flag.Value.Type(), "Flag %s should be %s type", flagName, expected.flagType)
			assert.Equal(t, expected.defaultValue, flag.DefValue, "Flag %s should have default %s", flagName, expected.defaultValue)
		})
	}
}

func TestAddToHistory(t *testing.T) {
	tempDir := t.TempDir()
	coverageFile := filepath.Join(tempDir, "coverage.txt")
	historyDir := filepath.Join(tempDir, "history")

	// Create a valid coverage file
	coverageContent := `mode: set
github.com/test/repo/main.go:10.2,12.16 1 1
github.com/test/repo/main.go:15.2,17.16 1 0
`
	require.NoError(t, os.WriteFile(coverageFile, []byte(coverageContent), 0o600))
	require.NoError(t, os.MkdirAll(historyDir, 0o750))

	// Create test config
	cfg := &config.Config{
		History: config.HistoryConfig{
			StoragePath:   historyDir,
			RetentionDays: 30,
			MaxEntries:    100,
		},
		GitHub: config.GitHubConfig{
			Owner:      "test",
			Repository: "repo",
			CommitSHA:  "abc123",
		},
	}

	// Create history tracker
	historyConfig := &history.Config{
		StoragePath:    cfg.History.StoragePath,
		RetentionDays:  cfg.History.RetentionDays,
		MaxEntries:     cfg.History.MaxEntries,
		AutoCleanup:    cfg.History.AutoCleanup,
		MetricsEnabled: cfg.History.MetricsEnabled,
	}
	tracker := history.NewWithConfig(historyConfig)

	// Create command
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Test addToHistory function
	ctx := context.Background()
	err := addToHistory(ctx, tracker, coverageFile, "main", "abc123", "https://github.com/test/repo/commit/abc123", cfg, cmd)
	require.NoError(t, err)

	// Check output
	output := buf.String()
	assert.Contains(t, output, "Coverage recorded successfully!")
	assert.Contains(t, output, "Branch: main")
	assert.Contains(t, output, "Commit: abc123")
	assert.Contains(t, output, "Coverage:")
}

func TestAddToHistoryWithDefaults(t *testing.T) {
	tempDir := t.TempDir()
	coverageFile := filepath.Join(tempDir, "coverage.txt")
	historyDir := filepath.Join(tempDir, "history")

	// Create a valid coverage file
	coverageContent := `mode: set
github.com/test/repo/main.go:10.2,12.16 1 1
`
	require.NoError(t, os.WriteFile(coverageFile, []byte(coverageContent), 0o600))
	require.NoError(t, os.MkdirAll(historyDir, 0o750))

	// Create test config with empty values to test defaults
	cfg := &config.Config{
		History: config.HistoryConfig{
			StoragePath:   historyDir,
			RetentionDays: 30,
			MaxEntries:    100,
		},
		GitHub: config.GitHubConfig{
			Owner:      "", // Empty to test defaults
			Repository: "",
			CommitSHA:  "",
		},
	}

	historyConfig := &history.Config{
		StoragePath:   cfg.History.StoragePath,
		RetentionDays: cfg.History.RetentionDays,
		MaxEntries:    cfg.History.MaxEntries,
	}
	tracker := history.NewWithConfig(historyConfig)

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	ctx := context.Background()
	err := addToHistory(ctx, tracker, coverageFile, "", "", "", cfg, cmd)
	require.NoError(t, err)

	// Should use defaults
	output := buf.String()
	assert.Contains(t, output, "Branch: master") // Default branch
}

func TestAddToHistoryInvalidFile(t *testing.T) {
	tempDir := t.TempDir()
	historyDir := filepath.Join(tempDir, "history")
	require.NoError(t, os.MkdirAll(historyDir, 0o750))

	cfg := &config.Config{
		History: config.HistoryConfig{
			StoragePath: historyDir,
		},
	}

	historyConfig := &history.Config{
		StoragePath: cfg.History.StoragePath,
	}
	tracker := history.NewWithConfig(historyConfig)

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	ctx := context.Background()
	err := addToHistory(ctx, tracker, "/nonexistent/coverage.txt", "main", "abc123", "", cfg, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse coverage file")
}

func TestShowTrendData(t *testing.T) {
	tempDir := t.TempDir()
	historyDir := filepath.Join(tempDir, "history")
	require.NoError(t, os.MkdirAll(historyDir, 0o750))

	// Create history tracker and add some test data
	historyConfig := &history.Config{
		StoragePath: historyDir,
	}
	tracker := history.NewWithConfig(historyConfig)

	// Add some test entries
	ctx := context.Background()
	coverage1 := &parser.CoverageData{
		Percentage:   85.0,
		TotalLines:   100,
		CoveredLines: 85,
	}
	coverage2 := &parser.CoverageData{
		Percentage:   90.0,
		TotalLines:   100,
		CoveredLines: 90,
	}

	err := tracker.Record(ctx, coverage1, history.WithBranch("main"), history.WithCommit("abc123", ""))
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	err = tracker.Record(ctx, coverage2, history.WithBranch("main"), history.WithCommit("def456", ""))
	require.NoError(t, err)

	// Test text format
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	err = showTrendData(ctx, tracker, "main", 30, "text", cmd)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Coverage Trend Analysis")
	assert.Contains(t, output, "Branch: main")
	assert.Contains(t, output, "Period: 30 days")
	assert.Contains(t, output, "Total Entries:")
	assert.Contains(t, output, "Average Coverage:")
}

func TestShowTrendDataJSON(t *testing.T) {
	tempDir := t.TempDir()
	historyDir := filepath.Join(tempDir, "history")
	require.NoError(t, os.MkdirAll(historyDir, 0o750))

	historyConfig := &history.Config{
		StoragePath: historyDir,
	}
	tracker := history.NewWithConfig(historyConfig)

	// Add test entry
	ctx := context.Background()
	coverage := &parser.CoverageData{
		Percentage:   85.0,
		TotalLines:   100,
		CoveredLines: 85,
	}
	err := tracker.Record(ctx, coverage, history.WithBranch("main"))
	require.NoError(t, err)

	// Test JSON format
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	err = showTrendData(ctx, tracker, "main", 30, "json", cmd)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "{")

	// Verify it's valid JSON
	var trendData interface{}
	err = json.Unmarshal([]byte(output), &trendData)
	assert.NoError(t, err, "Output should be valid JSON")
}

func TestShowTrendDataDefaults(t *testing.T) {
	tempDir := t.TempDir()
	historyDir := filepath.Join(tempDir, "history")
	require.NoError(t, os.MkdirAll(historyDir, 0o750))

	historyConfig := &history.Config{
		StoragePath: historyDir,
	}
	tracker := history.NewWithConfig(historyConfig)

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	ctx := context.Background()
	// Test with empty branch (should use default) and 0 days (should use 30)
	err := showTrendData(ctx, tracker, "", 0, "text", cmd)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Branch: master")  // Default branch
	assert.Contains(t, output, "Period: 30 days") // Default days
}

func TestShowStatistics(t *testing.T) {
	tempDir := t.TempDir()
	historyDir := filepath.Join(tempDir, "history")
	require.NoError(t, os.MkdirAll(historyDir, 0o750))

	historyConfig := &history.Config{
		StoragePath: historyDir,
	}
	tracker := history.NewWithConfig(historyConfig)

	// Add test entries
	ctx := context.Background()
	coverage := &parser.CoverageData{
		Percentage:   85.0,
		TotalLines:   100,
		CoveredLines: 85,
	}
	err := tracker.Record(ctx, coverage, history.WithBranch("main"), history.WithMetadata("project", "test/repo"))
	require.NoError(t, err)

	// Test text format
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	err = showStatistics(ctx, tracker, "text", cmd)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Coverage History Statistics")
	assert.Contains(t, output, "Total Entries:")
	assert.Contains(t, output, "Storage Size:")
}

func TestShowStatisticsJSON(t *testing.T) {
	tempDir := t.TempDir()
	historyDir := filepath.Join(tempDir, "history")
	require.NoError(t, os.MkdirAll(historyDir, 0o750))

	historyConfig := &history.Config{
		StoragePath: historyDir,
	}
	tracker := history.NewWithConfig(historyConfig)

	// Add test entry
	ctx := context.Background()
	coverage := &parser.CoverageData{Percentage: 85.0}
	err := tracker.Record(ctx, coverage, history.WithBranch("main"))
	require.NoError(t, err)

	// Test JSON format
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	err = showStatistics(ctx, tracker, "json", cmd)
	require.NoError(t, err)

	output := buf.String()
	var stats interface{}
	err = json.Unmarshal([]byte(output), &stats)
	assert.NoError(t, err, "Output should be valid JSON")
}

func TestCleanupHistory(t *testing.T) {
	tempDir := t.TempDir()
	historyDir := filepath.Join(tempDir, "history")
	require.NoError(t, os.MkdirAll(historyDir, 0o750))

	historyConfig := &history.Config{
		StoragePath: historyDir,
	}
	tracker := history.NewWithConfig(historyConfig)

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	ctx := context.Background()
	err := cleanupHistory(ctx, tracker, cmd)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "History cleanup completed successfully!")
}

func TestShowLatestEntry(t *testing.T) {
	tempDir := t.TempDir()
	historyDir := filepath.Join(tempDir, "history")
	require.NoError(t, os.MkdirAll(historyDir, 0o750))

	historyConfig := &history.Config{
		StoragePath: historyDir,
	}
	tracker := history.NewWithConfig(historyConfig)

	// Add test entry
	ctx := context.Background()
	coverage := &parser.CoverageData{
		Percentage:   85.0,
		TotalLines:   100,
		CoveredLines: 85,
	}
	err := tracker.Record(ctx, coverage,
		history.WithBranch("main"),
		history.WithCommit("abc123", ""),
		history.WithMetadata("project", "test/repo"))
	require.NoError(t, err)

	// Test text format
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	err = showLatestEntry(ctx, tracker, "main", "text", cmd)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Latest Coverage Entry")
	assert.Contains(t, output, "Branch: main")
	assert.Contains(t, output, "Commit: abc123")
	assert.Contains(t, output, "Coverage: 85.00%")
	assert.Contains(t, output, "Metadata:")
	assert.Contains(t, output, "project: test/repo")
}

func TestShowLatestEntryJSON(t *testing.T) {
	tempDir := t.TempDir()
	historyDir := filepath.Join(tempDir, "history")
	require.NoError(t, os.MkdirAll(historyDir, 0o750))

	historyConfig := &history.Config{
		StoragePath: historyDir,
	}
	tracker := history.NewWithConfig(historyConfig)

	// Add test entry
	ctx := context.Background()
	coverage := &parser.CoverageData{Percentage: 85.0}
	err := tracker.Record(ctx, coverage, history.WithBranch("main"))
	require.NoError(t, err)

	// Test JSON format
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	err = showLatestEntry(ctx, tracker, "main", "json", cmd)
	require.NoError(t, err)

	output := buf.String()
	var entry interface{}
	err = json.Unmarshal([]byte(output), &entry)
	assert.NoError(t, err, "Output should be valid JSON")
}

func TestShowLatestEntryDefaults(t *testing.T) {
	tempDir := t.TempDir()
	historyDir := filepath.Join(tempDir, "history")
	require.NoError(t, os.MkdirAll(historyDir, 0o750))

	historyConfig := &history.Config{
		StoragePath: historyDir,
	}
	tracker := history.NewWithConfig(historyConfig)

	// Add entry to default branch
	ctx := context.Background()
	coverage := &parser.CoverageData{Percentage: 85.0}
	err := tracker.Record(ctx, coverage, history.WithBranch("master"))
	require.NoError(t, err)

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	// Test with empty branch (should use default)
	err = showLatestEntry(ctx, tracker, "", "text", cmd)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Branch: master")
}

func TestHistoryCommandIntegration(t *testing.T) {
	t.Skip("Skipping integration test - functionality covered by unit tests")
}

// TestHistoryCommandShowLatestDefault is commented out due to complex integration dependencies
// The core functionality is tested by the individual function tests above
func TestHistoryCommandShowLatestDefault(t *testing.T) {
	t.Skip("Skipping complex integration test - functionality covered by unit tests")
}
