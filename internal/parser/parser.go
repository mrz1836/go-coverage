// Package parser processes Go coverage profile data
package parser

import (
	"bufio"
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/mrz1836/go-coverage/internal/urlutil"
)

// Static error definitions
var (
	ErrInvalidCoverageMode    = errors.New("invalid coverage file: first line must specify mode")
	ErrMissingModeDeclaration = errors.New("invalid coverage file: missing mode declaration")
	ErrInvalidStatementFormat = errors.New("invalid statement format")
	ErrMissingColon           = errors.New("invalid statement format: missing colon")
	ErrMissingComma           = errors.New("invalid position format: missing comma")
	ErrMissingDot             = errors.New("invalid position format: missing dot")
)

// CoverageData represents parsed coverage information
// IMPORTANT: Despite the field names, TotalLines and CoveredLines actually contain
// statement counts, not line counts. This is because Go coverage works at the
// statement level, not the line level. Each statement can span multiple lines.
type CoverageData struct {
	Mode         string                      `json:"mode"`
	Packages     map[string]*PackageCoverage `json:"packages"`
	TotalLines   int                         `json:"total_lines"`   // Actually contains total statement count
	CoveredLines int                         `json:"covered_lines"` // Actually contains covered statement count
	Percentage   float64                     `json:"percentage"`
	Timestamp    time.Time                   `json:"timestamp"`
}

// PackageCoverage represents coverage data for a single package
type PackageCoverage struct {
	Name         string                   `json:"name"`
	Files        map[string]*FileCoverage `json:"files"`
	TotalLines   int                      `json:"total_lines"`   // Actually contains total statement count
	CoveredLines int                      `json:"covered_lines"` // Actually contains covered statement count
	Percentage   float64                  `json:"percentage"`
}

// FileCoverage represents coverage data for a single file
type FileCoverage struct {
	Path         string      `json:"path"`
	Statements   []Statement `json:"statements"`
	TotalLines   int         `json:"total_lines"`   // Actually contains total statement count
	CoveredLines int         `json:"covered_lines"` // Actually contains covered statement count
	Percentage   float64     `json:"percentage"`
}

// Statement represents a coverage statement in Go coverage format
type Statement struct {
	StartLine int `json:"start_line"`
	StartCol  int `json:"start_col"`
	EndLine   int `json:"end_line"`
	EndCol    int `json:"end_col"`
	NumStmt   int `json:"num_stmt"`
	Count     int `json:"count"`
}

// Parser handles Go coverage profile parsing with exclusion logic
type Parser struct {
	config *Config
}

// Config holds parser configuration
type Config struct {
	ExcludePaths     []string
	ExcludeFiles     []string
	ExcludePackages  []string
	IncludeOnlyPaths []string
	ExcludeGenerated bool
	ExcludeTestFiles bool
	MinFileLines     int
}

// New creates a new parser instance with default configuration
func New() *Parser {
	return &Parser{
		config: &Config{
			ExcludePaths:     []string{"test/", "vendor/", "examples/", "third_party/", "testdata/"},
			ExcludeFiles:     []string{"*_test.go", "*.pb.go", "*_mock.go", "mock_*.go"},
			ExcludeGenerated: true,
			ExcludeTestFiles: true,
			MinFileLines:     10,
		},
	}
}

// NewWithConfig creates a new parser instance with custom configuration
func NewWithConfig(config *Config) *Parser {
	return &Parser{config: config}
}

// ParseFile parses a coverage profile file and returns structured coverage data
func (p *Parser) ParseFile(ctx context.Context, filename string) (*CoverageData, error) {
	file, err := os.Open(filename) //nolint:gosec // filename is controlled and validated by caller
	if err != nil {
		return nil, fmt.Errorf("failed to open coverage file %q: %w", filename, err)
	}
	defer func() { _ = file.Close() }()

	return p.Parse(ctx, file)
}

// StatementWithFile represents a coverage statement with its associated file
type StatementWithFile struct {
	Statement

	Filename string
}

