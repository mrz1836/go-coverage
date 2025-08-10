// Package urlutil provides utility functions for URL generation and formatting
// This file contains comprehensive fuzz tests for the urlutil package functions
package urlutil

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// FuzzBuildGitHubCommitURL tests the BuildGitHubCommitURL function with diverse inputs
// ensuring it never panics and produces valid output for valid inputs
func FuzzBuildGitHubCommitURL(f *testing.F) {
	// Seed corpus with typical and edge case inputs
	f.Add("", "", "")
	f.Add("owner", "repo", "abc123")
	f.Add("my-org", "my-repo", "1234567890abcdef")
	f.Add("owner with spaces", "repo", "commit")
	f.Add("owner", "repo with spaces", "commit")
	f.Add("owner", "repo", "commit with spaces")
	f.Add("\x00\x01\x02", "repo", "commit")  // null bytes
	f.Add("owner", "\x00\x01\x02", "commit") // null bytes in repo
	f.Add("owner", "repo", "\x00\x01\x02")   // null bytes in commit
	f.Add("unicode🔥", "repo", "commit")
	f.Add("owner", "unicode🔥", "commit")
	f.Add("owner", "repo", "unicode🔥")
	f.Add("../../../etc/passwd", "repo", "commit") // path traversal
	f.Add("owner", "../../../etc/passwd", "commit")
	f.Add("owner", "repo", "../../../etc/passwd")
	f.Add("javascript:alert(1)", "repo", "commit") // XSS attempt
	f.Add("owner", "javascript:alert(1)", "commit")
	f.Add("owner", "repo", "javascript:alert(1)")
	f.Add("\n\r\t", "repo", "commit") // control characters
	f.Add("owner", "\n\r\t", "commit")
	f.Add("owner", "repo", "\n\r\t")

	f.Fuzz(func(t *testing.T, owner, repo, commitSHA string) {
		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("BuildGitHubCommitURL panicked with inputs owner=%q, repo=%q, commitSHA=%q: %v", owner, repo, commitSHA, r)
			}
		}()

		result := BuildGitHubCommitURL(owner, repo, commitSHA)

		// Validate output format
		if owner == "" || repo == "" || commitSHA == "" {
			assert.Empty(t, result, "Should return empty string for empty inputs")
		} else {
			assert.NotEmpty(t, result, "Should return non-empty string for valid inputs")
			assert.True(t, strings.HasPrefix(result, "https://github.com/"), "Should start with GitHub URL prefix")
			assert.Contains(t, result, owner, "Should contain owner")
			assert.Contains(t, result, repo, "Should contain repo")
			assert.Contains(t, result, commitSHA, "Should contain commit SHA")
			assert.Contains(t, result, "/commit/", "Should contain commit path")

			// Check for proper structure
			expected := "https://github.com/" + owner + "/" + repo + "/commit/" + commitSHA
			assert.Equal(t, expected, result, "Should match expected URL format")
		}

		// Ensure result is valid UTF-8 if all inputs were valid UTF-8
		if utf8.ValidString(owner) && utf8.ValidString(repo) && utf8.ValidString(commitSHA) {
			assert.True(t, utf8.ValidString(result), "Result should be valid UTF-8 when inputs are valid UTF-8")
		}
	})
}

