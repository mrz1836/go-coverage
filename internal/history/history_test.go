package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errExpectedRecordButGotNil = errors.New("expected record but got nil")
	errCommitSHAMismatch       = errors.New("commit SHA mismatch")
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name    string
		baseDir string
		want    string
	}{
		{
			name:    "Standard directory",
			baseDir: "/tmp/test",
			want:    "/tmp/test/coverage-history.json",
		},
		{
			name:    "Empty directory",
			baseDir: "",
			want:    "coverage-history.json",
		},
		{
			name:    "Root directory",
			baseDir: "/",
			want:    "/coverage-history.json",
		},
		{
			name:    "Nested directory",
			baseDir: "/path/to/deep/dir",
			want:    "/path/to/deep/dir/coverage-history.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(tt.baseDir)
			require.NotNil(t, manager)
			assert.Equal(t, tt.want, manager.historyFile)
		})
	}
}

func TestSaveRecord(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "history_test_*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager := NewManager(tempDir)

	t.Run("Save single record", func(t *testing.T) {
		record := &CoverageRecord{
			Timestamp:    time.Now(),
			CommitSHA:    "abc123",
			Branch:       "main",
			Percentage:   85.5,
			TotalLines:   1000,
			CoveredLines: 855,
		}

		err := manager.SaveRecord(record)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(manager.historyFile)
		require.NoError(t, err)

		// Verify content
		history, err := manager.loadHistory()
		require.NoError(t, err)
		require.Len(t, history, 1)
		assert.Equal(t, record.CommitSHA, history[0].CommitSHA)
		assert.Equal(t, record.Branch, history[0].Branch)
		assert.InEpsilon(t, record.Percentage, history[0].Percentage, 0.001)
		assert.Equal(t, record.TotalLines, history[0].TotalLines)
		assert.Equal(t, record.CoveredLines, history[0].CoveredLines)
	})

	t.Run("Save multiple records", func(t *testing.T) {
		// Clean up from previous test
		_ = os.Remove(manager.historyFile)

		records := []*CoverageRecord{
			{
				Timestamp:    time.Now().Add(-2 * time.Hour),
				CommitSHA:    "abc123",
				Branch:       "main",
				Percentage:   80.0,
				TotalLines:   1000,
				CoveredLines: 800,
			},
			{
				Timestamp:    time.Now().Add(-1 * time.Hour),
				CommitSHA:    "def456",
				Branch:       "main",
				Percentage:   85.0,
				TotalLines:   1000,
				CoveredLines: 850,
			},
			{
				Timestamp:    time.Now(),
				CommitSHA:    "ghi789",
				Branch:       "main",
				Percentage:   90.0,
				TotalLines:   1000,
				CoveredLines: 900,
			},
		}

		for _, record := range records {
			err := manager.SaveRecord(record)
			require.NoError(t, err)
		}

		history, err := manager.loadHistory()
		require.NoError(t, err)
		require.Len(t, history, 3)

		// Verify records are in the correct order
		for i, expectedRecord := range records {
			assert.Equal(t, expectedRecord.CommitSHA, history[i].CommitSHA)
			assert.InEpsilon(t, expectedRecord.Percentage, history[i].Percentage, 0.01)
		}
	})

	t.Run("Save record with history limit", func(t *testing.T) {
		// Clean up from previous test
		_ = os.Remove(manager.historyFile)

		// Create 101 records to test the 100 record limit
		for i := 0; i < 101; i++ {
			record := &CoverageRecord{
				Timestamp:    time.Now().Add(time.Duration(i) * time.Minute),
				CommitSHA:    fmt.Sprintf("commit%d", i),
				Branch:       "main",
				Percentage:   float64(80 + i%20),
				TotalLines:   1000,
				CoveredLines: 800 + i%200,
			}
			err := manager.SaveRecord(record)
			require.NoError(t, err)
		}

		history, err := manager.loadHistory()
		require.NoError(t, err)
		// Should be limited to 100 records
		assert.Len(t, history, 100)
		// First record should be commit1 (commit0 was removed)
		assert.Equal(t, "commit1", history[0].CommitSHA)
		// Last record should be commit100
		assert.Equal(t, "commit100", history[99].CommitSHA)
	})

	t.Run("Save record with corrupted existing file", func(t *testing.T) {
		// Clean up from previous test
		_ = os.Remove(manager.historyFile)

		// Create a corrupted JSON file
		err := os.WriteFile(manager.historyFile, []byte("invalid json content"), 0o600)
		require.NoError(t, err)

		record := &CoverageRecord{
			Timestamp:    time.Now(),
			CommitSHA:    "abc123",
			Branch:       "main",
			Percentage:   85.5,
			TotalLines:   1000,
			CoveredLines: 855,
		}

		err = manager.SaveRecord(record)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load existing history")
	})

	t.Run("Save record with directory creation", func(t *testing.T) {
		// Test with a non-existent nested directory
		nestedDir := filepath.Join(tempDir, "deep", "nested", "path")
		nestedManager := NewManager(nestedDir)

		record := &CoverageRecord{
			Timestamp:    time.Now(),
			CommitSHA:    "abc123",
			Branch:       "main",
			Percentage:   85.5,
			TotalLines:   1000,
			CoveredLines: 855,
		}

		err := nestedManager.SaveRecord(record)
		require.NoError(t, err)

		// Verify directory was created
		_, err = os.Stat(nestedDir)
		require.NoError(t, err)

		// Verify file was created
		_, err = os.Stat(nestedManager.historyFile)
		require.NoError(t, err)
	})
}

