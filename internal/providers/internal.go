package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mrz1836/go-coverage/internal/analytics/report"
	"github.com/mrz1836/go-coverage/internal/badge"
	"github.com/mrz1836/go-coverage/internal/deployment"
	"github.com/mrz1836/go-coverage/internal/parser"
)

var (
	ErrInternalGitHubContextRequired     = errors.New("GitHub context is required for internal provider")
	ErrInternalCoverageDataRequired      = errors.New("coverage data is required")
	ErrInternalProviderNotInitialized    = errors.New("provider not properly initialized")
	ErrInternalNoCoverageFiles           = errors.New("no coverage files to deploy")
	ErrInternalConfigRequired            = errors.New("internal provider configuration is required")
	ErrInternalProviderConfigRequired    = errors.New("provider configuration is required")
	ErrInternalRepositoryRequired        = errors.New("repository is required for internal provider")
	ErrInternalGitHubTokenRequired       = errors.New("GitHub token is required for internal provider")
	ErrInternalRepositoryDetailsRequired = errors.New("repository owner and name are required for internal provider")
)

// InternalProvider implements the Provider interface for GitHub Pages deployment
type InternalProvider struct {
	config          *InternalProviderConfig
	deploymentMgr   deployment.DeploymentManager
	providerConfig  *Config
	coverageData    *CoverageData
	deploymentFiles map[string][]byte
	reportURL       string
}

// NewInternalProvider creates a new internal provider instance
func NewInternalProvider(config *InternalProviderConfig) *InternalProvider {
	return &InternalProvider{
		config:          config,
		deploymentFiles: make(map[string][]byte),
	}
}

// Name returns the provider name
func (p *InternalProvider) Name() string {
	return "internal"
}

// Initialize prepares the provider with the given configuration
func (p *InternalProvider) Initialize(ctx context.Context, config *Config) error {
	p.providerConfig = config

	if config.GitHubContext == nil {
		return ErrInternalGitHubContextRequired
	}

	// Create deployment manager
	deploymentMgr, err := deployment.NewManager(
		config.GitHubContext.Repository,
		config.GitHubContext.Token,
		config.DryRun,
		config.Debug,
	)
	if err != nil {
		return fmt.Errorf("failed to create deployment manager: %w", err)
	}

	p.deploymentMgr = deploymentMgr
	return nil
}

// Process handles the coverage data and prepares it for deployment
func (p *InternalProvider) Process(ctx context.Context, coverage *CoverageData) error {
	if coverage == nil {
		return ErrInternalCoverageDataRequired
	}

	p.coverageData = coverage

	// Generate coverage artifacts
	if err := p.generateCoverageArtifacts(ctx); err != nil {
		return fmt.Errorf("failed to generate coverage artifacts: %w", err)
	}

	return nil
}

// Upload deploys the processed coverage data to GitHub Pages
func (p *InternalProvider) Upload(ctx context.Context) (*UploadResult, error) {
	if p.deploymentMgr == nil {
		return nil, ErrInternalProviderNotInitialized
	}

	if len(p.deploymentFiles) == 0 {
		return nil, ErrInternalNoCoverageFiles
	}

	// Build deployment path
	deploymentPath := deployment.BuildDeploymentPath(
		p.providerConfig.GitHubContext.EventName,
		p.providerConfig.GitHubContext.Branch,
		p.providerConfig.GitHubContext.PRNumber,
	)

	// Create deployment options
	deploymentOpts := &deployment.DeploymentOptions{
		CoverageFiles:       p.deploymentFiles,
		Repository:          p.providerConfig.GitHubContext.Repository,
		Branch:              p.providerConfig.GitHubContext.Branch,
		CommitSHA:           p.providerConfig.GitHubContext.CommitSHA,
		PRNumber:            p.providerConfig.GitHubContext.PRNumber,
		EventName:           p.providerConfig.GitHubContext.EventName,
		TargetPath:          deploymentPath,
		CleanupPatterns:     p.config.CleanupPatterns,
		DryRun:              p.providerConfig.DryRun,
		Force:               p.providerConfig.Force,
		VerificationTimeout: p.config.VerificationTimeout,
	}

	// Perform deployment
	result, err := p.deploymentMgr.Deploy(ctx, deploymentOpts)
	if err != nil {
		return &UploadResult{
			Provider:   p.Name(),
			Success:    false,
			Error:      err,
			Message:    fmt.Sprintf("Deployment failed: %v", err),
			UploadTime: time.Now(),
		}, err
	}

	// Verify deployment if not in dry run mode
	if !p.providerConfig.DryRun && p.config.VerificationTimeout > 0 {
		if verifyErr := p.deploymentMgr.Verify(ctx, result); verifyErr != nil {
			// Don't fail the upload, but include warning
			result.Warnings = append(result.Warnings, fmt.Sprintf("Deployment verification failed: %v", verifyErr))
		}
	}

	// Store the report URL for later access
	p.reportURL = result.DeploymentURL

	// Create upload result
	uploadResult := &UploadResult{
		Provider:       p.Name(),
		Success:        true,
		ReportURL:      result.DeploymentURL,
		AdditionalURLs: result.AdditionalURLs,
		UploadTime:     result.DeploymentTime,
		CommitSHA:      result.CommitSHA,
		Branch:         p.providerConfig.GitHubContext.Branch,
		Metadata: map[string]interface{}{
			"files_deployed": result.FilesDeployed,
			"files_removed":  result.FilesRemoved,
			"backup_ref":     result.BackupRef,
			"warnings":       result.Warnings,
		},
	}

	return uploadResult, nil
}

