package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

// Static errors for testing
var (
	errTest    = errors.New("test error")
	errChain   = errors.New("chain error")
	errGeneric = errors.New("test")
)

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{Level(99), "INFO"}, // Unknown level defaults to INFO
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		var buf bytes.Buffer
		config := &Config{
			Level:  DebugLevel,
			Format: "json",
			Output: &buf,
		}

		logger := NewLogger(config)
		if logger == nil {
			t.Fatal("NewLogger returned nil")
		}

		// Test that it uses the provided config
		logger.Info("test message")
		output := buf.String()
		if !strings.Contains(output, "test message") {
			t.Errorf("Expected log output to contain 'test message', got: %s", output)
		}
	})

	t.Run("with nil config", func(t *testing.T) {
		logger := NewLogger(nil)
		if logger == nil {
			t.Fatal("NewLogger returned nil")
		}

		// Should use defaults without panicking
		logger.Info("test")
	})
}

func TestNewFromEnv(t *testing.T) {
	tests := []struct {
		name           string
		envLevel       string
		envFormat      string
		expectedLevel  Level
		expectedFormat string
	}{
		{"default", "", "", InfoLevel, "text"},
		{"debug level", "DEBUG", "", DebugLevel, "text"},
		{"warn level", "WARN", "", WarnLevel, "text"},
		{"json format", "", "json", InfoLevel, "json"},
		{"debug json", "DEBUG", "json", DebugLevel, "json"},
		{"invalid level", "INVALID", "", InfoLevel, "text"},
		{"invalid format", "", "invalid", InfoLevel, "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			if tt.envLevel != "" {
				_ = os.Setenv("GO_COVERAGE_LOG_LEVEL", tt.envLevel)
				defer func() { _ = os.Unsetenv("GO_COVERAGE_LOG_LEVEL") }()
			}
			if tt.envFormat != "" {
				_ = os.Setenv("GO_COVERAGE_LOG_FORMAT", tt.envFormat)
				defer func() { _ = os.Unsetenv("GO_COVERAGE_LOG_FORMAT") }()
			}

			var buf bytes.Buffer
			logger := NewFromEnv()

			// Cast to access internal config for testing
			if sl, ok := logger.(*simpleLogger); ok {
				sl.config.Output = &buf
				if sl.config.Level != tt.expectedLevel {
					t.Errorf("Expected level %v, got %v", tt.expectedLevel, sl.config.Level)
				}
				if sl.config.Format != tt.expectedFormat {
					t.Errorf("Expected format %s, got %s", tt.expectedFormat, sl.config.Format)
				}
			}
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: "text",
		Output: &buf,
	}

	logger := NewLogger(config)

	// These should be logged
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// This should be filtered out
	logger.Debug("debug message")

	output := buf.String()

	if !strings.Contains(output, "info message") {
		t.Error("Expected info message to be logged")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Expected warn message to be logged")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Expected error message to be logged")
	}
	if strings.Contains(output, "debug message") {
		t.Error("Expected debug message to be filtered out")
	}
}

func TestLoggerFormattedMessages(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: "text",
		Output: &buf,
	}

	logger := NewLogger(config)

	logger.Infof("formatted %s %d", "message", 123)

	output := buf.String()
	if !strings.Contains(output, "formatted message 123") {
		t.Errorf("Expected formatted message, got: %s", output)
	}
}

func TestWithField(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: "text",
		Output: &buf,
	}

	logger := NewLogger(config)

	logger.WithField("key", "value").Info("test message")

	output := buf.String()
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected field in output, got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected message in output, got: %s", output)
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: "text",
		Output: &buf,
	}

	logger := NewLogger(config)

	fields := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
	}

	logger.WithFields(fields).Info("test message")

	output := buf.String()
	if !strings.Contains(output, "field1=value1") {
		t.Errorf("Expected field1 in output, got: %s", output)
	}
	if !strings.Contains(output, "field2=42") {
		t.Errorf("Expected field2 in output, got: %s", output)
	}
}

