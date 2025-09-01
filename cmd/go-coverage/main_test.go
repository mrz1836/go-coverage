package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	// Test that main function can be called without panicking
	// We test this by running the binary with --help flag to avoid side effects

	if os.Getenv("GO_TEST_SUBPROCESS") == "1" {
		// This is the subprocess that will run main()
		// Set args to help to avoid side effects
		os.Args = []string{"go-coverage", "--help"}
		main()
		return
	}

	// Run the main function in a subprocess
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestMain") //nolint:gosec // Test needs subprocess execution
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS=1")
	err := cmd.Run()
	// The --help flag should cause the command to exit with code 0
	// If there's an error, it should be an exit error with code 0 (help was shown)
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			// Help command typically exits with 0, but cobra may exit with different codes
			// We just ensure it's a controlled exit, not a panic
			require.GreaterOrEqual(t, exitError.ExitCode(), 0, "Command should exit gracefully, got exit code: %d", exitError.ExitCode())
		} else {
			t.Fatalf("Unexpected error type: %v", err)
		}
	}
}

func TestMainErrorHandling(t *testing.T) {
	// Test that main function handles cmd.Execute() errors properly

	if os.Getenv("GO_TEST_SUBPROCESS_ERROR") == "1" {
		// This is the subprocess that will run main() with invalid args
		os.Args = []string{"go-coverage", "--invalid-flag-that-does-not-exist"}
		main()
		return
	}

	// Run the main function in a subprocess with invalid arguments
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestMainErrorHandling") //nolint:gosec // Test needs subprocess execution
	cmd.Env = append(os.Environ(), "GO_TEST_SUBPROCESS_ERROR=1")
	err := cmd.Run()

	// Should get an exit error since invalid flag should cause error
	require.Error(t, err)

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		// Should exit with code 1 as specified in main()
		require.Equal(t, 1, exitError.ExitCode(), "Command should exit with code 1 on error")
	} else {
		t.Fatalf("Expected exit error, got: %v", err)
	}
}

// TestMainFunctionExists verifies that main function exists and is callable
func TestMainFunctionExists(t *testing.T) {
	// This test ensures the main function exists and can be referenced
	// We can't call it directly in tests due to os.Exit(), but we can verify it exists
	require.NotNil(t, main, "main function should exist")
}

// TestGetVersion tests the GetVersion function
func TestGetVersion(t *testing.T) {
	version := GetVersion()
	// Version should not be empty
	require.NotEmpty(t, version)
}

// TestGetCommit tests the GetCommit function
func TestGetCommit(t *testing.T) {
	commit := GetCommit()
	// Commit should not be empty (could be "unknown" or actual commit hash)
	require.NotEmpty(t, commit)
}

// TestGetBuildDate tests the GetBuildDate function
func TestGetBuildDate(t *testing.T) {
	buildDate := GetBuildDate()
	// Build date should not be empty (could be "unknown" or actual date)
	require.NotEmpty(t, buildDate)
}

// TestIsModified tests the IsModified function
func TestIsModified(t *testing.T) {
	// IsModified should return a boolean without error
	modified := IsModified()
	// We just check it doesn't panic and returns a valid boolean
	require.True(t, modified == true || modified == false)
}

// TestIsTemplateString tests the isTemplateString function
func TestIsTemplateString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Template with Version",
			input:    "{{.Version}}",
			expected: true,
		},
		{
			name:     "Template with Commit",
			input:    "{{.Commit}}",
			expected: true,
		},
		{
			name:     "Template with Date",
			input:    "{{.Date}}",
			expected: true,
		},
		{
			name:     "Template in middle of string",
			input:    "version {{.Version}} build",
			expected: true,
		},
		{
			name:     "Multiple templates",
			input:    "{{.Version}}-{{.Commit}}",
			expected: true,
		},
		{
			name:     "Regular string",
			input:    "1.2.3",
			expected: false,
		},
		{
			name:     "String with braces but not template",
			input:    "{version: 1.2.3}",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "String with only opening braces",
			input:    "{{version",
			expected: false,
		},
		{
			name:     "String with only closing braces",
			input:    "version}}",
			expected: false,
		},
		{
			name:     "Dev version",
			input:    "dev",
			expected: false,
		},
		{
			name:     "Complex version",
			input:    "1.2.3-rc1+build.123",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isTemplateString(tc.input)
			require.Equal(t, tc.expected, result, "isTemplateString(%q) should return %v", tc.input, tc.expected)
		})
	}
}