// FuzzBuildCoverageBadgeURL tests the BuildCoverageBadgeURL function with diverse percentage inputs
func FuzzBuildCoverageBadgeURL(f *testing.F) {
	// Seed corpus with typical and edge case inputs
	f.Add(0.0)
	f.Add(50.0)
	f.Add(75.0)
	f.Add(85.0)
	f.Add(90.0)
	f.Add(95.0)
	f.Add(100.0)
	f.Add(-1.0) // negative
	f.Add(-100.0)
	f.Add(101.0) // over 100%
	f.Add(1000.0)
	f.Add(0.1)  // very small
	f.Add(99.9) // close to 100
	f.Add(49.9) // boundary case
	f.Add(50.1) // boundary case

	f.Fuzz(func(t *testing.T, percentage float64) {
		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("BuildCoverageBadgeURL panicked with percentage=%f: %v", percentage, r)
			}
		}()

		result := BuildCoverageBadgeURL(percentage)

		// Validate output format
		assert.NotEmpty(t, result, "Should always return non-empty string")
		assert.True(t, strings.HasPrefix(result, "https://img.shields.io/badge/coverage-"), "Should start with shields.io badge URL")
		assert.Contains(t, result, "%25-", "Should contain URL-encoded percent sign")

		// Validate color assignment based on percentage
		switch {
		case percentage >= 90:
			assert.Contains(t, result, "brightgreen", "Should use brightgreen for >= 90%")
		case percentage >= 80:
			assert.Contains(t, result, "green", "Should use green for >= 80%")
		case percentage >= 70:
			assert.Contains(t, result, "yellowgreen", "Should use yellowgreen for >= 70%")
		case percentage >= 60:
			assert.Contains(t, result, "yellow", "Should use yellow for >= 60%")
		case percentage >= 50:
			assert.Contains(t, result, "orange", "Should use orange for >= 50%")
		default:
			assert.Contains(t, result, "red", "Should use red for < 50%")
		}

		// Ensure result is always valid UTF-8 (BuildCoverageBadgeURL should always produce valid UTF-8)
		assert.True(t, utf8.ValidString(result), "Result should be valid UTF-8")

		// Ensure URL looks valid
		assert.NotContains(t, result, " ", "URL should not contain spaces")
		assert.NotContains(t, result, "\n", "URL should not contain newlines")
		assert.NotContains(t, result, "\t", "URL should not contain tabs")
	})
}

// FuzzBuildGitHubRepoURL tests the BuildGitHubRepoURL function
func FuzzBuildGitHubRepoURL(f *testing.F) {
	// Seed corpus with typical and edge case inputs
	f.Add("", "")
	f.Add("owner", "repo")
	f.Add("my-org", "my-repo")
	f.Add("owner with spaces", "repo")
	f.Add("owner", "repo with spaces")
	f.Add("\x00\x01\x02", "repo") // null bytes
	f.Add("owner", "\x00\x01\x02")
	f.Add("unicode🔥", "repo")
	f.Add("owner", "unicode🔥")
	f.Add("../../../etc/passwd", "repo") // path traversal
	f.Add("owner", "../../../etc/passwd")
	f.Add("javascript:alert(1)", "repo") // XSS attempt
	f.Add("owner", "javascript:alert(1)")
	f.Add("\n\r\t", "repo") // control characters
	f.Add("owner", "\n\r\t")

	f.Fuzz(func(t *testing.T, owner, repo string) {
		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("BuildGitHubRepoURL panicked with inputs owner=%q, repo=%q: %v", owner, repo, r)
			}
		}()

		result := BuildGitHubRepoURL(owner, repo)

		// Validate output format
		if owner == "" || repo == "" {
			assert.Empty(t, result, "Should return empty string for empty inputs")
		} else {
			assert.NotEmpty(t, result, "Should return non-empty string for valid inputs")
			assert.True(t, strings.HasPrefix(result, "https://github.com/"), "Should start with GitHub URL prefix")
			assert.Contains(t, result, owner, "Should contain owner")
			assert.Contains(t, result, repo, "Should contain repo")

			// Check for proper structure
			expected := "https://github.com/" + owner + "/" + repo
			assert.Equal(t, expected, result, "Should match expected URL format")
		}

		// Ensure result is valid UTF-8 if inputs were valid UTF-8
		if utf8.ValidString(owner) && utf8.ValidString(repo) {
			assert.True(t, utf8.ValidString(result), "Result should be valid UTF-8 when inputs are valid UTF-8")
		}
	})
}

