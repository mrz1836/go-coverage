// Package validation provides comprehensive input validation for coverage files,
// configuration, GitHub environment, and other inputs used throughout the system.
package validation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Error definitions for err113 linter compliance
var (
	ErrFileEmpty                  = errors.New("file is empty")
	ErrEmptyFile                  = errors.New("empty file")
	ErrMissingModeDeclaration     = errors.New("missing mode declaration")
	ErrInvalidCoverageMode        = errors.New("invalid coverage mode")
	ErrNoValidCoverageLines       = errors.New("no valid coverage lines found")
	ErrLCOVMissingSourceFiles     = errors.New("LCOV format missing source file entries (SF:)")
	ErrLCOVMissingLineData        = errors.New("LCOV format missing line data entries (DA:)")
	ErrXMLMissingCoverageElement  = errors.New("XML format missing coverage element")
	ErrInvalidOwnerRepoFormat     = errors.New("should be in format owner/repo")
	ErrInvalidSHALength           = errors.New("should be a 40-character SHA")
	ErrInvalidSHAFormat           = errors.New("should be a valid hexadecimal SHA")
	ErrTokenTooShort              = errors.New("token appears to be too short")
	ErrExpectedStringValue        = errors.New("expected string value")
	ErrStringTooShort             = errors.New("string too short")
	ErrStringTooLong              = errors.New("string too long")
	ErrStringPatternMismatch      = errors.New("string does not match required pattern")
	ErrExpectedIntegerNotFloat    = errors.New("expected integer, got float")
	ErrUnknownValidationType      = errors.New("unknown validation type")
	ErrCannotValidateNumericRange = errors.New("cannot validate numeric range")
	ErrValueBelowMinimum          = errors.New("value is below minimum")
	ErrValueAboveMaximum          = errors.New("value is above maximum")
	ErrCannotConvertToFloat64     = errors.New("cannot convert to float64")
	ErrValueNotInAllowedValues    = errors.New("value is not in allowed values")
)

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// ValidationError represents a validation error or warning
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Value   interface{} `json:"value,omitempty"`
}

// Validator defines the interface for input validators
type Validator interface {
	Validate(ctx context.Context) *ValidationResult
}

// CoverageFileValidator validates coverage file format and content
type CoverageFileValidator struct {
	FilePath     string
	MaxSize      int64 // Maximum file size in bytes
	RequiredMode bool  // Whether coverage data is required or optional
}

// NewCoverageFileValidator creates a new coverage file validator
func NewCoverageFileValidator(filePath string) *CoverageFileValidator {
	return &CoverageFileValidator{
		FilePath:     filePath,
		MaxSize:      50 * 1024 * 1024, // 50MB default
		RequiredMode: true,
	}
}

// Validate validates the coverage file
func (v *CoverageFileValidator) Validate(ctx context.Context) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check if context is canceled
	select {
	case <-ctx.Done():
		result.addError("context", "validation canceled", "CONTEXT_CANCELED", nil)
		return result
	default:
	}

	// Check if file path is provided
	if v.FilePath == "" {
		if v.RequiredMode {
			result.addError("file_path", "coverage file path is required", "MISSING_FILE_PATH", nil)
		}
		return result
	}

	// Check if file exists
	info, err := os.Stat(v.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			result.addError("file_path", "coverage file does not exist", "FILE_NOT_FOUND", v.FilePath)
		} else {
			result.addError("file_path", fmt.Sprintf("failed to access file: %v", err), "FILE_ACCESS_ERROR", v.FilePath)
		}
		return result
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		result.addError("file_path", "path is not a regular file", "NOT_REGULAR_FILE", v.FilePath)
		return result
	}

	// Check file size
	if v.MaxSize > 0 && info.Size() > v.MaxSize {
		result.addError("file_size", fmt.Sprintf("file size (%d bytes) exceeds maximum allowed (%d bytes)", info.Size(), v.MaxSize), "FILE_TOO_LARGE", info.Size())
	}

	// Warn if file is very small
	if info.Size() < 100 {
		result.addWarning("file_size", "coverage file is very small, may indicate incomplete coverage", "FILE_VERY_SMALL", info.Size())
	}

	// Validate file format by reading first few lines
	if err := v.validateCoverageFormat(ctx); err != nil {
		result.addError("file_format", fmt.Sprintf("invalid coverage file format: %v", err), "INVALID_FORMAT", nil)
	}

	return result
}

