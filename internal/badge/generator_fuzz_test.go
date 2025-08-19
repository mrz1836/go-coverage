// Package badge generates SVG coverage badges
// This file contains comprehensive fuzz tests for the badge package functions
package badge

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FuzzGetColorForPercentage tests the getColorForPercentage method with diverse percentage inputs
func FuzzGetColorForPercentage(f *testing.F) {
	// Seed corpus with typical and edge case inputs
	f.Add(0.0)
	f.Add(50.0)
	f.Add(75.0)
	f.Add(85.0)
	f.Add(90.0)
	f.Add(95.0)
	f.Add(100.0)
	f.Add(-1.0) // negative
	f.Add(-100.0)
	f.Add(-1000.0)
	f.Add(101.0) // over 100%
	f.Add(1000.0)
	f.Add(10000.0)
	f.Add(0.1) // very small
	f.Add(0.01)
	f.Add(0.001)
	f.Add(99.9) // close to 100
	f.Add(99.99)
	f.Add(99.999)
	f.Add(49.9) // boundary case
	f.Add(50.1) // boundary case
	f.Add(74.9) // just below threshold
	f.Add(75.0) // exactly at threshold
	f.Add(75.1) // just above threshold
	f.Add(84.9)
	f.Add(85.0)
	f.Add(85.1)
	f.Add(94.9)
	f.Add(95.0)
	f.Add(95.1)
	f.Add(math.Inf(1))                  // positive infinity
	f.Add(math.Inf(-1))                 // negative infinity
	f.Add(math.NaN())                   // NaN
	f.Add(math.MaxFloat64)              // maximum float64
	f.Add(-math.MaxFloat64)             // minimum float64
	f.Add(math.SmallestNonzeroFloat64)  // smallest positive float64
	f.Add(-math.SmallestNonzeroFloat64) // smallest negative float64

	f.Fuzz(func(t *testing.T, percentage float64) {
		generator := New()

		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("getColorForPercentage panicked with percentage=%f: %v", percentage, r)
			}
		}()

		result := generator.getColorForPercentage(percentage)

		// Validate output format
		assert.NotEmpty(t, result, "Should always return non-empty string")
		assert.True(t, strings.HasPrefix(result, "#"), "Should start with # (hex color format)")
		assert.True(t, len(result) == 7 || len(result) == 4, "Should be valid hex color length (7 or 4 characters)")

		// Validate hex characters after #
		hexPart := result[1:]
		for _, char := range hexPart {
			assert.True(t,
				(char >= '0' && char <= '9') ||
					(char >= 'a' && char <= 'f') ||
					(char >= 'A' && char <= 'F'),
				"Should contain only valid hex characters, got %c", char)
		}

		// Validate color assignment based on thresholds (default config)
		defaultConfig := &Config{
			ThresholdConfig: ThresholdConfig{
				Excellent:  95.0,
				Good:       85.0,
				Acceptable: 75.0,
				Low:        60.0,
			},
		}

		// Test color mapping logic (handle special float64 values)
		if math.IsNaN(percentage) {
			// NaN comparisons always return false, should fall to default case
			assert.Equal(t, "#dc3545", result, "NaN should use red (default case)")
		} else if math.IsInf(percentage, 1) {
			// Positive infinity should be >= excellent threshold
			assert.Equal(t, "#28a745", result, "Positive infinity should use bright green")
		} else if math.IsInf(percentage, -1) {
			// Negative infinity should fall to default case
			assert.Equal(t, "#dc3545", result, "Negative infinity should use red")
		} else {
			// Normal float64 values
			switch {
			case percentage >= defaultConfig.ThresholdConfig.Excellent:
				assert.Equal(t, "#28a745", result, "Should use bright green for >= 95%%")
			case percentage >= defaultConfig.ThresholdConfig.Good:
				assert.Equal(t, "#3fb950", result, "Should use green for >= 85%%")
			case percentage >= defaultConfig.ThresholdConfig.Acceptable:
				assert.Equal(t, "#ffc107", result, "Should use yellow for >= 75%%")
			case percentage >= defaultConfig.ThresholdConfig.Low:
				assert.Equal(t, "#fd7e14", result, "Should use orange for >= 60%%")
			default:
				assert.Equal(t, "#dc3545", result, "Should use red for < 60%%")
			}
		}

		// Ensure result is valid UTF-8
		assert.True(t, utf8.ValidString(result), "Result should be valid UTF-8")
	})
}

