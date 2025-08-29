package providers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mrz1836/go-coverage/internal/deployment"
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
	if err := p.generateCoverageArtifacts(); err != nil {
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
func (p *InternalProvider) generateCoverageArtifacts() error {
	if p.coverageData == nil {
		return ErrInternalCoverageDataRequired
	}

	// Generate HTML report
	htmlContent := p.generateHTMLReport()
	p.deploymentFiles["coverage.html"] = htmlContent

	// Generate SVG badge
	svgContent := p.generateSVGBadge()
	p.deploymentFiles["coverage.svg"] = svgContent

	// Generate JSON data
	jsonContent := p.generateJSONData()
	p.deploymentFiles["coverage.json"] = jsonContent

	return nil
}

// generateHTMLReport creates an HTML coverage report
func (p *InternalProvider) generateHTMLReport() []byte {
	// For now, generate a basic HTML report
	// In a full implementation, this would use the existing report generator
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Coverage Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .coverage { font-size: 24px; font-weight: bold; color: %s; }
        .summary { margin: 20px 0; }
        .package { margin: 10px 0; padding: 10px; background: #f5f5f5; }
    </style>
</head>
<body>
    <h1>Coverage Report</h1>
    <div class="summary">
        <div class="coverage">Coverage: %.1f%%</div>
        <div>Total Lines: %d</div>
        <div>Covered Lines: %d</div>
        <div>Branch: %s</div>
        <div>Commit: %s</div>
        <div>Generated: %s</div>
    </div>
    <h2>Package Coverage</h2>
    %s
</body>
</html>`,
		p.getCoverageColor(p.coverageData.Percentage),
		p.coverageData.Percentage,
		p.coverageData.TotalLines,
		p.coverageData.CoveredLines,
		p.coverageData.Branch,
		p.coverageData.CommitSHA[:8],
		p.coverageData.Timestamp.Format("2006-01-02 15:04:05 UTC"),
		p.generatePackageHTML(),
	)

	return []byte(html)
}

// generateSVGBadge creates an SVG coverage badge
func (p *InternalProvider) generateSVGBadge() []byte {
	percentage := p.coverageData.Percentage
	color := p.getCoverageColorHex(percentage)

	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="104" height="20">
    <linearGradient id="a" x2="0" y2="100%%">
        <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
        <stop offset="1" stop-opacity=".1"/>
    </linearGradient>
    <rect rx="3" width="104" height="20" fill="#555"/>
    <rect rx="3" x="63" width="41" height="20" fill="%s"/>
    <path fill="%s" d="m63 0h4v20h-4z"/>
    <rect rx="3" width="104" height="20" fill="url(#a)"/>
    <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
        <text x="32.5" y="15" fill="#010101" fill-opacity=".3">coverage</text>
        <text x="32.5" y="14">coverage</text>
        <text x="82.5" y="15" fill="#010101" fill-opacity=".3">%.0f%%</text>
        <text x="82.5" y="14">%.0f%%</text>
    </g>
</svg>`, color, color, percentage, percentage)

	return []byte(svg)
}

// generateJSONData creates JSON coverage data
func (p *InternalProvider) generateJSONData() []byte {
	json := fmt.Sprintf(`{
    "coverage": %.2f,
    "total_lines": %d,
    "covered_lines": %d,
    "branch": "%s",
    "commit_sha": "%s",
    "timestamp": "%s",
    "packages": %d,
    "files": %d
}`,
		p.coverageData.Percentage,
		p.coverageData.TotalLines,
		p.coverageData.CoveredLines,
		p.coverageData.Branch,
		p.coverageData.CommitSHA,
		p.coverageData.Timestamp.Format(time.RFC3339),
		len(p.coverageData.Packages),
		len(p.coverageData.Files),
	)

	return []byte(json)
}

// generatePackageHTML creates HTML for package coverage
func (p *InternalProvider) generatePackageHTML() string {
	html := ""
	for _, pkg := range p.coverageData.Packages {
		html += fmt.Sprintf(`
    <div class="package">
        <strong>%s</strong>: %.1f%% (%d/%d lines)
    </div>`, pkg.Name, pkg.Coverage, pkg.CoveredLines, pkg.TotalLines)
	}
	return html
}

// getCoverageColor returns a color name based on coverage percentage
func (p *InternalProvider) getCoverageColor(percentage float64) string {
	if percentage >= 80 {
		return "green"
	} else if percentage >= 60 {
		return "yellow"
	} else {
		return "red"
	}
}

// getCoverageColorHex returns a hex color based on coverage percentage
func (p *InternalProvider) getCoverageColorHex(percentage float64) string {
	if percentage >= 80 {
		return "#4c1"
	} else if percentage >= 60 {
		return "#dfb317"
	} else {
		return "#e05d44"
	}
}
