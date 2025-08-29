package deployment

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewFileCleanup(t *testing.T) {
	cleanup := NewFileCleanup(true, true)

	if !cleanup.dryRun {
		t.Error("Expected dryRun to be true")
	}

	if !cleanup.verbose {
		t.Error("Expected verbose to be true")
	}
}

func TestDefaultPreservePatterns(t *testing.T) {
	patterns := DefaultPreservePatterns()

	expectedPatterns := []string{"*.html", "*.svg", "*.css", "*.js", "*.json", ".nojekyll"}
	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range patterns {
			if pattern == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected preserve pattern %s not found", expected)
		}
	}

	// Ensure we have a reasonable number of patterns
	if len(patterns) < 10 {
		t.Errorf("Expected at least 10 preserve patterns, got %d", len(patterns))
	}
}

func TestShouldPreserve(t *testing.T) {
	cleanup := NewFileCleanup(true, false)

	tests := []struct {
		name     string
		relPath  string
		patterns []string
		expected bool
	}{
		{
			name:     "Critical file - .nojekyll",
			relPath:  ".nojekyll",
			patterns: []string{},
			expected: true,
		},
		{
			name:     "Critical file - index.html",
			relPath:  "index.html",
			patterns: []string{},
			expected: true,
		},
		{
			name:     "HTML file",
			relPath:  "coverage.html",
			patterns: []string{},
			expected: true,
		},
		{
			name:     "SVG file",
			relPath:  "coverage.svg",
			patterns: []string{},
			expected: true,
		},
		{
			name:     "CSS file",
			relPath:  "styles.css",
			patterns: []string{},
			expected: true,
		},
		{
			name:     "JavaScript file",
			relPath:  "script.js",
			patterns: []string{},
			expected: true,
		},
		{
			name:     "JSON file",
			relPath:  "data.json",
			patterns: []string{},
			expected: true,
		},
		{
			name:     "Go file",
			relPath:  "main.go",
			patterns: []string{},
			expected: false,
		},
		{
			name:     "Custom preserve pattern",
			relPath:  "special.txt",
			patterns: []string{"special.txt"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanup.shouldPreserve(tt.relPath, tt.patterns)
			if result != tt.expected {
				t.Errorf("Expected shouldPreserve(%s) = %v, got %v", tt.relPath, tt.expected, result)
			}
		})
	}
}

func TestShouldRemove(t *testing.T) {
	cleanup := NewFileCleanup(true, false)

	tests := []struct {
		name     string
		relPath  string
		patterns []string
		isDir    bool
		expected bool
	}{
		{
			name:     "Go file",
			relPath:  "main.go",
			patterns: []string{"*.go"},
			isDir:    false,
			expected: true,
		},
		{
			name:     "Yaml file",
			relPath:  "config.yml",
			patterns: []string{"*.yml"},
			isDir:    false,
			expected: true,
		},
		{
			name:     "Markdown file",
			relPath:  "README.md",
			patterns: []string{"*.md"},
			isDir:    false,
			expected: true,
		},
		{
			name:     "HTML file should be preserved",
			relPath:  "report.html",
			patterns: []string{"*.html"},
			isDir:    false,
			expected: false, // Preserved due to shouldPreserve
		},
		{
			name:     "Directory - cmd",
			relPath:  "cmd",
			patterns: []string{"cmd/"},
			isDir:    true,
			expected: true,
		},
		{
			name:     "Directory - internal",
			relPath:  "internal",
			patterns: []string{"internal/"},
			isDir:    true,
			expected: true,
		},
		{
			name:     "File in unwanted directory",
			relPath:  "cmd/main.go",
			patterns: []string{},
			isDir:    false,
			expected: true,
		},
		{
			name:     "Unwanted extension without pattern",
			relPath:  "test.log",
			patterns: []string{},
			isDir:    false,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &mockFileInfo{isDir: tt.isDir}
			result := cleanup.shouldRemove(tt.relPath, tt.patterns, info)
			if result != tt.expected {
				t.Errorf("Expected shouldRemove(%s) = %v, got %v", tt.relPath, tt.expected, result)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	cleanup := NewFileCleanup(true, false)

	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		{
			name:     "Wildcard extension match",
			path:     "main.go",
			pattern:  "*.go",
			expected: true,
		},
		{
			name:     "Wildcard extension no match",
			path:     "main.js",
			pattern:  "*.go",
			expected: false,
		},
		{
			name:     "Directory wildcard match",
			path:     "test/file.go",
			pattern:  "test/*",
			expected: true,
		},
		{
			name:     "Exact match",
			path:     "README.md",
			pattern:  "README.md",
			expected: true,
		},
		{
			name:     "Directory prefix match",
			path:     "cmd/main.go",
			pattern:  "cmd",
			expected: true,
		},
		{
			name:     "No match",
			path:     "coverage.html",
			pattern:  "*.go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanup.matchPattern(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("Expected matchPattern(%s, %s) = %v, got %v", tt.path, tt.pattern, tt.expected, result)
			}
		})
	}
}

