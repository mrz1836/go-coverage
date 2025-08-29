package deployment

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Static error definitions
var (
	ErrWorkDirNotExist        = errors.New("work directory does not exist")
	ErrCleanupPatternCritical = errors.New("cleanup pattern would remove critical coverage files")
)

// CleanupEngine defines the interface for aggressive file cleanup operations
type CleanupEngine interface {
	// CleanupFiles removes files matching the specified patterns
	CleanupFiles(workDir string, patterns, preservePatterns []string) (*CleanupResult, error)

	// ValidateCleanup validates that cleanup operations are safe
	ValidateCleanup(workDir string, patterns []string) error

	// PreviewCleanup returns a list of files that would be removed
	PreviewCleanup(workDir string, patterns, preservePatterns []string) ([]string, error)
}

// FileCleanup is the concrete implementation of CleanupEngine
type FileCleanup struct {
	dryRun  bool
	verbose bool
}

// CleanupResult contains information about the cleanup operation
type CleanupResult struct {
	// FilesRemoved is the number of files successfully removed
	FilesRemoved int

	// DirectoriesRemoved is the number of directories removed
	DirectoriesRemoved int

	// FilesPreserved is the number of files that were preserved
	FilesPreserved int

	// Errors contains any errors encountered during cleanup
	Errors []string

	// RemovedPaths contains the paths that were removed
	RemovedPaths []string

	// PreservedPaths contains the paths that were preserved
	PreservedPaths []string
}

// NewFileCleanup creates a new file cleanup engine
func NewFileCleanup(dryRun, verbose bool) *FileCleanup {
	return &FileCleanup{
		dryRun:  dryRun,
		verbose: verbose,
	}
}

// CleanupFiles removes files matching the specified patterns while preserving others
func (fc *FileCleanup) CleanupFiles(workDir string, patterns, preservePatterns []string) (*CleanupResult, error) {
	result := &CleanupResult{
		RemovedPaths:   make([]string, 0),
		PreservedPaths: make([]string, 0),
		Errors:         make([]string, 0),
	}

	// Walk through all files in the work directory
	err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error accessing %s: %v", path, err))
			return nil // Continue walking
		}

		// Skip the root directory
		if path == workDir {
			return nil
		}

		// Get relative path for pattern matching
		relPath, err := filepath.Rel(workDir, path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error getting relative path for %s: %v", path, err))
			return nil
		}

		// Skip .git directory completely
		if strings.HasPrefix(relPath, ".git/") || relPath == ".git" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be preserved
		if fc.shouldPreserve(relPath, preservePatterns) {
			result.FilesPreserved++
			result.PreservedPaths = append(result.PreservedPaths, relPath)
			return nil
		}

		// Check if file matches removal patterns
		if fc.shouldRemove(relPath, patterns, info) {
			if !fc.dryRun {
				if err := fc.removeFile(path, info); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to remove %s: %v", relPath, err))
					return nil
				}
			}

			if info.IsDir() {
				result.DirectoriesRemoved++
				return filepath.SkipDir // Skip walking into removed directory
			} else {
				result.FilesRemoved++
			}
			result.RemovedPaths = append(result.RemovedPaths, relPath)
		} else {
			result.FilesPreserved++
			result.PreservedPaths = append(result.PreservedPaths, relPath)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	return result, nil
}

// ValidateCleanup validates that cleanup operations are safe
func (fc *FileCleanup) ValidateCleanup(workDir string, patterns []string) error {
	// Check that work directory exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrWorkDirNotExist, workDir)
	}

	// Ensure patterns don't accidentally target critical files
	criticalPatterns := []string{".nojekyll", "index.html", "coverage.html", "coverage.svg", "*.html", "*.svg"}
	for _, pattern := range patterns {
		for _, critical := range criticalPatterns {
			if pattern == critical {
				return fmt.Errorf("%w: %s", ErrCleanupPatternCritical, pattern)
			}
		}
	}

	return nil
}

// PreviewCleanup returns a list of files that would be removed
func (fc *FileCleanup) PreviewCleanup(workDir string, patterns, preservePatterns []string) ([]string, error) {
	var toRemove []string

	err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log error and skip this path to continue processing
			log.Printf("Warning: skipping path due to error: %v", err)
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip root directory
		if path == workDir {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(workDir, path)
		if err != nil {
			// Log error and skip this path
			log.Printf("Warning: failed to get relative path for %s: %v", path, err)
			return nil
		}

		// Skip .git directory
		if strings.HasPrefix(relPath, ".git/") || relPath == ".git" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be preserved
		if fc.shouldPreserve(relPath, preservePatterns) {
			return nil
		}

		// Check if file matches removal patterns
		if fc.shouldRemove(relPath, patterns, info) {
			toRemove = append(toRemove, relPath)
			if info.IsDir() {
				return filepath.SkipDir
			}
		}

		return nil
	})

	return toRemove, err
}

// shouldPreserve checks if a file should be preserved based on preserve patterns
func (fc *FileCleanup) shouldPreserve(relPath string, preservePatterns []string) bool {
	// Always preserve critical GitHub Pages files
	criticalFiles := []string{".nojekyll", "index.html"}
	for _, critical := range criticalFiles {
		if relPath == critical {
			return true
		}
	}

	// Preserve coverage-related files
	coveragePatterns := []string{"*.html", "*.svg", "*.json", "*.css", "*.js"}
	for _, pattern := range coveragePatterns {
		if fc.matchPattern(relPath, pattern) {
			return true
		}
	}

	// Preserve files matching explicit preserve patterns
	for _, pattern := range preservePatterns {
		if fc.matchPattern(relPath, pattern) {
			return true
		}
	}

	return false
}

// shouldRemove checks if a file should be removed based on removal patterns
func (fc *FileCleanup) shouldRemove(relPath string, patterns []string, info os.FileInfo) bool {
	// Never remove if it's a preserved file type
	if fc.shouldPreserve(relPath, nil) {
		return false
	}

	// Check against removal patterns
	for _, pattern := range patterns {
		if fc.matchPattern(relPath, pattern) {
			return true
		}
	}

	// Remove common unwanted file types
	unwantedExtensions := []string{".go", ".mod", ".sum", ".yml", ".yaml", ".md", ".txt", ".log"}
	for _, ext := range unwantedExtensions {
		if strings.HasSuffix(relPath, ext) {
			return true
		}
	}

	// Remove common unwanted directories
	if info.IsDir() {
		unwantedDirs := []string{"cmd", "internal", "pkg", "test", "testdata", "docs", "examples", "scripts", "tools", ".github"}
		for _, dir := range unwantedDirs {
			if relPath == dir || strings.HasPrefix(relPath, dir+"/") {
				return true
			}
		}
	}

	return false
}

// matchPattern performs simple pattern matching with wildcards
func (fc *FileCleanup) matchPattern(path, pattern string) bool {
	// Handle simple wildcard patterns
	if strings.Contains(pattern, "*") {
		return fc.matchWildcard(path, pattern)
	}

	// Exact match or directory match
	return path == pattern || strings.HasPrefix(path, pattern+"/")
}

// matchWildcard performs wildcard pattern matching
func (fc *FileCleanup) matchWildcard(path, pattern string) bool {
	// Simple implementation for common patterns like "*.go"
	if strings.HasPrefix(pattern, "*.") {
		ext := pattern[1:] // Remove the *
		return strings.HasSuffix(path, ext)
	}

	// Pattern like "test/*"
	if strings.HasSuffix(pattern, "/*") {
		prefix := pattern[:len(pattern)-2]
		return strings.HasPrefix(path, prefix+"/")
	}

	// More complex patterns could be implemented here
	// For now, fall back to simple string matching
	return strings.Contains(path, strings.ReplaceAll(pattern, "*", ""))
}

// removeFile removes a file or directory
func (fc *FileCleanup) removeFile(path string, info os.FileInfo) error {
	if info.IsDir() {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

// DefaultPreservePatterns returns patterns for files that should always be preserved
func DefaultPreservePatterns() []string {
	return []string{
		"*.html",      // HTML reports
		"*.svg",       // Badge files
		"*.css",       // Stylesheets
		"*.js",        // JavaScript files
		"*.json",      // Data files
		"*.png",       // Images
		"*.jpg",       // Images
		"*.jpeg",      // Images
		"*.gif",       // Images
		"*.ico",       // Icons
		".nojekyll",   // GitHub Pages config
		"CNAME",       // Custom domain config
		"robots.txt",  // SEO config
		"sitemap.xml", // SEO config
		"favicon.*",   // Favicon files
		"manifest.*",  // Web manifest
	}
}
