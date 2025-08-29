package deployment

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewReportGenerator(t *testing.T) {
	repository := "owner/repo"
	baseURL := "https://owner.github.io/repo"

	generator := NewReportGenerator(repository, baseURL)

	if generator.repository != repository {
		t.Errorf("Expected repository %s, got %s", repository, generator.repository)
	}

	if generator.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, generator.baseURL)
	}
}

func TestGenerateIndexHTML(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "html-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	generator := NewReportGenerator("owner/repo", "https://owner.github.io/repo")

	// Create test reports
	reports := []*ReportInfo{
		{
			Type:        PathTypeMain,
			Name:        "Main Branch",
			Path:        "main/coverage.html",
			URL:         "https://owner.github.io/repo/main/coverage.html",
			LastUpdated: time.Now(),
			Branch:      "main",
		},
		{
			Type:        PathTypeBranch,
			Name:        "Branch: feature-test",
			Path:        "branch/feature-test/coverage.html",
			URL:         "https://owner.github.io/repo/branch/feature-test/coverage.html",
			LastUpdated: time.Now(),
			Branch:      "feature-test",
		},
		{
			Type:        PathTypePR,
			Name:        "PR #123",
			Path:        "pr/123/coverage.html",
			URL:         "https://owner.github.io/repo/pr/123/coverage.html",
			LastUpdated: time.Now(),
			PRNumber:    "123",
		},
	}

	err = generator.GenerateIndexHTML(tempDir, reports)
	if err != nil {
		t.Fatalf("GenerateIndexHTML failed: %v", err)
	}

	// Check that index.html was created
	indexPath := filepath.Join(tempDir, "index.html")
	if _, statErr := os.Stat(indexPath); os.IsNotExist(statErr) {
		t.Fatal("index.html was not created")
	}

	// Read and verify content
	// #nosec G304 - indexPath is constructed from controlled tempDir
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read index.html: %v", err)
	}

	contentStr := string(content)

	// Verify HTML structure
	if !strings.Contains(contentStr, "<!DOCTYPE html>") {
		t.Error("Generated HTML is missing DOCTYPE declaration")
	}

	if !strings.Contains(contentStr, "Coverage Reports - owner/repo") {
		t.Error("Generated HTML is missing expected title")
	}

	// Verify reports are included
	if !strings.Contains(contentStr, "Main Branch") {
		t.Error("Main branch report not found in generated HTML")
	}

	if !strings.Contains(contentStr, "Branch: feature-test") {
		t.Error("Feature branch report not found in generated HTML")
	}

	if !strings.Contains(contentStr, "PR #123") {
		t.Error("PR report not found in generated HTML")
	}

	// Verify links are correct
	if !strings.Contains(contentStr, "main/coverage.html") {
		t.Error("Main branch link not found in generated HTML")
	}

	if !strings.Contains(contentStr, "branch/feature-test/coverage.html") {
		t.Error("Feature branch link not found in generated HTML")
	}

	if !strings.Contains(contentStr, "pr/123/coverage.html") {
		t.Error("PR link not found in generated HTML")
	}
}

func TestGenerateReportHTML(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "html-report-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	generator := NewReportGenerator("owner/repo", "https://owner.github.io/repo")

	reportData := []byte("<html><body>Test Coverage Report</body></html>")
	targetPath := "main/coverage.html"

	err = generator.GenerateReportHTML(tempDir, targetPath, reportData)
	if err != nil {
		t.Fatalf("GenerateReportHTML failed: %v", err)
	}

	// Check that the report file was created
	reportPath := filepath.Join(tempDir, targetPath)
	if _, statErr := os.Stat(reportPath); os.IsNotExist(statErr) {
		t.Fatal("Report file was not created")
	}

	// Check that the directory was created
	reportDir := filepath.Dir(reportPath)
	if _, statErr2 := os.Stat(reportDir); os.IsNotExist(statErr2) {
		t.Fatal("Report directory was not created")
	}

	// Verify content
	// #nosec G304 - reportPath is constructed from controlled tempDir
	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	if string(content) != string(reportData) {
		t.Errorf("Report content mismatch. Expected: %s, Got: %s", string(reportData), string(content))
	}
}