func TestMatchWildcard(t *testing.T) {
	cleanup := NewFileCleanup(true, false)

	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		{
			name:     "Extension wildcard match",
			path:     "test.go",
			pattern:  "*.go",
			expected: true,
		},
		{
			name:     "Extension wildcard no match",
			path:     "test.js",
			pattern:  "*.go",
			expected: false,
		},
		{
			name:     "Directory wildcard match",
			path:     "test/file.txt",
			pattern:  "test/*",
			expected: true,
		},
		{
			name:     "Directory wildcard no match",
			path:     "src/file.txt",
			pattern:  "test/*",
			expected: false,
		},
		{
			name:     "Complex wildcard",
			path:     "testfile.backup",
			pattern:  "*test*",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanup.matchWildcard(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("Expected matchWildcard(%s, %s) = %v, got %v", tt.path, tt.pattern, tt.expected, result)
			}
		})
	}
}

func TestValidateCleanup(t *testing.T) {
	cleanup := NewFileCleanup(true, false)

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cleanup-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	tests := []struct {
		name      string
		workDir   string
		patterns  []string
		expectErr bool
	}{
		{
			name:      "Valid directory and patterns",
			workDir:   tempDir,
			patterns:  []string{"*.go", "*.md"},
			expectErr: false,
		},
		{
			name:      "Non-existent directory",
			workDir:   "/non/existent/path",
			patterns:  []string{"*.go"},
			expectErr: true,
		},
		{
			name:      "Dangerous pattern - .nojekyll",
			workDir:   tempDir,
			patterns:  []string{".nojekyll"},
			expectErr: true,
		},
		{
			name:      "Dangerous pattern - index.html",
			workDir:   tempDir,
			patterns:  []string{"index.html"},
			expectErr: true,
		},
		{
			name:      "Dangerous pattern - *.html",
			workDir:   tempDir,
			patterns:  []string{"*.html"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cleanup.ValidateCleanup(tt.workDir, tt.patterns)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got error: %v", tt.expectErr, err)
			}
		})
	}
}

func TestPreviewCleanup(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "cleanup-preview-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test files
	testFiles := []string{
		"main.go",
		"README.md",
		"coverage.html",
		"coverage.svg",
		"styles.css",
		".nojekyll",
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		if writeErr := os.WriteFile(filePath, []byte("test content"), 0o600); writeErr != nil {
			t.Fatalf("Failed to create test file %s: %v", file, writeErr)
		}
	}

	// Create test directory
	testDir := filepath.Join(tempDir, "cmd")
	if mkdirErr := os.MkdirAll(testDir, 0o750); mkdirErr != nil {
		t.Fatalf("Failed to create test directory: %v", mkdirErr)
	}

	cleanup := NewFileCleanup(true, false)
	patterns := []string{"*.go", "*.md", "cmd/"}
	preservePatterns := DefaultPreservePatterns()

	toRemove, err := cleanup.PreviewCleanup(tempDir, patterns, preservePatterns)
	if err != nil {
		t.Fatalf("PreviewCleanup failed: %v", err)
	}

	// Check that Go and Markdown files are marked for removal
	expectedRemove := []string{"main.go", "README.md", "cmd"}

	for _, expected := range expectedRemove {
		found := false
		for _, removed := range toRemove {
			if removed == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s to be marked for removal", expected)
		}
	}

	// Check that preserved files are NOT marked for removal
	expectedPreserve := []string{"coverage.html", "coverage.svg", "styles.css", ".nojekyll"}

	for _, expected := range expectedPreserve {
		for _, removed := range toRemove {
			if removed == expected {
				t.Errorf("Expected %s to be preserved, but it was marked for removal", expected)
			}
		}
	}
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	isDir bool
}

func (m *mockFileInfo) Name() string       { return "test" }
func (m *mockFileInfo) Size() int64        { return 100 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0o644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }
