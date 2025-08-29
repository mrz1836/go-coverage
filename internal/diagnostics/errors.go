// Package diagnostics provides enhanced error messages with diagnostic output
// for better troubleshooting and user experience.
package diagnostics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Error definitions for err113 linter compliance
var (
	ErrPATHNotSet         = errors.New("PATH environment variable not set")
	ErrExecutableNotFound = errors.New("executable not found in PATH")
)

// DiagnosticError represents an error with enhanced diagnostic information
type DiagnosticError struct {
	// Core error information
	Type        ErrorType `json:"type"`
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	Cause       error     `json:"-"` // Original error (not serialized)
	CauseString string    `json:"cause,omitempty"`

	// Context information
	Component string                 `json:"component"`
	Operation string                 `json:"operation"`
	Context   map[string]interface{} `json:"context,omitempty"`

	// Diagnostic information
	Suggestions   []string    `json:"suggestions,omitempty"`
	Documentation []DocLink   `json:"documentation,omitempty"`
	SystemState   SystemState `json:"system_state"`

	// Metadata
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"trace_id,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
}

// ErrorType categorizes different types of errors
type ErrorType string

const (
	ErrorTypeConfiguration ErrorType = "configuration"
	ErrorTypeNetwork       ErrorType = "network"
	ErrorTypePermission    ErrorType = "permission"
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeSystem        ErrorType = "system"
	ErrorTypeAPI           ErrorType = "api"
	ErrorTypeFile          ErrorType = "file"
	ErrorTypeGit           ErrorType = "git"
	ErrorTypeUser          ErrorType = "user"
	ErrorTypeInternal      ErrorType = "internal"
)

// DocLink represents a link to relevant documentation
type DocLink struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// SystemState captures relevant system state information
type SystemState struct {
	// Environment information
	GitHubActions    bool              `json:"github_actions"`
	Environment      map[string]string `json:"environment,omitempty"`
	WorkingDirectory string            `json:"working_directory"`

	// System resources
	DiskSpace *DiskSpaceInfo `json:"disk_space,omitempty"`
	Memory    *MemoryInfo    `json:"memory,omitempty"`

	// Application state
	ConfigValid   bool                  `json:"config_valid"`
	Dependencies  map[string]Dependency `json:"dependencies,omitempty"`
	LastOperation string                `json:"last_operation,omitempty"`

	// Timing information
	CapturedAt time.Time `json:"captured_at"`
}

// DiskSpaceInfo contains disk space information
type DiskSpaceInfo struct {
	TotalBytes  int64   `json:"total_bytes"`
	FreeBytes   int64   `json:"free_bytes"`
	UsedBytes   int64   `json:"used_bytes"`
	FreePercent float64 `json:"free_percent"`
	Path        string  `json:"path"`
}

// MemoryInfo contains memory usage information
type MemoryInfo struct {
	AllocBytes      uint64 `json:"alloc_bytes"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`
	SysBytes        uint64 `json:"sys_bytes"`
	NumGC           uint32 `json:"num_gc"`
	NumGoroutines   int    `json:"num_goroutines"`
}

// Dependency represents information about a system dependency
type Dependency struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
	Version   string `json:"version,omitempty"`
	Path      string `json:"path,omitempty"`
	Error     string `json:"error,omitempty"`
}

// DiagnosticErrorBuilder helps build comprehensive diagnostic errors
type DiagnosticErrorBuilder struct {
	diagError *DiagnosticError
}

// NewDiagnosticError creates a new diagnostic error builder
func NewDiagnosticError(errorType ErrorType, code, message string) *DiagnosticErrorBuilder {
	return &DiagnosticErrorBuilder{
		diagError: &DiagnosticError{
			Type:          errorType,
			Code:          code,
			Message:       message,
			Context:       make(map[string]interface{}),
			Suggestions:   make([]string, 0),
			Documentation: make([]DocLink, 0),
			SystemState: SystemState{
				Environment:  make(map[string]string),
				Dependencies: make(map[string]Dependency),
				CapturedAt:   time.Now(),
			},
			Timestamp: time.Now(),
		},
	}
}