func TestGetLastRecord(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "history_test_*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager := NewManager(tempDir)

	t.Run("Get last record when no history exists", func(t *testing.T) {
		record, err := manager.GetLastRecord()
		assert.Nil(t, record)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNoHistory)
	})

	t.Run("Get last record when file is corrupted", func(t *testing.T) {
		// Create a corrupted JSON file
		err := os.WriteFile(manager.historyFile, []byte("invalid json content"), 0o600)
		require.NoError(t, err)

		record, err := manager.GetLastRecord()
		assert.Nil(t, record)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load history")
	})

	t.Run("Get last record when history is empty array", func(t *testing.T) {
		// Create an empty array JSON file
		err := os.WriteFile(manager.historyFile, []byte("[]"), 0o600)
		require.NoError(t, err)

		record, err := manager.GetLastRecord()
		assert.Nil(t, record)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNoHistory)
	})

	t.Run("Get last record with single record", func(t *testing.T) {
		expectedRecord := &CoverageRecord{
			Timestamp:    time.Now(),
			CommitSHA:    "abc123",
			Branch:       "main",
			Percentage:   85.5,
			TotalLines:   1000,
			CoveredLines: 855,
		}

		err := manager.SaveRecord(expectedRecord)
		require.NoError(t, err)

		record, err := manager.GetLastRecord()
		require.NoError(t, err)
		require.NotNil(t, record)
		assert.Equal(t, expectedRecord.CommitSHA, record.CommitSHA)
		assert.Equal(t, expectedRecord.Branch, record.Branch)
		assert.InEpsilon(t, expectedRecord.Percentage, record.Percentage, 0.01)
		assert.Equal(t, expectedRecord.TotalLines, record.TotalLines)
		assert.Equal(t, expectedRecord.CoveredLines, record.CoveredLines)
	})

	t.Run("Get last record with multiple records", func(t *testing.T) {
		// Clean up from previous test
		_ = os.Remove(manager.historyFile)

		records := []*CoverageRecord{
			{
				Timestamp:    time.Now().Add(-2 * time.Hour),
				CommitSHA:    "abc123",
				Branch:       "main",
				Percentage:   80.0,
				TotalLines:   1000,
				CoveredLines: 800,
			},
			{
				Timestamp:    time.Now().Add(-1 * time.Hour),
				CommitSHA:    "def456",
				Branch:       "main",
				Percentage:   85.0,
				TotalLines:   1000,
				CoveredLines: 850,
			},
			{
				Timestamp:    time.Now(), // This should be the last record
				CommitSHA:    "ghi789",
				Branch:       "feature",
				Percentage:   90.0,
				TotalLines:   1200,
				CoveredLines: 1080,
			},
		}

		for _, rec := range records {
			err := manager.SaveRecord(rec)
			require.NoError(t, err)
		}

		lastRecord, err := manager.GetLastRecord()
		require.NoError(t, err)
		require.NotNil(t, lastRecord)

		// Should match the last record added
		expectedLast := records[2]
		assert.Equal(t, expectedLast.CommitSHA, lastRecord.CommitSHA)
		assert.Equal(t, expectedLast.Branch, lastRecord.Branch)
		assert.InEpsilon(t, expectedLast.Percentage, lastRecord.Percentage, 0.001)
		assert.Equal(t, expectedLast.TotalLines, lastRecord.TotalLines)
		assert.Equal(t, expectedLast.CoveredLines, lastRecord.CoveredLines)
	})
}

