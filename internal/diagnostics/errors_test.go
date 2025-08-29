package diagnostics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewDiagnosticError(t *testing.T) {
	builder := NewDiagnosticError(ErrorTypeNetwork, "NET001", "network connection failed")

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}

	if builder.diagError.Type != ErrorTypeNetwork {
		t.Errorf("Expected error type %s, got %s", ErrorTypeNetwork, builder.diagError.Type)
	}

	if builder.diagError.Code != "NET001" {
		t.Errorf("Expected error code NET001, got %s", builder.diagError.Code)
	}

	if builder.diagError.Message != "network connection failed" {
		t.Errorf("Expected message 'network connection failed', got '%s'", builder.diagError.Message)
	}
}

func TestDiagnosticErrorBuilder_WithCause(t *testing.T) {
	originalErr := errors.New("original error") //nolint:err113 // test error

	diagErr := NewDiagnosticError(ErrorTypeSystem, "SYS001", "system failure").
		WithCause(originalErr).
		Build()

	if !errors.Is(diagErr.Cause, originalErr) {
		t.Errorf("Expected cause to be preserved")
	}

	if diagErr.CauseString != originalErr.Error() {
		t.Errorf("Expected cause string '%s', got '%s'", originalErr.Error(), diagErr.CauseString)
	}
}

func TestDiagnosticErrorBuilder_WithComponent(t *testing.T) {
	diagErr := NewDiagnosticError(ErrorTypeAPI, "API001", "API error").
		WithComponent("github").
		Build()

	if diagErr.Component != "github" {
		t.Errorf("Expected component 'github', got '%s'", diagErr.Component)
	}
}

func TestDiagnosticErrorBuilder_WithOperation(t *testing.T) {
	diagErr := NewDiagnosticError(ErrorTypeAPI, "API001", "API error").
		WithOperation("list_repos").
		Build()

	if diagErr.Operation != "list_repos" {
		t.Errorf("Expected operation 'list_repos', got '%s'", diagErr.Operation)
	}
}

func TestDiagnosticErrorBuilder_WithContext(t *testing.T) {
	diagErr := NewDiagnosticError(ErrorTypeNetwork, "NET001", "network error").
		WithContext("url", "https://api.github.com").
		WithContext("timeout", 30).
		Build()

	if len(diagErr.Context) != 2 {
		t.Errorf("Expected 2 context items, got %d", len(diagErr.Context))
	}

	if diagErr.Context["url"] != "https://api.github.com" {
		t.Errorf("Expected context url to be 'https://api.github.com', got '%v'", diagErr.Context["url"])
	}

	if diagErr.Context["timeout"] != 30 {
		t.Errorf("Expected context timeout to be 30, got %v", diagErr.Context["timeout"])
	}
}

func TestDiagnosticErrorBuilder_WithSuggestion(t *testing.T) {
	diagErr := NewDiagnosticError(ErrorTypeNetwork, "NET001", "network error").
		WithSuggestion("Check your internet connection").
		WithSuggestion("Try again later").
		Build()

	// Build() adds 3 automatic suggestions for network errors plus 2 manual = 5 total
	if len(diagErr.Suggestions) != 5 {
		t.Errorf("Expected 5 suggestions, got %d", len(diagErr.Suggestions))
	}

	// Check the first two are our manual suggestions
	expectedManualSuggestions := []string{
		"Check your internet connection",
		"Try again later",
	}

	for i, expected := range expectedManualSuggestions {
		if diagErr.Suggestions[i] != expected {
			t.Errorf("Expected suggestion '%s', got '%s'", expected, diagErr.Suggestions[i])
		}
	}
}

