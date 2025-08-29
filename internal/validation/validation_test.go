package validation

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCoverageFileValidator_ValidFile(t *testing.T) {
	// Create a temporary valid coverage file
	tempDir := t.TempDir()
	coverageFile := filepath.Join(tempDir, "coverage.txt")

	validContent := `mode: set
github.com/test/pkg/file.go:10.1,12.2 1 1
github.com/test/pkg/file.go:15.1,17.5 2 0
github.com/test/pkg/other.go:20.1,22.3 1 1
`

	if err := os.WriteFile(coverageFile, []byte(validContent), 0o600); err != nil {
		t.Fatalf("Failed to create test coverage file: %v", err)
	}

	validator := NewCoverageFileValidator(coverageFile)
	result := validator.Validate(context.Background())

	if !result.Valid {
		t.Errorf("Expected valid result, got invalid. Errors: %v", result.ErrorMessages())
	}

	if len(result.Errors) > 0 {
		t.Errorf("Expected no errors, got: %v", result.ErrorMessages())
	}
}

func TestCoverageFileValidator_MissingFile(t *testing.T) {
	validator := NewCoverageFileValidator("/nonexistent/file.txt")
	result := validator.Validate(context.Background())

	if result.Valid {
		t.Error("Expected invalid result for missing file")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors for missing file")
	}

	// Check for specific error
	found := false
	for _, err := range result.Errors {
		if err.Code == "FILE_NOT_FOUND" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected FILE_NOT_FOUND error")
	}
}

func TestCoverageFileValidator_EmptyPath(t *testing.T) {
	validator := NewCoverageFileValidator("")
	result := validator.Validate(context.Background())

	if result.Valid {
		t.Error("Expected invalid result for empty path")
	}

	// Check for missing file path error
	found := false
	for _, err := range result.Errors {
		if err.Code == "MISSING_FILE_PATH" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected MISSING_FILE_PATH error")
	}
}

func TestCoverageFileValidator_EmptyPathNotRequired(t *testing.T) {
	validator := NewCoverageFileValidator("")
	validator.RequiredMode = false
	result := validator.Validate(context.Background())

	if !result.Valid {
		t.Errorf("Expected valid result when file not required, got errors: %v", result.ErrorMessages())
	}
}

func TestCoverageFileValidator_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	coverageFile := filepath.Join(tempDir, "large_coverage.txt")

	// Create a file that's larger than the max size
	validator := NewCoverageFileValidator(coverageFile)
	validator.MaxSize = 100 // 100 bytes max

	largeContent := strings.Repeat("mode: set\n", 50) // Much larger than 100 bytes
	if err := os.WriteFile(coverageFile, []byte(largeContent), 0o600); err != nil {
		t.Fatalf("Failed to create large coverage file: %v", err)
	}

	result := validator.Validate(context.Background())

	if result.Valid {
		t.Error("Expected invalid result for large file")
	}

	// Check for file too large error
	found := false
	for _, err := range result.Errors {
		if err.Code == "FILE_TOO_LARGE" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected FILE_TOO_LARGE error")
	}
}

func TestCoverageFileValidator_SmallFile(t *testing.T) {
	tempDir := t.TempDir()
	coverageFile := filepath.Join(tempDir, "small_coverage.txt")

	// Create a very small file
	if err := os.WriteFile(coverageFile, []byte("mode: set\n"), 0o600); err != nil {
		t.Fatalf("Failed to create small coverage file: %v", err)
	}

	validator := NewCoverageFileValidator(coverageFile)
	result := validator.Validate(context.Background())

	// Should be valid but have warning
	if !result.Valid {
		t.Errorf("Expected valid result for small file, got errors: %v", result.ErrorMessages())
	}

	// Check for warning about small file
	found := false
	for _, warn := range result.Warnings {
		if warn.Code == "FILE_VERY_SMALL" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected FILE_VERY_SMALL warning")
	}
}

func TestCoverageFileValidator_InvalidFormat(t *testing.T) {
	tempDir := t.TempDir()
	coverageFile := filepath.Join(tempDir, "invalid_coverage.txt")

	// Create an invalid coverage file
	invalidContent := `invalid format
not a coverage file
random content
`
	if err := os.WriteFile(coverageFile, []byte(invalidContent), 0o600); err != nil {
		t.Fatalf("Failed to create invalid coverage file: %v", err)
	}

	validator := NewCoverageFileValidator(coverageFile)
	result := validator.Validate(context.Background())

	if result.Valid {
		t.Error("Expected invalid result for invalid format file")
	}
}