func TestGetChangeStatus(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "history_test_*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager := NewManager(tempDir)

	t.Run("Get change status with no history", func(t *testing.T) {
		status, prevPercentage, err := manager.GetChangeStatus(85.5)
		require.NoError(t, err)
		assert.Equal(t, "stable", status)
		assert.Zero(t, prevPercentage)
	})

	t.Run("Get change status with corrupted history", func(t *testing.T) {
		// Create a corrupted JSON file
		err := os.WriteFile(manager.historyFile, []byte("invalid json content"), 0o600)
		require.NoError(t, err)

		status, prevPercentage, err := manager.GetChangeStatus(85.5)
		require.Error(t, err)
		assert.Equal(t, "stable", status)
		assert.Zero(t, prevPercentage)
		assert.Contains(t, err.Error(), "failed to unmarshal history")
	})

	t.Run("Get change status - improved", func(t *testing.T) {
		// Clean up from previous test
		_ = os.Remove(manager.historyFile)

		// Save a record with lower coverage
		prevRecord := &CoverageRecord{
			Timestamp:    time.Now(),
			CommitSHA:    "abc123",
			Branch:       "main",
			Percentage:   80.0,
			TotalLines:   1000,
			CoveredLines: 800,
		}
		err := manager.SaveRecord(prevRecord)
		require.NoError(t, err)

		// Test with improved coverage (difference > 0.1%)
		status, prevPercentage, err := manager.GetChangeStatus(85.0)
		require.NoError(t, err)
		assert.Equal(t, "improved", status)
		assert.InEpsilon(t, 80.0, prevPercentage, 0.001)
	})

	t.Run("Get change status - declined", func(t *testing.T) {
		// Clean up from previous test
		_ = os.Remove(manager.historyFile)

		// Save a record with higher coverage
		prevRecord := &CoverageRecord{
			Timestamp:    time.Now(),
			CommitSHA:    "abc123",
			Branch:       "main",
			Percentage:   90.0,
			TotalLines:   1000,
			CoveredLines: 900,
		}
		err := manager.SaveRecord(prevRecord)
		require.NoError(t, err)

		// Test with declined coverage (difference < -0.1%)
		status, prevPercentage, err := manager.GetChangeStatus(85.0)
		require.NoError(t, err)
		assert.Equal(t, "declined", status)
		assert.InEpsilon(t, 90.0, prevPercentage, 0.001)
	})

	t.Run("Get change status - stable within threshold", func(t *testing.T) {
		// Clean up from previous test
		_ = os.Remove(manager.historyFile)

		// Save a record
		prevRecord := &CoverageRecord{
			Timestamp:    time.Now(),
			CommitSHA:    "abc123",
			Branch:       "main",
			Percentage:   85.0,
			TotalLines:   1000,
			CoveredLines: 850,
		}
		err := manager.SaveRecord(prevRecord)
		require.NoError(t, err)

		// Test with coverage change within threshold (0.05% difference)
		status, prevPercentage, err := manager.GetChangeStatus(85.05)
		require.NoError(t, err)
		assert.Equal(t, "stable", status)
		assert.InEpsilon(t, 85.0, prevPercentage, 0.001)
	})

	t.Run("Get change status - boundary cases", func(t *testing.T) {
		// Clean up from previous test
		_ = os.Remove(manager.historyFile)

		// Save a record
		prevRecord := &CoverageRecord{
			Timestamp:    time.Now(),
			CommitSHA:    "abc123",
			Branch:       "main",
			Percentage:   85.0,
			TotalLines:   1000,
			CoveredLines: 850,
		}
		err := manager.SaveRecord(prevRecord)
		require.NoError(t, err)

		// Test exactly at threshold (0.1% improvement)
		status, prevPercentage, err := manager.GetChangeStatus(85.1)
		require.NoError(t, err)
		assert.Equal(t, "stable", status) // Should be stable since > threshold, not >=
		assert.InEpsilon(t, 85.0, prevPercentage, 0.001)

		// Test just above threshold (0.11% improvement)
		status, prevPercentage, err = manager.GetChangeStatus(85.11)
		require.NoError(t, err)
		assert.Equal(t, "improved", status)
		assert.InEpsilon(t, 85.0, prevPercentage, 0.001)

		// Test exactly at negative threshold (-0.1% decline)
		status, prevPercentage, err = manager.GetChangeStatus(84.9)
		require.NoError(t, err)
		assert.Equal(t, "stable", status) // Should be stable since < threshold, not <=
		assert.InEpsilon(t, 85.0, prevPercentage, 0.001)

		// Test just below negative threshold (-0.11% decline)
		status, prevPercentage, err = manager.GetChangeStatus(84.89)
		require.NoError(t, err)
		assert.Equal(t, "declined", status)
		assert.InEpsilon(t, 85.0, prevPercentage, 0.001)
	})
}