// validateCoverageFormat validates the coverage file format
func (v *CoverageFileValidator) validateCoverageFormat(_ context.Context) error {
	file, err := os.Open(v.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Read first 4KB to check format
	buffer := make([]byte, 4096)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to read file: %w", err)
	}

	content := string(buffer[:n])
	lines := strings.Split(content, "\n")

	// Check for mode line (first line should indicate mode)
	if len(lines) == 0 {
		return ErrFileEmpty
	}

	firstLine := strings.TrimSpace(lines[0])
	if firstLine == "" && len(lines) > 1 {
		firstLine = strings.TrimSpace(lines[1])
	}

	// Check for common coverage file formats
	if strings.HasPrefix(firstLine, "mode:") {
		// Go coverage format
		return v.validateGoCoverageFormat(lines)
	} else if strings.Contains(content, "SF:") || strings.Contains(content, "DA:") {
		// LCOV format
		return v.validateLCOVFormat(content)
	} else if strings.Contains(content, "<coverage") {
		// XML coverage format
		return v.validateXMLCoverageFormat(content)
	}

	// If no recognized format, it might still be valid but warn
	return nil
}

// validateGoCoverageFormat validates Go coverage format
func (v *CoverageFileValidator) validateGoCoverageFormat(lines []string) error {
	if len(lines) < 1 {
		return ErrEmptyFile
	}

	// First line should be mode declaration
	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(firstLine, "mode:") {
		return ErrMissingModeDeclaration
	}

	// Validate mode
	modePart := strings.TrimPrefix(firstLine, "mode:")
	mode := strings.TrimSpace(modePart)
	validModes := []string{"set", "count", "atomic"}
	validMode := false
	for _, validModeItem := range validModes {
		if mode == validModeItem {
			validMode = true
			break
		}
	}
	if !validMode {
		return fmt.Errorf("%w: %s", ErrInvalidCoverageMode, mode)
	}

	// Validate coverage lines format (at least check a few lines)
	coverageLineRegex := regexp.MustCompile(`^[^:]+:\d+\.\d+,\d+\.\d+ \d+ \d+$`)
	validLines := 0
	for i, line := range lines[1:] {
		if i > 5 { // Check first 5 coverage lines
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !coverageLineRegex.MatchString(line) {
			return fmt.Errorf("invalid coverage line format at line %d: %s", i+2, line) //nolint:err113 // dynamic context
		}
		validLines++
	}

	if validLines == 0 {
		return ErrNoValidCoverageLines
	}

	return nil
}

// validateLCOVFormat validates LCOV format coverage files
func (v *CoverageFileValidator) validateLCOVFormat(content string) error {
	// LCOV format should contain SF: (source file) and DA: (line data) entries
	if !strings.Contains(content, "SF:") {
		return ErrLCOVMissingSourceFiles
	}

	if !strings.Contains(content, "DA:") {
		return ErrLCOVMissingLineData
	}

	return nil
}

// validateXMLCoverageFormat validates XML format coverage files
func (v *CoverageFileValidator) validateXMLCoverageFormat(content string) error {
	// Basic XML coverage format validation
	if !strings.Contains(content, "<coverage") {
		return ErrXMLMissingCoverageElement
	}

	return nil
}

// GitHubEnvironmentValidator validates GitHub Actions environment
type GitHubEnvironmentValidator struct {
	RequireGitHubActions bool
	RequiredEnvVars      []string
}

// NewGitHubEnvironmentValidator creates a new GitHub environment validator
func NewGitHubEnvironmentValidator() *GitHubEnvironmentValidator {
	return &GitHubEnvironmentValidator{
		RequireGitHubActions: true,
		RequiredEnvVars: []string{
			"GITHUB_REPOSITORY",
			"GITHUB_SHA",
			"GITHUB_TOKEN",
		},
	}
}

// Validate validates the GitHub environment
func (v *GitHubEnvironmentValidator) Validate(ctx context.Context) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check if running in GitHub Actions
	if v.RequireGitHubActions {
		if os.Getenv("GITHUB_ACTIONS") != "true" {
			result.addError("github_actions", "not running in GitHub Actions environment", "NOT_GITHUB_ACTIONS", nil)
		}
	}

	// Check required environment variables
	for _, envVar := range v.RequiredEnvVars {
		value := os.Getenv(envVar)
		if value == "" {
			result.addError("env_var", fmt.Sprintf("required environment variable %s is missing", envVar), "MISSING_ENV_VAR", envVar)
			continue
		}

		// Validate specific environment variables
		if err := v.validateEnvVar(envVar, value); err != nil {
			result.addError("env_var", fmt.Sprintf("invalid value for %s: %v", envVar, err), "INVALID_ENV_VAR", envVar)
		}
	}

	return result
}