func TestDiagnosticErrorBuilder_WithDocumentation(t *testing.T) {
	diagErr := NewDiagnosticError(ErrorTypeConfiguration, "CFG001", "config error").
		WithDocumentation("Configuration Guide", "https://example.com/config", "How to configure").
		Build()

	// Build() adds 1 configuration-specific + 1 general troubleshooting doc + 1 manual = 3 total
	if len(diagErr.Documentation) != 3 {
		t.Errorf("Expected 3 documentation links, got %d", len(diagErr.Documentation))
	}

	// Check the first one is our manual documentation
	doc := diagErr.Documentation[0]
	if doc.Title != "Configuration Guide" {
		t.Errorf("Expected doc title 'Configuration Guide', got '%s'", doc.Title)
	}

	if doc.URL != "https://example.com/config" {
		t.Errorf("Expected doc URL 'https://example.com/config', got '%s'", doc.URL)
	}

	if doc.Description != "How to configure" {
		t.Errorf("Expected doc description 'How to configure', got '%s'", doc.Description)
	}
}

func TestDiagnosticErrorBuilder_WithTraceID(t *testing.T) {
	diagErr := NewDiagnosticError(ErrorTypeAPI, "API001", "API error").
		WithTraceID("trace-123").
		Build()

	if diagErr.TraceID != "trace-123" {
		t.Errorf("Expected trace ID 'trace-123', got '%s'", diagErr.TraceID)
	}
}

func TestDiagnosticErrorBuilder_WithSessionID(t *testing.T) {
	diagErr := NewDiagnosticError(ErrorTypeAPI, "API001", "API error").
		WithSessionID("session-456").
		Build()

	if diagErr.SessionID != "session-456" {
		t.Errorf("Expected session ID 'session-456', got '%s'", diagErr.SessionID)
	}
}

func TestDiagnosticErrorBuilder_AutomaticSuggestions(t *testing.T) {
	testCases := []struct {
		errorType        ErrorType
		expectedKeywords []string
	}{
		{
			ErrorTypeNetwork,
			[]string{"internet connection", "firewall", "temporary service"},
		},
		{
			ErrorTypePermission,
			[]string{"permissions", "token", "access"},
		},
		{
			ErrorTypeConfiguration,
			[]string{"configuration", "environment", "documentation"},
		},
		{
			ErrorTypeFile,
			[]string{"file exists", "locked", "disk space"},
		},
		{
			ErrorTypeGit,
			[]string{"git repository", "credentials", "remote repository"},
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.errorType), func(t *testing.T) {
			diagErr := NewDiagnosticError(tc.errorType, "TEST001", "test error").
				Build()

			if len(diagErr.Suggestions) == 0 {
				t.Errorf("Expected automatic suggestions for error type %s", tc.errorType)
			}

			// Check that expected keywords appear in suggestions
			allSuggestions := strings.Join(diagErr.Suggestions, " ")
			for _, keyword := range tc.expectedKeywords {
				if !strings.Contains(strings.ToLower(allSuggestions), strings.ToLower(keyword)) {
					t.Errorf("Expected suggestion to contain keyword '%s' for error type %s", keyword, tc.errorType)
				}
			}
		})
	}
}

func TestDiagnosticErrorBuilder_RelevantDocumentation(t *testing.T) {
	testCases := []struct {
		errorType    ErrorType
		expectedDocs []string
	}{
		{
			ErrorTypeConfiguration,
			[]string{"Configuration Guide"},
		},
		{
			ErrorTypeAPI,
			[]string{"GitHub API Documentation"},
		},
		{
			ErrorTypeGit,
			[]string{"Git Troubleshooting"},
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.errorType), func(t *testing.T) {
			diagErr := NewDiagnosticError(tc.errorType, "TEST001", "github api test error").
				Build()

			// All errors should have the general troubleshooting guide
			hasGeneralGuide := false
			for _, doc := range diagErr.Documentation {
				if doc.Title == "Troubleshooting Guide" {
					hasGeneralGuide = true
					break
				}
			}

			if !hasGeneralGuide {
				t.Errorf("Expected general troubleshooting guide for error type %s", tc.errorType)
			}

			// Check for specific documentation
			for _, expectedDoc := range tc.expectedDocs {
				hasExpectedDoc := false
				for _, doc := range diagErr.Documentation {
					if strings.Contains(doc.Title, expectedDoc) {
						hasExpectedDoc = true
						break
					}
				}

				if !hasExpectedDoc {
					t.Errorf("Expected documentation '%s' for error type %s", expectedDoc, tc.errorType)
				}
			}
		})
	}
}