func TestCoverageFileValidator_ValidateGoCoverageFormat(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		expectErr bool
	}{
		{
			name: "valid set mode",
			content: `mode: set
github.com/test/pkg/file.go:10.1,12.2 1 1
github.com/test/pkg/file.go:15.1,17.5 2 0`,
			expectErr: false,
		},
		{
			name: "valid count mode",
			content: `mode: count
github.com/test/pkg/file.go:10.1,12.2 1 5
github.com/test/pkg/file.go:15.1,17.5 2 0`,
			expectErr: false,
		},
		{
			name: "valid atomic mode",
			content: `mode: atomic
github.com/test/pkg/file.go:10.1,12.2 1 1
github.com/test/pkg/file.go:15.1,17.5 2 0`,
			expectErr: false,
		},
		{
			name: "invalid mode",
			content: `mode: invalid
github.com/test/pkg/file.go:10.1,12.2 1 1`,
			expectErr: true,
		},
		{
			name: "missing mode",
			content: `github.com/test/pkg/file.go:10.1,12.2 1 1
github.com/test/pkg/file.go:15.1,17.5 2 0`,
			expectErr: true,
		},
		{
			name: "invalid coverage line format",
			content: `mode: set
invalid line format
github.com/test/pkg/file.go:15.1,17.5 2 0`,
			expectErr: true,
		},
		{
			name:      "empty file",
			content:   "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			coverageFile := filepath.Join(tempDir, "coverage.txt")

			if err := os.WriteFile(coverageFile, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("Failed to create test coverage file: %v", err)
			}

			validator := NewCoverageFileValidator(coverageFile)
			result := validator.Validate(context.Background())

			if tt.expectErr {
				if result.Valid {
					t.Errorf("Expected invalid result for test %s", tt.name)
				}
			} else {
				if !result.Valid {
					t.Errorf("Expected valid result for test %s, got errors: %v", tt.name, result.ErrorMessages())
				}
			}
		})
	}
}

func TestGitHubEnvironmentValidator_ValidEnvironment(t *testing.T) {
	// Setup valid environment variables
	originalEnvVars := map[string]string{
		"GITHUB_ACTIONS":    os.Getenv("GITHUB_ACTIONS"),
		"GITHUB_REPOSITORY": os.Getenv("GITHUB_REPOSITORY"),
		"GITHUB_SHA":        os.Getenv("GITHUB_SHA"),
		"GITHUB_TOKEN":      os.Getenv("GITHUB_TOKEN"),
	}

	// Cleanup
	defer func() {
		for key, value := range originalEnvVars {
			if value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, value)
			}
		}
	}()

	// Set valid environment
	_ = os.Setenv("GITHUB_ACTIONS", "true")
	_ = os.Setenv("GITHUB_REPOSITORY", "owner/repo")
	_ = os.Setenv("GITHUB_SHA", "abcdef1234567890abcdef1234567890abcdef12")
	_ = os.Setenv("GITHUB_TOKEN", "valid_token_here_with_sufficient_length")

	validator := NewGitHubEnvironmentValidator()
	result := validator.Validate(context.Background())

	if !result.Valid {
		t.Errorf("Expected valid result for valid environment, got errors: %v", result.ErrorMessages())
	}
}

func TestGitHubEnvironmentValidator_MissingEnvironment(t *testing.T) {
	// Save current environment
	originalEnvVars := map[string]string{
		"GITHUB_ACTIONS":    os.Getenv("GITHUB_ACTIONS"),
		"GITHUB_REPOSITORY": os.Getenv("GITHUB_REPOSITORY"),
		"GITHUB_SHA":        os.Getenv("GITHUB_SHA"),
		"GITHUB_TOKEN":      os.Getenv("GITHUB_TOKEN"),
	}

	// Cleanup
	defer func() {
		for key, value := range originalEnvVars {
			if value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, value)
			}
		}
	}()

	// Clear environment variables
	_ = os.Unsetenv("GITHUB_ACTIONS")
	_ = os.Unsetenv("GITHUB_REPOSITORY")
	_ = os.Unsetenv("GITHUB_SHA")
	_ = os.Unsetenv("GITHUB_TOKEN")

	validator := NewGitHubEnvironmentValidator()
	result := validator.Validate(context.Background())

	if result.Valid {
		t.Error("Expected invalid result for missing environment")
	}

	// Should have multiple errors
	if len(result.Errors) < 3 { // At least 3 required env vars missing
		t.Errorf("Expected at least 3 errors, got %d: %v", len(result.Errors), result.ErrorMessages())
	}
}

