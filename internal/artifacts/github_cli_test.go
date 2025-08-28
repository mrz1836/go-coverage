package artifacts

import (
	"testing"
	"time"
)

func TestParseArtifactName(t *testing.T) {
	tests := []struct {
		name              string
		artifactName      string
		expectedBranch    string
		expectedCommitSHA string
		expectedPRNumber  string
	}{
		{
			name:              "Main branch latest",
			artifactName:      "coverage-history-main-latest",
			expectedBranch:    "main",
			expectedCommitSHA: "",
			expectedPRNumber:  "",
		},
		{
			name:              "Master branch latest",
			artifactName:      "coverage-history-master-latest",
			expectedBranch:    "master",
			expectedCommitSHA: "",
			expectedPRNumber:  "",
		},
		{
			name:              "PR artifact",
			artifactName:      "coverage-history-pr-123",
			expectedBranch:    "",
			expectedCommitSHA: "",
			expectedPRNumber:  "123",
		},
		{
			name:              "PR artifact with leading zeros",
			artifactName:      "coverage-history-pr-001",
			expectedBranch:    "",
			expectedCommitSHA: "",
			expectedPRNumber:  "001",
		},
		{
			name:              "Feature branch with SHA and timestamp",
			artifactName:      "coverage-history-feature-xyz-abc123d-1633024800",
			expectedBranch:    "feature-xyz",
			expectedCommitSHA: "abc123d",
			expectedPRNumber:  "",
		},
		{
			name:              "Main branch with SHA and timestamp",
			artifactName:      "coverage-history-main-def456e-1633024800",
			expectedBranch:    "main",
			expectedCommitSHA: "def456e",
			expectedPRNumber:  "",
		},
		{
			name:              "Branch only",
			artifactName:      "coverage-history-develop",
			expectedBranch:    "develop",
			expectedCommitSHA: "",
			expectedPRNumber:  "",
		},
		{
			name:              "Complex branch name",
			artifactName:      "coverage-history-feature-issue-123-fix",
			expectedBranch:    "feature-issue-123-fix",
			expectedCommitSHA: "",
			expectedPRNumber:  "",
		},
		{
			name:              "Branch with hyphens and SHA",
			artifactName:      "coverage-history-hot-fix-urgent-a1b2c3d-1633024800",
			expectedBranch:    "hot-fix-urgent",
			expectedCommitSHA: "a1b2c3d",
			expectedPRNumber:  "",
		},
		{
			name:              "Empty after prefix",
			artifactName:      "coverage-history-",
			expectedBranch:    "",
			expectedCommitSHA: "",
			expectedPRNumber:  "",
		},
		{
			name:              "Invalid name - no prefix",
			artifactName:      "not-a-coverage-artifact",
			expectedBranch:    "",
			expectedCommitSHA: "",
			expectedPRNumber:  "",
		},
		{
			name:              "Invalid name - wrong prefix",
			artifactName:      "other-history-main-abc123",
			expectedBranch:    "",
			expectedCommitSHA: "",
			expectedPRNumber:  "",
		},
		{
			name:              "Edge case - just prefix",
			artifactName:      "coverage-history",
			expectedBranch:    "",
			expectedCommitSHA: "",
			expectedPRNumber:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branch, commitSHA, prNumber := parseArtifactName(tt.artifactName)

			if branch != tt.expectedBranch {
				t.Errorf("Expected branch '%s', got '%s'", tt.expectedBranch, branch)
			}
			if commitSHA != tt.expectedCommitSHA {
				t.Errorf("Expected commitSHA '%s', got '%s'", tt.expectedCommitSHA, commitSHA)
			}
			if prNumber != tt.expectedPRNumber {
				t.Errorf("Expected prNumber '%s', got '%s'", tt.expectedPRNumber, prNumber)
			}
		})
	}
}

func TestGenerateArtifactNameEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		opts     *UploadOptions
		contains string // What the result should contain
		exact    string // Exact match (if specified)
	}{
		{
			name:     "Empty options",
			opts:     &UploadOptions{},
			contains: "coverage-history-",
		},
		{
			name: "Only branch",
			opts: &UploadOptions{
				Branch: "develop",
			},
			contains: "coverage-history-develop-",
		},
		{
			name: "Branch with special characters",
			opts: &UploadOptions{
				Branch: "feature/issue-123",
			},
			contains: "coverage-history-feature/issue-123-",
		},
		{
			name: "Main branch with short SHA",
			opts: &UploadOptions{
				Branch:    "main",
				CommitSHA: "abc",
			},
			contains: "coverage-history-main-abc-",
		},
		{
			name: "Main branch with full SHA",
			opts: &UploadOptions{
				Branch:    "main",
				CommitSHA: "abcdef1234567890abcdef1234567890abcdef12",
			},
			contains: "coverage-history-main-abcdef1-",
		},
		{
			name: "Master branch latest",
			opts: &UploadOptions{
				Branch: "master",
			},
			contains: "coverage-history-master-latest",
			exact:    "coverage-history-master-latest",
		},
		{
			name: "PR number only",
			opts: &UploadOptions{
				PRNumber: "42",
			},
			exact: "coverage-history-pr-42",
		},
		{
			name: "PR number with branch (PR takes precedence)",
			opts: &UploadOptions{
				PRNumber:  "99",
				Branch:    "feature",
				CommitSHA: "abc123",
			},
			exact: "coverage-history-pr-99",
		},
		{
			name: "Long branch name",
			opts: &UploadOptions{
				Branch: "very-long-branch-name-that-describes-the-feature",
			},
			contains: "coverage-history-very-long-branch-name-that-describes-the-feature-",
		},
		{
			name: "Custom name takes precedence",
			opts: &UploadOptions{
				Name:      "my-custom-artifact",
				Branch:    "main",
				CommitSHA: "abc123",
				PRNumber:  "456",
			},
			exact: "my-custom-artifact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateArtifactName(tt.opts)

			if tt.exact != "" {
				if result != tt.exact {
					t.Errorf("Expected exact match '%s', got '%s'", tt.exact, result)
				}
			} else if tt.contains != "" {
				if len(result) < len(tt.contains) || result[:len(tt.contains)] != tt.contains {
					t.Errorf("Expected result to start with '%s', got '%s'", tt.contains, result)
				}
			}

			// Ensure the result always starts with the prefix (unless it's a custom name)
			if tt.opts.Name == "" && !startsWithPrefix(result, "coverage-history-") {
				t.Errorf("Expected result to start with 'coverage-history-', got '%s'", result)
			}
		})
	}
}