func TestDiagnosticError_Error(t *testing.T) {
	testCases := []struct {
		name      string
		component string
		operation string
		code      string
		message   string
		expected  string
	}{
		{
			name:      "with component and operation",
			component: "github",
			operation: "list_repos",
			code:      "API001",
			message:   "API error",
			expected:  "[github:list_repos] API001: API error",
		},
		{
			name:      "with component only",
			component: "github",
			code:      "API001",
			message:   "API error",
			expected:  "[github] API001: API error",
		},
		{
			name:     "without component",
			code:     "API001",
			message:  "API error",
			expected: "API001: API error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewDiagnosticError(ErrorTypeAPI, tc.code, tc.message)

			if tc.component != "" {
				builder.WithComponent(tc.component)
			}

			if tc.operation != "" {
				builder.WithOperation(tc.operation)
			}

			diagErr := builder.Build()
			result := diagErr.Error()

			if result != tc.expected {
				t.Errorf("Expected error string '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestDiagnosticError_DetailedError(t *testing.T) {
	originalErr := errors.New("original cause") //nolint:err113 // test error

	diagErr := NewDiagnosticError(ErrorTypeNetwork, "NET001", "network failure").
		WithCause(originalErr).
		WithComponent("github").
		WithOperation("api_call").
		WithContext("url", "https://api.github.com").
		WithContext("timeout", 30).
		WithSuggestion("Check internet connection").
		WithSuggestion("Try again later").
		WithTraceID("trace-123").
		WithSessionID("session-456").
		Build()

	detailed := diagErr.DetailedError()

	// Check that all components are present
	expectedParts := []string{
		"Error:",
		"Cause:",
		"Context:",
		"url: https://api.github.com",
		"timeout: 30",
		"System Information:",
		"Troubleshooting Suggestions:",
		"Check internet connection",
		"Try again later",
		"Debugging Information:",
		"Trace ID: trace-123",
		"Session ID: session-456",
	}

	for _, part := range expectedParts {
		if !strings.Contains(detailed, part) {
			t.Errorf("Expected detailed error to contain '%s'", part)
		}
	}
}

func TestDiagnosticError_ToJSON(t *testing.T) {
	diagErr := NewDiagnosticError(ErrorTypeAPI, "API001", "API error").
		WithComponent("github").
		WithContext("status_code", 404).
		WithSuggestion("Check repository exists").
		Build()

	jsonStr, err := diagErr.ToJSON()
	if err != nil {
		t.Errorf("Expected no error from ToJSON, got: %v", err)
	}

	// Parse JSON to verify structure
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	if err != nil {
		t.Errorf("Expected valid JSON, got error: %v", err)
	}

	// Check key fields
	if parsed["type"] != string(ErrorTypeAPI) {
		t.Errorf("Expected type '%s' in JSON", ErrorTypeAPI)
	}

	if parsed["code"] != "API001" {
		t.Errorf("Expected code 'API001' in JSON")
	}

	if parsed["component"] != "github" {
		t.Errorf("Expected component 'github' in JSON")
	}
}

func TestDiagnosticError_IsType(t *testing.T) {
	diagErr := NewDiagnosticError(ErrorTypeNetwork, "NET001", "network error").Build()

	if !diagErr.IsType(ErrorTypeNetwork) {
		t.Error("Expected IsType to return true for matching type")
	}

	if diagErr.IsType(ErrorTypeAPI) {
		t.Error("Expected IsType to return false for non-matching type")
	}
}

func TestDiagnosticError_HasCode(t *testing.T) {
	diagErr := NewDiagnosticError(ErrorTypeNetwork, "NET001", "network error").Build()

	if !diagErr.HasCode("NET001") {
		t.Error("Expected HasCode to return true for matching code")
	}

	if diagErr.HasCode("API001") {
		t.Error("Expected HasCode to return false for non-matching code")
	}
}

func TestWrapWithDiagnostics(t *testing.T) {
	originalErr := errors.New("original error") //nolint:err113 // test error

	diagErr := WrapWithDiagnostics(originalErr, ErrorTypeNetwork, "github", "api_call")

	if diagErr == nil {
		t.Fatal("Expected non-nil diagnostic error")
	}

	if diagErr.Type != ErrorTypeNetwork {
		t.Errorf("Expected error type %s, got %s", ErrorTypeNetwork, diagErr.Type)
	}

	if diagErr.Component != "github" {
		t.Errorf("Expected component 'github', got '%s'", diagErr.Component)
	}

	if diagErr.Operation != "api_call" {
		t.Errorf("Expected operation 'api_call', got '%s'", diagErr.Operation)
	}

	if !errors.Is(diagErr.Cause, originalErr) {
		t.Error("Expected original error to be preserved as cause")
	}
}

func TestWrapWithDiagnostics_NilError(t *testing.T) {
	diagErr := WrapWithDiagnostics(nil, ErrorTypeNetwork, "github", "api_call")

	if diagErr != nil {
		t.Error("Expected nil diagnostic error for nil input")
	}
}

func TestWrapWithDiagnostics_AlreadyDiagnostic(t *testing.T) {
	originalDiagErr := NewDiagnosticError(ErrorTypeAPI, "API001", "API error").Build()

	result := WrapWithDiagnostics(originalDiagErr, ErrorTypeNetwork, "github", "api_call")

	// Should return the same diagnostic error
	if result != originalDiagErr {
		t.Error("Expected same diagnostic error to be returned when already diagnostic")
	}
}

func TestCreateNetworkError(t *testing.T) {
	cause := errors.New("connection timeout") //nolint:err113 // test error

	diagErr := CreateNetworkError("network connection failed", cause)

	if diagErr.Type != ErrorTypeNetwork {
		t.Errorf("Expected error type %s, got %s", ErrorTypeNetwork, diagErr.Type)
	}

	if diagErr.Code != "NETWORK_ERROR" {
		t.Errorf("Expected error code 'NETWORK_ERROR', got '%s'", diagErr.Code)
	}

	if !errors.Is(diagErr.Cause, cause) {
		t.Error("Expected cause to be preserved")
	}

	// Check for network-specific suggestions
	hasNetworkSuggestion := false
	for _, suggestion := range diagErr.Suggestions {
		if strings.Contains(strings.ToLower(suggestion), "connectivity") {
			hasNetworkSuggestion = true
			break
		}
	}

	if !hasNetworkSuggestion {
		t.Error("Expected network-specific suggestions")
	}
}

func TestCreateConfigurationError(t *testing.T) {
	cause := errors.New("invalid syntax") //nolint:err113 // test error

	diagErr := CreateConfigurationError("configuration file is invalid", cause)

	if diagErr.Type != ErrorTypeConfiguration {
		t.Errorf("Expected error type %s, got %s", ErrorTypeConfiguration, diagErr.Type)
	}

	if diagErr.Code != "CONFIG_ERROR" {
		t.Errorf("Expected error code 'CONFIG_ERROR', got '%s'", diagErr.Code)
	}
}

func TestCreatePermissionError(t *testing.T) {
	cause := errors.New("access denied") //nolint:err113 // test error

	diagErr := CreatePermissionError("insufficient permissions", cause)

	if diagErr.Type != ErrorTypePermission {
		t.Errorf("Expected error type %s, got %s", ErrorTypePermission, diagErr.Type)
	}

	if diagErr.Code != "PERMISSION_ERROR" {
		t.Errorf("Expected error code 'PERMISSION_ERROR', got '%s'", diagErr.Code)
	}
}

func TestCreateGitHubAPIError(t *testing.T) {
	testCases := []struct {
		statusCode          int
		expectedSuggestions []string
	}{
		{401, []string{"token", "expired"}},
		{403, []string{"permissions", "rate limits"}},
		{404, []string{"repository exists", "accessible"}},
		{422, []string{"payload", "required fields"}},
		{500, []string{"GitHub API", "issues"}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("status_%d", tc.statusCode), func(t *testing.T) {
			cause := errors.New("API request failed") //nolint:err113 // test error

			diagErr := CreateGitHubAPIError("GitHub API error", tc.statusCode, cause)

			if diagErr.Type != ErrorTypeAPI {
				t.Errorf("Expected error type %s, got %s", ErrorTypeAPI, diagErr.Type)
			}

			if diagErr.Context["status_code"] != tc.statusCode {
				t.Errorf("Expected status code %d in context", tc.statusCode)
			}

			if diagErr.Component != "github" {
				t.Errorf("Expected component 'github', got '%s'", diagErr.Component)
			}

			// Check for status-specific suggestions
			allSuggestions := strings.Join(diagErr.Suggestions, " ")
			for _, keyword := range tc.expectedSuggestions {
				if !strings.Contains(strings.ToLower(allSuggestions), strings.ToLower(keyword)) {
					t.Errorf("Expected suggestion to contain '%s' for status %d", keyword, tc.statusCode)
				}
			}
		})
	}
}

func TestErrorReporter(t *testing.T) {
	ctx := context.Background()
	reporter := NewErrorReporter(ctx)

	if reporter == nil {
		t.Fatal("Expected non-nil error reporter")
	}

	if reporter.HasErrors() {
		t.Error("Expected no errors initially")
	}

	// Add some errors
	err1 := NewDiagnosticError(ErrorTypeNetwork, "NET001", "network error").Build()
	err2 := NewDiagnosticError(ErrorTypeAPI, "API001", "API error").Build()

	reporter.Report(err1)
	reporter.Report(err2)

	if !reporter.HasErrors() {
		t.Error("Expected errors after reporting")
	}

	errors := reporter.GetErrors()
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}
}

func TestErrorReporter_GenerateReport(t *testing.T) {
	ctx := context.Background()
	reporter := NewErrorReporter(ctx)

	// Add errors with different timestamps
	err1 := NewDiagnosticError(ErrorTypeNetwork, "NET001", "first error").Build()
	err1.Timestamp = time.Now().Add(-2 * time.Minute)

	err2 := NewDiagnosticError(ErrorTypeAPI, "API001", "second error").Build()
	err2.Timestamp = time.Now().Add(-1 * time.Minute)

	reporter.Report(err1)
	reporter.Report(err2)

	report := reporter.GenerateReport()

	if !strings.Contains(report, "Diagnostic Error Report") {
		t.Error("Expected report header")
	}

	if !strings.Contains(report, "Total Errors: 2") {
		t.Error("Expected error count in report")
	}

	if !strings.Contains(report, "first error") {
		t.Error("Expected first error in report")
	}

	if !strings.Contains(report, "second error") {
		t.Error("Expected second error in report")
	}

	// Check ordering (should be chronological)
	firstPos := strings.Index(report, "first error")
	secondPos := strings.Index(report, "second error")

	if firstPos > secondPos {
		t.Error("Expected errors to be sorted chronologically")
	}
}

func TestErrorReporter_EmptyReport(t *testing.T) {
	ctx := context.Background()
	reporter := NewErrorReporter(ctx)

	report := reporter.GenerateReport()

	if report != "No errors to report" {
		t.Errorf("Expected empty report message, got: %s", report)
	}
}

// Benchmark tests
func BenchmarkNewDiagnosticError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewDiagnosticError(ErrorTypeNetwork, "NET001", "network error").Build()
	}
}

func BenchmarkDiagnosticError_DetailedError(b *testing.B) {
	diagErr := NewDiagnosticError(ErrorTypeNetwork, "NET001", "network error").
		WithComponent("github").
		WithOperation("api_call").
		WithContext("url", "https://api.github.com").
		WithSuggestion("Check connection").
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = diagErr.DetailedError()
	}
}