// WithCause adds the underlying cause error
func (b *DiagnosticErrorBuilder) WithCause(err error) *DiagnosticErrorBuilder {
	if err != nil {
		b.diagError.Cause = err
		b.diagError.CauseString = err.Error()
	}
	return b
}

// WithComponent specifies which component generated the error
func (b *DiagnosticErrorBuilder) WithComponent(component string) *DiagnosticErrorBuilder {
	b.diagError.Component = component
	return b
}

// WithOperation specifies the operation that failed
func (b *DiagnosticErrorBuilder) WithOperation(operation string) *DiagnosticErrorBuilder {
	b.diagError.Operation = operation
	return b
}

// WithContext adds contextual information
func (b *DiagnosticErrorBuilder) WithContext(key string, value interface{}) *DiagnosticErrorBuilder {
	b.diagError.Context[key] = value
	return b
}

// WithSuggestion adds a troubleshooting suggestion
func (b *DiagnosticErrorBuilder) WithSuggestion(suggestion string) *DiagnosticErrorBuilder {
	b.diagError.Suggestions = append(b.diagError.Suggestions, suggestion)
	return b
}

// WithDocumentation adds a documentation link
func (b *DiagnosticErrorBuilder) WithDocumentation(title, url, description string) *DiagnosticErrorBuilder {
	b.diagError.Documentation = append(b.diagError.Documentation, DocLink{
		Title:       title,
		URL:         url,
		Description: description,
	})
	return b
}

// WithTraceID adds a trace ID for request tracking
func (b *DiagnosticErrorBuilder) WithTraceID(traceID string) *DiagnosticErrorBuilder {
	b.diagError.TraceID = traceID
	return b
}

// WithSessionID adds a session ID for debugging
func (b *DiagnosticErrorBuilder) WithSessionID(sessionID string) *DiagnosticErrorBuilder {
	b.diagError.SessionID = sessionID
	return b
}

// WithSystemState captures current system state
func (b *DiagnosticErrorBuilder) WithSystemState() *DiagnosticErrorBuilder {
	b.diagError.SystemState = captureSystemState()
	return b
}

// Build creates the final diagnostic error
func (b *DiagnosticErrorBuilder) Build() *DiagnosticError {
	// Automatically capture system state if not already done
	if b.diagError.SystemState.CapturedAt.IsZero() {
		b.diagError.SystemState = captureSystemState()
	}

	// Add automatic suggestions based on error type
	b.addAutomaticSuggestions()

	// Add relevant documentation links
	b.addRelevantDocumentation()

	return b.diagError
}

// addAutomaticSuggestions adds suggestions based on error type and context
func (b *DiagnosticErrorBuilder) addAutomaticSuggestions() {
	switch b.diagError.Type {
	case ErrorTypeNetwork:
		b.diagError.Suggestions = append(b.diagError.Suggestions,
			"Check your internet connection",
			"Verify firewall settings allow outbound connections",
			"Try again in a few minutes in case of temporary service issues",
		)
	case ErrorTypePermission:
		b.diagError.Suggestions = append(b.diagError.Suggestions,
			"Check file and directory permissions",
			"Verify GitHub token has required permissions",
			"Ensure you have write access to the repository",
		)
	case ErrorTypeConfiguration:
		b.diagError.Suggestions = append(b.diagError.Suggestions,
			"Verify configuration file syntax",
			"Check environment variables are set correctly",
			"Review configuration documentation",
		)
	case ErrorTypeFile:
		b.diagError.Suggestions = append(b.diagError.Suggestions,
			"Check if the file exists",
			"Verify file is not locked by another process",
			"Ensure sufficient disk space is available",
		)
	case ErrorTypeGit:
		b.diagError.Suggestions = append(b.diagError.Suggestions,
			"Ensure git repository is initialized",
			"Check git credentials and authentication",
			"Verify remote repository is accessible",
		)
	case ErrorTypeValidation:
		b.diagError.Suggestions = append(b.diagError.Suggestions,
			"Check input data format and structure",
			"Verify all required fields are present",
			"Ensure data meets validation requirements",
		)
	case ErrorTypeSystem:
		b.diagError.Suggestions = append(b.diagError.Suggestions,
			"Check system resources (CPU, memory, disk)",
			"Verify system dependencies are installed",
			"Check system logs for additional information",
		)
	case ErrorTypeAPI:
		b.diagError.Suggestions = append(b.diagError.Suggestions,
			"Check API endpoint availability",
			"Verify authentication credentials",
			"Review API rate limiting and quotas",
		)
	case ErrorTypeUser:
		b.diagError.Suggestions = append(b.diagError.Suggestions,
			"Review the command usage and parameters",
			"Check the documentation for correct syntax",
			"Verify input values are valid",
		)
	case ErrorTypeInternal:
		b.diagError.Suggestions = append(b.diagError.Suggestions,
			"This is an internal error - please report it",
			"Try restarting the operation",
			"Check if the issue persists across restarts",
		)
	}

	// Add suggestions based on system state
	if b.diagError.SystemState.DiskSpace != nil && b.diagError.SystemState.DiskSpace.FreePercent < 5 {
		b.diagError.Suggestions = append(b.diagError.Suggestions,
			"Free up disk space - less than 5% available",
		)
	}

	if !b.diagError.SystemState.GitHubActions && strings.Contains(b.diagError.Message, "GITHUB_") {
		b.diagError.Suggestions = append(b.diagError.Suggestions,
			"This operation requires GitHub Actions environment",
			"Set required GITHUB_* environment variables manually for local testing",
		)
	}
}

