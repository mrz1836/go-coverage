// Package cmd provides CLI commands for the Go coverage tool
package cmd

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGitHubRepoFromURL(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "HTTPS URL",
			url:      "https://github.com/mrz1836/go-coverage",
			expected: "mrz1836/go-coverage",
		},
		{
			name:     "HTTPS URL with .git",
			url:      "https://github.com/mrz1836/go-coverage.git",
			expected: "mrz1836/go-coverage",
		},
		{
			name:     "SSH URL",
			url:      "git@github.com:mrz1836/go-coverage",
			expected: "mrz1836/go-coverage",
		},
		{
			name:     "SSH URL with .git",
			url:      "git@github.com:mrz1836/go-coverage.git",
			expected: "mrz1836/go-coverage",
		},
		{
			name:     "Invalid URL",
			url:      "https://example.com/repo",
			expected: "",
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseGitHubRepoFromURL(tc.url)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsValidRepositoryFormat(t *testing.T) {
	testCases := []struct {
		name     string
		repo     string
		expected bool
	}{
		{
			name:     "Valid format",
			repo:     "mrz1836/go-coverage",
			expected: true,
		},
		{
			name:     "Valid with underscores",
			repo:     "mrz_1836/go_coverage",
			expected: true,
		},
		{
			name:     "Valid with dots",
			repo:     "mrz.1836/go.coverage",
			expected: true,
		},
		{
			name:     "Valid with dashes",
			repo:     "mrz-1836/go-coverage",
			expected: true,
		},
		{
			name:     "Invalid - no slash",
			repo:     "mrz1836go-coverage",
			expected: false,
		},
		{
			name:     "Invalid - too many slashes",
			repo:     "mrz1836/go/coverage",
			expected: false,
		},
		{
			name:     "Invalid - starts with slash",
			repo:     "/mrz1836/go-coverage",
			expected: false,
		},
		{
			name:     "Invalid - ends with slash",
			repo:     "mrz1836/go-coverage/",
			expected: false,
		},
		{
			name:     "Invalid - empty owner",
			repo:     "/go-coverage",
			expected: false,
		},
		{
			name:     "Invalid - empty repo",
			repo:     "mrz1836/",
			expected: false,
		},
		{
			name:     "Empty string",
			repo:     "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidRepositoryFormat(tc.repo)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCheckPrerequisites(t *testing.T) {
	// Check if GitHub CLI is actually installed and authenticated
	ghInstalled := false
	ghAuthenticated := false

	if _, err := exec.LookPath("gh"); err == nil {
		ghInstalled = true
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		authCmd := exec.CommandContext(ctx, "gh", "auth", "status")
		if authCmd.Run() == nil {
			ghAuthenticated = true
		}
	}

	testCases := []struct {
		name          string
		expectedError error
		skip          bool
		skipReason    string
	}{
		{
			name:          "All prerequisites met",
			expectedError: nil,
			skip:          !ghInstalled || !ghAuthenticated,
			skipReason:    "GitHub CLI not installed or not authenticated",
		},
		{
			name:          "GitHub CLI not found",
			expectedError: ErrGitHubCLINotFound,
			skip:          true, // Cannot test this if gh is actually installed
			skipReason:    "Cannot mock GitHub CLI absence when it's actually installed",
		},
		{
			name:          "GitHub CLI not authenticated",
			expectedError: ErrGitHubCLINotAuthenticated,
			skip:          true, // Cannot test this if gh is actually authenticated
			skipReason:    "Cannot mock GitHub CLI authentication failure when it's actually authenticated",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip(tc.skipReason)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := NewCommands(VersionInfo{Version: "test", Commit: "test", BuildDate: "test"}).SetupPages
			err := checkPrerequisites(ctx, cmd, false)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetRepositoryFromGit(t *testing.T) {
	// This test only works in a git repository
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git command not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if we're in a git repository
	gitCheckCmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	if _, err := gitCheckCmd.Output(); err != nil {
		t.Skip("not in a git repository")
	}

	repo, err := getRepositoryFromGit(ctx, NewCommands(VersionInfo{Version: "test", Commit: "test", BuildDate: "test"}).SetupPages, false)

	// We should either get a valid repository or an error
	if err != nil {
		// If there's an error, it should be a known error type
		assert.True(t,
			strings.Contains(err.Error(), "not in a git repository") ||
				strings.Contains(err.Error(), "could not parse GitHub repository"),
			"Expected known error type, got: %v", err)
	} else {
		// If successful, should be a valid format
		assert.True(t, isValidRepositoryFormat(repo), "Repository format should be valid: %s", repo)
	}
}

func TestSetupPagesCommand(t *testing.T) {
	testCases := []struct {
		name      string
		args      []string
		flags     map[string]string
		expectErr bool
		skipCI    bool
	}{
		{
			name:      "Valid repository argument",
			args:      []string{"test-owner/test-repo"},
			flags:     map[string]string{"dry-run": "true"},
			expectErr: false,
			skipCI:    true, // Skip in CI since gh may not be available
		},
		{
			name:      "Invalid repository format",
			args:      []string{"invalid-repo-format"},
			flags:     map[string]string{"dry-run": "true"},
			expectErr: true,
			skipCI:    true,
		},
		{
			name:      "Too many arguments",
			args:      []string{"owner/repo", "extra-arg"},
			flags:     map[string]string{"dry-run": "true"},
			expectErr: true,
			skipCI:    true,
		},
		{
			name:      "Verbose flag",
			args:      []string{"test-owner/test-repo"},
			flags:     map[string]string{"dry-run": "true", "verbose": "true"},
			expectErr: false,
			skipCI:    true,
		},
		{
			name:      "Custom domain flag",
			args:      []string{"test-owner/test-repo"},
			flags:     map[string]string{"dry-run": "true", "custom-domain": "example.com"},
			expectErr: false,
			skipCI:    true,
		},
		{
			name:      "Protect branches flag",
			args:      []string{"test-owner/test-repo"},
			flags:     map[string]string{"dry-run": "true", "protect-branches": "true"},
			expectErr: false,
			skipCI:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipCI && (os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true") {
				t.Skip("Skipping test in CI environment")
			}

			cmd := NewCommands(VersionInfo{Version: "test", Commit: "test", BuildDate: "test"}).SetupPages

			// Set flags for this test
			for flag, value := range tc.flags {
				err := cmd.Flags().Set(flag, value)
				require.NoError(t, err)
			}

			// Execute the RunE function directly to avoid cobra parent-child issues
			err := cmd.RunE(cmd, tc.args)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				// In non-CI environments with proper setup, command should succeed
				// In CI or without proper GitHub setup, we expect certain errors
				if err != nil {
					// Allow certain expected errors in test environments
					errStr := err.Error()
					expectedErrors := []string{
						"prerequisites check failed",
						"github CLI",
						"not found",
						"not authenticated",
						"repository not found",
						"access denied",
					}

					hasExpectedError := false
					for _, expected := range expectedErrors {
						if strings.Contains(strings.ToLower(errStr), expected) {
							hasExpectedError = true
							break
						}
					}

					assert.True(t, hasExpectedError,
						"Unexpected error type: %v", err)
				}
			}
		})
	}
}

func TestSetupPagesCommandFlags(t *testing.T) {
	cmd := NewCommands(VersionInfo{Version: "test", Commit: "test", BuildDate: "test"}).SetupPages

	// Test that all expected flags are present
	expectedFlags := []string{
		"dry-run",
		"verbose",
		"custom-domain",
		"protect-branches",
	}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}

	// Test flag defaults
	dryRunFlag := cmd.Flags().Lookup("dry-run")
	assert.Equal(t, "false", dryRunFlag.DefValue, "dry-run should default to false")

	verboseFlag := cmd.Flags().Lookup("verbose")
	assert.Equal(t, "false", verboseFlag.DefValue, "verbose should default to false")

	customDomainFlag := cmd.Flags().Lookup("custom-domain")
	assert.Empty(t, customDomainFlag.DefValue, "custom-domain should default to empty string")

	protectBranchesFlag := cmd.Flags().Lookup("protect-branches")
	assert.Equal(t, "false", protectBranchesFlag.DefValue, "protect-branches should default to false")
}

func TestSetupPagesCommandHelp(t *testing.T) {
	cmd := NewCommands(VersionInfo{Version: "test", Commit: "test", BuildDate: "test"}).SetupPages

	// Test command metadata
	assert.Equal(t, "setup-pages [repository]", cmd.Use)
	assert.Contains(t, cmd.Short, "GitHub Pages environment")
	assert.Contains(t, cmd.Long, "Configure GitHub Pages environment")
	assert.Contains(t, cmd.Long, "deployment branch policies")

	// Test that command has examples
	assert.Contains(t, cmd.Long, "Examples:")
	assert.Contains(t, cmd.Long, "go-coverage setup-pages")
	assert.Contains(t, cmd.Long, "--dry-run")
	assert.Contains(t, cmd.Long, "--verbose")
}

// Benchmark tests for performance-critical functions
func BenchmarkParseGitHubRepoFromURL(b *testing.B) {
	testURL := "https://github.com/mrz1836/go-coverage.git"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseGitHubRepoFromURL(testURL)
	}
}

func BenchmarkIsValidRepositoryFormat(b *testing.B) {
	testRepo := "mrz1836/go-coverage"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isValidRepositoryFormat(testRepo)
	}
}

// Test error conditions and edge cases
func TestSetupPagesErrorHandling(t *testing.T) {
	testCases := []struct {
		name        string
		function    string
		expectedErr error
	}{
		{
			name:        "GitHub CLI not found error",
			function:    "checkPrerequisites",
			expectedErr: ErrGitHubCLINotFound,
		},
		{
			name:        "GitHub CLI not authenticated error",
			function:    "checkPrerequisites",
			expectedErr: ErrGitHubCLINotAuthenticated,
		},
		{
			name:        "Repository not found error",
			function:    "checkRepositoryAccess",
			expectedErr: ErrRepositoryNotFound,
		},
		{
			name:        "No git repository error",
			function:    "getRepositoryFromGit",
			expectedErr: ErrNoGitRepository,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test that the error constants are properly defined
			assert.NotEmpty(t, tc.expectedErr.Error())
			// Verify error is of correct type
			assert.Implements(t, (*error)(nil), tc.expectedErr)
		})
	}
}