func TestGitHubEnvironmentValidator_InvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		envVar  string
		value   string
		isValid bool
	}{
		{"valid repository", "GITHUB_REPOSITORY", "owner/repo", true},
		{"invalid repository format", "GITHUB_REPOSITORY", "invalid", false},
		{"empty repository parts", "GITHUB_REPOSITORY", "/", false},
		{"valid SHA", "GITHUB_SHA", "abcdef1234567890abcdef1234567890abcdef12", true},
		{"invalid SHA length", "GITHUB_SHA", "abcdef123", false},
		{"invalid SHA characters", "GITHUB_SHA", "ghijkl1234567890abcdef1234567890abcdef12", false},
		{"valid token", "GITHUB_TOKEN", "valid_token_here", true},
		{"invalid token", "GITHUB_TOKEN", "short", false},
	}

	// Save original environment
	originalEnvVars := map[string]string{
		"GITHUB_ACTIONS":    os.Getenv("GITHUB_ACTIONS"),
		"GITHUB_REPOSITORY": os.Getenv("GITHUB_REPOSITORY"),
		"GITHUB_SHA":        os.Getenv("GITHUB_SHA"),
		"GITHUB_TOKEN":      os.Getenv("GITHUB_TOKEN"),
	}

	defer func() {
		for key, value := range originalEnvVars {
			if value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, value)
			}
		}
	}()

	validator := NewGitHubEnvironmentValidator()
	validator.RequireGitHubActions = false // Focus on env var validation

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment with defaults
			_ = os.Setenv("GITHUB_REPOSITORY", "default/repo")
			_ = os.Setenv("GITHUB_SHA", "abcdef1234567890abcdef1234567890abcdef12")
			_ = os.Setenv("GITHUB_TOKEN", "default_token_here")

			// Override with test value
			_ = os.Setenv(tt.envVar, tt.value)

			result := validator.Validate(context.Background())

			if tt.isValid {
				if !result.Valid {
					t.Errorf("Expected valid result for %s, got errors: %v", tt.name, result.ErrorMessages())
				}
			} else {
				if result.Valid {
					t.Errorf("Expected invalid result for %s", tt.name)
				}
			}
		})
	}
}

func TestConfigValidator_ValidConfig(t *testing.T) {
	config := map[string]interface{}{
		"input_file":       "coverage.txt",
		"provider":         "internal",
		"threshold":        75.0,
		"debug":            true,
		"dry_run":          false,
		"timeout":          "30s",
		"max_history_size": 100,
	}

	validator := NewConfigValidator(config)
	result := validator.Validate(context.Background())

	if !result.Valid {
		t.Errorf("Expected valid result for valid config, got errors: %v", result.ErrorMessages())
	}
}

func TestConfigValidator_InvalidConfig(t *testing.T) {
	config := map[string]interface{}{
		"provider":         "invalid_provider",
		"threshold":        150.0, // Over 100%
		"debug":            "not_boolean",
		"timeout":          "invalid_duration",
		"max_history_size": -1,
	}

	validator := NewConfigValidator(config)
	result := validator.Validate(context.Background())

	if result.Valid {
		t.Error("Expected invalid result for invalid config")
	}

	// Should have multiple errors
	if len(result.Errors) < 4 {
		t.Errorf("Expected at least 4 errors, got %d: %v", len(result.Errors), result.ErrorMessages())
	}
}

func TestConfigValidator_MissingRequiredFields(t *testing.T) {
	// Create schema with required field
	validator := NewConfigValidator(map[string]interface{}{})
	validator.Schema = ValidationSchema{
		Fields: []FieldValidation{
			{
				Name:     "required_field",
				Type:     "string",
				Required: true,
			},
		},
	}

	result := validator.Validate(context.Background())

	if result.Valid {
		t.Error("Expected invalid result for missing required field")
	}

	// Should have missing field error
	found := false
	for _, err := range result.Errors {
		if err.Code == "MISSING_REQUIRED_FIELD" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected MISSING_REQUIRED_FIELD error")
	}
}

func TestValidationResult_Methods(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{Field: "field1", Message: "error1", Code: "ERR1"},
			{Field: "field2", Message: "error2", Code: "ERR2"},
		},
		Warnings: []ValidationError{
			{Field: "field3", Message: "warning1", Code: "WARN1"},
		},
	}

	// Test HasErrors
	if !result.HasErrors() {
		t.Error("Expected HasErrors to return true")
	}

	// Test HasWarnings
	if !result.HasWarnings() {
		t.Error("Expected HasWarnings to return true")
	}

	// Test ErrorMessages
	errorMessages := result.ErrorMessages()
	if len(errorMessages) != 2 {
		t.Errorf("Expected 2 error messages, got %d", len(errorMessages))
	}

	// Test WarningMessages
	warningMessages := result.WarningMessages()
	if len(warningMessages) != 1 {
		t.Errorf("Expected 1 warning message, got %d", len(warningMessages))
	}

	// Test String representation
	str := result.String()
	if !strings.Contains(str, "FAILED") {
		t.Error("Expected String() to contain 'FAILED'")
	}
	if !strings.Contains(str, "field1: error1") {
		t.Error("Expected String() to contain error details")
	}
}