func TestDiscoverReports(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "discover-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	generator := NewReportGenerator("owner/repo", "https://owner.github.io/repo")

	// Create test report structure
	testReports := []struct {
		path    string
		content string
	}{
		{"main/coverage.html", "<html>Main coverage</html>"},
		{"branch/feature-test/coverage.html", "<html>Branch coverage</html>"},
		{"pr/123/coverage.html", "<html>PR coverage</html>"},
		{"coverage.html", "<html>Root coverage</html>"},
		{"index.html", "<html>Navigation</html>"}, // Should be ignored
		{"styles.css", "body { color: black; }"},  // Should be ignored (not HTML)
	}

	for _, report := range testReports {
		reportPath := filepath.Join(tempDir, report.path)
		reportDir := filepath.Dir(reportPath)

		if mkdirErr := os.MkdirAll(reportDir, 0o750); mkdirErr != nil {
			t.Fatalf("Failed to create directory %s: %v", reportDir, mkdirErr)
		}

		if writeErr := os.WriteFile(reportPath, []byte(report.content), 0o600); writeErr != nil {
			t.Fatalf("Failed to create report file %s: %v", report.path, writeErr)
		}
	}

	reports, err := generator.DiscoverReports(tempDir)
	if err != nil {
		t.Fatalf("DiscoverReports failed: %v", err)
	}

	// Should find 4 HTML reports (excluding root index.html and non-HTML files)
	expectedCount := 4
	if len(reports) != expectedCount {
		t.Errorf("Expected %d reports, found %d", expectedCount, len(reports))
	}

	// Verify report types are correctly identified
	typeCount := make(map[PathType]int)
	for _, report := range reports {
		typeCount[report.Type]++
	}

	if typeCount[PathTypeMain] != 1 {
		t.Errorf("Expected 1 main report, found %d", typeCount[PathTypeMain])
	}

	if typeCount[PathTypeBranch] != 1 {
		t.Errorf("Expected 1 branch report, found %d", typeCount[PathTypeBranch])
	}

	if typeCount[PathTypePR] != 1 {
		t.Errorf("Expected 1 PR report, found %d", typeCount[PathTypePR])
	}

	if typeCount[PathTypeRoot] != 1 {
		t.Errorf("Expected 1 root report, found %d", typeCount[PathTypeRoot])
	}
}

func TestCreateReportInfo(t *testing.T) {
	generator := NewReportGenerator("owner/repo", "https://owner.github.io/repo")

	tests := []struct {
		name         string
		relPath      string
		expectedType PathType
		expectedName string
	}{
		{
			name:         "Main branch report",
			relPath:      "main/coverage.html",
			expectedType: PathTypeMain,
			expectedName: "Main Branch",
		},
		{
			name:         "Feature branch report",
			relPath:      "branch/feature-test/coverage.html",
			expectedType: PathTypeBranch,
			expectedName: "Branch: feature-test",
		},
		{
			name:         "PR report",
			relPath:      "pr/123/coverage.html",
			expectedType: PathTypePR,
			expectedName: "PR #123",
		},
		{
			name:         "Root level report",
			relPath:      "coverage.html",
			expectedType: PathTypeRoot,
			expectedName: "coverage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInfo := &mockFileInfo{isDir: false}
			report := generator.createReportInfo(tt.relPath, mockInfo)

			if report == nil {
				t.Fatal("createReportInfo returned nil")
			}

			if report.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, report.Type)
			}

			if report.Name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, report.Name)
			}

			if report.Path != tt.relPath {
				t.Errorf("Expected path %s, got %s", tt.relPath, report.Path)
			}

			expectedURL := generator.baseURL + "/" + tt.relPath
			if report.URL != expectedURL {
				t.Errorf("Expected URL %s, got %s", expectedURL, report.URL)
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	generator := NewReportGenerator("owner/repo", "https://owner.github.io/repo")

	template := `Hello {{.Name}}! Repository: {{.Repository}}`
	data := struct {
		Name       string
		Repository string
	}{
		Name:       "World",
		Repository: "test/repo",
	}

	result, err := generator.renderTemplate(template, data)
	if err != nil {
		t.Fatalf("renderTemplate failed: %v", err)
	}

	expected := "Hello World! Repository: test/repo"
	if result != expected {
		t.Errorf("Expected rendered template: %s, got: %s", expected, result)
	}
}

func TestRenderTemplateError(t *testing.T) {
	generator := NewReportGenerator("owner/repo", "https://owner.github.io/repo")

	// Invalid template syntax
	template := `Hello {{.Name}! Repository: {{.Repository}}`
	data := struct{ Name string }{Name: "World"}

	_, err := generator.renderTemplate(template, data)
	if err == nil {
		t.Error("Expected renderTemplate to fail with invalid template syntax")
	}
}

func TestGenerateIndexHTMLWithEmptyReports(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "empty-html-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	generator := NewReportGenerator("owner/repo", "https://owner.github.io/repo")

	// Generate with empty reports
	err = generator.GenerateIndexHTML(tempDir, []*ReportInfo{})
	if err != nil {
		t.Fatalf("GenerateIndexHTML with empty reports failed: %v", err)
	}

	// Check that index.html was created
	indexPath := filepath.Join(tempDir, "index.html")
	if _, statErr := os.Stat(indexPath); os.IsNotExist(statErr) {
		t.Fatal("index.html was not created")
	}

	// Read and verify content
	// #nosec G304 - indexPath is constructed from controlled tempDir
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read index.html: %v", err)
	}

	contentStr := string(content)

	// Should contain empty state message
	if !strings.Contains(contentStr, "No coverage reports found") {
		t.Error("Empty state message not found in generated HTML")
	}

	// Should still contain basic structure
	if !strings.Contains(contentStr, "Coverage Reports - owner/repo") {
		t.Error("Title not found in generated HTML for empty reports")
	}
}