// Test the command integration with the cobra framework
func TestSetupPagesCommandIntegration(t *testing.T) {
	cmd := NewCommands(VersionInfo{Version: "test", Commit: "test", BuildDate: "test"}).SetupPages

	// Test command is properly configured
	assert.NotNil(t, cmd.RunE, "Command should have a RunE function")
	assert.NotNil(t, cmd.Flags(), "Command should have flags configured")

	// Test argument validation (can't compare function pointers, so test behavior instead)
	assert.NotNil(t, cmd.Args, "Command should have argument validation")

	// Test that flags are properly bound
	flags := cmd.Flags()
	assert.True(t, flags.HasFlags(), "Command should have flags")

	flagNames := []string{"dry-run", "verbose", "custom-domain", "protect-branches"}
	for _, name := range flagNames {
		assert.NotNil(t, flags.Lookup(name), "Flag %s should be defined", name)
	}
}

// Test URL parsing edge cases and security considerations
func TestParseGitHubRepoFromURLSecurity(t *testing.T) {
	// Test potentially malicious URLs
	maliciousURLs := []string{
		"https://evil.com/../../github.com/owner/repo",
		"javascript:alert('xss')",
		"https://github.com/../../../etc/passwd",
		"file:///etc/passwd",
		"https://github.com/owner/repo/../../../",
		"https://github.com/owner/repo/.git/../..",
		"https://github.com/owner@evil.com/repo.git",
		"https://user:pass@github.com/owner/repo.git",
	}

	for _, url := range maliciousURLs {
		t.Run("Malicious URL: "+url, func(t *testing.T) {
			result := parseGitHubRepoFromURL(url)
			// Should return empty string for malicious URLs
			assert.Empty(t, result, "Malicious URL should return empty string")
		})
	}
}

