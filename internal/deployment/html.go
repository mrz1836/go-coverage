package deployment

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// HTMLGenerator defines the interface for HTML generation operations
type HTMLGenerator interface {
	// GenerateIndexHTML creates the main navigation index.html file
	GenerateIndexHTML(workDir string, reports []*ReportInfo) error

	// GenerateReportHTML creates an HTML report from coverage data
	GenerateReportHTML(workDir, targetPath string, reportData []byte) error

	// DiscoverReports discovers existing coverage reports in the deployment
	DiscoverReports(workDir string) ([]*ReportInfo, error)
}

// ReportGenerator is the concrete implementation of HTMLGenerator
type ReportGenerator struct {
	repository string
	baseURL    string
}

// ReportInfo contains information about a coverage report
type ReportInfo struct {
	// Type is the report type (main, branch, pr)
	Type PathType

	// Name is the display name for the report
	Name string

	// Path is the relative path to the report
	Path string

	// URL is the full URL to the report
	URL string

	// CoveragePercent is the coverage percentage (if available)
	CoveragePercent float64

	// LastUpdated is when the report was last updated
	LastUpdated time.Time

	// FileSize is the size of the report file
	FileSize int64

	// Branch is the source branch (for branch and PR reports)
	Branch string

	// PRNumber is the PR number (for PR reports)
	PRNumber string
}

// NewReportGenerator creates a new HTML report generator
func NewReportGenerator(repository, baseURL string) *ReportGenerator {
	return &ReportGenerator{
		repository: repository,
		baseURL:    baseURL,
	}
}

// GenerateIndexHTML creates the main navigation index.html file
func (rg *ReportGenerator) GenerateIndexHTML(workDir string, reports []*ReportInfo) error {
	// Sort reports by type and name
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].Type != reports[j].Type {
			// Order: main, branch, pr
			typeOrder := map[PathType]int{PathTypeMain: 0, PathTypeBranch: 1, PathTypePR: 2}
			return typeOrder[reports[i].Type] < typeOrder[reports[j].Type]
		}
		return reports[i].Name < reports[j].Name
	})

	// Prepare template data
	data := struct {
		Repository    string
		BaseURL       string
		Reports       []*ReportInfo
		MainReport    *ReportInfo
		BranchReports []*ReportInfo
		PRReports     []*ReportInfo
		UpdatedAt     string
		TotalReports  int
	}{
		Repository:   rg.repository,
		BaseURL:      rg.baseURL,
		Reports:      reports,
		UpdatedAt:    time.Now().Format("2006-01-02 15:04:05 UTC"),
		TotalReports: len(reports),
	}

	// Group reports by type
	for _, report := range reports {
		switch report.Type {
		case PathTypeMain:
			data.MainReport = report
		case PathTypeBranch:
			data.BranchReports = append(data.BranchReports, report)
		case PathTypePR:
			data.PRReports = append(data.PRReports, report)
		case PathTypeRoot:
			// Root level reports can be treated as main reports for navigation
			if data.MainReport == nil {
				data.MainReport = report
			}
		}
	}

	// Generate HTML from template
	html, err := rg.renderTemplate(indexTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render index template: %w", err)
	}

	// Write to index.html
	indexPath := filepath.Join(workDir, "index.html")
	if err := os.WriteFile(indexPath, []byte(html), 0o600); err != nil {
		return fmt.Errorf("failed to write index.html: %w", err)
	}

	return nil
}

// GenerateReportHTML creates an HTML report from coverage data
func (rg *ReportGenerator) GenerateReportHTML(workDir, targetPath string, reportData []byte) error {
	// Ensure target directory exists
	targetDir := filepath.Dir(filepath.Join(workDir, targetPath))
	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Write report data to target path
	fullPath := filepath.Join(workDir, targetPath)
	if err := os.WriteFile(fullPath, reportData, 0o600); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}

// DiscoverReports discovers existing coverage reports in the deployment
func (rg *ReportGenerator) DiscoverReports(workDir string) ([]*ReportInfo, error) {
	var reports []*ReportInfo

	// Walk through the work directory to find reports
	err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log error and skip this path to continue processing
			log.Printf("Warning: skipping path due to error: %v", err)
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process HTML files
		if !strings.HasSuffix(info.Name(), ".html") {
			return nil
		}

		// Skip the main index.html
		if info.Name() == "index.html" && filepath.Dir(path) == workDir {
			return nil
		}

		// Create report info
		relPath, err := filepath.Rel(workDir, path)
		if err != nil {
			// Log error and skip this file
			log.Printf("Warning: failed to get relative path for %s: %v", path, err)
			return nil
		}

		report := rg.createReportInfo(relPath, info)
		if report != nil {
			reports = append(reports, report)
		}

		return nil
	})

	return reports, err
}

// createReportInfo creates a ReportInfo from a file path
func (rg *ReportGenerator) createReportInfo(relPath string, info os.FileInfo) *ReportInfo {
	pathParts := strings.Split(relPath, "/")

	report := &ReportInfo{
		Path:        relPath,
		URL:         rg.baseURL + "/" + relPath,
		LastUpdated: info.ModTime(),
		FileSize:    info.Size(),
	}

	// Determine report type and details based on path
	if len(pathParts) >= 2 {
		switch pathParts[0] {
		case "main":
			report.Type = PathTypeMain
			report.Name = "Main Branch"
			report.Branch = "main"
		case "branch":
			report.Type = PathTypeBranch
			report.Name = fmt.Sprintf("Branch: %s", pathParts[1])
			report.Branch = pathParts[1]
		case "pr":
			report.Type = PathTypePR
			report.Name = fmt.Sprintf("PR #%s", pathParts[1])
			report.PRNumber = pathParts[1]
		default:
			// Root level report
			report.Type = PathTypeRoot
			report.Name = strings.TrimSuffix(filepath.Base(relPath), ".html")
		}
	} else {
		// Root level report
		report.Type = PathTypeRoot
		report.Name = strings.TrimSuffix(filepath.Base(relPath), ".html")
	}

	return report
}