// FuzzGetColorForPercentageWithCustomConfig tests getColorForPercentage with custom threshold configurations
func FuzzGetColorForPercentageWithCustomConfig(f *testing.F) {
	// Seed corpus with various percentages and custom thresholds
	f.Add(50.0, 90.0, 80.0, 70.0, 60.0)
	f.Add(95.0, 95.0, 85.0, 75.0, 65.0)
	f.Add(0.0, 100.0, 90.0, 80.0, 70.0)
	f.Add(100.0, 99.0, 95.0, 90.0, 85.0)
	f.Add(-10.0, 50.0, 40.0, 30.0, 20.0)
	f.Add(1000.0, 1000.0, 900.0, 800.0, 700.0)
	f.Add(50.0, 60.0, 50.0, 40.0, 30.0) // percentage at good threshold
	f.Add(40.0, 60.0, 50.0, 40.0, 30.0) // percentage at acceptable threshold
	f.Add(30.0, 60.0, 50.0, 40.0, 30.0) // percentage at low threshold
	f.Add(20.0, 60.0, 50.0, 40.0, 30.0) // percentage below low threshold
	f.Add(math.NaN(), 95.0, 85.0, 75.0, 65.0)
	f.Add(math.Inf(1), 95.0, 85.0, 75.0, 65.0)
	f.Add(math.Inf(-1), 95.0, 85.0, 75.0, 65.0)

	f.Fuzz(func(t *testing.T, percentage, excellent, good, acceptable, low float64) {
		config := &Config{
			ThresholdConfig: ThresholdConfig{
				Excellent:  excellent,
				Good:       good,
				Acceptable: acceptable,
				Low:        low,
			},
		}
		generator := NewWithConfig(config)

		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("getColorForPercentage with custom config panicked with percentage=%f, thresholds=[%f,%f,%f,%f]: %v", percentage, excellent, good, acceptable, low, r)
			}
		}()

		result := generator.getColorForPercentage(percentage)

		// Validate output format
		assert.NotEmpty(t, result, "Should always return non-empty string")
		assert.True(t, strings.HasPrefix(result, "#"), "Should start with # (hex color format)")
		assert.True(t, len(result) == 7 || len(result) == 4, "Should be valid hex color length")

		// Validate hex characters
		hexPart := result[1:]
		for _, char := range hexPart {
			assert.True(t,
				(char >= '0' && char <= '9') ||
					(char >= 'a' && char <= 'f') ||
					(char >= 'A' && char <= 'F'),
				"Should contain only valid hex characters")
		}

		// The result should be one of the expected colors
		expectedColors := []string{"#28a745", "#3fb950", "#ffc107", "#fd7e14", "#dc3545"}
		assert.Contains(t, expectedColors, result, "Should return one of the defined colors")

		// Ensure result is valid UTF-8
		assert.True(t, utf8.ValidString(result), "Result should be valid UTF-8")
	})
}

// FuzzGetColorByName tests the getColorByName method with diverse string inputs
func FuzzGetColorByName(f *testing.F) {
	// Seed corpus with valid and invalid color names
	f.Add("excellent")
	f.Add("good")
	f.Add("acceptable")
	f.Add("low")
	f.Add("poor")
	f.Add("") // empty string
	f.Add("invalid")
	f.Add("EXCELLENT") // uppercase
	f.Add("Good")      // mixed case
	f.Add("excellent ")
	f.Add(" excellent")
	f.Add("excellent\n")
	f.Add("excellent\t")
	f.Add("unicode🔥")
	f.Add("\x00\x01\x02")        // null bytes
	f.Add("../../../etc/passwd") // path traversal
	f.Add("javascript:alert(1)") // XSS attempt
	f.Add("\n\r\t")              // control characters
	f.Add("very long string that might cause issues with color name matching")
	f.Add("excellent good acceptable low poor") // multiple names
	f.Add("1234567890")                         // numeric string
	f.Add("!@#$%^&*()")                         // special characters
	f.Add("\U0001f600\U0001f601\U0001f602")     // emoji

	f.Fuzz(func(t *testing.T, name string) {
		generator := New()

		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("getColorByName panicked with name=%q: %v", name, r)
			}
		}()

		result := generator.getColorByName(name)

		// Validate output format
		assert.NotEmpty(t, result, "Should always return non-empty string")
		assert.True(t, strings.HasPrefix(result, "#"), "Should start with # (hex color format)")
		assert.True(t, len(result) == 7 || len(result) == 4, "Should be valid hex color length")

		// Validate hex characters
		hexPart := result[1:]
		for _, char := range hexPart {
			assert.True(t,
				(char >= '0' && char <= '9') ||
					(char >= 'a' && char <= 'f') ||
					(char >= 'A' && char <= 'F'),
				"Should contain only valid hex characters")
		}

		// Test specific color mappings
		switch name {
		case "excellent":
			assert.Equal(t, "#28a745", result, "excellent should return bright green")
		case "good":
			assert.Equal(t, "#3fb950", result, "good should return green")
		case "acceptable":
			assert.Equal(t, "#ffc107", result, "acceptable should return yellow")
		case "low":
			assert.Equal(t, "#fd7e14", result, "low should return orange")
		case "poor":
			assert.Equal(t, "#dc3545", result, "poor should return red")
		default:
			assert.Equal(t, "#8b949e", result, "unknown names should return neutral gray")
		}

		// Ensure result is valid UTF-8
		assert.True(t, utf8.ValidString(result), "Result should be valid UTF-8")
	})
}

