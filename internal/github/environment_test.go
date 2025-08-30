package github

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectEnvironment(t *testing.T) {
	// Save original environment
	originalEnv := saveEnv()
	defer restoreEnv(originalEnv)

	tests := []struct {
		name     string
		envVars  map[string]string
		expected *GitHubContext
	}{
		{
			name: "GitHub Actions environment",
			envVars: map[string]string{
				"GITHUB_ACTIONS":    "true",
				"GITHUB_REPOSITORY": "owner/repo",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_SHA":        "abc123",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_RUN_ID":     "123456",
				"GITHUB_TOKEN":      "secret",
			},
			expected: &GitHubContext{
				IsGitHubActions: true,
				Repository:      "owner/repo",
				Branch:          "main",
				CommitSHA:       "abc123",
				EventName:       "push",
				RunID:           "123456",
				Token:           "secret",
				PRNumber:        "",
			},
		},
		{
			name: "Local environment",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "false",
			},
			expected: &GitHubContext{
				IsGitHubActions: false,
			},
		},
		{
			name: "Pull request environment",
			envVars: map[string]string{
				"GITHUB_ACTIONS":    "true",
				"GITHUB_REPOSITORY": "owner/repo",
				"GITHUB_REF_NAME":   "feature-branch",
				"GITHUB_SHA":        "def456",
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_RUN_ID":     "789012",
				"GITHUB_TOKEN":      "secret",
			},
			expected: &GitHubContext{
				IsGitHubActions: true,
				Repository:      "owner/repo",
				Branch:          "feature-branch",
				CommitSHA:       "def456",
				EventName:       "pull_request",
				RunID:           "789012",
				Token:           "secret",
				PRNumber:        "", // Would be set if GITHUB_EVENT_PATH was valid
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			// Run the function
			result, err := DetectEnvironment()
			// Check for errors
			if err != nil {
				t.Fatalf("DetectEnvironment() returned error: %v", err)
			}

			// Verify results
			if result.IsGitHubActions != tt.expected.IsGitHubActions {
				t.Errorf("IsGitHubActions = %v, want %v", result.IsGitHubActions, tt.expected.IsGitHubActions)
			}
			if result.Repository != tt.expected.Repository {
				t.Errorf("Repository = %v, want %v", result.Repository, tt.expected.Repository)
			}
			if result.Branch != tt.expected.Branch {
				t.Errorf("Branch = %v, want %v", result.Branch, tt.expected.Branch)
			}
			if result.CommitSHA != tt.expected.CommitSHA {
				t.Errorf("CommitSHA = %v, want %v", result.CommitSHA, tt.expected.CommitSHA)
			}
			if result.EventName != tt.expected.EventName {
				t.Errorf("EventName = %v, want %v", result.EventName, tt.expected.EventName)
			}
			if result.RunID != tt.expected.RunID {
				t.Errorf("RunID = %v, want %v", result.RunID, tt.expected.RunID)
			}
			if result.Token != tt.expected.Token {
				t.Errorf("Token = %v, want %v", result.Token, tt.expected.Token)
			}
		})
	}
}

