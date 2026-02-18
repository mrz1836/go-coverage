package parser

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestDataPath returns the absolute path to a file in the testdata directory
func getTestDataPath(filename string) string {
	_, currentFile, _, ok := runtime.Caller(1)
	if !ok {
		panic("failed to get caller information")
	}
	testDataDir := filepath.Join(filepath.Dir(currentFile), "testdata")
	return filepath.Join(testDataDir, filename)
}

func TestNew(t *testing.T) {
	parser := New()
	assert.NotNil(t, parser)
	assert.NotNil(t, parser.config)
	assert.Equal(t, []string{"test/", "vendor/", "examples/", "third_party/", "testdata/"}, parser.config.ExcludePaths)
	assert.Equal(t, []string{"*_test.go", "*.pb.go", "*_mock.go", "mock_*.go"}, parser.config.ExcludeFiles)
	assert.True(t, parser.config.ExcludeGenerated)
	assert.True(t, parser.config.ExcludeTestFiles)
	assert.Equal(t, 10, parser.config.MinFileLines)
}

func TestNewWithConfig(t *testing.T) {
	config := &Config{
		ExcludePaths:     []string{"custom/"},
		ExcludeFiles:     []string{"*.custom"},
		ExcludeGenerated: false,
		ExcludeTestFiles: false,
		MinFileLines:     5,
	}

	parser := NewWithConfig(config)
	assert.NotNil(t, parser)
	assert.Equal(t, config, parser.config)
}

func TestParseFile(t *testing.T) {
	parser := New()
	ctx := context.Background()

	// Test valid coverage file using absolute path resolution
	testFile := getTestDataPath("coverage.txt")
	coverage, err := parser.ParseFile(ctx, testFile)
	require.NoError(t, err)
	assert.NotNil(t, coverage)
	assert.Equal(t, "atomic", coverage.Mode)
	assert.True(t, coverage.Percentage >= 0 && coverage.Percentage <= 100)
	assert.Positive(t, coverage.TotalLines)
	assert.WithinDuration(t, time.Now(), coverage.Timestamp, 5*time.Second)
}