// Parse parses coverage data from an io.Reader
func (p *Parser) Parse(ctx context.Context, reader io.Reader) (*CoverageData, error) {
	scanner := bufio.NewScanner(reader)

	var mode string
	var statements []StatementWithFile

	lineNum := 0
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		lineNum++

		if lineNum == 1 {
			// Parse mode line: "mode: atomic" or "mode: count"
			if !strings.HasPrefix(line, "mode:") {
				return nil, fmt.Errorf("%w, got %q", ErrInvalidCoverageMode, line)
			}
			mode = strings.TrimSpace(strings.TrimPrefix(line, "mode:"))
			continue
		}

		if line == "" {
			continue
		}

		// Parse coverage statement
		stmt, file, err := p.parseStatement(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}

		// Check if file should be excluded
		if p.shouldExcludeFile(file) {
			continue
		}

		statements = append(statements, StatementWithFile{
			Statement: stmt,
			Filename:  file,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading coverage data: %w", err)
	}

	// Check if we got a valid mode
	if mode == "" {
		return nil, ErrMissingModeDeclaration
	}

	return p.buildCoverageData(mode, statements)
}

// normalizeFilePath removes the Go module prefix from file paths to create
// relative paths, e.g. "github.com/owner/repo/internal/config/config.go" becomes
// "internal/config/config.go".
//
// It delegates to urlutil.CleanModulePath, the canonical module-path cleaner used
// throughout the tool, so path handling stays consistent for any repository
// (rather than special-casing a single hardcoded module).
func normalizeFilePath(fullPath string) string {
	return urlutil.CleanModulePath(fullPath)
}

// Percent returns covered/total expressed as a percentage in the range [0, 100].
// It guards against a zero (or negative) total so callers receive 0 instead of a
// NaN from dividing by zero, keeping coverage-percentage math consistent across
// the tool.
func Percent(covered, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(covered) / float64(total) * 100
}

// parseStatement parses a single coverage statement line
func (p *Parser) parseStatement(line string) (Statement, string, error) {
	// Format: filename:startLine.startCol,endLine.endCol numStmt count
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return Statement{}, "", fmt.Errorf("%w: expected 3 fields, got %d", ErrInvalidStatementFormat, len(parts))
	}

	// Parse filename and position
	colonIdx := strings.LastIndex(parts[0], ":")
	if colonIdx == -1 {
		return Statement{}, "", fmt.Errorf("%w in %q", ErrMissingColon, parts[0])
	}

	filename := parts[0][:colonIdx]
	position := parts[0][colonIdx+1:]

	// Parse start and end positions
	commaIdx := strings.Index(position, ",")
	if commaIdx == -1 {
		return Statement{}, "", fmt.Errorf("%w in %q", ErrMissingComma, position)
	}

	startPos := position[:commaIdx]
	endPos := position[commaIdx+1:]

	startLine, startCol, err := p.parsePosition(startPos)
	if err != nil {
		return Statement{}, "", fmt.Errorf("invalid start position %q: %w", startPos, err)
	}

	endLine, endCol, err := p.parsePosition(endPos)
	if err != nil {
		return Statement{}, "", fmt.Errorf("invalid end position %q: %w", endPos, err)
	}

	// Parse number of statements
	numStmt, err := strconv.Atoi(parts[1])
	if err != nil {
		return Statement{}, "", fmt.Errorf("invalid numStmt %q: %w", parts[1], err)
	}

	// Parse count
	count, err := strconv.Atoi(parts[2])
	if err != nil {
		return Statement{}, "", fmt.Errorf("invalid count %q: %w", parts[2], err)
	}

	return Statement{
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
		NumStmt:   numStmt,
		Count:     count,
	}, filename, nil
}

// parsePosition parses a position string like "10.15" into line and column
func (p *Parser) parsePosition(pos string) (int, int, error) {
	dotIdx := strings.Index(pos, ".")
	if dotIdx == -1 {
		return 0, 0, fmt.Errorf("%w in %q", ErrMissingDot, pos)
	}

	line, err := strconv.Atoi(pos[:dotIdx])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid line number %q: %w", pos[:dotIdx], err)
	}

	col, err := strconv.Atoi(pos[dotIdx+1:])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid column number %q: %w", pos[dotIdx+1:], err)
	}

	return line, col, nil
}

// shouldExcludeFile determines if a file should be excluded from coverage
func (p *Parser) shouldExcludeFile(filename string) bool {
	// Check include-only paths first
	if len(p.config.IncludeOnlyPaths) > 0 {
		included := false
		for _, path := range p.config.IncludeOnlyPaths {
			if strings.HasPrefix(filename, path) {
				included = true
				break
			}
		}
		if !included {
			return true
		}
	}

	// Check exclude paths
	for _, path := range p.config.ExcludePaths {
		if strings.Contains(filename, path) {
			return true
		}
	}

	// Check exclude file patterns
	basename := filepath.Base(filename)
	for _, pattern := range p.config.ExcludeFiles {
		if matched, _ := filepath.Match(pattern, basename); matched {
			return true
		}
	}

	// Check exclude test files
	if p.config.ExcludeTestFiles && strings.HasSuffix(basename, "_test.go") {
		return true
	}

	// Check exclude generated files
	if p.config.ExcludeGenerated && p.isGeneratedFile(filename) {
		return true
	}

	return false
}