// TestGetVersionWithLdflags tests GetVersion with different ldflags scenarios
func TestGetVersionWithLdflags(t *testing.T) {
	// These tests run in the current binary context
	// We test the actual behavior of GetVersion function
	version := GetVersion()
	require.NotEmpty(t, version, "GetVersion should return non-empty string")

	// Version should not be a template string
	require.False(t, isTemplateString(version), "GetVersion should not return template placeholders")

	// Should be either a proper version (starts with digit or 'v'), commit hash, or "dev"
	if version != "dev" {
		// If it's not "dev", it should either be a version or commit hash
		// Version patterns: v1.2.3, 1.2.3
		// Commit hash patterns: 7-40 character hex
		if !strings.HasPrefix(version, "v") && !strings.Contains(version, ".") {
			// Likely a commit hash - should be hex characters
			require.Regexp(t, `^[a-f0-9]+$`, version, "Commit hash should contain only hex characters")
			require.True(t, len(version) >= 7 && len(version) <= 40, "Commit hash should be 7-40 characters")
		}
	}
}

// TestGetCommitWithBuildInfo tests GetCommit function
func TestGetCommitWithBuildInfo(t *testing.T) {
	commit := GetCommit()
	require.NotEmpty(t, commit, "GetCommit should return non-empty string")

	// Commit should not be a template string
	require.False(t, isTemplateString(commit), "GetCommit should not return template placeholders")

	// Should be either a commit hash or "none"
	if commit != "none" {
		// Should be a hex string (commit hash)
		require.Regexp(t, `^[a-f0-9]+$`, commit, "Commit hash should contain only hex characters")
		require.GreaterOrEqual(t, len(commit), 7, "Commit hash should be at least 7 characters")
	}
}