func BenchmarkDiagnosticError_ToJSON(b *testing.B) {
	diagErr := NewDiagnosticError(ErrorTypeAPI, "API001", "API error").
		WithComponent("github").
		WithContext("status_code", 404).
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = diagErr.ToJSON()
	}
}

func BenchmarkWrapWithDiagnostics(b *testing.B) {
	err := errors.New("test error") //nolint:err113 // test error

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WrapWithDiagnostics(err, ErrorTypeNetwork, "component", "operation")
	}
}

func BenchmarkErrorReporter_Report(b *testing.B) {
	ctx := context.Background()
	reporter := NewErrorReporter(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := NewDiagnosticError(ErrorTypeNetwork, "NET001", "network error").Build()
		reporter.Report(err)
	}
}

// Test system state capture
func TestSystemStateCapture(t *testing.T) {
	diagErr := NewDiagnosticError(ErrorTypeSystem, "SYS001", "system error").
		WithSystemState().
		Build()

	state := diagErr.SystemState

	// Check that capture time is set
	if state.CapturedAt.IsZero() {
		t.Error("Expected CapturedAt to be set")
	}

	// Check that some basic state is captured
	if state.WorkingDirectory == "" {
		t.Error("Expected working directory to be captured")
	}

	// Memory info should be captured
	if state.Memory == nil {
		t.Error("Expected memory info to be captured")
	} else {
		if state.Memory.NumGoroutines == 0 {
			t.Error("Expected goroutine count to be captured")
		}
	}

	// Dependencies should be checked
	if len(state.Dependencies) == 0 {
		t.Error("Expected dependencies to be checked")
	}
}

