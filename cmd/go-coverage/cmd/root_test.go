package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecute(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		contains    []string
	}{
		{
			name:        "help command",
			args:        []string{"--help"},
			expectError: false,
			contains: []string{
				"Go Coverage is a self-contained",
				"Available Commands:",
				"complete",
				"history",
				"comment",
				"parse",
			},
		},
		{
			name:        "version flag",
			args:        []string{"--version"},
			expectError: false,
			contains:    []string{"go-coverage version 1.0.8"},
		},
		{
			name:        "debug flag",
			args:        []string{"--debug", "--help"},
			expectError: false,
			contains:    []string{"--debug", "Enable debug mode"},
		},
		{
			name:        "log level flag",
			args:        []string{"--log-level", "debug", "--help"},
			expectError: false,
			contains:    []string{"--log-level", "Log level"},
		},
		{
			name:        "log format flag",
			args:        []string{"--log-format", "json", "--help"},
			expectError: false,
			contains:    []string{"--log-format", "Log format"},
		},
		{
			name:        "invalid command",
			args:        []string{"invalid-command"},
			expectError: true,
			contains:    []string{"Error:", "unknown command"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			var buf bytes.Buffer

			// Create a new root command for this test to avoid interference
			testRootCmd := &cobra.Command{
				Use:   "go-coverage",
				Short: "Go-native coverage system for CI/CD",
				Long: `Go Coverage is a self-contained, Go-native coverage system that provides
professional coverage tracking, badge generation, and reporting while maintaining
the simplicity and performance that Go developers expect.

Built as a bolt-on solution completely encapsulated within the .github folder,
this tool replaces Codecov with zero external service dependencies.`,
				Version: "1.0.8",
			}

			// Add flags
			testRootCmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
			testRootCmd.PersistentFlags().StringP("log-level", "l", "info", "Log level (debug, info, warn, error)")
			testRootCmd.PersistentFlags().String("log-format", "text", "Log format (text, json, pretty)")

			// Add subcommands
			testRootCmd.AddCommand(completeCmd)
			testRootCmd.AddCommand(historyCmd)
			testRootCmd.AddCommand(commentCmd)
			testRootCmd.AddCommand(parseCmd)
			testRootCmd.AddCommand(setupPagesCmd)

			testRootCmd.SetOut(&buf)
			testRootCmd.SetErr(&buf)
			testRootCmd.SetArgs(tt.args)

			// Execute command
			err := testRootCmd.Execute()

			// Check error expectation
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Check output contains expected strings
			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

func TestRootCommandSetup(t *testing.T) {
	// Test that rootCmd is properly configured
	assert.Equal(t, "go-coverage", rootCmd.Use)
	assert.Equal(t, "Go-native coverage system for CI/CD", rootCmd.Short)
	assert.Contains(t, rootCmd.Long, "Go Coverage is a self-contained")
	assert.Equal(t, "1.0.8", rootCmd.Version)

	// Test that all expected flags exist
	flagNames := []string{"debug", "log-level", "log-format"}
	for _, flagName := range flagNames {
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}

	// Test flag defaults
	debugFlag := rootCmd.PersistentFlags().Lookup("debug")
	assert.Equal(t, "false", debugFlag.DefValue)

	logLevelFlag := rootCmd.PersistentFlags().Lookup("log-level")
	assert.Equal(t, "info", logLevelFlag.DefValue)

	logFormatFlag := rootCmd.PersistentFlags().Lookup("log-format")
	assert.Equal(t, "text", logFormatFlag.DefValue)
}

func TestRootCommandSubcommands(t *testing.T) {
	// Test that all expected subcommands are added
	expectedCommands := []string{"complete", "history", "comment", "parse", "setup-pages"}
	actualCommands := make([]string, 0)

	for _, cmd := range rootCmd.Commands() {
		actualCommands = append(actualCommands, cmd.Name())
	}

	// Check each expected command exists
	for _, expected := range expectedCommands {
		assert.Contains(t, actualCommands, expected, "Command %s should be added", expected)
	}
}

func TestExecuteWithoutArgs(t *testing.T) {
	// Test Execute function directly
	// This tests the module-level Execute function
	var buf bytes.Buffer

	// Save original and restore after test
	originalRootCmd := rootCmd
	defer func() {
		rootCmd = originalRootCmd
	}()

	// Create a test root command with help output
	rootCmd = &cobra.Command{
		Use:   "go-coverage",
		Short: "Go-native coverage system for CI/CD",
		Long: `Go Coverage is a self-contained, Go-native coverage system that provides
professional coverage tracking, badge generation, and reporting while maintaining
the simplicity and performance that Go developers expect.`,
		Version: "1.0.8",
		Args:    cobra.NoArgs,
	}

	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"--help"})

	err := Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Go Coverage is a self-contained")
}

func TestExecuteWithError(t *testing.T) {
	// Save original and restore after test
	originalRootCmd := rootCmd
	defer func() {
		rootCmd = originalRootCmd
	}()

	// Create a test root command that returns an error
	rootCmd = &cobra.Command{
		Use: "go-coverage",
		RunE: func(cmd *cobra.Command, args []string) error {
			return assert.AnError
		},
	}

	err := Execute()
	require.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestFlagParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flagName string
		expected string
	}{
		{
			name:     "debug flag true",
			args:     []string{"--debug", "--help"},
			flagName: "debug",
			expected: "true",
		},
		{
			name:     "log level custom",
			args:     []string{"--log-level", "debug", "--help"},
			flagName: "log-level",
			expected: "debug",
		},
		{
			name:     "log level short flag",
			args:     []string{"-l", "warn", "--help"},
			flagName: "log-level",
			expected: "warn",
		},
		{
			name:     "log format json",
			args:     []string{"--log-format", "json", "--help"},
			flagName: "log-format",
			expected: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRootCmd := &cobra.Command{
				Use: "go-coverage",
			}

			// Add flags
			testRootCmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
			testRootCmd.PersistentFlags().StringP("log-level", "l", "info", "Log level")
			testRootCmd.PersistentFlags().String("log-format", "text", "Log format")

			testRootCmd.SetArgs(tt.args)

			err := testRootCmd.Execute()
			require.NoError(t, err)

			// Check flag value
			if tt.flagName == "debug" {
				value, err := testRootCmd.PersistentFlags().GetBool(tt.flagName)
				require.NoError(t, err)
				assert.Equal(t, tt.expected == "true", value)
			} else {
				value, err := testRootCmd.PersistentFlags().GetString(tt.flagName)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, value)
			}
		})
	}
}

func TestRootCommandValidation(t *testing.T) {
	// Test environment variable isolation
	originalEnv := os.Getenv("GO_COVERAGE_DEBUG")
	defer func() {
		if originalEnv != "" {
			require.NoError(t, os.Setenv("GO_COVERAGE_DEBUG", originalEnv))
		} else {
			require.NoError(t, os.Unsetenv("GO_COVERAGE_DEBUG"))
		}
	}()

	// Set environment variable
	require.NoError(t, os.Setenv("GO_COVERAGE_DEBUG", "true"))

	// Test that command still works with environment variables
	var buf bytes.Buffer
	testCmd := &cobra.Command{
		Use: "go-coverage",
	}
	testCmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)
	testCmd.SetArgs([]string{"--help"})

	err := testCmd.Execute()
	require.NoError(t, err)
}