// validateEnvVar validates specific environment variables
func (v *GitHubEnvironmentValidator) validateEnvVar(name, value string) error {
	switch name {
	case "GITHUB_REPOSITORY":
		// Should be in format owner/repo
		if !strings.Contains(value, "/") {
			return ErrInvalidOwnerRepoFormat
		}
		parts := strings.SplitN(value, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return ErrInvalidOwnerRepoFormat
		}

	case "GITHUB_SHA":
		// Should be a valid git SHA (40 character hex string)
		if len(value) != 40 {
			return ErrInvalidSHALength
		}
		if matched, _ := regexp.MatchString("^[a-f0-9]{40}$", value); !matched {
			return ErrInvalidSHAFormat
		}

	case "GITHUB_TOKEN":
		// Basic validation - should not be empty and should look like a token
		if len(value) < 10 {
			return ErrTokenTooShort
		}
		// Don't log the actual token value for security
	}

	return nil
}

// ConfigValidator validates configuration parameters
type ConfigValidator struct {
	Config map[string]interface{}
	Schema ValidationSchema
}

// ValidationSchema defines validation rules for configuration
type ValidationSchema struct {
	Fields []FieldValidation `json:"fields"`
}

// FieldValidation defines validation rules for a single field
type FieldValidation struct {
	Name          string        `json:"name"`
	Type          string        `json:"type"` // "string", "int", "float", "bool", "duration"
	Required      bool          `json:"required"`
	MinValue      interface{}   `json:"min_value,omitempty"`
	MaxValue      interface{}   `json:"max_value,omitempty"`
	MinLength     int           `json:"min_length,omitempty"`
	MaxLength     int           `json:"max_length,omitempty"`
	Pattern       string        `json:"pattern,omitempty"`
	AllowedValues []interface{} `json:"allowed_values,omitempty"`
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator(config map[string]interface{}) *ConfigValidator {
	return &ConfigValidator{
		Config: config,
		Schema: GetDefaultSchema(),
	}
}

// Validate validates the configuration
func (v *ConfigValidator) Validate(ctx context.Context) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate each field according to schema
	for _, fieldValidation := range v.Schema.Fields {
		value, exists := v.Config[fieldValidation.Name]

		// Check if required field is missing
		if fieldValidation.Required && !exists {
			result.addError(fieldValidation.Name, "required field is missing", "MISSING_REQUIRED_FIELD", nil)
			continue
		}

		// Skip validation if field is not present and not required
		if !exists {
			continue
		}

		// Validate field value
		if err := v.validateField(fieldValidation, value); err != nil {
			result.addError(fieldValidation.Name, err.Error(), "INVALID_FIELD_VALUE", value)
		}
	}

	return result
}