func TestSystemStateEnvironmentCapture(t *testing.T) {
	// Set a test environment variable
	testEnvVar := "TEST_DIAGNOSTIC_VAR"
	testValue := "test_value_12345"

	originalValue := os.Getenv(testEnvVar)
	_ = os.Setenv(testEnvVar, testValue)

	defer func() {
		if originalValue == "" {
			_ = os.Unsetenv(testEnvVar)
		} else {
			_ = os.Setenv(testEnvVar, originalValue)
		}
	}()

	diagErr := NewDiagnosticError(ErrorTypeSystem, "SYS001", "system error").
		WithSystemState().
		Build()

	// The test env var won't be in the captured environment since it's not in the relevant list
	// But we can check that some environment variables are captured
	if len(diagErr.SystemState.Environment) == 0 {
		t.Error("Expected some environment variables to be captured")
	}

	// Check that sensitive values would be redacted (test PATH which is typically present)
	if path, exists := diagErr.SystemState.Environment["PATH"]; exists {
		// PATH should be captured (it's in the relevant list)
		if len(path) > 100 {
			// Should be truncated if very long
			if !strings.HasSuffix(path, "...") {
				t.Error("Expected very long PATH to be truncated")
			}
		}
	}
}

// Test helper to compare error types
func TestErrorTypes(t *testing.T) {
	expectedTypes := []ErrorType{
		ErrorTypeConfiguration,
		ErrorTypeNetwork,
		ErrorTypePermission,
		ErrorTypeValidation,
		ErrorTypeSystem,
		ErrorTypeAPI,
		ErrorTypeFile,
		ErrorTypeGit,
		ErrorTypeUser,
		ErrorTypeInternal,
	}

	for _, errorType := range expectedTypes {
		diagErr := NewDiagnosticError(errorType, "TEST001", "test error").Build()

		if !diagErr.IsType(errorType) {
			t.Errorf("Expected error to be of type %s", errorType)
		}

		// Each error type should have some automatic suggestions
		if len(diagErr.Suggestions) == 0 {
			t.Errorf("Expected automatic suggestions for error type %s", errorType)
		}
	}
}