// addRelevantDocumentation adds documentation links based on error context
func (b *DiagnosticErrorBuilder) addRelevantDocumentation() {
	baseURL := "https://github.com/mrz1836/go-coverage"

	switch b.diagError.Type {
	case ErrorTypeConfiguration:
		b.diagError.Documentation = append(b.diagError.Documentation, DocLink{
			Title:       "Configuration Guide",
			URL:         baseURL + "/docs/configuration.md",
			Description: "Complete configuration reference and examples",
		})
	case ErrorTypeAPI:
		if strings.Contains(b.diagError.Message, "github") {
			b.diagError.Documentation = append(b.diagError.Documentation, DocLink{
				Title:       "GitHub API Documentation",
				URL:         "https://docs.github.com/en/rest",
				Description: "Official GitHub API documentation",
			})
		}
	case ErrorTypeGit:
		b.diagError.Documentation = append(b.diagError.Documentation, DocLink{
			Title:       "Git Troubleshooting",
			URL:         baseURL + "/docs/git-troubleshooting.md",
			Description: "Common git issues and solutions",
		})
	case ErrorTypeNetwork:
		b.diagError.Documentation = append(b.diagError.Documentation, DocLink{
			Title:       "Network Troubleshooting",
			URL:         baseURL + "/docs/network-issues.md",
			Description: "Network connectivity and firewall issues",
		})
	case ErrorTypePermission:
		b.diagError.Documentation = append(b.diagError.Documentation, DocLink{
			Title:       "Permission Issues",
			URL:         baseURL + "/docs/permissions.md",
			Description: "File, directory, and API permission troubleshooting",
		})
	case ErrorTypeValidation:
		b.diagError.Documentation = append(b.diagError.Documentation, DocLink{
			Title:       "Data Validation",
			URL:         baseURL + "/docs/validation.md",
			Description: "Input validation and data format requirements",
		})
	case ErrorTypeSystem:
		b.diagError.Documentation = append(b.diagError.Documentation, DocLink{
			Title:       "System Requirements",
			URL:         baseURL + "/docs/system-requirements.md",
			Description: "System dependencies and environment setup",
		})
	case ErrorTypeFile:
		b.diagError.Documentation = append(b.diagError.Documentation, DocLink{
			Title:       "File System Issues",
			URL:         baseURL + "/docs/file-system.md",
			Description: "File access and storage troubleshooting",
		})
	case ErrorTypeUser:
		b.diagError.Documentation = append(b.diagError.Documentation, DocLink{
			Title:       "Usage Guide",
			URL:         baseURL + "/docs/usage.md",
			Description: "Command usage and parameter reference",
		})
	case ErrorTypeInternal:
		b.diagError.Documentation = append(b.diagError.Documentation, DocLink{
			Title:       "Bug Reports",
			URL:         baseURL + "/issues",
			Description: "Report internal errors and bugs",
		})
	}

	// Always add general troubleshooting guide
	b.diagError.Documentation = append(b.diagError.Documentation, DocLink{
		Title:       "Troubleshooting Guide",
		URL:         baseURL + "/docs/troubleshooting.md",
		Description: "General troubleshooting steps and FAQ",
	})
}