// isGeneratedFile checks if a file appears to be generated
func (p *Parser) isGeneratedFile(filename string) bool {
	// Common patterns for generated files
	generatedPatterns := []string{
		"// Code generated",
		"// This file was automatically generated",
		"// Code generated by protoc-gen-go",
		"// This file is generated",
	}

	file, err := os.Open(filename) //nolint:gosec // filename is controlled and validated by caller
	if err != nil {
		return false
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() && lineCount < 10 { // Check first 10 lines
		line := scanner.Text()
		for _, pattern := range generatedPatterns {
			if strings.Contains(line, pattern) {
				return true
			}
		}
		lineCount++
	}

	return false
}

// buildCoverageData constructs the final coverage data structure
func (p *Parser) buildCoverageData(mode string, statements []StatementWithFile) (*CoverageData, error) {
	packages := make(map[string]*PackageCoverage)

	// Group statements by file (normalize filenames for relative paths)
	fileStatements := make(map[string][]Statement)
	for _, stmt := range statements {
		normalizedFilename := normalizeFilePath(stmt.Filename)
		fileStatements[normalizedFilename] = append(fileStatements[normalizedFilename], stmt.Statement)
	}

	// Build coverage data structure
	totalLines := 0
	coveredLines := 0

	for filename, stmts := range fileStatements {
		pkg := p.extractPackageName(filename)

		if packages[pkg] == nil {
			packages[pkg] = &PackageCoverage{
				Name:  pkg,
				Files: make(map[string]*FileCoverage),
			}
		}

		fileCov := p.calculateFileCoverage(filename, stmts)
		packages[pkg].Files[filename] = fileCov

		packages[pkg].TotalLines += fileCov.TotalLines
		packages[pkg].CoveredLines += fileCov.CoveredLines

		totalLines += fileCov.TotalLines
		coveredLines += fileCov.CoveredLines
	}

	// Calculate package percentages
	for _, pkg := range packages {
		pkg.Percentage = Percent(pkg.CoveredLines, pkg.TotalLines)
	}

	// Calculate total percentage
	percentage := Percent(coveredLines, totalLines)

	return &CoverageData{
		Mode:         mode,
		Packages:     packages,
		TotalLines:   totalLines,
		CoveredLines: coveredLines,
		Percentage:   percentage,
		Timestamp:    time.Now(),
	}, nil
}

// extractPackageName extracts the Go package name from a file path
func (p *Parser) extractPackageName(filename string) string {
	dir := filepath.Dir(filename)
	if dir == "." {
		return "master"
	}
	return filepath.Base(dir)
}

// DiscoverEligibleFiles discovers all Go files that should be included in coverage based on exclusion rules
func (p *Parser) DiscoverEligibleFiles(ctx context.Context, rootPath string) ([]string, error) {
	var eligibleFiles []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only consider Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Convert to relative path for consistent exclusion checking
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			relPath = path
		}

		// Check if file should be excluded using the same logic as coverage parsing
		// Use relPath for pattern matching, but pass absolute path for isGeneratedFile
		if !p.shouldExcludeFileForDiscovery(relPath, path) {
			eligibleFiles = append(eligibleFiles, relPath)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to discover Go files: %w", err)
	}

	return eligibleFiles, nil
}

// shouldExcludeFileForDiscovery determines if a file should be excluded from file discovery
// It takes both relative and absolute paths for different checks
func (p *Parser) shouldExcludeFileForDiscovery(relPath, absPath string) bool {
	// Check include-only paths first (using relative path)
	if len(p.config.IncludeOnlyPaths) > 0 {
		included := false
		for _, path := range p.config.IncludeOnlyPaths {
			if strings.HasPrefix(relPath, path) {
				included = true
				break
			}
		}
		if !included {
			return true
		}
	}

	// Check exclude paths (using relative path)
	for _, path := range p.config.ExcludePaths {
		if strings.Contains(relPath, path) {
			return true
		}
	}

	// Check exclude file patterns (using relative path)
	basename := filepath.Base(relPath)
	for _, pattern := range p.config.ExcludeFiles {
		if matched, _ := filepath.Match(pattern, basename); matched {
			return true
		}
	}

	// Check exclude test files (using relative path)
	if p.config.ExcludeTestFiles && strings.HasSuffix(basename, "_test.go") {
		return true
	}

	// Check exclude generated files (using absolute path for file reading)
	if p.config.ExcludeGenerated && p.isGeneratedFile(absPath) {
		return true
	}

	return false
}

// calculateFileCoverage calculates coverage statistics for a single file
func (p *Parser) calculateFileCoverage(filename string, statements []Statement) *FileCoverage {
	slices.SortFunc(statements, func(a, b Statement) int {
		return cmp.Compare(a.StartLine, b.StartLine)
	})

	totalStmts := 0
	coveredStmts := 0

	for _, stmt := range statements {
		totalStmts += stmt.NumStmt
		if stmt.Count > 0 {
			coveredStmts += stmt.NumStmt
		}
	}

	percentage := Percent(coveredStmts, totalStmts)

	return &FileCoverage{
		Path:         filename,
		Statements:   statements,
		TotalLines:   totalStmts,
		CoveredLines: coveredStmts,
		Percentage:   percentage,
	}
}
