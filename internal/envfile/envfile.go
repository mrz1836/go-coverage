// Package envfile provides loading and parsing of .env files into environment variables.
// It supports both individual file loading and directory-based loading with deterministic
// lexicographic ordering for modular environment configuration.
package envfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Sentinel errors for env file operations
var (
	ErrNotDirectory = errors.New("path is not a directory")
	ErrNoEnvFiles   = errors.New("no .env files found in directory")
)

// Load reads an env file and sets environment variables without overriding existing values.
func Load(filename string) error {
	return loadFile(filename, false)
}

// Overload reads an env file and sets environment variables, overriding existing values.
func Overload(filename string) error {
	return loadFile(filename, true)
}

// LoadDir loads all *.env files from the given directory in lexicographic order using
// Overload semantics (last-wins). When skipLocal is true, 99-local.env is skipped
// (intended for CI environments).
func LoadDir(dirPath string, skipLocal bool) error {
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("failed to access directory %s: %w", dirPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s: %w", dirPath, ErrNotDirectory)
	}

	pattern := filepath.Join(dirPath, "*.env")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob %s: %w", pattern, err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("%s: %w", dirPath, ErrNoEnvFiles)
	}

	sort.Strings(matches)

	for _, file := range matches {
		if skipLocal && filepath.Base(file) == "99-local.env" {
			continue
		}
		if err := Overload(file); err != nil {
			return fmt.Errorf("failed to load %s: %w", file, err)
		}
	}

	return nil
}

// loadFile reads and parses an env file, setting environment variables.
// When overload is false, existing environment variables are preserved.
func loadFile(filename string, overload bool) error {
	content, err := os.ReadFile(filename) //nolint:gosec // filename is provided by caller
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filename, err)
	}

	envMap := parse(string(content))

	for key, value := range envMap {
		if overload {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("failed to set %s: %w", key, err)
			}
		} else {
			if _, exists := os.LookupEnv(key); !exists {
				if err := os.Setenv(key, value); err != nil {
					return fmt.Errorf("failed to set %s: %w", key, err)
				}
			}
		}
	}

	return nil
}

// parse parses the content of an env file and returns a map of key-value pairs.
func parse(content string) map[string]string {
	envMap := make(map[string]string)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Find the first equals sign
		eqIndex := strings.Index(line, "=")
		if eqIndex < 0 {
			continue
		}

		key := strings.TrimSpace(line[:eqIndex])
		if key == "" {
			continue
		}

		// Strip optional "export " prefix
		key = strings.TrimPrefix(key, "export ")
		key = strings.TrimSpace(key)

		value := line[eqIndex+1:]
		value = processValue(value)

		envMap[key] = value
	}

	return envMap
}

// processValue processes a raw value from an env file, handling quotes and inline comments.
func processValue(value string) string {
	value = strings.TrimSpace(value)

	if len(value) == 0 {
		return value
	}

	// Handle double-quoted values
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1]
	}

	// Handle single-quoted values (no variable expansion, no inline comment stripping)
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		return value[1 : len(value)-1]
	}

	// For unquoted values, strip inline comments (# preceded by whitespace)
	if idx := strings.Index(value, " #"); idx >= 0 {
		value = strings.TrimSpace(value[:idx])
	}
	if idx := strings.Index(value, "\t#"); idx >= 0 {
		value = strings.TrimSpace(value[:idx])
	}

	return value
}
