// Package health provides comprehensive health checks for system readiness,
// including GitHub API connectivity, permissions, disk space, and network connectivity.
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Name        string        `json:"name"`
	Status      CheckStatus   `json:"status"`
	Message     string        `json:"message"`
	Duration    time.Duration `json:"duration"`
	Details     interface{}   `json:"details,omitempty"`
	Suggestions []string      `json:"suggestions,omitempty"`
}

// CheckStatus represents the status of a health check
type CheckStatus string

const (
	StatusHealthy   CheckStatus = "healthy"
	StatusWarning   CheckStatus = "warning"
	StatusUnhealthy CheckStatus = "unhealthy"
	StatusSkipped   CheckStatus = "skipped"
)

// HealthReport contains the results of all health checks
type HealthReport struct {
	OverallStatus CheckStatus   `json:"overall_status"`
	Timestamp     time.Time     `json:"timestamp"`
	Checks        []CheckResult `json:"checks"`
	Summary       HealthSummary `json:"summary"`
}

// HealthSummary provides a summary of health check results
type HealthSummary struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Warnings  int `json:"warnings"`
	Unhealthy int `json:"unhealthy"`
	Skipped   int `json:"skipped"`
}

// HealthChecker defines the interface for health checks
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) *CheckResult
}

// Manager manages and executes health checks
type Manager struct {
	checkers []HealthChecker
	timeout  time.Duration
}

// NewManager creates a new health check manager
func NewManager() *Manager {
	return &Manager{
		checkers: make([]HealthChecker, 0),
		timeout:  30 * time.Second,
	}
}

// AddChecker adds a health checker to the manager
func (m *Manager) AddChecker(checker HealthChecker) *Manager {
	m.checkers = append(m.checkers, checker)
	return m
}

// SetTimeout sets the timeout for health checks
func (m *Manager) SetTimeout(timeout time.Duration) *Manager {
	m.timeout = timeout
	return m
}

// CheckAll runs all registered health checks
func (m *Manager) CheckAll(ctx context.Context) *HealthReport {
	report := &HealthReport{
		Timestamp: time.Now(),
		Checks:    make([]CheckResult, 0, len(m.checkers)),
		Summary:   HealthSummary{},
	}

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	// Run all health checks
	for _, checker := range m.checkers {
		select {
		case <-checkCtx.Done():
			// Context canceled or timeout
			result := &CheckResult{
				Name:     checker.Name(),
				Status:   StatusUnhealthy,
				Message:  "Health check timed out",
				Duration: m.timeout,
			}
			report.Checks = append(report.Checks, *result)
			report.Summary.Unhealthy++
		default:
			result := m.runSingleCheck(checkCtx, checker)
			report.Checks = append(report.Checks, *result)

			// Update summary
			switch result.Status {
			case StatusHealthy:
				report.Summary.Healthy++
			case StatusWarning:
				report.Summary.Warnings++
			case StatusUnhealthy:
				report.Summary.Unhealthy++
			case StatusSkipped:
				report.Summary.Skipped++
			}
		}
	}

	report.Summary.Total = len(report.Checks)

	// Determine overall status
	report.OverallStatus = m.calculateOverallStatus(&report.Summary)

	return report
}

// runSingleCheck runs a single health check with timing
func (m *Manager) runSingleCheck(ctx context.Context, checker HealthChecker) *CheckResult {
	start := time.Now()

	// Run the check with a defer to catch panics
	var result *CheckResult

	defer func() {
		if r := recover(); r != nil {
			result = &CheckResult{
				Name:     checker.Name(),
				Status:   StatusUnhealthy,
				Message:  fmt.Sprintf("Health check panicked: %v", r),
				Duration: time.Since(start),
			}
		}
	}()

	result = checker.Check(ctx)
	if result == nil {
		result = &CheckResult{
			Name:     checker.Name(),
			Status:   StatusUnhealthy,
			Message:  "Health check returned nil result",
			Duration: time.Since(start),
		}
	} else {
		result.Duration = time.Since(start)
	}

	return result
}

// calculateOverallStatus determines the overall health status
func (m *Manager) calculateOverallStatus(summary *HealthSummary) CheckStatus {
	if summary.Unhealthy > 0 {
		return StatusUnhealthy
	}
	if summary.Warnings > 0 {
		return StatusWarning
	}
	if summary.Healthy > 0 {
		return StatusHealthy
	}
	return StatusSkipped
}