// renderTemplate renders a template with the given data
func (rg *ReportGenerator) renderTemplate(tmplText string, data interface{}) (string, error) {
	tmpl, err := template.New("index").Parse(tmplText)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// indexTemplate is the HTML template for the main navigation page
const indexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Coverage Reports - {{.Repository}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background-color: #f6f8fa;
            color: #24292e;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 6px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: #24292e;
            color: white;
            padding: 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
        }
        .header p {
            margin: 5px 0 0 0;
            opacity: 0.8;
        }
        .main-report {
            background: #28a745;
            color: white;
            padding: 20px;
            text-align: center;
        }
        .main-report h2 {
            margin: 0 0 10px 0;
        }
        .main-report a {
            color: white;
            text-decoration: none;
            font-weight: bold;
            padding: 8px 16px;
            background: rgba(255,255,255,0.2);
            border-radius: 4px;
            display: inline-block;
            margin-top: 10px;
        }
        .main-report a:hover {
            background: rgba(255,255,255,0.3);
        }
        .section {
            padding: 20px;
            border-bottom: 1px solid #e1e4e8;
        }
        .section:last-child {
            border-bottom: none;
        }
        .section h3 {
            margin: 0 0 15px 0;
            color: #24292e;
            border-bottom: 2px solid #e1e4e8;
            padding-bottom: 5px;
        }
        .report-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 15px;
        }
        .report-card {
            border: 1px solid #e1e4e8;
            border-radius: 6px;
            padding: 15px;
            background: #fff;
        }
        .report-card h4 {
            margin: 0 0 8px 0;
            color: #0366d6;
        }
        .report-card a {
            color: #0366d6;
            text-decoration: none;
        }
        .report-card a:hover {
            text-decoration: underline;
        }
        .report-meta {
            font-size: 12px;
            color: #586069;
            margin-top: 8px;
        }
        .coverage-badge {
            display: inline-block;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 11px;
            font-weight: bold;
            margin-left: 8px;
        }
        .coverage-good { background: #28a745; color: white; }
        .coverage-fair { background: #ffc107; color: black; }
        .coverage-poor { background: #dc3545; color: white; }
        .stats {
            display: flex;
            justify-content: space-around;
            background: #f6f8fa;
            padding: 15px;
            margin: 20px 0;
            border-radius: 6px;
        }
        .stat {
            text-align: center;
        }
        .stat-number {
            display: block;
            font-size: 24px;
            font-weight: bold;
            color: #0366d6;
        }
        .stat-label {
            font-size: 12px;
            color: #586069;
            text-transform: uppercase;
        }
        .footer {
            text-align: center;
            padding: 15px;
            font-size: 12px;
            color: #586069;
            background: #f6f8fa;
        }
        .empty-state {
            text-align: center;
            padding: 40px;
            color: #586069;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 Coverage Reports</h1>
            <p>{{.Repository}}</p>
        </div>

        {{if .MainReport}}
        <div class="main-report">
            <h2>🌟 Latest Coverage</h2>
            <p>{{.MainReport.Name}}</p>
            <a href="{{.MainReport.Path}}">View Full Report</a>
        </div>
        {{end}}

        <div class="stats">
            <div class="stat">
                <span class="stat-number">{{.TotalReports}}</span>
                <span class="stat-label">Total Reports</span>
            </div>
            <div class="stat">
                <span class="stat-number">{{len .BranchReports}}</span>
                <span class="stat-label">Branch Reports</span>
            </div>
            <div class="stat">
                <span class="stat-number">{{len .PRReports}}</span>
                <span class="stat-label">PR Reports</span>
            </div>
        </div>

        {{if .BranchReports}}
        <div class="section">
            <h3>🌿 Branch Coverage</h3>
            <div class="report-grid">
                {{range .BranchReports}}
                <div class="report-card">
                    <h4><a href="{{.Path}}">{{.Name}}</a></h4>
                    <div class="report-meta">
                        Updated: {{.LastUpdated.Format "Jan 2, 2006 15:04"}}
                    </div>
                </div>
                {{end}}
            </div>
        </div>
        {{end}}

        {{if .PRReports}}
        <div class="section">
            <h3>🔀 Pull Request Coverage</h3>
            <div class="report-grid">
                {{range .PRReports}}
                <div class="report-card">
                    <h4><a href="{{.Path}}">{{.Name}}</a></h4>
                    <div class="report-meta">
                        Updated: {{.LastUpdated.Format "Jan 2, 2006 15:04"}}
                    </div>
                </div>
                {{end}}
            </div>
        </div>
        {{end}}

        {{if eq .TotalReports 0}}
        <div class="empty-state">
            <h3>No coverage reports found</h3>
            <p>Coverage reports will appear here after your first deployment.</p>
        </div>
        {{end}}

        <div class="footer">
            <p>Last updated: {{.UpdatedAt}} | Generated by <a href="https://github.com/mrz1836/go-coverage">go-coverage</a></p>
        </div>
    </div>
</body>
</html>`