// GenerateReports creates additional reports or artifacts
func (p *InternalProvider) GenerateReports(ctx context.Context) error {
	// The internal provider generates reports as part of the deployment process
	// This method can be used for any post-deployment report generation

	if p.config.EnableTrends && p.coverageData != nil {
		// Generate trend visualization (placeholder for future implementation)
		// This could generate trend charts, historical analysis, etc.
		// TODO: Implement trend generation functionality
		_ = ctx // Use ctx to avoid unused parameter warning
	}

	return nil
}

// GetReportURL returns the URL where coverage reports can be accessed
func (p *InternalProvider) GetReportURL() string {
	if p.reportURL != "" {
		return p.reportURL
	}

	// Fallback: construct URL based on GitHub context
	if p.providerConfig != nil && p.providerConfig.GitHubContext != nil {
		ctx := p.providerConfig.GitHubContext
		if ctx.Owner != "" && ctx.Repo != "" {
			baseURL := fmt.Sprintf("https://%s.github.io/%s", ctx.Owner, ctx.Repo)

			// Build path based on deployment context
			deploymentPath := deployment.BuildDeploymentPath(ctx.EventName, ctx.Branch, ctx.PRNumber)
			if deploymentPath.String() == "" {
				return baseURL + "/coverage.html"
			}
			return baseURL + "/" + deploymentPath.String() + "/coverage.html"
		}
	}

	return ""
}

// Cleanup performs any necessary cleanup operations
func (p *InternalProvider) Cleanup(ctx context.Context) error {
	// Clear deployment files from memory
	p.deploymentFiles = make(map[string][]byte)

	// The deployment manager handles its own cleanup
	// No additional cleanup needed for internal provider
	return nil
}

// Validate checks if the provider is properly configured
func (p *InternalProvider) Validate() error {
	if p.config == nil {
		return ErrInternalConfigRequired
	}

	if p.providerConfig == nil {
		return ErrInternalProviderConfigRequired
	}

	if p.providerConfig.GitHubContext == nil {
		return ErrInternalGitHubContextRequired
	}

	ctx := p.providerConfig.GitHubContext
	if ctx.Repository == "" {
		return ErrInternalRepositoryRequired
	}

	if ctx.Token == "" {
		return ErrInternalGitHubTokenRequired
	}

	if ctx.Owner == "" || ctx.Repo == "" {
		return ErrInternalRepositoryDetailsRequired
	}

	return nil
}

// Capabilities returns the capabilities supported by this provider
func (p *InternalProvider) Capabilities() ProviderCapabilities {
	return ProviderCapabilities{
		SupportsHistory:    true,
		SupportsPRComments: false, // PR comments are handled separately
		SupportsBadges:     true,
		SupportsReports:    true,
		SupportsDeployment: true,
		RequiresToken:      true,
	}
}

// generateCoverageArtifacts creates the coverage files for deployment
func (p *InternalProvider) generateCoverageArtifacts(ctx context.Context) error {
	if p.coverageData == nil {
		return ErrInternalCoverageDataRequired
	}

	// Generate HTML report using the proper report generator
	htmlContent, err := p.generateHTMLReport(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}
	p.deploymentFiles["coverage.html"] = htmlContent

	// Generate SVG badge using proper badge generator
	svgContent, err := p.generateSVGBadge(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate SVG badge: %w", err)
	}
	p.deploymentFiles["coverage.svg"] = svgContent

	// Generate JSON data using proper marshaling
	jsonContent, err := p.generateJSONData()
	if err != nil {
		return fmt.Errorf("failed to generate JSON data: %w", err)
	}
	p.deploymentFiles["coverage.json"] = jsonContent

	return nil
}