// GitHubAPIChecker checks GitHub API connectivity and permissions
type GitHubAPIChecker struct {
	Token      string
	Repository string // Format: "owner/repo"
	BaseURL    string
}

// NewGitHubAPIChecker creates a new GitHub API health checker
func NewGitHubAPIChecker(token, repository string) *GitHubAPIChecker {
	return &GitHubAPIChecker{
		Token:      token,
		Repository: repository,
		BaseURL:    "https://api.github.com",
	}
}

// Name returns the name of this health checker
func (c *GitHubAPIChecker) Name() string {
	return "GitHub API"
}

// Check performs the GitHub API health check
func (c *GitHubAPIChecker) Check(ctx context.Context) *CheckResult {
	result := &CheckResult{
		Name:   c.Name(),
		Status: StatusHealthy,
	}

	// Check if token is provided
	if c.Token == "" {
		result.Status = StatusUnhealthy
		result.Message = "GitHub token is not provided"
		result.Suggestions = []string{
			"Set GITHUB_TOKEN environment variable",
			"Ensure token has required permissions",
		}
		return result
	}

	// Check if repository is provided
	if c.Repository == "" {
		result.Status = StatusUnhealthy
		result.Message = "GitHub repository is not specified"
		result.Suggestions = []string{
			"Set GITHUB_REPOSITORY environment variable",
			"Ensure format is 'owner/repo'",
		}
		return result
	}

	// Test basic API connectivity
	apiURL := fmt.Sprintf("%s/user", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Failed to create API request: %v", err)
		return result
	}

	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Failed to connect to GitHub API: %v", err)
		result.Suggestions = []string{
			"Check internet connectivity",
			"Verify GitHub API is accessible",
			"Check firewall settings",
		}
		return result
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	if resp.StatusCode == http.StatusUnauthorized {
		result.Status = StatusUnhealthy
		result.Message = "GitHub token is invalid or expired"
		result.Suggestions = []string{
			"Verify token is correct",
			"Check token expiration",
			"Regenerate token if needed",
		}
		return result
	}

	if resp.StatusCode >= 400 {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("GitHub API returned error: %d", resp.StatusCode)
		result.Suggestions = []string{
			"Check GitHub API status",
			"Verify token permissions",
			"Try again later if rate limited",
		}
		return result
	}

	// Test repository access
	repoCheckResult := c.checkRepositoryAccess(ctx, client)
	if repoCheckResult.Status != StatusHealthy {
		return repoCheckResult
	}

	result.Status = StatusHealthy
	result.Message = "GitHub API is accessible and token has required permissions"
	result.Details = map[string]interface{}{
		"api_url":      c.BaseURL,
		"repository":   c.Repository,
		"token_length": len(c.Token),
	}

	return result
}

// checkRepositoryAccess verifies access to the specified repository
func (c *GitHubAPIChecker) checkRepositoryAccess(ctx context.Context, client *http.Client) *CheckResult {
	result := &CheckResult{
		Name: c.Name(),
	}

	// Test repository access
	repoURL := fmt.Sprintf("%s/repos/%s", c.BaseURL, c.Repository)
	req, err := http.NewRequestWithContext(ctx, "GET", repoURL, nil)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Failed to create repository request: %v", err)
		return result
	}

	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Failed to access repository: %v", err)
		result.Suggestions = []string{
			"Check repository exists",
			"Verify token has repository permissions",
		}
		return result
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		result.Status = StatusUnhealthy
		result.Message = "Repository not found or no access"
		result.Suggestions = []string{
			"Verify repository name format (owner/repo)",
			"Check token has repository access",
			"Ensure repository exists",
		}
		return result
	}

	if resp.StatusCode >= 400 {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Repository access error: %d", resp.StatusCode)
		return result
	}

	result.Status = StatusHealthy
	result.Message = "Repository is accessible"
	return result
}

// DiskSpaceChecker checks available disk space
type DiskSpaceChecker struct {
	Path           string  // Path to check disk space for
	MinFreeBytes   int64   // Minimum free bytes required
	MinFreePercent float64 // Minimum free percentage required
}