// Error implements the error interface
func (de *DiagnosticError) Error() string {
	if de.Component != "" && de.Operation != "" {
		return fmt.Sprintf("[%s:%s] %s: %s", de.Component, de.Operation, de.Code, de.Message)
	} else if de.Component != "" {
		return fmt.Sprintf("[%s] %s: %s", de.Component, de.Code, de.Message)
	}
	return fmt.Sprintf("%s: %s", de.Code, de.Message)
}

// DetailedError returns a detailed error message with diagnostics
func (de *DiagnosticError) DetailedError() string {
	var parts []string

	// Main error
	parts = append(parts, fmt.Sprintf("Error: %s", de.Error()))

	// Cause
	if de.CauseString != "" {
		parts = append(parts, fmt.Sprintf("Cause: %s", de.CauseString))
	}

	// Context
	if len(de.Context) > 0 {
		parts = append(parts, "\nContext:")
		for k, v := range de.Context {
			parts = append(parts, fmt.Sprintf("  %s: %v", k, v))
		}
	}

	// System state summary
	if !de.SystemState.CapturedAt.IsZero() {
		parts = append(parts, "\nSystem Information:")
		parts = append(parts, fmt.Sprintf("  GitHub Actions: %t", de.SystemState.GitHubActions))
		parts = append(parts, fmt.Sprintf("  Working Directory: %s", de.SystemState.WorkingDirectory))

		if de.SystemState.DiskSpace != nil {
			parts = append(parts, fmt.Sprintf("  Disk Space: %.1f%% free", de.SystemState.DiskSpace.FreePercent))
		}

		if len(de.SystemState.Dependencies) > 0 {
			parts = append(parts, "  Dependencies:")
			for name, dep := range de.SystemState.Dependencies {
				status := "✓"
				if !dep.Available {
					status = "✗"
				}
				parts = append(parts, fmt.Sprintf("    %s %s", status, name))
			}
		}
	}

	// Suggestions
	if len(de.Suggestions) > 0 {
		parts = append(parts, "\nTroubleshooting Suggestions:")
		for i, suggestion := range de.Suggestions {
			parts = append(parts, fmt.Sprintf("  %d. %s", i+1, suggestion))
		}
	}

	// Documentation
	if len(de.Documentation) > 0 {
		parts = append(parts, "\nHelpful Documentation:")
		for _, doc := range de.Documentation {
			parts = append(parts, fmt.Sprintf("  - %s: %s", doc.Title, doc.URL))
		}
	}

	// Metadata
	if de.TraceID != "" || de.SessionID != "" {
		parts = append(parts, "\nDebugging Information:")
		if de.TraceID != "" {
			parts = append(parts, fmt.Sprintf("  Trace ID: %s", de.TraceID))
		}
		if de.SessionID != "" {
			parts = append(parts, fmt.Sprintf("  Session ID: %s", de.SessionID))
		}
	}

	return strings.Join(parts, "\n")
}

// ToJSON returns the diagnostic error as JSON
func (de *DiagnosticError) ToJSON() (string, error) {
	data, err := json.MarshalIndent(de, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal diagnostic error: %w", err)
	}
	return string(data), nil
}

// IsType checks if the error is of a specific type
func (de *DiagnosticError) IsType(errorType ErrorType) bool {
	return de.Type == errorType
}

// HasCode checks if the error has a specific error code
func (de *DiagnosticError) HasCode(code string) bool {
	return de.Code == code
}