// FuzzResolveLogo tests the resolveLogo method with diverse logo input strings
func FuzzResolveLogo(f *testing.F) {
	// Seed corpus with valid and invalid logo inputs
	f.Add("")
	f.Add("example")
	f.Add("Example") // uppercase
	f.Add("EXAMPLE")
	f.Add("GitHub")
	f.Add("GITHUB")
	f.Add("http://example.com/logo.svg")
	f.Add("https://example.com/logo.svg")
	f.Add("data:image/svg+xml;base64,PHN2Zw==")
	f.Add("invalid")
	f.Add("unicode🔥")
	f.Add("\x00\x01\x02")        // null bytes
	f.Add("../../../etc/passwd") // path traversal
	f.Add("javascript:alert(1)") // XSS attempt
	f.Add("\n\r\t")              // control characters
	f.Add("very long logo name that might cause issues")
	f.Add("http://")  // incomplete URL
	f.Add("https://") // incomplete URL
	f.Add("data:")
	f.Add("ftp://example.com/logo.svg") // non-http protocol
	f.Add("file:///path/to/logo.svg")
	f.Add("go github")                      // multiple logos
	f.Add("1234567890")                     // numeric string
	f.Add("!@#$%^&*()")                     // special characters
	f.Add("\U0001f600\U0001f601\U0001f602") // emoji

	f.Fuzz(func(t *testing.T, logo string) {
		generator := New()

		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("resolveLogo panicked with logo=%q: %v", logo, r)
			}
		}()

		result := generator.resolveLogo(logo)

		// Validate output (can be empty string for invalid logos)
		assert.IsType(t, "", result, "Should return string")

		// Test specific logo mappings
		lowerLogo := strings.ToLower(logo)
		switch lowerLogo {
		case "example":
			assert.True(t, strings.HasPrefix(result, "data:image/svg+xml;base64,"), "example logo should return data URI")
			assert.NotEmpty(t, result, "example logo should not be empty")
		case "":
			assert.Empty(t, result, "empty logo should return empty string")
		default:
			if strings.HasPrefix(logo, "http") || strings.HasPrefix(logo, "data:") {
				assert.Equal(t, logo, result, "valid URL/data URI should be returned as-is")
			} else if isValidSimpleIconName(strings.ToLower(logo)) {
				// Valid Simple Icons name should return a CDN URL
				expectedURL := fmt.Sprintf("https://cdn.simpleicons.org/%s", strings.ToLower(logo))
				assert.Equal(t, expectedURL, result, "valid Simple Icons name should return CDN URL")
			} else {
				assert.Empty(t, result, "invalid logo should return empty string")
			}
		}

		// If result is not empty, ensure it's valid UTF-8
		// For pass-through cases (URLs and data URIs), we accept the input as-is
		// Only check UTF-8 validity for our predefined logos
		if result != "" {
			lowerLogo := strings.ToLower(logo)
			if lowerLogo == "example" {
				// These are our built-in logos, should always be valid UTF-8
				assert.True(t, utf8.ValidString(result), "Built-in logo result should be valid UTF-8")
			}
			// For pass-through cases, we don't enforce UTF-8 validity as they might contain binary data
		}
	})
}

