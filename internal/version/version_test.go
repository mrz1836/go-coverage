package version

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetLatestRelease tests the GetLatestRelease function with HTTP integration
// Since the function has a hardcoded GitHub API URL, this is effectively an integration test
// that will make real HTTP requests to GitHub. In a production environment, this would be refactored
// to inject an HTTP client interface for proper mocking.
func TestGetLatestRelease(t *testing.T) {
	// Test error handling for malformed owner/repo
	t.Run("InvalidOwnerRepo", func(t *testing.T) {
		tests := []struct {
			name  string
			owner string
			repo  string
		}{
			{
				name:  "empty owner",
				owner: "",
				repo:  "test-repo",
			},
			{
				name:  "empty repo",
				owner: "test-owner",
				repo:  "",
			},
			{
				name:  "nonexistent repository",
				owner: "nonexistent-owner-12345",
				repo:  "nonexistent-repo-12345",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := GetLatestRelease(tt.owner, tt.repo)
				require.Error(t, err)
				// Should get a GitHub API error (likely 404 or similar)
			})
		}
	})

	// Test with a well-known repository that should always have releases
	// This is more of an integration test but necessary given the current function design
	t.Run("ValidRepository", func(t *testing.T) {
		// Use a well-known repository that should always have releases
		// Using the go-coverage repository itself for testing
		release, err := GetLatestRelease("mrz1836", "go-coverage")
		if err != nil {
			// Allow network-related errors in CI environments
			errStr := strings.ToLower(err.Error())
			networkErrors := []string{
				"connection",
				"timeout",
				"network",
				"dns",
				"no such host",
				"temporary failure",
				"rate limit",
			}

			hasNetworkError := false
			for _, netErr := range networkErrors {
				if strings.Contains(errStr, netErr) {
					hasNetworkError = true
					break
				}
			}

			if hasNetworkError {
				t.Skip("Skipping test due to network connectivity issues")
				return
			}

			// If it's not a network error, fail the test
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify the release structure
		require.NotNil(t, release)
		require.NotEmpty(t, release.TagName)
		require.NotEmpty(t, release.Name)
		// PublishedAt should be a valid time
		require.False(t, release.PublishedAt.IsZero())
	})
}

func TestCompareVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		// Basic version comparisons
		{
			name:     "V1GreaterThanV2",
			v1:       "1.2.3",
			v2:       "1.2.2",
			expected: 1,
		},
		{
			name:     "V1LessThanV2",
			v1:       "1.2.2",
			v2:       "1.2.3",
			expected: -1,
		},
		{
			name:     "V1EqualsV2",
			v1:       "1.2.3",
			v2:       "1.2.3",
			expected: 0,
		},
		// Version with 'v' prefix
		{
			name:     "VersionWithVPrefix",
			v1:       "v1.2.3",
			v2:       "v1.2.2",
			expected: 1,
		},
		{
			name:     "MixedVPrefix",
			v1:       "v1.2.3",
			v2:       "1.2.3",
			expected: 0,
		},
		// Major version differences
		{
			name:     "MajorVersionDifference",
			v1:       "2.0.0",
			v2:       "1.9.9",
			expected: 1,
		},
		{
			name:     "MajorVersionLower",
			v1:       "1.9.9",
			v2:       "2.0.0",
			expected: -1,
		},
		// Minor version differences
		{
			name:     "MinorVersionDifference",
			v1:       "1.3.0",
			v2:       "1.2.9",
			expected: 1,
		},
		// Development versions and commit hashes
		{
			name:     "DevVersionVsRelease",
			v1:       "dev",
			v2:       "1.2.3",
			expected: -1,
		},
		{
			name:     "ReleaseVsDevVersion",
			v1:       "1.2.3",
			v2:       "dev",
			expected: 1,
		},
		{
			name:     "BothDevVersions",
			v1:       "dev",
			v2:       "dev",
			expected: 0,
		},
		{
			name:     "DevDirtyVsRelease",
			v1:       "dev-dirty",
			v2:       "1.2.3",
			expected: -1,
		},
		{
			name:     "ReleaseVsDevDirty",
			v1:       "1.2.3",
			v2:       "dev-dirty",
			expected: 1,
		},
		{
			name:     "DevVsDevDirty",
			v1:       "dev",
			v2:       "dev-dirty",
			expected: 0,
		},
		{
			name:     "DevDirtyVsDev",
			v1:       "dev-dirty",
			v2:       "dev",
			expected: 0,
		},
		{
			name:     "EmptyVersionVsRelease",
			v1:       "",
			v2:       "1.2.3",
			expected: -1,
		},
		{
			name:     "CommitHashVsRelease",
			v1:       "abc123def456",
			v2:       "1.2.3",
			expected: -1,
		},
		{
			name:     "TwoCommitHashes",
			v1:       "abc123def456",
			v2:       "def456abc123",
			expected: 0,
		},
		// Different version lengths
		{
			name:     "DifferentVersionLengths",
			v1:       "1.2.3.4",
			v2:       "1.2.3",
			expected: 0, // Only compares first 3 parts
		},
		{
			name:     "ShortVersion",
			v1:       "1.2",
			v2:       "1.2.0",
			expected: 0,
		},
		{
			name:     "VeryShortVersion",
			v1:       "2",
			v2:       "1.9.9",
			expected: 1,
		},
		// Edge cases
		{
			name:     "BothEmpty",
			v1:       "",
			v2:       "",
			expected: 0,
		},
		{
			name:     "VersionWithSuffix",
			v1:       "1.2.3-rc1",
			v2:       "1.2.3",
			expected: 0, // Suffixes are ignored in comparison
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CompareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result, "CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
		})
	}
}

func TestParseVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version  string
		expected []int
	}{
		{
			name:     "StandardVersion",
			version:  "1.2.3",
			expected: []int{1, 2, 3},
		},
		{
			name:     "VersionWithSuffix",
			version:  "1.2.3-rc1",
			expected: []int{1, 2, 3},
		},
		{
			name:     "VersionWithBuildMetadata",
			version:  "1.2.3+build.1",
			expected: []int{1, 2, 3},
		},
		{
			name:     "TwoPartVersion",
			version:  "1.2",
			expected: []int{1, 2},
		},
		{
			name:     "SinglePartVersion",
			version:  "1",
			expected: []int{1},
		},
		{
			name:     "FourPartVersion",
			version:  "1.2.3.4",
			expected: []int{1, 2, 3, 4},
		},
		{
			name:     "EmptyVersion",
			version:  "",
			expected: []int{},
		},
		{
			name:     "InvalidVersion",
			version:  "a.b.c",
			expected: []int{},
		},
		{
			name:     "MixedValidInvalid",
			version:  "1.a.3",
			expected: []int{1, 3},
		},
		{
			name:     "LeadingZeros",
			version:  "01.02.03",
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := parseVersion(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNewerVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		expected       bool
	}{
		{
			name:           "NewerVersionAvailable",
			currentVersion: "1.2.2",
			latestVersion:  "1.2.3",
			expected:       true,
		},
		{
			name:           "SameVersion",
			currentVersion: "1.2.3",
			latestVersion:  "1.2.3",
			expected:       false,
		},
		{
			name:           "OlderVersionProvided",
			currentVersion: "1.2.3",
			latestVersion:  "1.2.2",
			expected:       false,
		},
		{
			name:           "DevVersionCurrent",
			currentVersion: "dev",
			latestVersion:  "1.2.3",
			expected:       true,
		},
		{
			name:           "DevDirtyVersionCurrent",
			currentVersion: "dev-dirty",
			latestVersion:  "1.2.3",
			expected:       true,
		},
		{
			name:           "DevVersionLatest",
			currentVersion: "1.2.3",
			latestVersion:  "dev",
			expected:       false,
		},
		{
			name:           "MajorVersionUpgrade",
			currentVersion: "1.9.9",
			latestVersion:  "2.0.0",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsNewerVersion(tt.currentVersion, tt.latestVersion)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "StandardVersion",
			version:  "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "VersionWithVPrefix",
			version:  "v1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "VersionWithSuffix",
			version:  "1.2.3-rc1",
			expected: "1.2.3",
		},
		{
			name:     "VersionWithWhitespace",
			version:  "  1.2.3  ",
			expected: "1.2.3",
		},
		{
			name:     "VersionWithVPrefixAndSuffix",
			version:  "v1.2.3-dirty",
			expected: "1.2.3",
		},
		{
			name:     "EmptyVersion",
			version:  "",
			expected: "",
		},
		{
			name:     "OnlyVPrefix",
			version:  "v",
			expected: "",
		},
		{
			name:     "OnlyWhitespace",
			version:  "   ",
			expected: "",
		},
		{
			name:     "VersionWithMultipleDashes",
			version:  "1.2.3-rc1-dirty",
			expected: "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NormalizeVersion(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsCommitHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "ValidShortHash",
			input:    "abc123d",
			expected: true,
		},
		{
			name:     "ValidLongHash",
			input:    "abc123def456789012345678901234567890abcd",
			expected: true,
		},
		{
			name:     "ValidMixedCaseHash",
			input:    "AbC123DeF456",
			expected: true,
		},
		{
			name:     "TooShort",
			input:    "abc12",
			expected: false,
		},
		{
			name:     "TooLong",
			input:    "abc123def456789012345678901234567890abcdef",
			expected: false,
		},
		{
			name:     "ContainsInvalidCharacters",
			input:    "abc123xyz",
			expected: false,
		},
		{
			name:     "ContainsSpecialCharacters",
			input:    "abc123-def",
			expected: false,
		},
		{
			name:     "EmptyString",
			input:    "",
			expected: false,
		},
		{
			name:     "OnlyNumbers",
			input:    "1234567890",
			expected: true,
		},
		{
			name:     "OnlyLetters",
			input:    "abcdefabcde",
			expected: true,
		},
		{
			name:     "VersionString",
			input:    "1.2.3",
			expected: false,
		},
		{
			name:     "DevString",
			input:    "dev",
			expected: false,
		},
		{
			name:     "ExactMinLength",
			input:    "abc123d", // 7 characters
			expected: true,
		},
		{
			name:     "ExactMaxLength",
			input:    "0123456789abcdef0123456789abcdef01234567", // 40 characters
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isCommitHash(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitHubReleaseStruct(t *testing.T) {
	t.Parallel()

	// Test JSON unmarshaling
	jsonData := `{
		"tag_name": "v1.2.3",
		"name": "Release v1.2.3",
		"draft": false,
		"prerelease": true,
		"published_at": "2025-01-01T12:00:00Z",
		"body": "Release notes here"
	}`

	var release GitHubRelease
	err := json.Unmarshal([]byte(jsonData), &release)
	require.NoError(t, err)

	assert.Equal(t, "v1.2.3", release.TagName)
	assert.Equal(t, "Release v1.2.3", release.Name)
	assert.False(t, release.Draft)
	assert.True(t, release.Prerelease)
	assert.Equal(t, "Release notes here", release.Body)

	expectedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedTime, release.PublishedAt)
}

func TestInfoStruct(t *testing.T) {
	t.Parallel()

	info := Info{
		Current: "1.2.2",
		Latest:  "1.2.3",
		IsNewer: true,
	}

	assert.Equal(t, "1.2.2", info.Current)
	assert.Equal(t, "1.2.3", info.Latest)
	assert.True(t, info.IsNewer)
}

func TestErrGitHubAPIFailed(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "GitHub API request failed", ErrGitHubAPIFailed.Error())
}

// Integration test for version comparison workflow
func TestVersionComparisonWorkflow(t *testing.T) {
	t.Parallel()

	// Test typical version comparison workflow
	currentVersion := "1.2.2"
	latestVersion := "1.2.3"

	// Normalize versions
	normalizedCurrent := NormalizeVersion(currentVersion)
	normalizedLatest := NormalizeVersion(latestVersion)

	// Compare versions
	comparison := CompareVersions(normalizedCurrent, normalizedLatest)
	isNewer := IsNewerVersion(normalizedCurrent, normalizedLatest)

	assert.Equal(t, -1, comparison)
	assert.True(t, isNewer)

	// Test with development version
	devVersion := "dev"
	normalizedDev := NormalizeVersion(devVersion)
	isDevNewer := IsNewerVersion(normalizedDev, normalizedLatest)

	assert.True(t, isDevNewer)
}

func TestIsDevelopmentVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "ExactDevVersion",
			version:  "dev",
			expected: true,
		},
		{
			name:     "DevWithDirtySuffix",
			version:  "dev-dirty",
			expected: true,
		},
		{
			name:     "DevWithOtherSuffix",
			version:  "dev-alpha",
			expected: true,
		},
		{
			name:     "DevWithMultipleSuffixes",
			version:  "dev-dirty-modified",
			expected: true,
		},
		{
			name:     "StandardVersion",
			version:  "1.2.3",
			expected: false,
		},
		{
			name:     "VersionWithV",
			version:  "v1.2.3",
			expected: false,
		},
		{
			name:     "ReleaseCandidate",
			version:  "1.2.3-rc1",
			expected: false,
		},
		{
			name:     "CommitHash",
			version:  "abc123def456",
			expected: false,
		},
		{
			name:     "EmptyVersion",
			version:  "",
			expected: false,
		},
		{
			name:     "DevInMiddle",
			version:  "1.2.3-dev-4",
			expected: false,
		},
		{
			name:     "DevAtEnd",
			version:  "1.2.3-dev",
			expected: false,
		},
		{
			name:     "CaseSensitive",
			version:  "Dev",
			expected: false,
		},
		{
			name:     "DevLikeButNot",
			version:  "development",
			expected: false,
		},
		{
			name:     "DevWithDashOnly",
			version:  "dev-",
			expected: true,
		},
		{
			name:     "DevWithNumbers",
			version:  "dev-123",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isDevelopmentVersion(tt.version)
			assert.Equal(t, tt.expected, result, "isDevelopmentVersion(%q) should return %v", tt.version, tt.expected)
		})
	}
}