// captureSystemState captures current system state information
func captureSystemState() SystemState {
	state := SystemState{
		GitHubActions: os.Getenv("GITHUB_ACTIONS") == "true",
		Environment:   make(map[string]string),
		Dependencies:  make(map[string]Dependency),
		CapturedAt:    time.Now(),
	}

	// Capture working directory
	if wd, err := os.Getwd(); err == nil {
		state.WorkingDirectory = wd
	}

	// Capture relevant environment variables
	relevantEnvVars := []string{
		"GITHUB_ACTIONS", "GITHUB_REPOSITORY", "GITHUB_SHA", "GITHUB_TOKEN",
		"GITHUB_REF", "GITHUB_HEAD_REF", "GITHUB_BASE_REF", "GITHUB_EVENT_NAME",
		"RUNNER_OS", "RUNNER_ARCH", "RUNNER_TEMP", "RUNNER_WORKSPACE",
		"PATH", "HOME", "TMPDIR", "PWD",
	}

	for _, envVar := range relevantEnvVars {
		if value := os.Getenv(envVar); value != "" {
			// Mask sensitive values
			if strings.Contains(strings.ToLower(envVar), "token") ||
				strings.Contains(strings.ToLower(envVar), "password") {
				state.Environment[envVar] = "[REDACTED]"
			} else if len(value) > 100 {
				state.Environment[envVar] = value[:97] + "..."
			} else {
				state.Environment[envVar] = value
			}
		}
	}

	// Capture disk space information
	if wd, err := os.Getwd(); err == nil {
		if diskInfo := getDiskUsage(wd); diskInfo != nil {
			state.DiskSpace = diskInfo
		}
	}

	// Capture memory information
	state.Memory = getMemoryInfo()

	// Check dependencies
	state.Dependencies["git"] = checkGitAvailability()
	state.Dependencies["gh"] = checkGHCLIAvailability()

	return state
}

// getDiskUsage returns disk usage information for a path
func getDiskUsage(path string) *DiskSpaceInfo {
	// This would be implemented using platform-specific code
	// For now, return nil to indicate unavailable
	return nil
}

// getMemoryInfo returns current memory usage information
func getMemoryInfo() *MemoryInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &MemoryInfo{
		AllocBytes:      memStats.Alloc,
		TotalAllocBytes: memStats.TotalAlloc,
		SysBytes:        memStats.Sys,
		NumGC:           memStats.NumGC,
		NumGoroutines:   runtime.NumGoroutine(),
	}
}

// checkGitAvailability checks if git is available
func checkGitAvailability() Dependency {
	dep := Dependency{Name: "git"}

	// Try to run git --version
	cmd := "git"
	if path, err := findExecutable(cmd); err == nil {
		dep.Available = true
		dep.Path = path
		// Could capture version here
	} else {
		dep.Available = false
		dep.Error = err.Error()
	}

	return dep
}

// checkGHCLIAvailability checks if GitHub CLI is available
func checkGHCLIAvailability() Dependency {
	dep := Dependency{Name: "gh"}

	// Try to find gh executable
	cmd := "gh"
	if path, err := findExecutable(cmd); err == nil {
		dep.Available = true
		dep.Path = path
	} else {
		dep.Available = false
		dep.Error = err.Error()
	}

	return dep
}

// findExecutable looks for an executable in PATH
func findExecutable(name string) (string, error) {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return "", ErrPATHNotSet
	}

	paths := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, path := range paths {
		fullPath := filepath.Join(path, name)

		// On Windows, try with .exe extension
		if runtime.GOOS == "windows" {
			fullPath += ".exe"
		}

		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			// Check if executable (simplified)
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("%w: %s", ErrExecutableNotFound, name)
}

// Convenience functions for creating common diagnostic errors

// WrapWithDiagnostics wraps a regular error with diagnostic information
func WrapWithDiagnostics(err error, errorType ErrorType, component, operation string) *DiagnosticError {
	if err == nil {
		return nil
	}

	// Check if it's already a diagnostic error
	var diagErr *DiagnosticError
	if errors.As(err, &diagErr) {
		return diagErr
	}

	code := "UNKNOWN_ERROR"
	if errorType != "" {
		code = strings.ToUpper(string(errorType)) + "_ERROR"
	}

	return NewDiagnosticError(errorType, code, err.Error()).
		WithCause(err).
		WithComponent(component).
		WithOperation(operation).
		WithSystemState().
		Build()
}

