package deployment

import (
	"testing"
	"time"
)

func TestDefaultCleanupPatterns(t *testing.T) {
	patterns := DefaultCleanupPatterns()

	// Check that we have expected patterns
	expectedPatterns := []string{"*.go", "*.mod", "*.sum", "*.yml", "*.yaml", "*.md"}
	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range patterns {
			if pattern == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern %s not found in default cleanup patterns", expected)
		}
	}

	// Ensure we have a reasonable number of patterns
	if len(patterns) < 10 {
		t.Errorf("Expected at least 10 cleanup patterns, got %d", len(patterns))
	}
}

func TestDefaultDeploymentOptions(t *testing.T) {
	opts := DefaultDeploymentOptions()

	if opts == nil {
		t.Fatal("DefaultDeploymentOptions returned nil")
	}

	if opts.DryRun {
		t.Error("Expected DryRun to be false by default")
	}

	if opts.Force {
		t.Error("Expected Force to be false by default")
	}

	if opts.VerificationTimeout != 30*time.Second {
		t.Errorf("Expected VerificationTimeout to be 30s, got %v", opts.VerificationTimeout)
	}

	if len(opts.CleanupPatterns) == 0 {
		t.Error("Expected default cleanup patterns to be set")
	}
}

func TestBuildDeploymentPath(t *testing.T) {
	tests := []struct {
		name      string
		eventName string
		branch    string
		prNumber  string
		expected  DeploymentPath
	}{
		{
			name:      "Pull request deployment",
			eventName: "pull_request",
			branch:    "feature-branch",
			prNumber:  "123",
			expected: DeploymentPath{
				Type:       PathTypePR,
				Root:       "pr",
				Identifier: "123",
			},
		},
		{
			name:      "Main branch deployment",
			eventName: "push",
			branch:    "main",
			prNumber:  "",
			expected: DeploymentPath{
				Type:       PathTypeMain,
				Root:       "main",
				Identifier: "main",
			},
		},
		{
			name:      "Master branch deployment",
			eventName: "push",
			branch:    "master",
			prNumber:  "",
			expected: DeploymentPath{
				Type:       PathTypeMain,
				Root:       "main",
				Identifier: "master",
			},
		},
		{
			name:      "Feature branch deployment",
			eventName: "push",
			branch:    "feature/new-functionality",
			prNumber:  "",
			expected: DeploymentPath{
				Type:       PathTypeBranch,
				Root:       "branch",
				Identifier: "feature-new-functionality",
			},
		},
		{
			name:      "PR with merge ref - should use PR path",
			eventName: "pull_request",
			branch:    "23/merge",
			prNumber:  "23",
			expected: DeploymentPath{
				Type:       PathTypePR,
				Root:       "pr",
				Identifier: "23",
			},
		},
		{
			name:      "PR target with merge ref - should use PR path",
			eventName: "pull_request_target",
			branch:    "42/merge",
			prNumber:  "42",
			expected: DeploymentPath{
				Type:       PathTypePR,
				Root:       "pr",
				Identifier: "42",
			},
		},
		{
			name:      "Branch with merge in name but no PR number",
			eventName: "push",
			branch:    "feature/merge-conflicts",
			prNumber:  "",
			expected: DeploymentPath{
				Type:       PathTypeBranch,
				Root:       "branch",
				Identifier: "feature-merge-conflicts",
			},
		},
		{
			name:      "Merge ref with PR number but wrong event - should use PR path",
			eventName: "push",
			branch:    "123/merge",
			prNumber:  "123",
			expected: DeploymentPath{
				Type:       PathTypePR,
				Root:       "pr",
				Identifier: "123",
			},
		},
		{
			name:      "Empty branch with PR number - should use PR path",
			eventName: "pull_request",
			branch:    "",
			prNumber:  "99",
			expected: DeploymentPath{
				Type:       PathTypePR,
				Root:       "pr",
				Identifier: "99",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeploymentPath(tt.eventName, tt.branch, tt.prNumber)

			if result.Type != tt.expected.Type {
				t.Errorf("Expected type %s, got %s", tt.expected.Type, result.Type)
			}

			if result.Root != tt.expected.Root {
				t.Errorf("Expected root %s, got %s", tt.expected.Root, result.Root)
			}

			if result.Identifier != tt.expected.Identifier {
				t.Errorf("Expected identifier %s, got %s", tt.expected.Identifier, result.Identifier)
			}
		})
	}
}