func TestWithError(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: "text",
		Output: &buf,
	}

	logger := NewLogger(config)

	logger.WithError(errTest).Warn("something went wrong")

	output := buf.String()
	if !strings.Contains(output, "error=test error") {
		t.Errorf("Expected error in output, got: %s", output)
	}
}

func TestWithContext(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: "text",
		Output: &buf,
	}

	logger := NewLogger(config)
	ctx := context.Background()

	// Normal context should work
	logger.WithContext(ctx).Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected message with context, got: %s", output)
	}

	// Canceled context should skip logging
	buf.Reset()
	canceledCtx, cancel := context.WithCancel(ctx)
	cancel()

	logger.WithContext(canceledCtx).Info("canceled message")

	output = buf.String()
	if strings.Contains(output, "canceled message") {
		t.Errorf("Expected canceled context to skip logging, got: %s", output)
	}
}

func TestChaining(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: "text",
		Output: &buf,
	}

	logger := NewLogger(config)

	// Test method chaining
	logger.WithField("key1", "value1").
		WithFields(map[string]interface{}{"key2": "value2"}).
		WithError(errChain).
		Info("chained message")

	output := buf.String()
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("Expected key1 in chained output, got: %s", output)
	}
	if !strings.Contains(output, "key2=value2") {
		t.Errorf("Expected key2 in chained output, got: %s", output)
	}
	if !strings.Contains(output, "error=chain error") {
		t.Errorf("Expected error in chained output, got: %s", output)
	}
	if !strings.Contains(output, "chained message") {
		t.Errorf("Expected message in chained output, got: %s", output)
	}
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: "json",
		Output: &buf,
	}

	logger := NewLogger(config)

	logger.WithField("test_field", "test_value").Info("json test")

	output := buf.String()

	// Parse as JSON to verify format
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v, output: %s", err, output)
	}

	if logEntry["message"] != "json test" {
		t.Errorf("Expected message 'json test', got: %v", logEntry["message"])
	}

	if logEntry["level"] != "INFO" {
		t.Errorf("Expected level 'INFO', got: %v", logEntry["level"])
	}

	fields, ok := logEntry["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields object in JSON")
	}

	if fields["test_field"] != "test_value" {
		t.Errorf("Expected test_field 'test_value', got: %v", fields["test_field"])
	}
}

func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: "text",
		Output: &buf,
	}

	logger := NewLogger(config)

	logger.WithField("test_field", "test_value").Info("text test")

	output := buf.String()

	// Check text format structure
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Expected [INFO] in text output, got: %s", output)
	}
	if !strings.Contains(output, "text test") {
		t.Errorf("Expected message in text output, got: %s", output)
	}
	if !strings.Contains(output, "test_field=test_value") {
		t.Errorf("Expected field in text output, got: %s", output)
	}
}

func TestLoggerInterface(_ *testing.T) {
	// Verify that our types implement the Logger interface
	var _ Logger = (*simpleLogger)(nil)
	var _ Logger = (*entry)(nil)

	// Test that we can use the interface polymorphically
	logger := NewLogger(nil)

	// All these should compile and work
	logger.Info("test")
	logger.WithField("key", "value").Warn("test")
	logger.WithError(errGeneric).Error("test")
	logger.WithContext(context.Background()).Debug("test")
}

// TestLoggerFormattedMethods tests the formatted logging methods that have 0% coverage
func TestLoggerFormattedMethods(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  DebugLevel,
		Format: "text",
		Output: &buf,
	}

	logger := NewLogger(config)

	// Test Debugf method
	logger.Debugf("debug message with %s and %d", "string", 42)
	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Expected [DEBUG] in output for Debugf, got: %s", output)
	}
	if !strings.Contains(output, "debug message with string and 42") {
		t.Errorf("Expected formatted message in Debugf output, got: %s", output)
	}

	buf.Reset()

	// Test Warnf method
	logger.Warnf("warning message with %s", "format")
	output = buf.String()
	if !strings.Contains(output, "[WARN]") {
		t.Errorf("Expected [WARN] in output for Warnf, got: %s", output)
	}
	if !strings.Contains(output, "warning message with format") {
		t.Errorf("Expected formatted message in Warnf output, got: %s", output)
	}

	buf.Reset()

	// Test Errorf method
	logger.Errorf("error message with %s and %v", "text", errTest)
	output = buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("Expected [ERROR] in output for Errorf, got: %s", output)
	}
	if !strings.Contains(output, "error message with text and") {
		t.Errorf("Expected formatted message in Errorf output, got: %s", output)
	}
}