// validateField validates a single field according to its validation rules
func (v *ConfigValidator) validateField(fieldValidation FieldValidation, value interface{}) error {
	// Type validation
	if err := v.validateFieldType(fieldValidation.Type, value); err != nil {
		return fmt.Errorf("type validation failed: %w", err)
	}

	// String-specific validations
	if fieldValidation.Type == "string" {
		str, ok := value.(string)
		if !ok {
			return ErrExpectedStringValue
		}

		if fieldValidation.MinLength > 0 && len(str) < fieldValidation.MinLength {
			return fmt.Errorf("%w (minimum length: %d)", ErrStringTooShort, fieldValidation.MinLength)
		}

		if fieldValidation.MaxLength > 0 && len(str) > fieldValidation.MaxLength {
			return fmt.Errorf("%w (maximum length: %d)", ErrStringTooLong, fieldValidation.MaxLength)
		}

		if fieldValidation.Pattern != "" {
			matched, err := regexp.MatchString(fieldValidation.Pattern, str)
			if err != nil {
				return fmt.Errorf("pattern validation error: %w", err)
			}
			if !matched {
				return fmt.Errorf("%w: %s", ErrStringPatternMismatch, fieldValidation.Pattern)
			}
		}
	}

	// Numeric validations
	if fieldValidation.Type == "int" || fieldValidation.Type == "float" {
		if err := v.validateNumericRange(fieldValidation, value); err != nil {
			return err
		}
	}

	// Allowed values validation
	if len(fieldValidation.AllowedValues) > 0 {
		if err := v.validateAllowedValues(fieldValidation.AllowedValues, value); err != nil {
			return err
		}
	}

	return nil
}

// validateFieldType validates the type of a field value
func (v *ConfigValidator) validateFieldType(expectedType string, value interface{}) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value) //nolint:err113 // type info needed
		}
	case "int":
		switch v := value.(type) {
		case int, int32, int64:
			return nil
		case float64:
			// JSON numbers are parsed as float64, check if it's actually an integer
			if v == float64(int(v)) {
				return nil
			}
			return ErrExpectedIntegerNotFloat
		default:
			return fmt.Errorf("expected integer, got %T", value) //nolint:err113 // type info needed
		}
	case "float":
		switch value.(type) {
		case float32, float64, int, int32, int64:
			return nil
		default:
			return fmt.Errorf("expected number, got %T", value) //nolint:err113 // type info needed
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value) //nolint:err113 // type info needed
		}
	case "duration":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected duration string, got %T", value) //nolint:err113 // type info needed
		}
		if _, err := time.ParseDuration(str); err != nil {
			return fmt.Errorf("invalid duration format: %w", err)
		}
	default:
		return fmt.Errorf("%w: %s", ErrUnknownValidationType, expectedType)
	}

	return nil
}

// validateNumericRange validates numeric values against min/max constraints
func (v *ConfigValidator) validateNumericRange(fieldValidation FieldValidation, value interface{}) error {
	var numValue float64

	switch val := value.(type) {
	case int:
		numValue = float64(val)
	case int32:
		numValue = float64(val)
	case int64:
		numValue = float64(val)
	case float32:
		numValue = float64(val)
	case float64:
		numValue = val
	default:
		return fmt.Errorf("%w for type %T", ErrCannotValidateNumericRange, value)
	}

	// Check minimum value
	if fieldValidation.MinValue != nil {
		minVal, err := v.convertToFloat64(fieldValidation.MinValue)
		if err != nil {
			return fmt.Errorf("invalid min value in schema: %w", err)
		}
		if numValue < minVal {
			return fmt.Errorf("%w: %v (minimum %v)", ErrValueBelowMinimum, numValue, minVal)
		}
	}

	// Check maximum value
	if fieldValidation.MaxValue != nil {
		maxVal, err := v.convertToFloat64(fieldValidation.MaxValue)
		if err != nil {
			return fmt.Errorf("invalid max value in schema: %w", err)
		}
		if numValue > maxVal {
			return fmt.Errorf("%w: %v (maximum %v)", ErrValueAboveMaximum, numValue, maxVal)
		}
	}

	return nil
}