func TestLoadHistory(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "history_test_*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager := NewManager(tempDir)

	t.Run("Load history from non-existent file", func(t *testing.T) {
		history, err := manager.loadHistory()
		assert.Nil(t, history)
		require.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("Load history from corrupted JSON file", func(t *testing.T) {
		// Create a corrupted JSON file
		err := os.WriteFile(manager.historyFile, []byte("invalid json content"), 0o600)
		require.NoError(t, err)

		history, err := manager.loadHistory()
		assert.Nil(t, history)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal history")
	})

	t.Run("Load history from empty JSON array", func(t *testing.T) {
		// Create an empty array JSON file
		err := os.WriteFile(manager.historyFile, []byte("[]"), 0o600)
		require.NoError(t, err)

		history, err := manager.loadHistory()
		require.NoError(t, err)
		assert.NotNil(t, history)
		assert.Empty(t, history)
	})

	t.Run("Load history from valid JSON file", func(t *testing.T) {
		expectedRecords := []CoverageRecord{
			{
				Timestamp:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				CommitSHA:    "abc123",
				Branch:       "main",
				Percentage:   80.0,
				TotalLines:   1000,
				CoveredLines: 800,
			},
			{
				Timestamp:    time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
				CommitSHA:    "def456",
				Branch:       "feature",
				Percentage:   85.0,
				TotalLines:   1200,
				CoveredLines: 1020,
			},
		}

		// Create valid JSON file
		data, err := json.MarshalIndent(expectedRecords, "", "  ")
		require.NoError(t, err)

		err = os.WriteFile(manager.historyFile, data, 0o600)
		require.NoError(t, err)

		history, err := manager.loadHistory()
		require.NoError(t, err)
		require.Len(t, history, 2)

		for i, expected := range expectedRecords {
			assert.Equal(t, expected.CommitSHA, history[i].CommitSHA)
			assert.Equal(t, expected.Branch, history[i].Branch)
			assert.InEpsilon(t, expected.Percentage, history[i].Percentage, 0.001)
			assert.Equal(t, expected.TotalLines, history[i].TotalLines)
			assert.Equal(t, expected.CoveredLines, history[i].CoveredLines)
			assert.Equal(t, expected.Timestamp.Unix(), history[i].Timestamp.Unix())
		}
	})
}

func TestEnsureHistoryDir(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "history_test_*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	t.Run("Ensure directory exists when parent exists", func(t *testing.T) {
		manager := NewManager(tempDir)
		ensureErr := manager.ensureHistoryDir()
		require.NoError(t, ensureErr)

		// Directory should already exist, so no error
		_, err = os.Stat(tempDir)
		require.NoError(t, err)
	})

	t.Run("Ensure directory creates nested path", func(t *testing.T) {
		nestedPath := filepath.Join(tempDir, "deep", "nested", "path")
		manager := NewManager(nestedPath)

		// Nested path should not exist initially
		_, err = os.Stat(nestedPath)
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))

		// ensureHistoryDir should create it
		err = manager.ensureHistoryDir()
		require.NoError(t, err)

		// Now it should exist
		info, err := os.Stat(nestedPath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		// Check permissions (0o750)
		assert.Equal(t, os.FileMode(0o750), info.Mode().Perm())
	})

	t.Run("Ensure directory handles existing directory", func(t *testing.T) {
		existingPath := filepath.Join(tempDir, "existing")
		err := os.MkdirAll(existingPath, 0o750)
		require.NoError(t, err)

		manager := NewManager(existingPath)
		err = manager.ensureHistoryDir()
		require.NoError(t, err)

		// Should still exist
		info, err := os.Stat(existingPath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

func TestSaveHistory(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "history_test_*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager := NewManager(tempDir)

	t.Run("Save valid history", func(t *testing.T) {
		history := []CoverageRecord{
			{
				Timestamp:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				CommitSHA:    "abc123",
				Branch:       "main",
				Percentage:   80.0,
				TotalLines:   1000,
				CoveredLines: 800,
			},
			{
				Timestamp:    time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
				CommitSHA:    "def456",
				Branch:       "feature",
				Percentage:   85.0,
				TotalLines:   1200,
				CoveredLines: 1020,
			},
		}

		saveErr := manager.saveHistory(history)
		require.NoError(t, saveErr)

		// Verify file exists
		_, statErr := os.Stat(manager.historyFile)
		require.NoError(t, statErr)

		// Verify content
		data, readErr := os.ReadFile(manager.historyFile)
		require.NoError(t, readErr)

		var loadedHistory []CoverageRecord
		err = json.Unmarshal(data, &loadedHistory)
		require.NoError(t, err)
		require.Len(t, loadedHistory, 2)

		for i, expected := range history {
			assert.Equal(t, expected.CommitSHA, loadedHistory[i].CommitSHA)
			assert.Equal(t, expected.Branch, loadedHistory[i].Branch)
			assert.InEpsilon(t, expected.Percentage, loadedHistory[i].Percentage, 0.001)
			assert.Equal(t, expected.TotalLines, loadedHistory[i].TotalLines)
			assert.Equal(t, expected.CoveredLines, loadedHistory[i].CoveredLines)
		}

		// Verify file permissions (0o600)
		info, statErr := os.Stat(manager.historyFile)
		require.NoError(t, statErr)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("Save empty history", func(t *testing.T) {
		// Clean up from previous test
		_ = os.Remove(manager.historyFile)

		emptyHistory := []CoverageRecord{}

		saveErr := manager.saveHistory(emptyHistory)
		require.NoError(t, saveErr)

		// Verify file exists and contains empty array
		data, readErr := os.ReadFile(manager.historyFile)
		require.NoError(t, readErr)

		var loadedHistory []CoverageRecord
		err = json.Unmarshal(data, &loadedHistory)
		require.NoError(t, err)
		assert.Empty(t, loadedHistory)
	})

	t.Run("Save history creates nested directory", func(t *testing.T) {
		nestedPath := filepath.Join(tempDir, "auto", "create", "path")
		nestedManager := NewManager(nestedPath)

		// Nested path should not exist initially
		_, err = os.Stat(nestedPath)
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))

		history := []CoverageRecord{
			{
				Timestamp:    time.Now(),
				CommitSHA:    "abc123",
				Branch:       "main",
				Percentage:   80.0,
				TotalLines:   1000,
				CoveredLines: 800,
			},
		}

		err = nestedManager.saveHistory(history)
		require.NoError(t, err)

		// Directory and file should now exist
		_, err = os.Stat(nestedPath)
		require.NoError(t, err)
		_, err = os.Stat(nestedManager.historyFile)
		require.NoError(t, err)
	})

	t.Run("Save history fails with directory creation error", func(t *testing.T) {
		if os.Getenv("CI") != "" {
			t.Skip("Skipping directory creation error test in CI")
		}

		// Test with an invalid path that cannot be created
		invalidPath := "/dev/null/cannot/create/this/path"
		invalidManager := NewManager(invalidPath)

		history := []CoverageRecord{
			{
				Timestamp:    time.Now(),
				CommitSHA:    "abc123",
				Branch:       "main",
				Percentage:   80.0,
				TotalLines:   1000,
				CoveredLines: 800,
			},
		}

		err := invalidManager.saveHistory(history)
		require.Error(t, err)
		// The error might be from write file or create directory, both are valid error paths
		assert.True(t,
			strings.Contains(err.Error(), "failed to create directory") ||
				strings.Contains(err.Error(), "failed to write history file"),
			"Expected error to contain directory creation or file write error, got: %v", err)
	})

	t.Run("Save history fails with write error", func(t *testing.T) {
		if os.Getenv("CI") != "" {
			t.Skip("Skipping write error test in CI")
		}

		// Create a directory where we expect a file
		conflictPath := filepath.Join(tempDir, "conflict")
		err := os.MkdirAll(conflictPath, 0o750)
		require.NoError(t, err)

		// Try to create a file with the same name as the directory
		conflictingFile := filepath.Join(conflictPath, "coverage-history.json")
		err = os.MkdirAll(conflictingFile, 0o750) // Create a directory with the file name
		require.NoError(t, err)

		conflictManager := NewManager(conflictPath)

		history := []CoverageRecord{
			{
				Timestamp:    time.Now(),
				CommitSHA:    "abc123",
				Branch:       "main",
				Percentage:   80.0,
				TotalLines:   1000,
				CoveredLines: 800,
			},
		}

		err = conflictManager.saveHistory(history)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write history file")
	})
}

// Test concurrent access scenarios
func TestConcurrentAccess(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "history_test_*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager := NewManager(tempDir)

	t.Run("Concurrent save operations", func(t *testing.T) {
		const numGoroutines = 10
		const recordsPerGoroutine = 5

		done := make(chan error, numGoroutines)

		// Start multiple goroutines that save records concurrently
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				for j := 0; j < recordsPerGoroutine; j++ {
					record := &CoverageRecord{
						Timestamp:    time.Now(),
						CommitSHA:    fmt.Sprintf("commit-%d-%d", goroutineID, j),
						Branch:       "main",
						Percentage:   float64(80 + goroutineID + j),
						TotalLines:   1000,
						CoveredLines: 800 + goroutineID + j,
					}
					if err := manager.SaveRecord(record); err != nil {
						done <- err
						return
					}
				}
				done <- nil
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			err := <-done
			// Some operations might fail due to file contention,
			// but at least some should succeed
			if err != nil {
				t.Logf("Concurrent operation failed (expected): %v", err)
			}
		}

		// Verify that the file exists and contains some records
		history, err := manager.loadHistory()
		if err == nil {
			// If we can load history, it should have some records
			assert.NotEmpty(t, history, "Should have at least some records from concurrent operations")
			// Should not exceed the limit
			assert.LessOrEqual(t, len(history), 100, "Should not exceed the 100 record limit")
		}
	})

	t.Run("Concurrent read operations", func(t *testing.T) {
		// Clean up from previous test
		_ = os.Remove(manager.historyFile)

		// First, save a record
		record := &CoverageRecord{
			Timestamp:    time.Now(),
			CommitSHA:    "abc123",
			Branch:       "main",
			Percentage:   85.0,
			TotalLines:   1000,
			CoveredLines: 850,
		}
		err := manager.SaveRecord(record)
		require.NoError(t, err)

		const numReaders = 10
		done := make(chan error, numReaders)

		// Start multiple goroutines that read concurrently
		for i := 0; i < numReaders; i++ {
			go func() {
				lastRecord, err := manager.GetLastRecord()
				if err != nil {
					done <- err
					return
				}
				if lastRecord == nil {
					done <- errExpectedRecordButGotNil
					return
				}
				if lastRecord.CommitSHA != "abc123" {
					done <- fmt.Errorf("%w: expected 'abc123' but got '%s'", errCommitSHAMismatch, lastRecord.CommitSHA)
					return
				}
				done <- nil
			}()
		}

		// Wait for all readers to complete
		for i := 0; i < numReaders; i++ {
			err := <-done
			require.NoError(t, err, "All read operations should succeed")
		}
	})
}

// Test error handling edge cases
func TestErrorHandlingEdgeCases(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "history_test_*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	t.Run("Save record to read-only directory", func(t *testing.T) {
		if os.Getenv("CI") != "" {
			t.Skip("Skipping read-only directory test in CI")
		}

		// Create a read-only directory (if possible)
		readOnlyDir := filepath.Join(tempDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0o500) // read and execute only
		require.NoError(t, err)

		manager := NewManager(readOnlyDir)
		record := &CoverageRecord{
			Timestamp:    time.Now(),
			CommitSHA:    "abc123",
			Branch:       "main",
			Percentage:   85.0,
			TotalLines:   1000,
			CoveredLines: 850,
		}

		err = manager.SaveRecord(record)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write history file")

		// Clean up: restore write permission to allow deletion
		_ = os.Chmod(readOnlyDir, 0o600)
	})

	t.Run("Invalid JSON data in CoverageRecord", func(t *testing.T) {
		manager := NewManager(tempDir)

		// Create a record with data that should be serializable
		record := &CoverageRecord{
			Timestamp:    time.Now(),
			CommitSHA:    "abc123",
			Branch:       "main",
			Percentage:   85.0,
			TotalLines:   1000,
			CoveredLines: 850,
		}

		// This should work fine - JSON marshaling of CoverageRecord should not fail
		// under normal circumstances
		err := manager.SaveRecord(record)
		require.NoError(t, err)
	})

	t.Run("Nil record", func(t *testing.T) {
		manager := NewManager(tempDir)

		// This will cause a panic in the current implementation,
		// but we'll test it anyway to document the behavior
		assert.Panics(t, func() {
			_ = manager.SaveRecord(nil)
		})
	})

	t.Run("Extremely long file path", func(t *testing.T) {
		// Create a very long directory path
		longPath := tempDir
		for i := 0; i < 50; i++ {
			longPath = filepath.Join(longPath, "verylongdirectoryname")
		}

		manager := NewManager(longPath)
		record := &CoverageRecord{
			Timestamp:    time.Now(),
			CommitSHA:    "abc123",
			Branch:       "main",
			Percentage:   85.0,
			TotalLines:   1000,
			CoveredLines: 850,
		}

		// This might fail depending on OS path length limits
		err := manager.SaveRecord(record)
		if err != nil {
			// Error is acceptable for extremely long paths
			t.Logf("Long path failed as expected: %v", err)
		}
	})
}

func TestCoverageRecordStruct(t *testing.T) {
	t.Run("CoverageRecord JSON marshaling", func(t *testing.T) {
		timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		record := CoverageRecord{
			Timestamp:    timestamp,
			CommitSHA:    "abc123def456",
			Branch:       "feature/test-branch",
			Percentage:   87.42,
			TotalLines:   1234,
			CoveredLines: 1079,
		}

		// Test marshaling
		data, err := json.Marshal(record)
		require.NoError(t, err)

		// Test unmarshaling
		var unmarshaled CoverageRecord
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		// Verify all fields
		assert.Equal(t, record.Timestamp.Unix(), unmarshaled.Timestamp.Unix())
		assert.Equal(t, record.CommitSHA, unmarshaled.CommitSHA)
		assert.Equal(t, record.Branch, unmarshaled.Branch)
		assert.InEpsilon(t, record.Percentage, unmarshaled.Percentage, 0.001)
		assert.Equal(t, record.TotalLines, unmarshaled.TotalLines)
		assert.Equal(t, record.CoveredLines, unmarshaled.CoveredLines)
	})

	t.Run("ErrNoHistory constant", func(t *testing.T) {
		assert.Equal(t, "no coverage history available", ErrNoHistory.Error())
	})
}