func TestValidateAll(t *testing.T) {
	ctx := context.Background()

	// Create multiple validators
	tempDir := t.TempDir()
	validCoverageFile := filepath.Join(tempDir, "coverage.txt")
	validContent := "mode: set\ngithub.com/test/pkg/file.go:10.1,12.2 1 1\n"
	if err := os.WriteFile(validCoverageFile, []byte(validContent), 0o600); err != nil {
		t.Fatalf("Failed to create test coverage file: %v", err)
	}

	coverageValidator := NewCoverageFileValidator(validCoverageFile)

	configValidator := NewConfigValidator(map[string]interface{}{
		"provider": "internal",
		"debug":    true,
	})

	// GitHub environment validator (will likely fail in test environment)
	githubValidator := NewGitHubEnvironmentValidator()
	githubValidator.RequireGitHubActions = false // Don't require GitHub Actions for test
	githubValidator.RequiredEnvVars = []string{} // No required env vars for test

	// Run ValidateAll
	result := ValidateAll(ctx, coverageValidator, configValidator, githubValidator)

	// Should be valid (coverage and config are valid, GitHub has no requirements)
	if !result.Valid {
		t.Errorf("Expected valid result from ValidateAll, got errors: %v", result.ErrorMessages())
	}
}

func TestValidateAll_WithErrors(t *testing.T) {
	ctx := context.Background()

	// Create validators with errors
	coverageValidator := NewCoverageFileValidator("/nonexistent/file.txt")

	configValidator := NewConfigValidator(map[string]interface{}{
		"provider":  "invalid_provider",
		"threshold": 150.0,
	})

	result := ValidateAll(ctx, coverageValidator, configValidator)

	if result.Valid {
		t.Error("Expected invalid result from ValidateAll with errors")
	}

	// Should have errors from both validators
	if len(result.Errors) < 2 {
		t.Errorf("Expected at least 2 errors, got %d: %v", len(result.Errors), result.ErrorMessages())
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	validator := NewCoverageFileValidator("test.txt")
	result := validator.Validate(ctx)

	if result.Valid {
		t.Error("Expected invalid result when context is canceled")
	}

	// Should have context cancellation error
	found := false
	for _, err := range result.Errors {
		if err.Code == "CONTEXT_CANCELED" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected CONTEXT_CANCELED error")
	}
}

func TestGetDefaultSchema(t *testing.T) {
	schema := GetDefaultSchema()

	if len(schema.Fields) == 0 {
		t.Error("Expected default schema to have fields")
	}

	// Check for some expected fields
	expectedFields := []string{"input_file", "provider", "threshold", "debug"}
	fieldMap := make(map[string]bool)

	for _, field := range schema.Fields {
		fieldMap[field.Name] = true
	}

	for _, expected := range expectedFields {
		if !fieldMap[expected] {
			t.Errorf("Expected field %s in default schema", expected)
		}
	}
}

func BenchmarkCoverageFileValidation(b *testing.B) {
	tempDir := b.TempDir()
	coverageFile := filepath.Join(tempDir, "coverage.txt")

	validContent := `mode: set
github.com/test/pkg/file.go:10.1,12.2 1 1
github.com/test/pkg/file.go:15.1,17.5 2 0
github.com/test/pkg/other.go:20.1,22.3 1 1
`

	if err := os.WriteFile(coverageFile, []byte(validContent), 0o600); err != nil {
		b.Fatalf("Failed to create test coverage file: %v", err)
	}

	validator := NewCoverageFileValidator(coverageFile)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(ctx)
	}
}

func BenchmarkConfigValidation(b *testing.B) {
	config := map[string]interface{}{
		"input_file":       "coverage.txt",
		"provider":         "internal",
		"threshold":        75.0,
		"debug":            true,
		"dry_run":          false,
		"timeout":          "30s",
		"max_history_size": 100,
	}

	validator := NewConfigValidator(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(ctx)
	}
}

func BenchmarkValidateAll(b *testing.B) {
	tempDir := b.TempDir()
	coverageFile := filepath.Join(tempDir, "coverage.txt")
	validContent := "mode: set\ngithub.com/test/pkg/file.go:10.1,12.2 1 1\n"
	if err := os.WriteFile(coverageFile, []byte(validContent), 0o600); err != nil {
		b.Fatalf("Failed to create test coverage file: %v", err)
	}

	coverageValidator := NewCoverageFileValidator(coverageFile)
	configValidator := NewConfigValidator(map[string]interface{}{
		"provider": "internal",
		"debug":    true,
	})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateAll(ctx, coverageValidator, configValidator)
	}
}