// convertToFloat64 converts various numeric types to float64
func (v *ConfigValidator) convertToFloat64(value interface{}) (float64, error) {
	switch val := value.(type) {
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("%w: %T", ErrCannotConvertToFloat64, value)
	}
}

// validateAllowedValues validates that a value is in the list of allowed values
func (v *ConfigValidator) validateAllowedValues(allowedValues []interface{}, value interface{}) error {
	for _, allowed := range allowedValues {
		if allowed == value {
			return nil
		}
	}

	return fmt.Errorf("%w: %v (allowed: %v)", ErrValueNotInAllowedValues, value, allowedValues)
}

// GetDefaultSchema returns the default validation schema for go-coverage configuration
func GetDefaultSchema() ValidationSchema {
	return ValidationSchema{
		Fields: []FieldValidation{
			{
				Name:      "input_file",
				Type:      "string",
				Required:  false,
				MinLength: 1,
				MaxLength: 500,
			},
			{
				Name:          "provider",
				Type:          "string",
				Required:      false,
				AllowedValues: []interface{}{"auto", "internal", "codecov"},
			},
			{
				Name:     "threshold",
				Type:     "float",
				Required: false,
				MinValue: 0.0,
				MaxValue: 100.0,
			},
			{
				Name:     "debug",
				Type:     "bool",
				Required: false,
			},
			{
				Name:     "dry_run",
				Type:     "bool",
				Required: false,
			},
			{
				Name:     "timeout",
				Type:     "duration",
				Required: false,
			},
			{
				Name:     "max_history_size",
				Type:     "int",
				Required: false,
				MinValue: 1,
				MaxValue: 10000,
			},
		},
	}
}

// Helper methods for ValidationResult

func (r *ValidationResult) addError(field, message, code string, value interface{}) {
	r.Valid = false
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
		Value:   value,
	})
}

func (r *ValidationResult) addWarning(field, message, code string, value interface{}) {
	r.Warnings = append(r.Warnings, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
		Value:   value,
	})
}

// HasErrors returns true if there are any validation errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are any validation warnings
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// ErrorMessages returns all error messages as a slice of strings
func (r *ValidationResult) ErrorMessages() []string {
	messages := make([]string, len(r.Errors))
	for i, err := range r.Errors {
		messages[i] = fmt.Sprintf("%s: %s", err.Field, err.Message)
	}
	return messages
}

// WarningMessages returns all warning messages as a slice of strings
func (r *ValidationResult) WarningMessages() []string {
	messages := make([]string, len(r.Warnings))
	for i, warn := range r.Warnings {
		messages[i] = fmt.Sprintf("%s: %s", warn.Field, warn.Message)
	}
	return messages
}

// String returns a string representation of the validation result
func (r *ValidationResult) String() string {
	var parts []string

	if r.Valid {
		parts = append(parts, "Validation: PASSED")
	} else {
		parts = append(parts, "Validation: FAILED")
	}

	if len(r.Errors) > 0 {
		parts = append(parts, fmt.Sprintf("Errors (%d):", len(r.Errors)))
		for _, err := range r.Errors {
			parts = append(parts, fmt.Sprintf("  - %s: %s", err.Field, err.Message))
		}
	}

	if len(r.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("Warnings (%d):", len(r.Warnings)))
		for _, warn := range r.Warnings {
			parts = append(parts, fmt.Sprintf("  - %s: %s", warn.Field, warn.Message))
		}
	}

	return strings.Join(parts, "\n")
}

// ValidateAll runs multiple validators and combines their results
func ValidateAll(ctx context.Context, validators ...Validator) *ValidationResult {
	combined := &ValidationResult{Valid: true}

	for _, validator := range validators {
		result := validator.Validate(ctx)

		// Combine errors and warnings
		combined.Errors = append(combined.Errors, result.Errors...)
		combined.Warnings = append(combined.Warnings, result.Warnings...)

		// Overall validity is false if any validator fails
		if !result.Valid {
			combined.Valid = false
		}
	}

	return combined
}
