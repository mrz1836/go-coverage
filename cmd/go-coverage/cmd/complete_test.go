package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMainBranches(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected []string
	}{
		{
			name:     "default branches",
			envValue: "",
			expected: []string{"master", "main"},
		},
		{
			name:     "single custom branch",
			envValue: "develop",
			expected: []string{"develop"},
		},
		{
			name:     "multiple custom branches",
			envValue: "master,main,develop",
			expected: []string{"master", "main", "develop"},
		},
		{
			name:     "branches with spaces",
			envValue: "master, main , develop ",
			expected: []string{"master", "main", "develop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			original := os.Getenv("MAIN_BRANCHES")
			defer func() {
				if original != "" {
					require.NoError(t, os.Setenv("MAIN_BRANCHES", original))
				} else {
					require.NoError(t, os.Unsetenv("MAIN_BRANCHES"))
				}
			}()

			if tt.envValue != "" {
				require.NoError(t, os.Setenv("MAIN_BRANCHES", tt.envValue))
			} else {
				require.NoError(t, os.Unsetenv("MAIN_BRANCHES"))
			}

			actual := getMainBranches()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestGetPrimaryMainBranch(t *testing.T) {
	tests := []struct {
		name              string
		defaultMainBranch string
		mainBranches      string
		expected          string
	}{
		{
			name:              "default branch env set",
			defaultMainBranch: "develop",
			mainBranches:      "master,main",
			expected:          "develop",
		},
		{
			name:              "use first main branch",
			defaultMainBranch: "",
			mainBranches:      "develop,main,master",
			expected:          "develop",
		},
		{
			name:              "fallback to master",
			defaultMainBranch: "",
			mainBranches:      "",
			expected:          "master",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			originalDefault := os.Getenv("DEFAULT_MAIN_BRANCH")
			originalMain := os.Getenv("MAIN_BRANCHES")
			defer func() {
				if originalDefault != "" {
					require.NoError(t, os.Setenv("DEFAULT_MAIN_BRANCH", originalDefault))
				} else {
					require.NoError(t, os.Unsetenv("DEFAULT_MAIN_BRANCH"))
				}
				if originalMain != "" {
					require.NoError(t, os.Setenv("MAIN_BRANCHES", originalMain))
				} else {
					require.NoError(t, os.Unsetenv("MAIN_BRANCHES"))
				}
			}()

			if tt.defaultMainBranch != "" {
				require.NoError(t, os.Setenv("DEFAULT_MAIN_BRANCH", tt.defaultMainBranch))
			} else {
				require.NoError(t, os.Unsetenv("DEFAULT_MAIN_BRANCH"))
			}

			if tt.mainBranches != "" {
				require.NoError(t, os.Setenv("MAIN_BRANCHES", tt.mainBranches))
			} else {
				require.NoError(t, os.Unsetenv("MAIN_BRANCHES"))
			}

			actual := getPrimaryMainBranch()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestGetDefaultBranch(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "github ref name set",
			envValue: "feature-branch",
			expected: "feature-branch",
		},
		{
			name:     "no env variable set",
			envValue: "",
			expected: "master", // history.DefaultBranch
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			original := os.Getenv("GITHUB_REF_NAME")
			defer func() {
				if original != "" {
					require.NoError(t, os.Setenv("GITHUB_REF_NAME", original))
				} else {
					require.NoError(t, os.Unsetenv("GITHUB_REF_NAME"))
				}
			}()

			if tt.envValue != "" {
				require.NoError(t, os.Setenv("GITHUB_REF_NAME", tt.envValue))
			} else {
				require.NoError(t, os.Unsetenv("GITHUB_REF_NAME"))
			}

			actual := getDefaultBranch()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestGetStatusIcon(t *testing.T) {
	tests := []struct {
		name      string
		coverage  float64
		threshold float64
		expected  string
	}{
		{
			name:      "below threshold",
			coverage:  70.0,
			threshold: 80.0,
			expected:  "🔴 Below Threshold",
		},
		{
			name:      "excellent coverage",
			coverage:  95.0,
			threshold: 80.0,
			expected:  "🟢 Excellent",
		},
		{
			name:      "good coverage",
			coverage:  85.0,
			threshold: 80.0,
			expected:  "🟡 Good",
		},
		{
			name:      "fair coverage",
			coverage:  75.0,
			threshold: 70.0,
			expected:  "🟠 Fair",
		},
		{
			name:      "needs improvement",
			coverage:  65.0,
			threshold: 60.0,
			expected:  "🔴 Needs Improvement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getStatusIcon(tt.coverage, tt.threshold)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestCompleteCommandFlags(t *testing.T) {
	// Create a Commands instance for testing
	versionInfo := VersionInfo{
		Version:   "test",
		Commit:    "test-commit",
		BuildDate: "test-date",
	}
	commands := NewCommands(versionInfo)

	// Test that all expected flags exist and have correct defaults
	expectedFlags := map[string]struct {
		flagType     string
		defaultValue string
	}{
		"input":        {"string", ""},
		"output":       {"string", ""},
		"skip-history": {"bool", "false"},
		"skip-github":  {"bool", "false"},
		"dry-run":      {"bool", "false"},
	}

	for flagName, expected := range expectedFlags {
		t.Run(fmt.Sprintf("flag_%s", flagName), func(t *testing.T) {
			flag := commands.Complete.Flags().Lookup(flagName)
			require.NotNil(t, flag, "Flag %s should exist", flagName)
			assert.Equal(t, expected.flagType, flag.Value.Type(), "Flag %s should be %s type", flagName, expected.flagType)
			assert.Equal(t, expected.defaultValue, flag.DefValue, "Flag %s should have default %s", flagName, expected.defaultValue)
		})
	}
}

func TestCompleteCommandMetadata(t *testing.T) {
	// Test command metadata
	// Create a Commands instance for testing
	versionInfo := VersionInfo{
		Version:   "test",
		Commit:    "test-commit",
		BuildDate: "test-date",
	}
	commands := NewCommands(versionInfo)

	assert.Equal(t, "complete", commands.Complete.Use)
	assert.Equal(t, "Run complete coverage pipeline", commands.Complete.Short)
	assert.Contains(t, commands.Complete.Long, "Run the complete coverage pipeline")
	assert.NotNil(t, commands.Complete.RunE)
}

func TestCompleteCommandDryRun(t *testing.T) {
	// Test dry run functionality with a simple coverage file
	tempDir := t.TempDir()
	coverageFile := filepath.Join(tempDir, "coverage.txt")
	outputDir := filepath.Join(tempDir, "output")

	// Create a simple coverage file with good coverage
	coverageContent := `mode: set
github.com/test/repo/main.go:10.2,12.16 2 2
github.com/test/repo/main.go:15.2,17.16 3 3
github.com/test/repo/utils.go:5.1,7.2 2 2
github.com/test/repo/utils.go:8.1,10.2 2 2
`
	require.NoError(t, os.WriteFile(coverageFile, []byte(coverageContent), 0o600))

	// Create command with dry-run flag
	var buf bytes.Buffer
	// Create a Commands instance for testing
	versionInfo := VersionInfo{
		Version:   "test",
		Commit:    "test-commit",
		BuildDate: "test-date",
	}
	commands := NewCommands(versionInfo)

	testCmd := &cobra.Command{
		Use:  "complete",
		RunE: commands.Complete.RunE,
	}
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)

	// Add flags
	testCmd.Flags().AddFlagSet(commands.Complete.Flags())

	// Set arguments for dry run
	testCmd.SetArgs([]string{
		"--input", coverageFile,
		"--output", outputDir,
		"--dry-run",
		"--skip-github",
	})

	// Set required environment variables to avoid config validation errors
	originalEnvs := map[string]string{
		"GITHUB_REPOSITORY":       os.Getenv("GITHUB_REPOSITORY"),
		"GITHUB_REPOSITORY_OWNER": os.Getenv("GITHUB_REPOSITORY_OWNER"),
		"GITHUB_SHA":              os.Getenv("GITHUB_SHA"),
		"GITHUB_TOKEN":            os.Getenv("GITHUB_TOKEN"),
		"GO_COVERAGE_THRESHOLD":   os.Getenv("GO_COVERAGE_THRESHOLD"),
	}
	defer func() {
		for key, val := range originalEnvs {
			if val != "" {
				require.NoError(t, os.Setenv(key, val))
			} else {
				require.NoError(t, os.Unsetenv(key))
			}
		}
	}()

	require.NoError(t, os.Setenv("GITHUB_REPOSITORY", "test/repo"))
	require.NoError(t, os.Setenv("GITHUB_REPOSITORY_OWNER", "test"))
	require.NoError(t, os.Setenv("GITHUB_SHA", "abc123"))
	require.NoError(t, os.Setenv("GITHUB_TOKEN", "test-token"))
	require.NoError(t, os.Setenv("GO_COVERAGE_THRESHOLD", "0.0")) // Disable threshold for test

	// Execute command
	err := testCmd.Execute()

	// In dry run mode, it should complete successfully
	require.NoError(t, err)

	// Check that output contains dry run indicators
	output := buf.String()
	assert.Contains(t, output, "Mode: DRY RUN")
	assert.Contains(t, output, "Starting Go Coverage Pipeline")
	assert.Contains(t, output, "Step 1: Parsing coverage data")
	assert.Contains(t, output, "Step 2: Generating coverage badge")
	assert.Contains(t, output, "Coverage:")

	// Verify no actual files were created in dry run mode
	_, err = os.Stat(outputDir)
	assert.True(t, os.IsNotExist(err), "Output directory should not be created in dry run")
}

func TestErrCoverageBelowThreshold(t *testing.T) {
	assert.Equal(t, "coverage is below threshold", ErrCoverageBelowThreshold.Error())
}

func TestErrEmptyIndexHTML(t *testing.T) {
	assert.Equal(t, "generated index.html is empty", ErrEmptyIndexHTML.Error())
}

// Test the copyDir and copyFile functions
func TestCopyDir(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	// Create source directory structure
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "subdir"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0o600))

	// Create a dummy command for logging
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Test copyDir
	err := copyDir(cmd, srcDir, dstDir)
	require.NoError(t, err)

	// Verify files were copied
	content1, err := os.ReadFile(filepath.Clean(filepath.Join(dstDir, "file1.txt")))
	require.NoError(t, err)
	assert.Equal(t, "content1", string(content1))

	content2, err := os.ReadFile(filepath.Clean(filepath.Join(dstDir, "subdir", "file2.txt")))
	require.NoError(t, err)
	assert.Equal(t, "content2", string(content2))
}

func TestCopyDirErrors(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Test with non-existent source
	err := copyDir(cmd, "/nonexistent", "/tmp/test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to stat source directory")
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "src.txt")
	dstFile := filepath.Join(tempDir, "dst.txt")

	// Create source file
	content := "test content"
	require.NoError(t, os.WriteFile(srcFile, []byte(content), 0o600))

	// Create a dummy command for logging
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Test copyFile
	err := copyFile(cmd, srcFile, dstFile)
	require.NoError(t, err)

	// Verify file was copied
	copiedContent, err := os.ReadFile(filepath.Clean(dstFile))
	require.NoError(t, err)
	assert.Equal(t, content, string(copiedContent))
}

func TestCopyFileErrors(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Test with non-existent source
	err := copyFile(cmd, "/nonexistent.txt", "/tmp/test.txt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open source file")
}