// FuzzGenerateBadge tests the Generate method with diverse percentage inputs and options
func FuzzGenerateBadge(f *testing.F) {
	// Seed corpus with various percentages
	f.Add(0.0)
	f.Add(50.0)
	f.Add(75.0)
	f.Add(85.0)
	f.Add(90.0)
	f.Add(95.0)
	f.Add(100.0)
	f.Add(-1.0)
	f.Add(101.0)
	f.Add(math.NaN())
	f.Add(math.Inf(1))
	f.Add(math.Inf(-1))
	f.Add(math.MaxFloat64)
	f.Add(-math.MaxFloat64)

	f.Fuzz(func(t *testing.T, percentage float64) {
		generator := New()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Generate panicked with percentage=%f: %v", percentage, r)
			}
		}()

		result, err := generator.Generate(ctx, percentage)

		// Should not return error for any percentage value
		require.NoError(t, err, "Generate should not return error")

		// Validate SVG output
		assert.NotEmpty(t, result, "Should return non-empty SVG")
		svgString := string(result)
		assert.True(t, strings.HasPrefix(svgString, "<svg"), "Should start with <svg tag")
		assert.True(t, strings.HasSuffix(svgString, "</svg>"), "Should end with </svg tag")
		assert.Contains(t, svgString, "coverage", "Should contain coverage label")

		// Validate percentage formatting in SVG
		if !math.IsNaN(percentage) && !math.IsInf(percentage, 0) {
			// For normal float values, check percentage formatting
			if percentage >= -1000000 && percentage <= 1000000 {
				// Only check formatting for reasonable values
				percentageStr := strings.Contains(svgString, "%")
				assert.True(t, percentageStr, "Should contain percentage symbol")
			}
		}

		// Ensure SVG is valid UTF-8
		assert.True(t, utf8.ValidString(svgString), "SVG should be valid UTF-8")

		// Validate SVG structure
		assert.Contains(t, svgString, "width=", "Should contain width attribute")
		assert.Contains(t, svgString, "height=", "Should contain height attribute")
		assert.Contains(t, svgString, "role=\"img\"", "Should contain accessibility attributes")
	})
}

// FuzzGenerateBadgeWithOptions tests Generate with various option combinations
func FuzzGenerateBadgeWithOptions(f *testing.F) {
	// Seed corpus with different option combinations
	f.Add(75.0, "flat", "coverage", "", "white")
	f.Add(85.0, "flat-square", "test", "example", "blue")
	f.Add(50.0, "invalid-style", "cov", "invalid-logo", "green")
	f.Add(0.0, "", "", "", "")
	f.Add(100.0, "unicode🔥", "unicode🔥", "unicode🔥", "unicode🔥")
	f.Add(math.NaN(), "flat", "coverage", "example", "white")

	f.Fuzz(func(t *testing.T, percentage float64, style, label, logo, logoColor string) {
		generator := New()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Create options
		options := []Option{}
		if style != "" {
			options = append(options, WithStyle(style))
		}
		if label != "" {
			options = append(options, WithLabel(label))
		}
		if logo != "" {
			options = append(options, WithLogo(logo))
		}
		if logoColor != "" {
			options = append(options, WithLogoColor(logoColor))
		}

		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Generate with options panicked with inputs percentage=%f, style=%q, label=%q, logo=%q, logoColor=%q: %v", percentage, style, label, logo, logoColor, r)
			}
		}()

		result, err := generator.Generate(ctx, percentage, options...)

		// Should not return error
		require.NoError(t, err, "Generate with options should not return error")

		// Validate SVG output
		assert.NotEmpty(t, result, "Should return non-empty SVG")
		svgString := string(result)
		assert.True(t, strings.HasPrefix(svgString, "<svg"), "Should start with <svg tag")
		assert.True(t, strings.HasSuffix(svgString, "</svg>"), "Should end with </svg tag")

		// Check if custom label is used (if provided and valid UTF-8)
		if label != "" && utf8.ValidString(label) {
			// For "for-the-badge" style, labels are converted to uppercase
			if style == "for-the-badge" {
				assert.Contains(t, svgString, strings.ToUpper(label), "Should contain custom label in uppercase for for-the-badge style")
			} else {
				assert.Contains(t, svgString, label, "Should contain custom label")
			}
		}

		// Ensure SVG is valid UTF-8
		assert.True(t, utf8.ValidString(svgString), "SVG should be valid UTF-8")
	})
}