// FuzzExtractRepoNameFromURL tests the ExtractRepoNameFromURL function
func FuzzExtractRepoNameFromURL(f *testing.F) {
	// Seed corpus with typical and edge case inputs
	f.Add("")
	f.Add("repo")
	f.Add("owner/repo")
	f.Add("github.com/owner/repo")
	f.Add("https://github.com/owner/repo")
	f.Add("owner/repo/extra/path")
	f.Add("/")
	f.Add("//")
	f.Add("///")
	f.Add("owner/")
	f.Add("/repo")
	f.Add("owner\\repo") // Windows path separator
	f.Add("owner/repo/")
	f.Add("unicode🔥/repo")
	f.Add("owner/unicode🔥")
	f.Add("owner/repo/unicode🔥")
	f.Add("\x00\x01\x02/repo") // null bytes
	f.Add("owner/\x00\x01\x02")
	f.Add("../../../etc/passwd/repo") // path traversal
	f.Add("owner/../../../etc/passwd")
	f.Add("javascript:alert(1)/repo") // XSS attempt
	f.Add("owner/javascript:alert(1)")
	f.Add("\n\r\t/repo") // control characters
	f.Add("owner/\n\r\t")

	f.Fuzz(func(t *testing.T, fullName string) {
		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("ExtractRepoNameFromURL panicked with input fullName=%q: %v", fullName, r)
			}
		}()

		result := ExtractRepoNameFromURL(fullName)

		// Validate output
		assert.NotNil(t, result, "Should never return nil")

		// If input has slashes, result should be the last part
		parts := strings.Split(fullName, "/")
		if len(parts) >= 2 {
			expected := parts[len(parts)-1]
			assert.Equal(t, expected, result, "Should return last part after splitting on slash")
		} else {
			assert.Equal(t, fullName, result, "Should return original string if no slashes")
		}

		// Ensure result is valid UTF-8 if input was valid UTF-8
		if utf8.ValidString(fullName) {
			assert.True(t, utf8.ValidString(result), "Result should be valid UTF-8 when input is valid UTF-8")
		}
	})
}

// FuzzCleanModulePath tests the CleanModulePath function with diverse path inputs
func FuzzCleanModulePath(f *testing.F) {
	// Seed corpus with typical and edge case inputs
	f.Add("")
	f.Add("file.go")
	f.Add("internal/file.go")
	f.Add("github.com/owner/repo/file.go")
	f.Add("github.com/owner/repo/internal/file.go")
	f.Add("github.com/owner/repo/cli/internal/cli/cancel.go")
	f.Add("domain.com/owner/repo/internal/file.go")
	f.Add("example.org/user/project/cmd/main.go")
	f.Add("localhost/project/file.go")
	f.Add("github.com")
	f.Add("github.com/")
	f.Add("github.com/owner")
	f.Add("github.com/owner/")
	f.Add("github.com/owner/repo")
	f.Add("github.com/owner/repo/")
	f.Add("/github.com/owner/repo/file.go")
	f.Add("github.com\\owner\\repo\\file.go") // Windows paths
	f.Add("github.com/owner/repo/unicode🔥/file.go")
	f.Add("github.com/unicode🔥/repo/file.go")
	f.Add("unicode🔥.com/owner/repo/file.go")
	f.Add("\x00\x01\x02.com/owner/repo/file.go") // null bytes
	f.Add("github.com/\x00\x01\x02/repo/file.go")
	f.Add("github.com/owner/\x00\x01\x02/file.go")
	f.Add("../../../etc/passwd")
	f.Add("javascript:alert(1)/file.go")
	f.Add("\n\r\t/file.go")
	f.Add("very/long/path/with/many/segments/that/could/cause/issues/file.go")

	f.Fuzz(func(t *testing.T, fullPath string) {
		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("CleanModulePath panicked with input fullPath=%q: %v", fullPath, r)
			}
		}()

		result := CleanModulePath(fullPath)

		// Validate output
		assert.NotNil(t, result, "Should never return nil")

		// Result should be a valid path (no double slashes except possibly at start)
		assert.NotContains(t, strings.TrimPrefix(result, "//"), "//", "Should not contain double slashes (except possibly at start)")

		// If no github.com pattern found, should return cleaned path
		if !strings.Contains(fullPath, "github.com") {
			// The path should be cleaned (no double slashes, UTF-8 issues, etc.)
			if strings.Contains(fullPath, "//") || strings.HasSuffix(fullPath, "/") || !utf8.ValidString(fullPath) {
				// If the path needs cleaning (including UTF-8 sanitization), we expect it to be cleaned
				assert.NotContains(t, strings.TrimPrefix(result, "//"), "//", "Should clean double slashes")
				// Result must be valid UTF-8
				assert.True(t, utf8.ValidString(result), "Result should be valid UTF-8")
			} else if !strings.Contains(fullPath, ".") {
				// No domain pattern and no cleaning needed (including valid UTF-8)
				assert.Equal(t, fullPath, result, "Should return original path if no cleaning needed")
			}
		}

		// Result must always be valid UTF-8 (we sanitize invalid UTF-8)
		assert.True(t, utf8.ValidString(result), "Result should always be valid UTF-8")

		// Result length check: when UTF-8 sanitization occurs, result might be longer
		// (invalid bytes are replaced with 3-byte replacement character)
		if utf8.ValidString(fullPath) {
			assert.LessOrEqual(t, len(result), len(fullPath), "Result should not be longer than valid UTF-8 input")
		}
	})
}