func TestNewGitHubCLI(t *testing.T) {
	// This test will fail in normal test environment since we're not in GitHub Actions
	cli, err := NewGitHubCLI()

	if err != nil {
		// Expected in test environment
		t.Logf("Expected error in non-GitHub Actions environment: %v", err)

		if cli != nil {
			t.Error("Expected nil client when error occurs")
		}
	} else {
		// If somehow we're in GitHub Actions, validate the client
		if cli == nil {
			t.Fatal("Expected non-nil client")
		}

		if cli.repository == "" {
			t.Error("Expected non-empty repository")
		}

		if cli.token == "" {
			t.Error("Expected non-empty token")
		}
	}
}

func TestGitHubArtifactStructure(t *testing.T) {
	// Test that our GitHubArtifact struct can handle realistic data
	now := time.Now()

	artifact := &GitHubArtifact{
		ID:                123456789,
		Name:              "coverage-history-main-latest",
		SizeInBytes:       2048,
		ArchiveURL:        "https://api.github.com/repos/owner/repo/actions/artifacts/123456789/zip",
		Expired:           false,
		CreatedAt:         now,
		ExpiresAt:         now.Add(30 * 24 * time.Hour),
		WorkflowRunID:     987654321,
		WorkflowRunHead:   "abc123def456",
		WorkflowRunBranch: "main",
	}

	if artifact.ID != 123456789 {
		t.Errorf("Expected ID 123456789, got %d", artifact.ID)
	}

	if artifact.Name != "coverage-history-main-latest" {
		t.Errorf("Expected name 'coverage-history-main-latest', got '%s'", artifact.Name)
	}

	if artifact.SizeInBytes != 2048 {
		t.Errorf("Expected size 2048, got %d", artifact.SizeInBytes)
	}

	if artifact.Expired {
		t.Error("Expected artifact to not be expired")
	}

	if artifact.WorkflowRunBranch != "main" {
		t.Errorf("Expected branch 'main', got '%s'", artifact.WorkflowRunBranch)
	}
}

func TestGitHubArtifactsResponse(t *testing.T) {
	// Test that our response struct can handle empty and populated responses

	t.Run("Empty response", func(t *testing.T) {
		response := &GitHubArtifactsResponse{
			TotalCount: 0,
			Artifacts:  []*GitHubArtifact{},
		}

		if response.TotalCount != 0 {
			t.Errorf("Expected total count 0, got %d", response.TotalCount)
		}

		if len(response.Artifacts) != 0 {
			t.Errorf("Expected 0 artifacts, got %d", len(response.Artifacts))
		}
	})

	t.Run("Populated response", func(t *testing.T) {
		now := time.Now()

		response := &GitHubArtifactsResponse{
			TotalCount: 2,
			Artifacts: []*GitHubArtifact{
				{
					ID:          1,
					Name:        "coverage-history-main-latest",
					CreatedAt:   now,
					SizeInBytes: 1024,
				},
				{
					ID:          2,
					Name:        "coverage-history-pr-123",
					CreatedAt:   now.Add(-time.Hour),
					SizeInBytes: 2048,
				},
			},
		}

		if response.TotalCount != 2 {
			t.Errorf("Expected total count 2, got %d", response.TotalCount)
		}

		if len(response.Artifacts) != 2 {
			t.Errorf("Expected 2 artifacts, got %d", len(response.Artifacts))
		}

		if response.Artifacts[0].Name != "coverage-history-main-latest" {
			t.Errorf("Expected first artifact name 'coverage-history-main-latest', got '%s'", response.Artifacts[0].Name)
		}

		if response.Artifacts[1].Name != "coverage-history-pr-123" {
			t.Errorf("Expected second artifact name 'coverage-history-pr-123', got '%s'", response.Artifacts[1].Name)
		}
	})
}

// Helper function to check if a string starts with a prefix
func startsWithPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