// NewDiskSpaceChecker creates a new disk space health checker
func NewDiskSpaceChecker(path string, minFreeBytes int64, minFreePercent float64) *DiskSpaceChecker {
	return &DiskSpaceChecker{
		Path:           path,
		MinFreeBytes:   minFreeBytes,
		MinFreePercent: minFreePercent,
	}
}

// Name returns the name of this health checker
func (c *DiskSpaceChecker) Name() string {
	return "Disk Space"
}

// Check performs the disk space health check
func (c *DiskSpaceChecker) Check(ctx context.Context) *CheckResult {
	result := &CheckResult{
		Name:   c.Name(),
		Status: StatusHealthy,
	}

	// Ensure the path exists
	path := c.Path
	if path == "" {
		path = "."
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		result.Status = StatusWarning
		result.Message = fmt.Sprintf("Failed to get absolute path: %v", err)
		return result
	}

	// Check if path exists, if not create it temporarily for checking
	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		// Try parent directory
		absPath = filepath.Dir(absPath)
	}

	// Get disk usage information
	usage, err := c.getDiskUsage(absPath)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Failed to get disk usage: %v", err)
		result.Suggestions = []string{
			"Check path exists and is accessible",
			"Verify filesystem permissions",
		}
		return result
	}

	// Check free bytes
	if c.MinFreeBytes > 0 && usage.Free < c.MinFreeBytes {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Insufficient free disk space: %s available, %s required",
			formatBytes(usage.Free), formatBytes(c.MinFreeBytes))
		result.Suggestions = []string{
			"Free up disk space",
			"Move to a different location with more space",
			"Clean up temporary files",
		}
	} else if c.MinFreePercent > 0 {
		freePercent := float64(usage.Free) / float64(usage.Total) * 100
		if freePercent < c.MinFreePercent {
			if result.Status == StatusHealthy {
				result.Status = StatusWarning
				result.Message = fmt.Sprintf("Low free disk space: %.1f%% available, %.1f%% required",
					freePercent, c.MinFreePercent)
				result.Suggestions = []string{
					"Monitor disk usage",
					"Consider cleaning up old files",
				}
			}
		}
	}

	if result.Status == StatusHealthy {
		freePercent := float64(usage.Free) / float64(usage.Total) * 100
		result.Message = fmt.Sprintf("Sufficient disk space available: %s free (%.1f%%)",
			formatBytes(usage.Free), freePercent)
	}

	result.Details = map[string]interface{}{
		"path":         absPath,
		"total_bytes":  usage.Total,
		"free_bytes":   usage.Free,
		"used_bytes":   usage.Total - usage.Free,
		"free_percent": float64(usage.Free) / float64(usage.Total) * 100,
	}

	return result
}

// DiskUsage represents disk usage information
type DiskUsage struct {
	Total int64 `json:"total"`
	Free  int64 `json:"free"`
}

// NetworkConnectivityChecker checks network connectivity to required services
type NetworkConnectivityChecker struct {
	Endpoints []string // List of endpoints to check
	Timeout   time.Duration
}

// NewNetworkConnectivityChecker creates a new network connectivity health checker
func NewNetworkConnectivityChecker(endpoints []string) *NetworkConnectivityChecker {
	return &NetworkConnectivityChecker{
		Endpoints: endpoints,
		Timeout:   5 * time.Second,
	}
}

// Name returns the name of this health checker
func (c *NetworkConnectivityChecker) Name() string {
	return "Network Connectivity"
}