// FuzzGenerateTrendBadge tests the GenerateTrendBadge method with diverse current/previous percentage combinations
func FuzzGenerateTrendBadge(f *testing.F) {
	// Seed corpus with various trend scenarios
	f.Add(75.0, 70.0) // upward trend
	f.Add(70.0, 75.0) // downward trend
	f.Add(75.0, 75.0) // stable
	f.Add(75.1, 75.0) // small increase
	f.Add(75.0, 75.1) // small decrease
	f.Add(80.0, 70.0) // large increase
	f.Add(70.0, 80.0) // large decrease
	f.Add(0.0, 0.0)
	f.Add(100.0, 100.0)
	f.Add(-10.0, 10.0)
	f.Add(10.0, -10.0)
	f.Add(math.NaN(), 75.0)
	f.Add(75.0, math.NaN())
	f.Add(math.NaN(), math.NaN())
	f.Add(math.Inf(1), 75.0)
	f.Add(75.0, math.Inf(1))
	f.Add(math.Inf(1), math.Inf(1))
	f.Add(math.Inf(-1), 75.0)
	f.Add(75.0, math.Inf(-1))
	f.Add(math.MaxFloat64, 75.0)
	f.Add(75.0, math.MaxFloat64)

	f.Fuzz(func(t *testing.T, current, previous float64) {
		generator := New()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("GenerateTrendBadge panicked with current=%f, previous=%f: %v", current, previous, r)
			}
		}()

		result, err := generator.GenerateTrendBadge(ctx, current, previous)

		// Should not return error
		require.NoError(t, err, "GenerateTrendBadge should not return error")

		// Validate SVG output
		assert.NotEmpty(t, result, "Should return non-empty SVG")
		svgString := string(result)
		assert.True(t, strings.HasPrefix(svgString, "<svg"), "Should start with <svg tag")
		assert.True(t, strings.HasSuffix(svgString, "</svg>"), "Should end with </svg tag")
		assert.Contains(t, svgString, "trend", "Should contain trend label")

		// Check trend indicators (only for normal float values)
		if !math.IsNaN(current) && !math.IsNaN(previous) &&
			!math.IsInf(current, 0) && !math.IsInf(previous, 0) {
			diff := current - previous
			if math.Abs(diff) < math.MaxFloat64 { // Avoid overflow
				if diff > 0.1 {
					assert.Contains(t, svgString, "↑", "Should contain up arrow for positive trend")
				} else if diff < -0.1 {
					assert.Contains(t, svgString, "↓", "Should contain down arrow for negative trend")
				} else {
					assert.Contains(t, svgString, "→", "Should contain stable arrow for stable trend")
				}
			}
		}

		// Ensure SVG is valid UTF-8
		assert.True(t, utf8.ValidString(svgString), "SVG should be valid UTF-8")
	})
}

// FuzzCalculateTextWidth tests the calculateTextWidth helper method
func FuzzCalculateTextWidth(f *testing.F) {
	// Seed corpus with various text inputs
	f.Add("")
	f.Add("coverage")
	f.Add("95.5%")
	f.Add("a")
	f.Add("very long text that might cause width calculation issues")
	f.Add("unicode🔥text")
	f.Add("\x00\x01\x02")
	f.Add("../../../etc/passwd")
	f.Add("javascript:alert(1)")
	f.Add("\n\r\t")
	f.Add("1234567890")
	f.Add("!@#$%^&*()")
	f.Add("    ")                           // spaces
	f.Add("\U0001f600\U0001f601\U0001f602") // emoji
	f.Add(strings.Repeat("x", 1000))        // very long string
	f.Add(strings.Repeat("🔥", 100))         // long unicode string

	f.Fuzz(func(t *testing.T, text string) {
		generator := New()

		// Function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("calculateTextWidth panicked with text=%q: %v", text, r)
			}
		}()

		result := generator.calculateTextWidth(text)

		// Validate result
		assert.GreaterOrEqual(t, result, 0, "Width should be non-negative")
		assert.IsType(t, 0, result, "Should return int")

		// For empty string, width should be 0
		if text == "" {
			assert.Equal(t, 0, result, "Empty string should have zero width")
		}

		// Width should be proportional to string length (rough estimate)
		if len(text) > 0 {
			assert.Positive(t, result, "Non-empty string should have positive width")
		}

		// Width should not be unreasonably large (only check for non-empty strings)
		if utf8.ValidString(text) && len(text) > 0 && len(text) < 10000 {
			assert.Less(t, result, len(text)*20, "Width should be reasonable compared to string length")
		}
	})
}