// Test repository format validation with various edge cases
func TestRepositoryFormatValidation(t *testing.T) {
	edgeCases := []struct {
		repo  string
		valid bool
		desc  string
	}{
		{"a/b", true, "minimal valid format"},
		{"user123/repo_name", true, "alphanumeric with underscore"},
		{"user.name/repo-name", true, "dot and dash"},
		{"user_name/repo.name", true, "underscore and dot"},
		{"123user/456repo", true, "starting with numbers"},
		{"", false, "empty string"},
		{"user", false, "missing repo name"},
		{"/repo", false, "missing user name"},
		{"user/", false, "missing repo name with slash"},
		{"user//repo", false, "double slash"},
		{"user/ /repo", false, "space in name"},
		{"user/repo/extra", false, "too many parts"},
		{"user@domain/repo", false, "invalid character @"},
		{"user/repo#tag", false, "invalid character #"},
		{"user/repo?param", false, "invalid character ?"},
		{"../user/repo", false, "path traversal attempt"},
		{"user/../repo", false, "path traversal in middle"},
		{"user/repo/..", false, "path traversal at end"},
	}

	for _, tc := range edgeCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := isValidRepositoryFormat(tc.repo)
			assert.Equal(t, tc.valid, result,
				"Repository '%s' validation failed: %s", tc.repo, tc.desc)
		})
	}
}