func TestGetBranch(t *testing.T) {
	// Save original environment
	originalEnv := saveEnv()
	defer restoreEnv(originalEnv)

	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name: "GITHUB_REF_NAME available",
			envVars: map[string]string{
				"GITHUB_REF_NAME": "feature-branch",
				"GITHUB_REF":      "refs/heads/other-branch",
			},
			expected: "feature-branch",
		},
		{
			name: "GITHUB_REF with heads prefix",
			envVars: map[string]string{
				"GITHUB_REF": "refs/heads/main",
			},
			expected: "main",
		},
		{
			name: "GITHUB_REF with tags prefix",
			envVars: map[string]string{
				"GITHUB_REF": "refs/tags/v1.0.0",
			},
			expected: "v1.0.0",
		},
		{
			name:     "No branch information",
			envVars:  map[string]string{},
			expected: "",
		},
		{
			name: "PR with merge ref - should use GITHUB_HEAD_REF",
			envVars: map[string]string{
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_REF_NAME":   "23/merge",
				"GITHUB_HEAD_REF":   "feat/github-actions-integration",
			},
			expected: "feat/github-actions-integration",
		},
		{
			name: "PR with merge ref but no GITHUB_HEAD_REF - should return empty",
			envVars: map[string]string{
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_REF_NAME":   "23/merge",
			},
			expected: "",
		},
		{
			name: "PR target with merge ref - should use GITHUB_HEAD_REF",
			envVars: map[string]string{
				"GITHUB_EVENT_NAME": "pull_request_target",
				"GITHUB_REF_NAME":   "42/merge",
				"GITHUB_HEAD_REF":   "bugfix/coverage-urls",
			},
			expected: "bugfix/coverage-urls",
		},
		{
			name: "Non-PR event with merge ref - should return empty",
			envVars: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_REF_NAME":   "23/merge",
			},
			expected: "",
		},
		{
			name: "GITHUB_REF with merge ref - should skip",
			envVars: map[string]string{
				"GITHUB_REF": "refs/heads/23/merge",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			// Run the function
			result := getBranch()

			// Verify result
			if result != tt.expected {
				t.Errorf("getBranch() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractPRNumber(t *testing.T) {
	// Save original environment
	originalEnv := saveEnv()
	defer restoreEnv(originalEnv)

	// Create temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		eventPath   string
		eventData   interface{}
		expected    string
		expectError bool
	}{
		{
			name:      "Valid PR event",
			eventPath: filepath.Join(tempDir, "pr_event.json"),
			eventData: map[string]interface{}{
				"pull_request": map[string]interface{}{
					"number": 42,
				},
			},
			expected:    "42",
			expectError: false,
		},
		{
			name:      "No pull request in event",
			eventPath: filepath.Join(tempDir, "push_event.json"),
			eventData: map[string]interface{}{
				"ref": "refs/heads/main",
			},
			expected:    "",
			expectError: true,
		},
		{
			name:        "No event path",
			eventPath:   "",
			eventData:   nil,
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv()

			if tt.eventPath != "" {
				// Create event file
				if tt.eventData != nil {
					data, err := json.Marshal(tt.eventData)
					if err != nil {
						t.Fatalf("Failed to marshal test data: %v", err)
					}
					if err := os.WriteFile(tt.eventPath, data, 0o600); err != nil {
						t.Fatalf("Failed to write test file: %v", err)
					}
				}
				_ = os.Setenv("GITHUB_EVENT_PATH", tt.eventPath)
			}

			// Run the function
			result, err := extractPRNumber()

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("extractPRNumber() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("extractPRNumber() unexpected error: %v", err)
			}

			// Check result
			if result != tt.expected {
				t.Errorf("extractPRNumber() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsGitHubActions(t *testing.T) {
	// Save original environment
	originalEnv := saveEnv()
	defer restoreEnv(originalEnv)

	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "GitHub Actions true",
			envValue: "true",
			expected: true,
		},
		{
			name:     "GitHub Actions false",
			envValue: "false",
			expected: false,
		},
		{
			name:     "GitHub Actions empty",
			envValue: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()
			if tt.envValue != "" {
				_ = os.Setenv("GITHUB_ACTIONS", tt.envValue)
			}

			result := IsGitHubActions()
			if result != tt.expected {
				t.Errorf("IsGitHubActions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsPullRequest(t *testing.T) {
	// Save original environment
	originalEnv := saveEnv()
	defer restoreEnv(originalEnv)

	tests := []struct {
		name      string
		eventName string
		expected  bool
	}{
		{
			name:      "Pull request event",
			eventName: "pull_request",
			expected:  true,
		},
		{
			name:      "Pull request target event",
			eventName: "pull_request_target",
			expected:  true,
		},
		{
			name:      "Push event",
			eventName: "push",
			expected:  false,
		},
		{
			name:      "No event name",
			eventName: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()
			if tt.eventName != "" {
				_ = os.Setenv("GITHUB_EVENT_NAME", tt.eventName)
			}

			result := IsPullRequest()
			if result != tt.expected {
				t.Errorf("IsPullRequest() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateEnvironment(t *testing.T) {
	// Save original environment
	originalEnv := saveEnv()
	defer restoreEnv(originalEnv)

	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
	}{
		{
			name: "Valid environment",
			envVars: map[string]string{
				"GITHUB_ACTIONS":    "true",
				"GITHUB_REPOSITORY": "owner/repo",
				"GITHUB_SHA":        "abc123",
				"GITHUB_TOKEN":      "secret",
			},
			expectError: false,
		},
		{
			name: "Not GitHub Actions",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "false",
			},
			expectError: true,
		},
		{
			name: "Missing repository",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
				"GITHUB_SHA":     "abc123",
				"GITHUB_TOKEN":   "secret",
			},
			expectError: true,
		},
		{
			name: "Missing token",
			envVars: map[string]string{
				"GITHUB_ACTIONS":    "true",
				"GITHUB_REPOSITORY": "owner/repo",
				"GITHUB_SHA":        "abc123",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()

			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			err := ValidateEnvironment()
			if tt.expectError && err == nil {
				t.Errorf("ValidateEnvironment() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ValidateEnvironment() unexpected error: %v", err)
			}
		})
	}
}

// Helper functions for environment management in tests
func saveEnv() map[string]string {
	env := make(map[string]string)
	for _, kv := range os.Environ() {
		key, value := splitEnvVar(kv)
		if key != "" {
			env[key] = value
		}
	}
	return env
}

func restoreEnv(env map[string]string) {
	clearEnv()
	for key, value := range env {
		_ = os.Setenv(key, value)
	}
}

func clearEnv() {
	githubVars := []string{
		"GITHUB_ACTIONS",
		"GITHUB_REPOSITORY",
		"GITHUB_REF_NAME",
		"GITHUB_REF",
		"GITHUB_HEAD_REF",
		"GITHUB_SHA",
		"GITHUB_EVENT_NAME",
		"GITHUB_EVENT_PATH",
		"GITHUB_RUN_ID",
		"GITHUB_TOKEN",
	}

	for _, v := range githubVars {
		_ = os.Unsetenv(v)
	}
}

func splitEnvVar(kv string) (string, string) {
	for i := 0; i < len(kv); i++ {
		if kv[i] == '=' {
			return kv[:i], kv[i+1:]
		}
	}
	return kv, ""
}

func TestIsMergeRef(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		expected bool
	}{
		{
			name:     "Standard merge ref",
			ref:      "23/merge",
			expected: true,
		},
		{
			name:     "Another merge ref",
			ref:      "456/merge",
			expected: true,
		},
		{
			name:     "Merge ref with prefix",
			ref:      "feature/123/merge",
			expected: true,
		},
		{
			name:     "Regular branch name",
			ref:      "feature-branch",
			expected: false,
		},
		{
			name:     "Empty string",
			ref:      "",
			expected: false,
		},
		{
			name:     "Main branch",
			ref:      "main",
			expected: false,
		},
		{
			name:     "Branch with merge in name but no slash",
			ref:      "merge-feature",
			expected: false, // Doesn't contain "/merge"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMergeRef(tt.ref)
			if result != tt.expected {
				t.Errorf("isMergeRef(%q) = %v, want %v", tt.ref, result, tt.expected)
			}
		})
	}
}