func TestDeploymentPathString(t *testing.T) {
	tests := []struct {
		name     string
		path     DeploymentPath
		expected string
	}{
		{
			name:     "Root path",
			path:     DeploymentPath{Type: PathTypeRoot},
			expected: "",
		},
		{
			name:     "Main branch path",
			path:     DeploymentPath{Type: PathTypeMain, Root: "main", Identifier: "main"},
			expected: "main/main",
		},
		{
			name:     "Feature branch path",
			path:     DeploymentPath{Type: PathTypeBranch, Root: "branch", Identifier: "feature-test"},
			expected: "branch/feature-test",
		},
		{
			name:     "PR path",
			path:     DeploymentPath{Type: PathTypePR, Root: "pr", Identifier: "123"},
			expected: "pr/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.path.String()
			if result != tt.expected {
				t.Errorf("Expected path string %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSanitizeBranchName(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected string
	}{
		{
			name:     "Simple branch name",
			branch:   "feature-branch",
			expected: "feature-branch",
		},
		{
			name:     "Branch with slashes",
			branch:   "feature/new-functionality",
			expected: "feature-new-functionality",
		},
		{
			name:     "Branch with multiple special characters",
			branch:   "fix/issue-123:urgent",
			expected: "fix-issue-123-urgent",
		},
		{
			name:     "Branch with spaces",
			branch:   "fix issue 123",
			expected: "fix-issue-123",
		},
		{
			name:     "Branch with Windows problematic characters",
			branch:   "feature\\fix<test>",
			expected: "feature-fix-test-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeBranchName(tt.branch)
			if result != tt.expected {
				t.Errorf("Expected sanitized name %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestReplaceAll(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		old      string
		new      string
		expected string
	}{
		{
			name:     "Simple replacement",
			s:        "hello world",
			old:      "world",
			new:      "Go",
			expected: "hello Go",
		},
		{
			name:     "Multiple replacements",
			s:        "test/test/test",
			old:      "/",
			new:      "-",
			expected: "test-test-test",
		},
		{
			name:     "No replacement needed",
			s:        "hello",
			old:      "world",
			new:      "Go",
			expected: "hello",
		},
		{
			name:     "Empty old string",
			s:        "hello",
			old:      "",
			new:      "-",
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceAll(tt.s, tt.old, tt.new)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDeploymentStatus(t *testing.T) {
	// Test that status constants are defined
	statuses := []DeploymentStatus{
		StatusPending,
		StatusSuccess,
		StatusFailed,
		StatusRolledBack,
	}

	expectedStrings := []string{
		"pending",
		"success",
		"failed",
		"rolled_back",
	}

	for i, status := range statuses {
		if string(status) != expectedStrings[i] {
			t.Errorf("Expected status %s, got %s", expectedStrings[i], string(status))
		}
	}
}

func TestPathType(t *testing.T) {
	// Test that path type constants are defined
	types := []PathType{
		PathTypeMain,
		PathTypeBranch,
		PathTypePR,
		PathTypeRoot,
	}

	expectedStrings := []string{
		"main",
		"branch",
		"pr",
		"root",
	}

	for i, pathType := range types {
		if string(pathType) != expectedStrings[i] {
			t.Errorf("Expected path type %s, got %s", expectedStrings[i], string(pathType))
		}
	}
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
			name:     "Branch with merge in name",
			ref:      "fix/merge-conflicts",
			expected: true, // Contains "/merge"
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

func TestCleanMergeRef(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected string
	}{
		{
			name:     "Standard merge ref",
			branch:   "23/merge",
			expected: "23",
		},
		{
			name:     "Another merge ref",
			branch:   "456/merge",
			expected: "456",
		},
		{
			name:     "Regular branch name",
			branch:   "feature-branch",
			expected: "feature-branch",
		},
		{
			name:     "Empty string",
			branch:   "",
			expected: "",
		},
		{
			name:     "Branch with merge in middle",
			branch:   "fix/merge/conflicts",
			expected: "fix/merge/conflicts", // Only removes /merge suffix
		},
		{
			name:     "Main branch",
			branch:   "main",
			expected: "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanMergeRef(tt.branch)
			if result != tt.expected {
				t.Errorf("cleanMergeRef(%q) = %q, want %q", tt.branch, result, tt.expected)
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "String contains substring",
			s:        "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "String does not contain substring",
			s:        "hello world",
			substr:   "foo",
			expected: false,
		},
		{
			name:     "Empty substring",
			s:        "hello world",
			substr:   "",
			expected: true,
		},
		{
			name:     "Empty string",
			s:        "",
			substr:   "foo",
			expected: false,
		},
		{
			name:     "Both empty",
			s:        "",
			substr:   "",
			expected: true,
		},
		{
			name:     "Substring longer than string",
			s:        "hi",
			substr:   "hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsString(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("containsString(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}
