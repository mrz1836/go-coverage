package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrCodecovTokenRequired         = errors.New("codecov token is required")
	ErrCodecovDataRequired          = errors.New("coverage data is required")
	ErrCodecovNoCoverageData        = errors.New("no coverage data to upload")
	ErrCodecovConfigRequired        = errors.New("codecov provider configuration is required")
	ErrCodecovAPIURLRequired        = errors.New("codecov API URL is required")
	ErrCodecovTimeoutMustBePositive = errors.New("codecov timeout must be positive")
	ErrCodecovUploadFailed          = errors.New("upload failed")
	ErrCodecovUploadError           = errors.New("codecov upload error")
)

// CodecovProvider implements the Provider interface for Codecov integration
type CodecovProvider struct {
	config         *CodecovProviderConfig
	providerConfig *Config
	coverageData   *CoverageData
	httpClient     *http.Client
	reportURL      string
}

// CodecovUploadResponse represents the response from Codecov upload API
type CodecovUploadResponse struct {
	URL     string `json:"url"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// NewCodecovProvider creates a new Codecov provider instance
func NewCodecovProvider(config *CodecovProviderConfig) *CodecovProvider {
	return &CodecovProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Name returns the provider name
func (p *CodecovProvider) Name() string {
	return "codecov"
}

// Initialize prepares the provider with the given configuration
func (p *CodecovProvider) Initialize(ctx context.Context, config *Config) error {
	p.providerConfig = config

	// Validate that we have the necessary configuration
	if p.config.Token == "" {
		return ErrCodecovTokenRequired
	}

	return nil
}

// Process handles the coverage data and prepares it for upload
func (p *CodecovProvider) Process(ctx context.Context, coverage *CoverageData) error {
	if coverage == nil {
		return ErrCodecovDataRequired
	}

	p.coverageData = coverage
	return nil
}

// Upload sends the coverage data to Codecov
func (p *CodecovProvider) Upload(ctx context.Context) (*UploadResult, error) {
	if p.coverageData == nil {
		return nil, ErrCodecovNoCoverageData
	}

	// Convert coverage data to Codecov format
	codecovData := p.formatCoverageForCodecov()

	// Upload to Codecov
	uploadURL, err := p.uploadToCodecov(ctx, codecovData)
	if err != nil {
		return &UploadResult{
			Provider:   p.Name(),
			Success:    false,
			Error:      err,
			Message:    fmt.Sprintf("Upload failed: %v", err),
			UploadTime: time.Now(),
			CommitSHA:  p.coverageData.CommitSHA,
			Branch:     p.coverageData.Branch,
		}, err
	}

	// Store the report URL
	p.reportURL = p.buildReportURL()

	// Create successful upload result
	return &UploadResult{
		Provider:   p.Name(),
		Success:    true,
		ReportURL:  p.reportURL,
		UploadTime: time.Now(),
		CommitSHA:  p.coverageData.CommitSHA,
		Branch:     p.coverageData.Branch,
		Metadata: map[string]interface{}{
			"upload_url": uploadURL,
			"flags":      p.config.Flags,
			"build":      p.config.Build,
		},
	}, nil
}

// GenerateReports creates additional reports or artifacts
func (p *CodecovProvider) GenerateReports(ctx context.Context) error {
	// Codecov handles report generation on their side
	// Nothing to do here
	return nil
}

// GetReportURL returns the URL where coverage reports can be accessed
func (p *CodecovProvider) GetReportURL() string {
	if p.reportURL != "" {
		return p.reportURL
	}
	return p.buildReportURL()
}

// Cleanup performs any necessary cleanup operations
func (p *CodecovProvider) Cleanup(ctx context.Context) error {
	// Clear coverage data from memory
	p.coverageData = nil
	return nil
}

// Validate checks if the provider is properly configured
func (p *CodecovProvider) Validate() error {
	if p.config == nil {
		return ErrCodecovConfigRequired
	}

	if p.config.Token == "" {
		return ErrCodecovTokenRequired
	}

	if p.config.APIURL == "" {
		return ErrCodecovAPIURLRequired
	}

	if p.config.Timeout <= 0 {
		return ErrCodecovTimeoutMustBePositive
	}

	return nil
}

// Capabilities returns the capabilities supported by this provider
func (p *CodecovProvider) Capabilities() ProviderCapabilities {
	return ProviderCapabilities{
		SupportsHistory:    true,
		SupportsPRComments: p.config.EnablePRComments,
		SupportsBadges:     true,
		SupportsReports:    true,
		SupportsDeployment: false, // Codecov doesn't handle deployment
		RequiresToken:      true,
	}
}

// formatCoverageForCodecov converts coverage data to Codecov format
func (p *CodecovProvider) formatCoverageForCodecov() string {
	// Codecov expects coverage data in a specific format
	// For now, we'll generate a simple format that Codecov can understand

	var buffer bytes.Buffer

	// Add header information
	buffer.WriteString("# path=coverage.out\n")
	buffer.WriteString(fmt.Sprintf("# timestamp=%s\n", p.coverageData.Timestamp.Format(time.RFC3339)))
	buffer.WriteString(fmt.Sprintf("# branch=%s\n", p.coverageData.Branch))
	buffer.WriteString(fmt.Sprintf("# commit=%s\n", p.coverageData.CommitSHA))

	// Add coverage data for each file
	for _, file := range p.coverageData.Files {
		// Convert to Codecov format: filename:line_number:coverage
		for i := int64(1); i <= file.TotalLines; i++ {
			covered := "1"
			// Check if this line is in missed lines
			for _, missedLine := range file.MissedLines {
				if int64(missedLine) == i {
					covered = "0"
					break
				}
			}
			buffer.WriteString(fmt.Sprintf("%s:%d:%s\n", file.Filename, i, covered))
		}
	}

	return buffer.String()
}

// uploadToCodecov sends the formatted coverage data to Codecov
func (p *CodecovProvider) uploadToCodecov(ctx context.Context, coverageData string) (string, error) {
	// Build upload URL
	uploadURL := p.config.APIURL + "/upload/v4"

	// Prepare form data
	formData := url.Values{}
	formData.Set("token", p.config.Token)
	formData.Set("commit", p.coverageData.CommitSHA)
	formData.Set("branch", p.coverageData.Branch)

	if p.config.Build != "" {
		formData.Set("build", p.config.Build)
	}

	// Add flags if specified
	if len(p.config.Flags) > 0 {
		formData.Set("flags", strings.Join(p.config.Flags, ","))
	}

	// Add GitHub context if available
	if p.providerConfig != nil && p.providerConfig.GitHubContext != nil {
		ctx := p.providerConfig.GitHubContext
		if ctx.Repository != "" {
			formData.Set("slug", ctx.Repository)
		}
		if ctx.PRNumber != "" {
			formData.Set("pr", ctx.PRNumber)
		}
	}

	// Create multipart form data with coverage data
	var requestBody bytes.Buffer
	requestBody.WriteString(formData.Encode())
	requestBody.WriteString("&coverage=")
	requestBody.WriteString(url.QueryEscape(coverageData))

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, &requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "go-coverage/1.0")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send upload request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read upload response: %w", err)
	}

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%w with status %d: %s", ErrCodecovUploadFailed, resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var uploadResp CodecovUploadResponse
	if err := json.Unmarshal(bodyBytes, &uploadResp); err != nil {
		// If JSON parsing fails, just return the raw response
		return string(bodyBytes), nil
	}

	if uploadResp.Error != "" {
		return "", fmt.Errorf("%w: %s", ErrCodecovUploadError, uploadResp.Error)
	}

	return uploadResp.URL, nil
}

// buildReportURL constructs the Codecov report URL
func (p *CodecovProvider) buildReportURL() string {
	if p.providerConfig == nil || p.providerConfig.GitHubContext == nil {
		return ""
	}

	ctx := p.providerConfig.GitHubContext
	if ctx.Repository == "" {
		return ""
	}

	// Build Codecov report URL
	baseURL := strings.TrimSuffix(p.config.APIURL, "/upload/v4")
	baseURL = strings.TrimSuffix(baseURL, "/api/v4")

	if ctx.PRNumber != "" {
		// PR-specific report
		return fmt.Sprintf("%s/github/%s/pull/%s", baseURL, ctx.Repository, ctx.PRNumber)
	} else {
		// Branch report
		return fmt.Sprintf("%s/github/%s/branch/%s", baseURL, ctx.Repository, ctx.Branch)
	}
}