// TestLoggerEntryFormattedMethods tests the formatted methods on entry objects
func TestLoggerEntryFormattedMethods(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  DebugLevel,
		Format: "text",
		Output: &buf,
	}

	logger := NewLogger(config)
	entry := logger.WithField("test", "value")

	// Test Debugf method on entry
	entry.Debugf("entry debug with %d items", 5)
	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Expected [DEBUG] in output for entry Debugf, got: %s", output)
	}
	if !strings.Contains(output, "entry debug with 5 items") {
		t.Errorf("Expected formatted message in entry Debugf output, got: %s", output)
	}

	buf.Reset()

	// Test Infof method on entry
	entry.Infof("entry info with %s", "parameter")
	output = buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Expected [INFO] in output for entry Infof, got: %s", output)
	}
	if !strings.Contains(output, "entry info with parameter") {
		t.Errorf("Expected formatted message in entry Infof output, got: %s", output)
	}

	buf.Reset()

	// Test Warnf method on entry
	entry.Warnf("entry warn with %.2f percent", 85.67)
	output = buf.String()
	if !strings.Contains(output, "[WARN]") {
		t.Errorf("Expected [WARN] in output for entry Warnf, got: %s", output)
	}
	if !strings.Contains(output, "entry warn with 85.67 percent") {
		t.Errorf("Expected formatted message in entry Warnf output, got: %s", output)
	}

	buf.Reset()

	// Test Errorf method on entry
	entry.Errorf("entry error with %v", errChain)
	output = buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("Expected [ERROR] in output for entry Errorf, got: %s", output)
	}
	if !strings.Contains(output, "entry error with chain error") {
		t.Errorf("Expected formatted message in entry Errorf output, got: %s", output)
	}
}

// TestLoggerEntryWithContext tests the WithContext method that has 0% coverage
func TestLoggerEntryWithContext(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: "json",
		Output: &buf,
	}

	logger := NewLogger(config)
	entry := logger.WithField("base", "value")

	// Test WithContext method on entry
	type requestIDKey string
	ctx := context.WithValue(context.Background(), requestIDKey("request-id"), "test-123")
	contextEntry := entry.WithContext(ctx)

	// Verify it returns a new entry and doesn't panic
	if contextEntry == nil {
		t.Error("WithContext should return a non-nil entry")
	}

	// Log a message to verify it works
	contextEntry.Info("test message with context")
	output := buf.String()

	// Parse JSON output to verify structure
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if logEntry["message"] != "test message with context" {
		t.Errorf("Expected message to be preserved, got: %v", logEntry["message"])
	}
}

func TestEntryImmutability(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: "text",
		Output: &buf,
	}

	logger := NewLogger(config)

	// Create base entry
	entry1 := logger.WithField("base", "value")

	// Create derived entries
	entry2 := entry1.WithField("extra", "value2")
	entry3 := entry1.WithField("different", "value3")

	// Log from each entry
	entry2.Info("entry2 message")
	entry3.Info("entry3 message")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 2 {
		t.Fatalf("Expected 2 log lines, got %d", len(lines))
	}

	// entry2 should have both base and extra
	if !strings.Contains(lines[0], "base=value") || !strings.Contains(lines[0], "extra=value2") {
		t.Errorf("entry2 missing expected fields: %s", lines[0])
	}
	if strings.Contains(lines[0], "different=value3") {
		t.Errorf("entry2 has field from entry3: %s", lines[0])
	}

	// entry3 should have base and different
	if !strings.Contains(lines[1], "base=value") || !strings.Contains(lines[1], "different=value3") {
		t.Errorf("entry3 missing expected fields: %s", lines[1])
	}
	if strings.Contains(lines[1], "extra=value2") {
		t.Errorf("entry3 has field from entry2: %s", lines[1])
	}
}
