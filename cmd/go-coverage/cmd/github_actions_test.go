package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewGitHubActionsCmd(t *testing.T) {
	// Create commands instance
	cmds := &Commands{}
	cmd := cmds.newGitHubActionsCmd()

	// Test basic command properties
	if cmd.Use != "github-actions" {
		t.Errorf("Expected Use to be 'github-actions', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected Long description to be set")
	}

	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}

	// Test flags are present
	flags := []string{"input", "provider", "dry-run", "debug", "auto-detect", "force"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag '%s' to be present", flag)
		}
	}
}

func TestGitHubActionsFlags(t *testing.T) {
	cmds := &Commands{}
	cmd := cmds.newGitHubActionsCmd()

	// Test default values
	tests := []struct {
		flag         string
		expectedType string
		defaultValue interface{}
	}{
		{"input", "string", ""},
		{"provider", "string", "auto"},
		{"dry-run", "bool", false},
		{"debug", "bool", false},
		{"auto-detect", "bool", true},
		{"force", "bool", false},
	}

	for _, test := range tests {
		flag := cmd.Flags().Lookup(test.flag)
		if flag == nil {
			t.Errorf("Flag '%s' not found", test.flag)
			continue
		}

		if flag.Value.Type() != test.expectedType {
			t.Errorf("Flag '%s' expected type %s, got %s", test.flag, test.expectedType, flag.Value.Type())
		}

		if flag.DefValue != stringValue(test.defaultValue) {
			t.Errorf("Flag '%s' expected default value %v, got %s", test.flag, test.defaultValue, flag.DefValue)
		}
	}
}

func TestLoadEnvFiles(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	// Create .github directory
	githubDir := filepath.Join(tempDir, ".github")
	if err := os.MkdirAll(githubDir, 0o750); err != nil {
		t.Fatalf("Failed to create .github directory: %v", err)
	}

	tests := []struct {
		name       string
		baseFile   string
		customFile string
		expected   map[string]string
	}{
		{
			name:     "No files",
			expected: map[string]string{},
		},
		{
			name:     "Base file only",
			baseFile: "GO_COVERAGE_THRESHOLD=80\nGO_COVERAGE_BADGE_STYLE=flat",
			expected: map[string]string{
				"GO_COVERAGE_THRESHOLD":   "80",
				"GO_COVERAGE_BADGE_STYLE": "flat",
			},
		},
		{
			name:       "Both files with override",
			baseFile:   "GO_COVERAGE_THRESHOLD=80\nGO_COVERAGE_BADGE_STYLE=flat",
			customFile: "GO_COVERAGE_THRESHOLD=90\nGO_COVERAGE_DEBUG=true",
			expected: map[string]string{
				"GO_COVERAGE_THRESHOLD":   "90", // overridden by custom
				"GO_COVERAGE_BADGE_STYLE": "flat",
				"GO_COVERAGE_DEBUG":       "true",
			},
		},
		{
			name:     "File with comments and empty lines",
			baseFile: "# This is a comment\nGO_COVERAGE_THRESHOLD=80\n\n# Another comment\nGO_COVERAGE_BADGE_STYLE=flat\n",
			expected: map[string]string{
				"GO_COVERAGE_THRESHOLD":   "80",
				"GO_COVERAGE_BADGE_STYLE": "flat",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up previous test files
			_ = os.Remove(filepath.Join(githubDir, ".env.base"))
			_ = os.Remove(filepath.Join(githubDir, ".env.custom"))

			// Clear environment variables
			for key := range tt.expected {
				_ = os.Unsetenv(key)
			}

			// Create test files
			if tt.baseFile != "" {
				if err := os.WriteFile(filepath.Join(githubDir, ".env.base"), []byte(tt.baseFile), 0o600); err != nil {
					t.Fatalf("Failed to create base file: %v", err)
				}
			}
			if tt.customFile != "" {
				if err := os.WriteFile(filepath.Join(githubDir, ".env.custom"), []byte(tt.customFile), 0o600); err != nil {
					t.Fatalf("Failed to create custom file: %v", err)
				}
			}

			// Run loadEnvFiles
			err := loadEnvFiles()
			if err != nil {
				t.Fatalf("loadEnvFiles() returned error: %v", err)
			}

			// Check environment variables
			for key, expectedValue := range tt.expected {
				actualValue := os.Getenv(key)
				if actualValue != expectedValue {
					t.Errorf("Environment variable %s = %s, want %s", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestParseEnvContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name:     "Empty content",
			content:  "",
			expected: map[string]string{},
		},
		{
			name:    "Simple key-value pairs",
			content: "KEY1=value1\nKEY2=value2",
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name:    "With comments and empty lines",
			content: "# Comment\nKEY1=value1\n\n# Another comment\nKEY2=value2\n",
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name:    "With whitespace",
			content: " KEY1 = value1 \n KEY2=value2\n",
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name:    "Malformed lines",
			content: "KEY1=value1\nMALFORMED_LINE\nKEY2=value2\n",
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			for key := range tt.expected {
				_ = os.Unsetenv(key)
			}

			// Parse content
			parseEnvContent(tt.content)

			// Check environment variables
			for key, expectedValue := range tt.expected {
				actualValue := os.Getenv(key)
				if actualValue != expectedValue {
					t.Errorf("Environment variable %s = %s, want %s", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "Single line",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "Multiple lines",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "With empty lines",
			input:    "line1\n\nline3",
			expected: []string{"line1", "", "line3"},
		},
		{
			name:     "Trailing newline",
			input:    "line1\nline2\n",
			expected: []string{"line1", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitLines() returned %d lines, want %d", len(result), len(tt.expected))
			}

			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("Missing line %d: want %s", i, expected)
					continue
				}
				if result[i] != expected {
					t.Errorf("Line %d: got %s, want %s", i, result[i], expected)
				}
			}
		})
	}
}

func TestSplitOnFirst(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      byte
		expected []string
	}{
		{
			name:     "No separator",
			input:    "hello",
			sep:      '=',
			expected: []string{"hello"},
		},
		{
			name:     "Single separator",
			input:    "key=value",
			sep:      '=',
			expected: []string{"key", "value"},
		},
		{
			name:     "Multiple separators",
			input:    "key=value=with=equals",
			sep:      '=',
			expected: []string{"key", "value=with=equals"},
		},
		{
			name:     "Empty parts",
			input:    "=value",
			sep:      '=',
			expected: []string{"", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitOnFirst(tt.input, tt.sep)
			if len(result) != len(tt.expected) {
				t.Errorf("splitOnFirst() returned %d parts, want %d", len(result), len(tt.expected))
			}

			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("Missing part %d: want %s", i, expected)
					continue
				}
				if result[i] != expected {
					t.Errorf("Part %d: got %s, want %s", i, result[i], expected)
				}
			}
		})
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No whitespace",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "Leading spaces",
			input:    "   hello",
			expected: "hello",
		},
		{
			name:     "Trailing spaces",
			input:    "hello   ",
			expected: "hello",
		},
		{
			name:     "Both sides",
			input:    "  hello  ",
			expected: "hello",
		},
		{
			name:     "Tabs and spaces",
			input:    "\t  hello  \t",
			expected: "hello",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only whitespace",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimSpace(tt.input)
			if result != tt.expected {
				t.Errorf("trimSpace(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to convert interface{} to string for comparison
func stringValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}