// Check performs the network connectivity health check
func (c *NetworkConnectivityChecker) Check(ctx context.Context) *CheckResult {
	result := &CheckResult{
		Name:   c.Name(),
		Status: StatusHealthy,
	}

	if len(c.Endpoints) == 0 {
		result.Status = StatusSkipped
		result.Message = "No endpoints configured for connectivity check"
		return result
	}

	client := &http.Client{
		Timeout: c.Timeout,
	}

	var failedEndpoints []string
	endpointResults := make([]map[string]interface{}, 0, len(c.Endpoints))

	for _, endpoint := range c.Endpoints {
		endpointResult := c.checkEndpoint(ctx, client, endpoint)
		endpointResults = append(endpointResults, endpointResult)

		if !endpointResult["accessible"].(bool) {
			failedEndpoints = append(failedEndpoints, endpoint)
		}
	}

	result.Details = map[string]interface{}{
		"endpoints": endpointResults,
		"timeout":   c.Timeout,
	}

	if len(failedEndpoints) > 0 {
		if len(failedEndpoints) == len(c.Endpoints) {
			result.Status = StatusUnhealthy
			result.Message = "All network endpoints are unreachable"
			result.Suggestions = []string{
				"Check internet connectivity",
				"Verify firewall settings",
				"Check DNS resolution",
			}
		} else {
			result.Status = StatusWarning
			result.Message = fmt.Sprintf("%d of %d endpoints unreachable: %s",
				len(failedEndpoints), len(c.Endpoints), strings.Join(failedEndpoints, ", "))
		}
	} else {
		result.Message = fmt.Sprintf("All %d endpoints are reachable", len(c.Endpoints))
	}

	return result
}

// checkEndpoint checks connectivity to a single endpoint
func (c *NetworkConnectivityChecker) checkEndpoint(ctx context.Context, client *http.Client, endpoint string) map[string]interface{} {
	result := map[string]interface{}{
		"endpoint":   endpoint,
		"accessible": false,
		"error":      "",
		"duration":   time.Duration(0),
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "HEAD", endpoint, nil)
	if err != nil {
		result["error"] = fmt.Sprintf("Failed to create request: %v", err)
		result["duration"] = time.Since(start)
		return result
	}

	resp, err := client.Do(req)
	if err != nil {
		result["error"] = err.Error()
		result["duration"] = time.Since(start)
		return result
	}
	defer func() { _ = resp.Body.Close() }()

	result["accessible"] = true
	result["status_code"] = resp.StatusCode
	result["duration"] = time.Since(start)

	return result
}

// DefaultHealthCheckers returns a set of default health checkers for the go-coverage system
func DefaultHealthCheckers() []HealthChecker {
	checkers := make([]HealthChecker, 0)

	// GitHub API checker
	token := os.Getenv("GITHUB_TOKEN")
	repository := os.Getenv("GITHUB_REPOSITORY")
	if token != "" && repository != "" {
		checkers = append(checkers, NewGitHubAPIChecker(token, repository))
	}

	// Disk space checker (require 100MB and 5% free space)
	tempDir := os.TempDir()
	checkers = append(checkers, NewDiskSpaceChecker(tempDir, 100*1024*1024, 5.0))

	// Network connectivity checker
	endpoints := []string{
		"https://api.github.com",
		"https://github.com",
	}
	checkers = append(checkers, NewNetworkConnectivityChecker(endpoints))

	return checkers
}

// Helper functions

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// String returns a formatted string representation of the health report
func (r *HealthReport) String() string {
	var parts []string

	// Overall status
	parts = append(parts, fmt.Sprintf("Health Status: %s", strings.ToUpper(string(r.OverallStatus))))
	parts = append(parts, fmt.Sprintf("Timestamp: %s", r.Timestamp.Format(time.RFC3339)))

	// Summary
	parts = append(parts, fmt.Sprintf("\nSummary: %d total checks (%d healthy, %d warnings, %d unhealthy, %d skipped)",
		r.Summary.Total, r.Summary.Healthy, r.Summary.Warnings, r.Summary.Unhealthy, r.Summary.Skipped))

	// Individual check results
	if len(r.Checks) > 0 {
		parts = append(parts, "\nCheck Results:")
		for _, check := range r.Checks {
			status := strings.ToUpper(string(check.Status))
			duration := check.Duration.Round(time.Millisecond)
			parts = append(parts, fmt.Sprintf("  %s [%s] (%s): %s", check.Name, status, duration, check.Message))

			// Add suggestions for unhealthy checks
			if check.Status == StatusUnhealthy && len(check.Suggestions) > 0 {
				for _, suggestion := range check.Suggestions {
					parts = append(parts, fmt.Sprintf("    → %s", suggestion))
				}
			}
		}
	}

	return strings.Join(parts, "\n")
}

// ToJSON returns the health report as JSON
func (r *HealthReport) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal health report: %w", err)
	}
	return string(data), nil
}