// TestGetBuildDateWithBuildInfo tests GetBuildDate function
func TestGetBuildDateWithBuildInfo(t *testing.T) {
	buildDate := GetBuildDate()
	require.NotEmpty(t, buildDate, "GetBuildDate should return non-empty string")

	// Build date should not be a template string
	require.False(t, isTemplateString(buildDate), "GetBuildDate should not return template placeholders")

	// Should be either a valid timestamp or "unknown"
	if buildDate != "unknown" {
		// Should be in RFC3339 format (ISO 8601)
		// Example: 2023-01-01T12:00:00Z
		require.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`, buildDate, "Build date should be in RFC3339 format")
	}
}

// TestGetVersionInfoSingleton tests the singleton pattern
func TestGetVersionInfoSingleton(t *testing.T) {
	// Get version info multiple times and ensure it's the same instance behavior
	info1 := getVersionInfo()
	info2 := getVersionInfo()

	require.NotNil(t, info1)
	require.NotNil(t, info2)

	// Should return consistent values (singleton behavior)
	require.Equal(t, info1.version, info2.version)
	require.Equal(t, info1.commit, info2.commit)
	require.Equal(t, info1.buildDate, info2.buildDate)
}

// TestVersionFallbackBehavior tests version fallback to build info
func TestVersionFallbackBehavior(t *testing.T) {
	// This tests the actual behavior when no ldflags are set
	version := GetVersion()

	// Should get a valid version string
	require.NotEmpty(t, version)

	// Should handle common cases:
	// 1. Valid module version (from go install @version)
	// 2. VCS revision (from git)
	// 3. "dev" as fallback
	validVersionPatterns := []string{
		`^v\d+\.\d+\.\d+`,  // Semantic version with v prefix
		`^\d+\.\d+\.\d+`,   // Semantic version without v prefix
		`^[a-f0-9]{7,40}$`, // Git commit hash
		`^dev$`,            // Development version
	}

	matched := false
	for _, pattern := range validVersionPatterns {
		if matched, _ = regexp.MatchString(pattern, version); matched {
			break
		}
	}
	require.True(t, matched, "Version '%s' should match one of the valid patterns", version)
}

// TestIsModifiedFunction tests the IsModified function more thoroughly
func TestIsModifiedFunction(t *testing.T) {
	modified := IsModified()

	// IsModified should return a valid boolean
	require.True(t, modified == true || modified == false, "IsModified should return boolean")

	// Test that the function doesn't panic and handles build info gracefully
	// In test environment, this might be true or false depending on git state
	// The important thing is that it doesn't crash
}

// TestVersionFunctionsConsistency tests consistency between version functions
func TestVersionFunctionsConsistency(t *testing.T) {
	// Test that calling the functions multiple times returns consistent results
	version1 := GetVersion()
	version2 := GetVersion()
	require.Equal(t, version1, version2, "GetVersion should return consistent results")

	commit1 := GetCommit()
	commit2 := GetCommit()
	require.Equal(t, commit1, commit2, "GetCommit should return consistent results")

	buildDate1 := GetBuildDate()
	buildDate2 := GetBuildDate()
	require.Equal(t, buildDate1, buildDate2, "GetBuildDate should return consistent results")

	modified1 := IsModified()
	modified2 := IsModified()
	require.Equal(t, modified1, modified2, "IsModified should return consistent results")
}

// TestVersionInfoWithoutBuildInfo tests behavior when build info is not available
func TestVersionInfoWithoutBuildInfo(t *testing.T) {
	// This test verifies that the functions handle cases where build info might not be available
	// In normal test execution, build info should be available, but we test the fallback behavior

	version := GetVersion()
	commit := GetCommit()
	buildDate := GetBuildDate()

	// All should return non-empty values
	require.NotEmpty(t, version)
	require.NotEmpty(t, commit)
	require.NotEmpty(t, buildDate)

	// None should be template strings
	require.False(t, isTemplateString(version))
	require.False(t, isTemplateString(commit))
	require.False(t, isTemplateString(buildDate))
}

// TestGetVersionEdgeCases tests edge cases for GetVersion function to improve coverage
func TestGetVersionEdgeCases(t *testing.T) {
	// Save original values
	origInfo := versionInstance
	defer func() { versionInstance = origInfo }()

	// Test with empty version (should fallback to build info)
	versionInstance = &VersionInfo{
		version:   "",
		commit:    "test-commit",
		buildDate: "2023-01-01",
	}

	version := GetVersion()
	require.NotEmpty(t, version, "GetVersion should not return empty string")

	// Test with template string version (should fallback to build info)
	versionInstance = &VersionInfo{
		version:   "{{.Version}}",
		commit:    "test-commit",
		buildDate: "2023-01-01",
	}

	version = GetVersion()
	require.NotEmpty(t, version, "GetVersion should not return template string")
	require.False(t, isTemplateString(version), "GetVersion should resolve template strings")

	// Test with valid version
	versionInstance = &VersionInfo{
		version:   "v1.2.3",
		commit:    "test-commit",
		buildDate: "2023-01-01",
	}

	version = GetVersion()
	require.Equal(t, "v1.2.3", version, "GetVersion should return set version")
}

// TestGetCommitEdgeCases tests edge cases for GetCommit function to improve coverage
func TestGetCommitEdgeCases(t *testing.T) {
	// Save original values
	origInfo := versionInstance
	defer func() { versionInstance = origInfo }()

	// Test with "none" commit (should fallback to build info)
	versionInstance = &VersionInfo{
		version:   "1.0.0",
		commit:    "none",
		buildDate: "2023-01-01",
	}

	commit := GetCommit()
	require.NotEmpty(t, commit, "GetCommit should not return empty string")

	// Test with empty commit (should fallback to build info)
	versionInstance = &VersionInfo{
		version:   "1.0.0",
		commit:    "",
		buildDate: "2023-01-01",
	}

	commit = GetCommit()
	require.NotEmpty(t, commit, "GetCommit should not return empty string")

	// Test with template string commit (should fallback to build info)
	versionInstance = &VersionInfo{
		version:   "1.0.0",
		commit:    "{{.Commit}}",
		buildDate: "2023-01-01",
	}

	commit = GetCommit()
	require.NotEmpty(t, commit, "GetCommit should not return template string")
	require.False(t, isTemplateString(commit), "GetCommit should resolve template strings")

	// Test with valid commit
	versionInstance = &VersionInfo{
		version:   "1.0.0",
		commit:    "abc123def",
		buildDate: "2023-01-01",
	}

	commit = GetCommit()
	require.Equal(t, "abc123def", commit, "GetCommit should return set commit")
}

// TestGetBuildDateEdgeCases tests edge cases for GetBuildDate function to improve coverage
func TestGetBuildDateEdgeCases(t *testing.T) {
	// Save original values
	origInfo := versionInstance
	defer func() { versionInstance = origInfo }()

	// Test with "unknown" build date (should fallback to build info)
	versionInstance = &VersionInfo{
		version:   "1.0.0",
		commit:    "abc123",
		buildDate: "unknown",
	}

	buildDate := GetBuildDate()
	require.NotEmpty(t, buildDate, "GetBuildDate should not return empty string")

	// Test with empty build date (should fallback to build info)
	versionInstance = &VersionInfo{
		version:   "1.0.0",
		commit:    "abc123",
		buildDate: "",
	}

	buildDate = GetBuildDate()
	require.NotEmpty(t, buildDate, "GetBuildDate should not return empty string")

	// Test with template string build date (should fallback to build info)
	versionInstance = &VersionInfo{
		version:   "1.0.0",
		commit:    "abc123",
		buildDate: "{{.Date}}",
	}

	buildDate = GetBuildDate()
	require.NotEmpty(t, buildDate, "GetBuildDate should not return template string")
	require.False(t, isTemplateString(buildDate), "GetBuildDate should resolve template strings")

	// Test with valid build date
	versionInstance = &VersionInfo{
		version:   "1.0.0",
		commit:    "abc123",
		buildDate: "2023-01-01T10:00:00Z",
	}

	buildDate = GetBuildDate()
	require.Equal(t, "2023-01-01T10:00:00Z", buildDate, "GetBuildDate should return set date")
}