// FuzzBuildGitHubFileURL tests the BuildGitHubFileURL function
func FuzzBuildGitHubFileURL(f *testing.F) {
	// Seed corpus with typical and edge case inputs
	f.Add("", "", "", "")
	f.Add("owner", "repo", "main", "file.go")
	f.Add("owner", "repo", "feature-branch", "internal/file.go")
	f.Add("owner", "repo", "main", "github.com/owner/repo/internal/file.go")
	f.Add("owner with spaces", "repo", "main", "file.go")
	f.Add("owner", "repo with spaces", "main", "file.go")
	f.Add("owner", "repo", "branch with spaces", "file.go")
	f.Add("owner", "repo", "main", "file with spaces.go")
	f.Add("unicode🔥", "repo", "main", "file.go")
	f.Add("owner", "unicode🔥", "main", "file.go")
	f.Add("owner", "repo", "unicode🔥", "file.go")
	f.Add("owner", "repo", "main", "unicode🔥.go")
	f.Add("\x00\x01\x02", "repo", "main", "file.go") // null bytes
	f.Add("owner", "\x00\x01\x02", "main", "file.go")
	f.Add("owner", "repo", "\x00\x01\x02", "file.go")
	f.Add("owner", "repo", "main", "\x00\x01\x02.go")
	f.Add("../../../etc/passwd", "repo", "main", "file.go") // path traversal
	f.Add("owner", "../../../etc/passwd", "main", "file.go")
	f.Add("owner", "repo", "../../../etc/passwd", "file.go")
	f.Add("owner", "repo", "main", "../../../etc/passwd")
	f.Add("javascript:alert(1)", "repo", "main", "file.go") // XSS attempt
	f.Add("owner", "javascript:alert(1)", "main", "file.go")
	f.Add("owner", "repo", "javascript:alert(1)", "file.go")
	f.Add("owner", "repo", "main", "javascript:alert(1).go")
	f.Add("\n\r\t", "repo", "main", "file.go") // control characters
	f.Add("owner", "\n\r\t", "main", "file.go")
	f.Add("owner", "repo", "\n\r\t", "file.go")
	f.Add("owner", "repo", "main", "\n\r\t.go")

	f.Fuzz(func(t *testing.T, owner, repo, branch, filePath string) {
		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("BuildGitHubFileURL panicked with inputs owner=%q, repo=%q, branch=%q, filePath=%q: %v", owner, repo, branch, filePath, r)
			}
		}()

		result := BuildGitHubFileURL(owner, repo, branch, filePath)

		// Validate output format
		if owner == "" || repo == "" || branch == "" || filePath == "" {
			assert.Empty(t, result, "Should return empty string for empty inputs")
		} else {
			assert.NotEmpty(t, result, "Should return non-empty string for valid inputs")
			assert.True(t, strings.HasPrefix(result, "https://github.com/"), "Should start with GitHub URL prefix")
			assert.Contains(t, result, owner, "Should contain owner")
			assert.Contains(t, result, repo, "Should contain repo")
			assert.Contains(t, result, branch, "Should contain branch")
			assert.Contains(t, result, "/blob/", "Should contain blob path")

			// The file path should be processed through CleanModulePath
			cleanedPath := CleanModulePath(filePath)
			assert.Contains(t, result, cleanedPath, "Should contain cleaned file path")
		}

		// Ensure result is valid UTF-8 if inputs were valid UTF-8
		if utf8.ValidString(owner) && utf8.ValidString(repo) && utf8.ValidString(branch) && utf8.ValidString(filePath) {
			assert.True(t, utf8.ValidString(result), "Result should be valid UTF-8 when inputs are valid UTF-8")
		}
	})
}