// generateHTMLReport creates an HTML coverage report using the proper report generator
func (p *InternalProvider) generateHTMLReport(ctx context.Context) ([]byte, error) {
	// Convert provider data to parser format
	parserData := p.convertToParserData()

	// Create temporary directory for report generation
	tempDir, err := os.MkdirTemp("", "coverage-report-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			// Log the error but don't fail the function
			_ = removeErr
		}
	}()

	// Configure report generator
	reportConfig := &report.Config{
		OutputDir:       tempDir,
		RepositoryOwner: p.providerConfig.GitHubContext.Owner,
		RepositoryName:  p.providerConfig.GitHubContext.Repo,
		BranchName:      p.coverageData.Branch,
		CommitSHA:       p.coverageData.CommitSHA,
		PRNumber:        p.providerConfig.GitHubContext.PRNumber,
	}

	// Create and run generator
	generator := report.NewGenerator(reportConfig)
	if genErr := generator.Generate(ctx, parserData); genErr != nil {
		return nil, fmt.Errorf("failed to generate report: %w", genErr)
	}

	// Read generated HTML file
	htmlPath := filepath.Join(tempDir, "coverage.html")
	// #nosec G304 -- htmlPath is safely constructed from our own tempDir
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read generated HTML: %w", err)
	}

	return htmlContent, nil
}

// generateSVGBadge creates an SVG coverage badge using the proper badge generator
func (p *InternalProvider) generateSVGBadge(ctx context.Context) ([]byte, error) {
	// Create badge generator
	badgeGen := badge.New()

	// Generate badge with default options
	svgContent, err := badgeGen.Generate(ctx, p.coverageData.Percentage)
	if err != nil {
		return nil, fmt.Errorf("failed to generate badge: %w", err)
	}

	return svgContent, nil
}

// generateJSONData creates JSON coverage data using proper marshaling
func (p *InternalProvider) generateJSONData() ([]byte, error) {
	// Create structured data for JSON output
	data := struct {
		Coverage     float64   `json:"coverage"`
		TotalLines   int64     `json:"total_lines"`
		CoveredLines int64     `json:"covered_lines"`
		Branch       string    `json:"branch"`
		CommitSHA    string    `json:"commit_sha"`
		Timestamp    time.Time `json:"timestamp"`
		Packages     int       `json:"packages"`
		Files        int       `json:"files"`
	}{
		Coverage:     p.coverageData.Percentage,
		TotalLines:   p.coverageData.TotalLines,
		CoveredLines: p.coverageData.CoveredLines,
		Branch:       p.coverageData.Branch,
		CommitSHA:    p.coverageData.CommitSHA,
		Timestamp:    p.coverageData.Timestamp,
		Packages:     len(p.coverageData.Packages),
		Files:        len(p.coverageData.Files),
	}

	// Marshal to JSON with proper formatting
	jsonContent, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return jsonContent, nil
}

// convertToParserData converts provider data to parser.CoverageData format
func (p *InternalProvider) convertToParserData() *parser.CoverageData {
	// Convert packages to map format expected by parser
	packages := make(map[string]*parser.PackageCoverage)

	for _, pkg := range p.coverageData.Packages {
		// Convert files to map format
		files := make(map[string]*parser.FileCoverage)
		for _, filename := range pkg.Files {
			// Find the corresponding file coverage data
			for _, fileCov := range p.coverageData.Files {
				if strings.Contains(fileCov.Filename, filename) {
					files[fileCov.Filename] = &parser.FileCoverage{
						Path:         fileCov.Filename,
						TotalLines:   int(fileCov.TotalLines),
						CoveredLines: int(fileCov.CoveredLines),
						Percentage:   fileCov.Coverage,
						Statements:   []parser.Statement{}, // Empty for now
					}
				}
			}
		}

		packages[pkg.Name] = &parser.PackageCoverage{
			Name:         pkg.Name,
			Files:        files,
			TotalLines:   int(pkg.TotalLines),
			CoveredLines: int(pkg.CoveredLines),
			Percentage:   pkg.Coverage,
		}
	}

	return &parser.CoverageData{
		Mode:         "set", // Default mode
		Packages:     packages,
		TotalLines:   int(p.coverageData.TotalLines),
		CoveredLines: int(p.coverageData.CoveredLines),
		Percentage:   p.coverageData.Percentage,
		Timestamp:    p.coverageData.Timestamp,
	}
}