// CreateNetworkError creates a network-related diagnostic error
func CreateNetworkError(message string, cause error) *DiagnosticError {
	return NewDiagnosticError(ErrorTypeNetwork, "NETWORK_ERROR", message).
		WithCause(cause).
		WithSuggestion("Check internet connectivity").
		WithSuggestion("Verify firewall settings").
		WithSuggestion("Try again later if service is temporarily unavailable").
		WithSystemState().
		Build()
}

// CreateConfigurationError creates a configuration-related diagnostic error
func CreateConfigurationError(message string, cause error) *DiagnosticError {
	return NewDiagnosticError(ErrorTypeConfiguration, "CONFIG_ERROR", message).
		WithCause(cause).
		WithSuggestion("Check configuration file syntax").
		WithSuggestion("Verify all required settings are provided").
		WithSuggestion("Review environment variables").
		WithSystemState().
		Build()
}

// CreatePermissionError creates a permission-related diagnostic error
func CreatePermissionError(message string, cause error) *DiagnosticError {
	return NewDiagnosticError(ErrorTypePermission, "PERMISSION_ERROR", message).
		WithCause(cause).
		WithSuggestion("Check file and directory permissions").
		WithSuggestion("Verify GitHub token permissions").
		WithSuggestion("Ensure repository access rights").
		WithSystemState().
		Build()
}

// CreateGitHubAPIError creates a GitHub API-related diagnostic error
func CreateGitHubAPIError(message string, statusCode int, cause error) *DiagnosticError {
	builder := NewDiagnosticError(ErrorTypeAPI, "GITHUB_API_ERROR", message).
		WithCause(cause).
		WithContext("status_code", statusCode).
		WithComponent("github").
		WithSystemState()

	// Add specific suggestions based on status code
	switch statusCode {
	case 401:
		builder.WithSuggestion("Check GitHub token is valid and not expired").
			WithSuggestion("Ensure token has required permissions")
	case 403:
		builder.WithSuggestion("Token may have insufficient permissions").
			WithSuggestion("Check API rate limits")
	case 404:
		builder.WithSuggestion("Verify repository exists and is accessible").
			WithSuggestion("Check repository name format (owner/repo)")
	case 422:
		builder.WithSuggestion("Check request payload format").
			WithSuggestion("Verify all required fields are provided")
	default:
		if statusCode >= 500 {
			builder.WithSuggestion("GitHub API may be experiencing issues").
				WithSuggestion("Try again in a few minutes")
		}
	}

	return builder.Build()
}

// ErrorReporter helps collect and report diagnostic errors
type ErrorReporter struct {
	errors []DiagnosticError
}

// NewErrorReporter creates a new error reporter
func NewErrorReporter(ctx context.Context) *ErrorReporter {
	return &ErrorReporter{
		errors: make([]DiagnosticError, 0),
	}
}

// Report adds a diagnostic error to the reporter
func (er *ErrorReporter) Report(err *DiagnosticError) {
	if err != nil {
		er.errors = append(er.errors, *err)
	}
}

// GetErrors returns all collected errors
func (er *ErrorReporter) GetErrors() []DiagnosticError {
	return er.errors
}

// HasErrors returns true if any errors have been collected
func (er *ErrorReporter) HasErrors() bool {
	return len(er.errors) > 0
}

// GenerateReport generates a comprehensive error report
func (er *ErrorReporter) GenerateReport() string {
	if len(er.errors) == 0 {
		return "No errors to report"
	}

	parts := make([]string, 0, 3+len(er.errors)*2)
	parts = append(parts, fmt.Sprintf("Diagnostic Error Report - %s", time.Now().Format(time.RFC3339)))
	parts = append(parts, fmt.Sprintf("Total Errors: %d", len(er.errors)))
	parts = append(parts, strings.Repeat("=", 50))

	// Sort errors by timestamp
	sort.Slice(er.errors, func(i, j int) bool {
		return er.errors[i].Timestamp.Before(er.errors[j].Timestamp)
	})

	for i, err := range er.errors {
		parts = append(parts, fmt.Sprintf("\nError %d/%d:", i+1, len(er.errors)))
		parts = append(parts, err.DetailedError())
		if i < len(er.errors)-1 {
			parts = append(parts, strings.Repeat("-", 40))
		}
	}

	return strings.Join(parts, "\n")
}