// FuzzCleanupExtraPathPrefixes tests the cleanupExtraPathPrefixes helper function
// This function is not exported, but we can test it through CleanModulePath
func FuzzCleanupExtraPathPrefixesViaCleanModulePath(f *testing.F) {
	// Seed corpus with paths that would trigger the cleanup logic
	f.Add("cli/internal/cli/cancel.go")
	f.Add("cmd/internal/cmd/server.go")
	f.Add("api/internal/api/handler.go")
	f.Add("test/internal/test/utils.go")
	f.Add("a/b/a/c.go")
	f.Add("x/y/z/x/file.go")
	f.Add("single")
	f.Add("a/b")
	f.Add("")
	f.Add("/")
	f.Add("a/")
	f.Add("/a")
	f.Add("a/a")
	f.Add("a/b/a")
	f.Add("very/long/path/very/long/deep/structure.go")
	f.Add("unicode🔥/internal/unicode🔥/file.go")
	f.Add("\x00\x01\x02/internal/\x00\x01\x02/file.go")
	f.Add("../../../etc/passwd/internal/../../../etc/passwd/file.go")

	f.Fuzz(func(t *testing.T, path string) {
		// Function should never panic when called via CleanModulePath
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("cleanupExtraPathPrefixes (via CleanModulePath) panicked with input path=%q: %v", path, r)
			}
		}()

		// Test via CleanModulePath since cleanupExtraPathPrefixes is not exported
		// We'll construct a path that forces the use of cleanupExtraPathPrefixes
		testPath := "github.com/owner/repo/" + path
		result := CleanModulePath(testPath)

		// Validate that result is reasonable
		assert.True(t, utf8.ValidString(result), "Result should be valid UTF-8")

		// When UTF-8 sanitization occurs, the result might be longer than input
		// (invalid bytes are replaced with the 3-byte replacement character)
		// So we only check length when the input was valid UTF-8
		if utf8.ValidString(testPath) {
			assert.LessOrEqual(t, len(result), len(testPath), "Result should not be longer than valid UTF-8 input")
		}

		// If the original path is not empty, result might still be empty if the path
		// resolves to nothing valid (e.g., "../../.." going outside the repo)
		if path != "" && !strings.Contains(path, "..") {
			assert.NotEmpty(t, result, "Result should not be empty for non-empty input without ..")
		}
	})
}

// Additional edge case fuzz test for URL building functions with very long inputs
func FuzzURLBuildingWithLongInputs(f *testing.F) {
	// Create some long strings for testing
	longString := strings.Repeat("a", 1000)
	veryLongString := strings.Repeat("x", 10000)
	unicodeString := strings.Repeat("🔥", 100)

	f.Add(longString, "repo", "commit")
	f.Add("owner", longString, "commit")
	f.Add("owner", "repo", longString)
	f.Add(veryLongString, "repo", "commit")
	f.Add("owner", veryLongString, "commit")
	f.Add("owner", "repo", veryLongString)
	f.Add(unicodeString, "repo", "commit")
	f.Add("owner", unicodeString, "commit")
	f.Add("owner", "repo", unicodeString)

	f.Fuzz(func(t *testing.T, str1, str2, str3 string) {
		// Test all URL building functions with potentially long inputs
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("URL building functions panicked with long inputs: %v", r)
			}
		}()

		// Test BuildGitHubCommitURL
		commitURL := BuildGitHubCommitURL(str1, str2, str3)
		assert.True(t, utf8.ValidString(commitURL), "Commit URL should be valid UTF-8")

		// Test BuildGitHubRepoURL
		repoURL := BuildGitHubRepoURL(str1, str2)
		assert.True(t, utf8.ValidString(repoURL), "Repo URL should be valid UTF-8")

		// Test BuildGitHubFileURL
		fileURL := BuildGitHubFileURL(str1, str2, str3, "file.go")
		assert.True(t, utf8.ValidString(fileURL), "File URL should be valid UTF-8")

		// Test ExtractRepoNameFromURL
		repoName := ExtractRepoNameFromURL(str1 + "/" + str2 + "/" + str3)
		assert.True(t, utf8.ValidString(repoName), "Repo name should be valid UTF-8")

		// Test CleanModulePath
		cleanPath := CleanModulePath(str1 + "/" + str2 + "/" + str3)
		assert.True(t, utf8.ValidString(cleanPath), "Clean path should be valid UTF-8")
	})
}
