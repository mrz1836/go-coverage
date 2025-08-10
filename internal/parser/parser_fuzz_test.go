// Package parser processes Go coverage profile data
// This file contains simplified fuzz tests for the parser package functions
package parser

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// FuzzParseStatementSimple tests the parseStatement method with basic validation
func FuzzParseStatementSimple(f *testing.F) {
	// Seed corpus with typical inputs
	f.Add("file.go:1.1,2.2 1 0")
	f.Add("file.go:10.5,15.20 5 3")
	f.Add("internal/package/file.go:100.1,200.50 10 15")
	f.Add("")                    // empty string
	f.Add("file.go:1.1,2.2")     // missing fields
	f.Add("file.go 1.1,2.2 1 0") // missing colon
	f.Add(":1.1,2.2 1 0")        // empty filename

	f.Fuzz(func(t *testing.T, line string) {
		parser := New()

		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("parseStatement panicked with input line=%q: %v", line, r)
			}
		}()

		_, filename, err := parser.parseStatement(line)

		// Basic validation: malformed inputs should return errors
		fields := strings.Fields(line)
		if line == "" || len(fields) != 3 || !strings.Contains(line, ":") {
			assert.Error(t, err, "Should return error for malformed input")
			return
		}

		// For well-formed inputs, either success or appropriate error
		if err == nil {
			// Success case: validate filename extraction
			colonIdx := strings.LastIndex(fields[0], ":")
			if colonIdx != -1 {
				expectedFilename := fields[0][:colonIdx]
				assert.Equal(t, expectedFilename, filename, "Filename should match expected")
			}

			// Ensure filename is valid UTF-8 if line was valid UTF-8
			if utf8.ValidString(line) {
				assert.True(t, utf8.ValidString(filename), "Filename should be valid UTF-8 when line is valid UTF-8")
			}
		}
	})
}

// FuzzShouldExcludeFileSimple tests the shouldExcludeFile method with basic validation
func FuzzShouldExcludeFileSimple(f *testing.F) {
	// Seed corpus with typical inputs
	f.Add("file.go")
	f.Add("internal/file.go")
	f.Add("test/file.go")
	f.Add("file_test.go")
	f.Add("file.pb.go")
	f.Add("") // empty string

	f.Fuzz(func(t *testing.T, filename string) {
		parser := New()

		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("shouldExcludeFile panicked with input filename=%q: %v", filename, r)
			}
		}()

		result := parser.shouldExcludeFile(filename)

		// Validate boolean result
		assert.IsType(t, false, result, "Should return boolean")

		// Test some known exclusion patterns
		if strings.HasSuffix(filename, "_test.go") {
			assert.True(t, result, "Should exclude test files")
		}
		if strings.Contains(filename, "test/") {
			assert.True(t, result, "Should exclude files in test/ directory")
		}
	})
}