func TestParseFileNotExists(t *testing.T) {
	parser := New()
	ctx := context.Background()

	_, err := parser.ParseFile(ctx, getTestDataPath("nonexistent.txt"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open coverage file")
}

func TestParseValidCoverage(t *testing.T) {
	parser := New()
	ctx := context.Background()

	coverageData := `mode: atomic
github.com/example/pkg/file.go:10.1,12.2 2 1
github.com/example/pkg/file.go:15.1,17.16 2 0
github.com/example/pkg/other.go:20.1,22.2 1 1`

	reader := strings.NewReader(coverageData)
	coverage, err := parser.Parse(ctx, reader)

	require.NoError(t, err)
	assert.Equal(t, "atomic", coverage.Mode)
	assert.Equal(t, 5, coverage.TotalLines)            // 2 + 2 + 1 = 5 total statements
	assert.Equal(t, 3, coverage.CoveredLines)          // 2 + 0 + 1 = 3 covered statements
	assert.InDelta(t, 60.0, coverage.Percentage, 0.01) // 3/5 = 60%

	// Check packages
	assert.Len(t, coverage.Packages, 1)
	pkg, exists := coverage.Packages["pkg"]
	assert.True(t, exists)
	assert.Equal(t, "pkg", pkg.Name)
	assert.Len(t, pkg.Files, 2)
}

func TestParseInvalidMode(t *testing.T) {
	parser := New()
	ctx := context.Background()

	invalidData := `invalid mode line
github.com/example/pkg/file.go:10.1,12.2 2 1`

	reader := strings.NewReader(invalidData)
	_, err := parser.Parse(ctx, reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid coverage file: first line must specify mode")
}

func TestParseInvalidStatement(t *testing.T) {
	parser := New()
	ctx := context.Background()

	invalidData := `mode: atomic
invalid statement format`

	reader := strings.NewReader(invalidData)
	_, err := parser.Parse(ctx, reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse line")
}

func TestParseStatementValid(t *testing.T) {
	parser := New()

	stmt, filename, err := parser.parseStatement("github.com/example/pkg/file.go:10.5,12.10 2 1")

	require.NoError(t, err)
	assert.Equal(t, "github.com/example/pkg/file.go", filename)
	assert.Equal(t, 10, stmt.StartLine)
	assert.Equal(t, 5, stmt.StartCol)
	assert.Equal(t, 12, stmt.EndLine)
	assert.Equal(t, 10, stmt.EndCol)
	assert.Equal(t, 2, stmt.NumStmt)
	assert.Equal(t, 1, stmt.Count)
}

func TestParseStatementInvalidFormat(t *testing.T) {
	parser := New()

	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "too few fields",
			line: "file.go:10.5,12.10",
			want: "invalid statement format: expected 3 fields",
		},
		{
			name: "missing colon",
			line: "file.go 10.5,12.10 2 1",
			want: "invalid statement format: expected 3 fields",
		},
		{
			name: "missing comma",
			line: "file.go:10.5 12.10 2 1",
			want: "invalid statement format: expected 3 fields",
		},
		{
			name: "invalid start position",
			line: "file.go:invalid,12.10 2 1",
			want: "invalid start position",
		},
		{
			name: "invalid numStmt",
			line: "file.go:10.5,12.10 invalid 1",
			want: "invalid numStmt",
		},
		{
			name: "invalid count",
			line: "file.go:10.5,12.10 2 invalid",
			want: "invalid count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parser.parseStatement(tt.line)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestParsePosition(t *testing.T) {
	parser := New()

	line, col, err := parser.parsePosition("10.15")
	require.NoError(t, err)
	assert.Equal(t, 10, line)
	assert.Equal(t, 15, col)
}

func TestParsePositionInvalid(t *testing.T) {
	parser := New()

	tests := []struct {
		name string
		pos  string
		want string
	}{
		{
			name: "missing dot",
			pos:  "1015",
			want: "invalid position format: missing dot",
		},
		{
			name: "invalid line",
			pos:  "invalid.15",
			want: "invalid line number",
		},
		{
			name: "invalid column",
			pos:  "10.invalid",
			want: "invalid column number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parser.parsePosition(tt.pos)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestShouldExcludeFile(t *testing.T) {
	parser := New()

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "exclude test file",
			filename: "internal/config/config_test.go",
			want:     true,
		},
		{
			name:     "exclude vendor path",
			filename: "vendor/github.com/lib/pkg.go",
			want:     true,
		},
		{
			name:     "exclude testdata path",
			filename: "internal/testdata/file.go",
			want:     true,
		},
		{
			name:     "exclude protobuf file",
			filename: "internal/proto/service.pb.go",
			want:     true,
		},
		{
			name:     "exclude mock file",
			filename: "internal/mocks/mock_service.go",
			want:     true,
		},
		{
			name:     "include regular file",
			filename: "internal/config/config.go",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.shouldExcludeFile(tt.filename)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestShouldExcludeFileIncludeOnly(t *testing.T) {
	config := &Config{
		IncludeOnlyPaths: []string{"internal/", "pkg/"},
	}
	parser := NewWithConfig(config)

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "include internal path",
			filename: "internal/config/config.go",
			want:     false,
		},
		{
			name:     "include pkg path",
			filename: "pkg/utils/helper.go",
			want:     false,
		},
		{
			name:     "exclude cmd path",
			filename: "cmd/main.go",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.shouldExcludeFile(tt.filename)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestExtractPackageName(t *testing.T) {
	parser := New()

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "nested package",
			filename: "github.com/example/internal/config/config.go",
			want:     "config",
		},
		{
			name:     "root package",
			filename: "main.go",
			want:     "master",
		},
		{
			name:     "single level",
			filename: "pkg/file.go",
			want:     "pkg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.extractPackageName(tt.filename)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestCalculateFileCoverage(t *testing.T) {
	parser := New()

	statements := []Statement{
		{StartLine: 10, NumStmt: 2, Count: 1},
		{StartLine: 15, NumStmt: 3, Count: 0},
		{StartLine: 20, NumStmt: 1, Count: 2},
	}

	fileCov := parser.calculateFileCoverage("test.go", statements)

	assert.Equal(t, "test.go", fileCov.Path)
	assert.Equal(t, 6, fileCov.TotalLines)
	assert.Equal(t, 3, fileCov.CoveredLines)
	assert.InDelta(t, 50.0, fileCov.Percentage, 0.01)
	assert.Len(t, fileCov.Statements, 3)

	// Check statements are sorted by line number
	assert.Equal(t, 10, fileCov.Statements[0].StartLine)
	assert.Equal(t, 15, fileCov.Statements[1].StartLine)
	assert.Equal(t, 20, fileCov.Statements[2].StartLine)
}

func TestParseContextCancellation(t *testing.T) {
	parser := New()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	coverageData := `mode: atomic
github.com/example/pkg/file.go:10.1,12.2 2 1`

	reader := strings.NewReader(coverageData)
	_, err := parser.Parse(ctx, reader)

	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestParseComplexCoverage(t *testing.T) {
	parser := New()
	ctx := context.Background()

	// Test with complex coverage file
	coverage, err := parser.ParseFile(ctx, getTestDataPath("complex.txt"))
	require.NoError(t, err)

	assert.Equal(t, "count", coverage.Mode)
	assert.Positive(t, coverage.Percentage)
	assert.NotEmpty(t, coverage.Packages)

	// Verify vendor files are excluded
	for _, pkg := range coverage.Packages {
		for filename := range pkg.Files {
			assert.NotContains(t, filename, "vendor/")
			assert.NotContains(t, filename, "_test.go")
		}
	}
}

func TestIsGeneratedFile(t *testing.T) {
	parser := New()

	// Create a temporary generated file
	tmpFile, err := os.CreateTemp("", "generated_test_*.go")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }() //nolint:gosec // G703: tmpFile.Name() is from os.CreateTemp, not user-controlled

	// Write generated file content
	_, err = tmpFile.WriteString("// Code generated by protoc-gen-go. DO NOT EDIT.\npackage test\n")
	require.NoError(t, err)
	_ = tmpFile.Close()

	assert.True(t, parser.isGeneratedFile(tmpFile.Name()))

	// Test with non-generated file
	tmpFile2, err := os.CreateTemp("", "regular_test_*.go")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile2.Name()) }() //nolint:gosec // G703: tmpFile2.Name() is from os.CreateTemp, not user-controlled

	_, err = tmpFile2.WriteString("package test\n\nfunc main() {}\n")
	require.NoError(t, err)
	_ = tmpFile2.Close()

	assert.False(t, parser.isGeneratedFile(tmpFile2.Name()))
}

func TestParseEmptyFile(t *testing.T) {
	parser := New()
	ctx := context.Background()

	reader := strings.NewReader("")
	_, err := parser.Parse(ctx, reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid coverage file: missing mode declaration")
}

func TestParseOnlyMode(t *testing.T) {
	parser := New()
	ctx := context.Background()

	reader := strings.NewReader("mode: atomic")
	coverage, err := parser.Parse(ctx, reader)

	require.NoError(t, err)
	assert.Equal(t, "atomic", coverage.Mode)
	assert.Equal(t, 0, coverage.TotalLines)
	assert.Equal(t, 0, coverage.CoveredLines)
	assert.InDelta(t, 0.0, coverage.Percentage, 0.001)
}

// TestDiscoverEligibleFiles tests the DiscoverEligibleFiles function with various scenarios
func TestDiscoverEligibleFiles(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		setupFiles    func() (string, func()) // returns tmpDir and cleanup function
		expectedFiles []string
		expectedError bool
	}{
		{
			name: "basic Go files discovery",
			config: &Config{
				ExcludePaths:     []string{},
				ExcludeFiles:     []string{},
				ExcludeGenerated: false,
				ExcludeTestFiles: false,
			},
			setupFiles: func() (string, func()) {
				tmpDir := t.TempDir()

				// Create test files
				err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc main() {}"), 0o600)
				require.NoError(t, err)

				err = os.WriteFile(filepath.Join(tmpDir, "config.go"), []byte("package config\nfunc Config() {}"), 0o600)
				require.NoError(t, err)

				// Create non-Go file (should be ignored)
				err = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# README"), 0o600)
				require.NoError(t, err)

				return tmpDir, func() {}
			},
			expectedFiles: []string{"config.go", "main.go"},
			expectedError: false,
		},
		{
			name: "exclude test files",
			config: &Config{
				ExcludePaths:     []string{},
				ExcludeFiles:     []string{},
				ExcludeGenerated: false,
				ExcludeTestFiles: true,
			},
			setupFiles: func() (string, func()) {
				tmpDir := t.TempDir()

				err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc main() {}"), 0o600)
				require.NoError(t, err)

				err = os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte("package main\nfunc TestMain() {}"), 0o600)
				require.NoError(t, err)

				return tmpDir, func() {}
			},
			expectedFiles: []string{"main.go"},
			expectedError: false,
		},
		{
			name: "exclude specific paths",
			config: &Config{
				ExcludePaths:     []string{"vendor/", "testdata/"},
				ExcludeFiles:     []string{},
				ExcludeGenerated: false,
				ExcludeTestFiles: false,
			},
			setupFiles: func() (string, func()) {
				tmpDir := t.TempDir()

				// Create regular file
				err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc main() {}"), 0o600)
				require.NoError(t, err)

				// Create vendor directory and file
				vendorDir := filepath.Join(tmpDir, "vendor")
				err = os.MkdirAll(vendorDir, 0o750)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(vendorDir, "lib.go"), []byte("package lib\nfunc Lib() {}"), 0o600)
				require.NoError(t, err)

				// Create testdata directory and file
				testdataDir := filepath.Join(tmpDir, "testdata")
				err = os.MkdirAll(testdataDir, 0o750)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(testdataDir, "data.go"), []byte("package data\nfunc Data() {}"), 0o600)
				require.NoError(t, err)

				return tmpDir, func() {}
			},
			expectedFiles: []string{"main.go"},
			expectedError: false,
		},
		{
			name: "exclude file patterns",
			config: &Config{
				ExcludePaths:     []string{},
				ExcludeFiles:     []string{"*.pb.go", "*_mock.go"},
				ExcludeGenerated: false,
				ExcludeTestFiles: false,
			},
			setupFiles: func() (string, func()) {
				tmpDir := t.TempDir()

				err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc main() {}"), 0o600)
				require.NoError(t, err)

				err = os.WriteFile(filepath.Join(tmpDir, "proto.pb.go"), []byte("package proto\nfunc Proto() {}"), 0o600)
				require.NoError(t, err)

				err = os.WriteFile(filepath.Join(tmpDir, "service_mock.go"), []byte("package mocks\nfunc Mock() {}"), 0o600)
				require.NoError(t, err)

				return tmpDir, func() {}
			},
			expectedFiles: []string{"main.go"},
			expectedError: false,
		},
		{
			name: "include only specific paths",
			config: &Config{
				ExcludePaths:     []string{},
				ExcludeFiles:     []string{},
				IncludeOnlyPaths: []string{"internal/"},
				ExcludeGenerated: false,
				ExcludeTestFiles: false,
			},
			setupFiles: func() (string, func()) {
				tmpDir := t.TempDir()

				// Create file in root (should be excluded)
				err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc main() {}"), 0o600)
				require.NoError(t, err)

				// Create internal directory and file (should be included)
				internalDir := filepath.Join(tmpDir, "internal")
				err = os.MkdirAll(internalDir, 0o750)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(internalDir, "config.go"), []byte("package config\nfunc Config() {}"), 0o600)
				require.NoError(t, err)

				return tmpDir, func() {}
			},
			expectedFiles: []string{"internal/config.go"},
			expectedError: false,
		},
		{
			name: "exclude generated files",
			config: &Config{
				ExcludePaths:     []string{},
				ExcludeFiles:     []string{},
				ExcludeGenerated: true,
				ExcludeTestFiles: false,
			},
			setupFiles: func() (string, func()) {
				tmpDir := t.TempDir()

				err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc main() {}"), 0o600)
				require.NoError(t, err)

				// Create generated file with proper header - use first line as pattern match
				generatedContent := "// Code generated by protoc-gen-go. DO NOT EDIT.\npackage proto\nfunc Generated() {}"
				err = os.WriteFile(filepath.Join(tmpDir, "generated.go"), []byte(generatedContent), 0o600)
				require.NoError(t, err)

				return tmpDir, func() {}
			},
			expectedFiles: []string{"main.go"}, // generated.go should be excluded due to generated header
			expectedError: false,
		},
		{
			name: "nested directories",
			config: &Config{
				ExcludePaths:     []string{},
				ExcludeFiles:     []string{},
				ExcludeGenerated: false,
				ExcludeTestFiles: false,
			},
			setupFiles: func() (string, func()) {
				tmpDir := t.TempDir()

				// Create nested structure
				pkgADir := filepath.Join(tmpDir, "pkg", "a")
				err := os.MkdirAll(pkgADir, 0o750)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(pkgADir, "a.go"), []byte("package a\nfunc A() {}"), 0o600)
				require.NoError(t, err)

				pkgBDir := filepath.Join(tmpDir, "pkg", "b")
				err = os.MkdirAll(pkgBDir, 0o750)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(pkgBDir, "b.go"), []byte("package b\nfunc B() {}"), 0o600)
				require.NoError(t, err)

				return tmpDir, func() {}
			},
			expectedFiles: []string{"pkg/a/a.go", "pkg/b/b.go"},
			expectedError: false,
		},
		{
			name: "empty directory",
			config: &Config{
				ExcludePaths:     []string{},
				ExcludeFiles:     []string{},
				ExcludeGenerated: false,
				ExcludeTestFiles: false,
			},
			setupFiles: func() (string, func()) {
				tmpDir := t.TempDir()
				return tmpDir, func() {}
			},
			expectedFiles: []string{},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewWithConfig(tt.config)
			tmpDir, cleanup := tt.setupFiles()
			defer cleanup()

			ctx := context.Background()
			files, err := parser.DiscoverEligibleFiles(ctx, tmpDir)

			if tt.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, files, len(tt.expectedFiles))

			// Check that all expected files are present
			for _, expectedFile := range tt.expectedFiles {
				assert.Contains(t, files, expectedFile, "Expected file %s not found in result", expectedFile)
			}

			// Verify all returned files are Go files
			for _, file := range files {
				assert.True(t, strings.HasSuffix(file, ".go"), "Non-Go file returned: %s", file)
			}
		})
	}
}

// TestDiscoverEligibleFilesCancellation tests context cancellation
func TestDiscoverEligibleFilesCancellation(t *testing.T) {
	parser := New()
	tmpDir := t.TempDir()

	// Create some files
	err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc main() {}"), 0o600)
	require.NoError(t, err)

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = parser.DiscoverEligibleFiles(ctx, tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// TestDiscoverEligibleFilesNonExistentDirectory tests with non-existent directory
func TestDiscoverEligibleFilesNonExistentDirectory(t *testing.T) {
	parser := New()
	ctx := context.Background()

	_, err := parser.DiscoverEligibleFiles(ctx, "/non/existent/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to discover Go files")
}
