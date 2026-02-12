package badge

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	generator := New()
	assert.NotNil(t, generator)
	assert.NotNil(t, generator.config)
	assert.Equal(t, "flat", generator.config.Style)
	assert.Equal(t, "coverage", generator.config.Label)
	assert.InDelta(t, 95.0, generator.config.ThresholdConfig.Excellent, 0.001)
	assert.InDelta(t, 85.0, generator.config.ThresholdConfig.Good, 0.001)
	assert.InDelta(t, 75.0, generator.config.ThresholdConfig.Acceptable, 0.001)
	assert.InDelta(t, 60.0, generator.config.ThresholdConfig.Low, 0.001)
}

func TestNewWithConfig(t *testing.T) {
	config := &Config{
		Style: "flat-square",
		Label: "test",
		ThresholdConfig: ThresholdConfig{
			Excellent:  95.0,
			Good:       85.0,
			Acceptable: 75.0,
			Low:        65.0,
		},
	}

	generator := NewWithConfig(config)
	assert.Equal(t, config, generator.config)
}

func TestGenerate(t *testing.T) {
	generator := New()
	ctx := context.Background()

	svg, err := generator.Generate(ctx, 85.5)
	require.NoError(t, err)
	assert.NotEmpty(t, svg)

	svgStr := string(svg)
	assert.Contains(t, svgStr, "<svg")
	assert.Contains(t, svgStr, "85.5%")
	assert.Contains(t, svgStr, "coverage")
	assert.Contains(t, svgStr, "</svg>")
	assert.Contains(t, svgStr, `role="img"`)
	assert.Contains(t, svgStr, `aria-label="Code coverage: 85.5 percent"`)
}

func TestGenerateWithOptions(t *testing.T) {
	generator := New()
	ctx := context.Background()

	svg, err := generator.Generate(ctx, 92.3,
		WithStyle("flat-square"),
		WithLabel("test coverage"),
		WithLogo("https://example.com/logo.svg"),
	)
	require.NoError(t, err)

	svgStr := string(svg)
	assert.Contains(t, svgStr, "92.3%")
	assert.Contains(t, svgStr, "test coverage")
	assert.Contains(t, svgStr, "https://example.com/logo.svg")
	assert.Contains(t, svgStr, `shape-rendering="crispEdges"`) // flat-square style
}

func TestGenerateTrendBadge(t *testing.T) {
	generator := New()
	ctx := context.Background()

	tests := []struct {
		name     string
		current  float64
		previous float64
		expected string
	}{
		{
			name:     "upward trend",
			current:  85.0,
			previous: 80.0,
			expected: "↑ +5.0%",
		},
		{
			name:     "downward trend",
			current:  75.0,
			previous: 82.0,
			expected: "↓ -7.0%",
		},
		{
			name:     "stable trend",
			current:  80.0,
			previous: 80.0,
			expected: "→ stable",
		},
		{
			name:     "small increase",
			current:  80.05,
			previous: 80.0,
			expected: "→ stable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := generator.GenerateTrendBadge(ctx, tt.current, tt.previous)
			require.NoError(t, err)

			svgStr := string(svg)
			assert.Contains(t, svgStr, tt.expected)
			assert.Contains(t, svgStr, "trend")
		})
	}
}

func TestGetColorForPercentage(t *testing.T) {
	generator := New()

	tests := []struct {
		percentage float64
		expected   string
	}{
		{96.0, "#28a745"},  // excellent (bright green)
		{87.0, "#3fb950"},  // good (green)
		{77.0, "#ffc107"},  // acceptable (yellow)
		{67.0, "#fd7e14"},  // low (orange)
		{55.0, "#dc3545"},  // poor (red)
		{100.0, "#28a745"}, // perfect (bright green)
		{0.0, "#dc3545"},   // zero (red)
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.1f%%", tt.percentage), func(t *testing.T) {
			color := generator.getColorForPercentage(tt.percentage)
			assert.Equal(t, tt.expected, color)
		})
	}
}

func TestGetColorByName(t *testing.T) {
	generator := New()

	tests := []struct {
		name     string
		expected string
	}{
		{"excellent", "#28a745"},
		{"good", "#3fb950"},
		{"acceptable", "#ffc107"},
		{"low", "#fd7e14"},
		{"poor", "#dc3545"},
		{"unknown", "#8b949e"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := generator.getColorByName(tt.name)
			assert.Equal(t, tt.expected, color)
		})
	}
}

func TestRenderSVGStyles(t *testing.T) {
	generator := New()
	ctx := context.Background()

	badgeData := Data{
		Label:     "coverage",
		Message:   "85.5%",
		Color:     "#3fb950",
		AriaLabel: "Code coverage: 85.5 percent",
	}

	tests := []struct {
		style        string
		expectedText string
	}{
		{"flat", `rx="3"`}, // rounded corners
		{"flat-square", `shape-rendering="crispEdges"`}, // crisp edges
		{"for-the-badge", `font-weight="bold"`},         // bold text
	}

	for _, tt := range tests {
		t.Run(tt.style, func(t *testing.T) {
			badgeData.Style = tt.style
			svg, err := generator.renderSVG(ctx, badgeData)
			require.NoError(t, err)

			svgStr := string(svg)
			assert.Contains(t, svgStr, tt.expectedText)
		})
	}
}

func TestCalculateTextWidth(t *testing.T) {
	generator := New()

	tests := []struct {
		text     string
		expected int
	}{
		{"", 0},
		{"a", 7},         // ceil(1 * 6.5) = 7
		{"abc", 20},      // ceil(3 * 6.5) = 20
		{"coverage", 52}, // ceil(8 * 6.5) = 52
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			width := generator.calculateTextWidth(tt.text)
			assert.Equal(t, tt.expected, width)
		})
	}
}

func TestOptions(t *testing.T) {
	opts := &Options{}

	WithStyle("flat-square")(opts)
	assert.Equal(t, "flat-square", opts.Style)

	WithLabel("test")(opts)
	assert.Equal(t, "test", opts.Label)

	WithLogo("logo.svg")(opts)
	assert.Equal(t, "logo.svg", opts.Logo)

	WithLogoColor("blue")(opts)
	assert.Equal(t, "blue", opts.LogoColor)
}

func TestRenderFlatBadge(t *testing.T) {
	generator := New()

	data := Data{
		Label:     "coverage",
		Message:   "85.5%",
		Color:     "#3fb950",
		Style:     "flat",
		AriaLabel: "Code coverage: 85.5 percent",
	}

	svg := generator.renderFlatBadge(data, 100, 50, 40, 0)
	svgStr := string(svg)

	assert.Contains(t, svgStr, `xmlns="http://www.w3.org/2000/svg"`)
	assert.Contains(t, svgStr, `width="100"`)
	assert.Contains(t, svgStr, `height="20"`)
	assert.Contains(t, svgStr, "coverage")
	assert.Contains(t, svgStr, "85.5%")
	assert.Contains(t, svgStr, "#3fb950")
	assert.Contains(t, svgStr, `rx="3"`)         // rounded corners
	assert.Contains(t, svgStr, `linearGradient`) // gradient effect
}

func TestRenderFlatSquareBadge(t *testing.T) {
	generator := New()

	data := Data{
		Label:     "coverage",
		Message:   "85.5%",
		Color:     "#3fb950",
		Style:     "flat-square",
		AriaLabel: "Code coverage: 85.5 percent",
	}

	svg := generator.renderFlatSquareBadge(data, 100, 20, 50, 40, 0)
	svgStr := string(svg)

	assert.Contains(t, svgStr, `xmlns="http://www.w3.org/2000/svg"`)
	assert.Contains(t, svgStr, `shape-rendering="crispEdges"`)
	assert.Contains(t, svgStr, "coverage")
	assert.Contains(t, svgStr, "85.5%")
	assert.Contains(t, svgStr, "#3fb950")
	assert.NotContains(t, svgStr, `rx="3"`) // no rounded corners
}

func TestRenderForTheBadge(t *testing.T) {
	generator := New()

	data := Data{
		Label:     "coverage",
		Message:   "85.5%",
		Color:     "#3fb950",
		Style:     "for-the-badge",
		AriaLabel: "Code coverage: 85.5 percent",
	}

	svg := generator.renderForTheBadge(data, 100, 28, 50, 40, 0)
	svgStr := string(svg)

	assert.Contains(t, svgStr, `xmlns="http://www.w3.org/2000/svg"`)
	assert.Contains(t, svgStr, `font-weight="bold"`)
	assert.Contains(t, svgStr, "COVERAGE") // uppercase
	assert.Contains(t, svgStr, "85.5%")
	assert.Contains(t, svgStr, "#3fb950")
}

func TestRenderWithLogo(t *testing.T) {
	generator := New()

	data := Data{
		Label:     "coverage",
		Message:   "85.5%",
		Color:     "#3fb950",
		Style:     "flat",
		Logo:      "https://example.com/logo.svg",
		AriaLabel: "Code coverage: 85.5 percent",
	}

	svg := generator.renderFlatBadge(data, 120, 50, 40, 16)
	svgStr := string(svg)

	assert.Contains(t, svgStr, `<image`)
	assert.Contains(t, svgStr, `xlink:href="https://example.com/logo.svg"`)
	assert.Contains(t, svgStr, `width="14"`)
	assert.Contains(t, svgStr, `height="14"`)
}

func TestGenerateContextCancellation(t *testing.T) {
	generator := New()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	_, err := generator.Generate(ctx, 85.5)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestGenerateValidSVG(t *testing.T) {
	generator := New()
	ctx := context.Background()

	// Test various coverage percentages
	coverages := []float64{0.0, 25.5, 50.0, 75.3, 100.0}

	for _, coverage := range coverages {
		t.Run(fmt.Sprintf("coverage_%.1f", coverage), func(t *testing.T) {
			svg, err := generator.Generate(ctx, coverage)
			require.NoError(t, err)

			svgStr := string(svg)
			// Basic SVG validation
			assert.True(t, strings.HasPrefix(svgStr, "<svg"))
			assert.True(t, strings.HasSuffix(svgStr, "</svg>"))
			assert.Contains(t, svgStr, fmt.Sprintf("%.1f%%", coverage))

			// Accessibility checks
			assert.Contains(t, svgStr, `role="img"`)
			assert.Contains(t, svgStr, `aria-label=`)
			assert.Contains(t, svgStr, "<title>")
		})
	}
}

func TestGenerateCustomThresholds(t *testing.T) {
	config := &Config{
		Style: "flat",
		Label: "coverage",
		ThresholdConfig: ThresholdConfig{
			Excellent:  95.0,
			Good:       85.0,
			Acceptable: 75.0,
			Low:        65.0,
		},
	}

	generator := NewWithConfig(config)
	ctx := context.Background()

	tests := []struct {
		percentage float64
		expected   string
	}{
		{96.0, "#28a745"}, // excellent
		{87.0, "#3fb950"}, // good
		{77.0, "#ffc107"}, // acceptable
		{67.0, "#fd7e14"}, // low
		{55.0, "#dc3545"}, // poor
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.1f%%", tt.percentage), func(t *testing.T) {
			svg, err := generator.Generate(ctx, tt.percentage)
			require.NoError(t, err)

			svgStr := string(svg)
			assert.Contains(t, svgStr, tt.expected)
		})
	}
}

func TestGenerateEdgeCases(t *testing.T) {
	generator := New()
	ctx := context.Background()

	tests := []struct {
		name       string
		percentage float64
	}{
		{"zero coverage", 0.0},
		{"perfect coverage", 100.0},
		{"high precision", 87.123456},
		{"very low", 0.1},
		{"very high", 99.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := generator.Generate(ctx, tt.percentage)
			require.NoError(t, err)
			assert.NotEmpty(t, svg)

			svgStr := string(svg)
			// Should format to 1 decimal place
			expected := fmt.Sprintf("%.1f%%", tt.percentage)
			assert.Contains(t, svgStr, expected)
		})
	}
}

func TestResolveLogo(t *testing.T) {
	generator := New()

	tests := []struct {
		name     string
		input    string
		expected string
		contains string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "example logo",
			input:    "example",
			contains: "data:image/svg+xml;base64,",
		},
		{
			name:     "case insensitive example",
			input:    "EXAMPLE",
			contains: "data:image/svg+xml;base64,",
		},
		{
			name:     "valid HTTP URL",
			input:    "https://example.com/logo.svg",
			expected: "https://example.com/logo.svg",
		},
		{
			name:     "valid data URI",
			input:    "data:image/svg+xml;base64,PHN2Zw==",
			expected: "data:image/svg+xml;base64,PHN2Zw==",
		},
		{
			name:     "valid simple icon name (becomes base64 or empty on network failure)",
			input:    "invalid-logo",
			expected: "", // Will be empty string if network fetch fails
		},
		{
			name:     "actually invalid logo name",
			input:    "invalid logo with spaces",
			expected: "",
		},
		{
			name:     "relative path - invalid",
			input:    "./logo.svg",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.resolveLogo(context.Background(), tt.input, "")

			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			} else if tt.contains != "" {
				assert.Contains(t, result, tt.contains)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestGenerateWithResolvedLogos(t *testing.T) {
	// Create mock Simple Icons CDN server
	mockServer := createMockSimpleIconsServer(t)
	defer mockServer.Close()

	// Create custom HTTP client that redirects to mock server
	mockClient := &http.Client{
		Transport: &mockTransport{
			mockServerURL: mockServer.URL,
		},
		Timeout: 3 * time.Second,
	}

	// Create generator with injected HTTP client
	generator := NewWithConfig(&Config{
		Style:     "flat",
		Label:     "coverage",
		Logo:      "",
		LogoColor: "white",
		ThresholdConfig: ThresholdConfig{
			Excellent:  95.0,
			Good:       85.0,
			Acceptable: 75.0,
			Low:        60.0,
		},
		HTTPClient: mockClient,
	})

	ctx := context.Background()

	tests := []struct {
		name     string
		logo     string
		hasImage bool
	}{
		{"no logo", "", false},
		{"example logo", "example", true},
		{"valid simple icon", "github", true},
		{"actually invalid logo", "invalid logo name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := generator.Generate(ctx, 85.5, WithLogo(tt.logo))
			require.NoError(t, err)
			assert.NotEmpty(t, svg)

			svgStr := string(svg)
			if tt.hasImage {
				assert.Contains(t, svgStr, "<image")
				// Check for data URI (mocked icons return data URIs)
				assert.Contains(t, svgStr, "xlink:href=\"data:")
			} else {
				assert.NotContains(t, svgStr, "<image")
			}
		})
	}
}

func TestGenerateWithRealSimpleIcons(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test with external CDN in short mode")
	}

	// This test uses the real CDN and is skipped in CI via -short flag
	generator := New()
	ctx := context.Background()

	svg, err := generator.Generate(ctx, 85.5, WithLogo("github"))
	require.NoError(t, err)
	assert.NotEmpty(t, svg)

	svgStr := string(svg)
	assert.Contains(t, svgStr, "<image")
	// Should have either data URI or CDN URL
	hasValidImage := strings.Contains(svgStr, "xlink:href=\"data:") ||
		strings.Contains(svgStr, "xlink:href=\"https://cdn.simpleicons.org")
	assert.True(t, hasValidImage, "Expected valid image reference")
}

func TestProcessLogoColor(t *testing.T) {
	generator := New()

	tests := []struct {
		name         string
		logoURL      string
		color        string
		expectedSVG  string
		shouldModify bool
	}{
		{
			name:         "no color specified",
			logoURL:      "data:image/svg+xml;base64,PHN2ZyBmaWxsPSJjdXJyZW50Q29sb3IiPjwvc3ZnPg==", // <svg fill="currentColor"></svg>
			color:        "",
			shouldModify: false,
		},
		{
			name:         "white color replaces currentColor",
			logoURL:      "data:image/svg+xml;base64,PHN2ZyBmaWxsPSJjdXJyZW50Q29sb3IiPjwvc3ZnPg==",
			color:        "white",
			expectedSVG:  `<svg fill="white"></svg>`,
			shouldModify: true,
		},
		{
			name:         "red color replaces currentColor",
			logoURL:      "data:image/svg+xml;base64,PHN2ZyBmaWxsPSJjdXJyZW50Q29sb3IiPjwvc3ZnPg==", // <svg fill="currentColor"></svg>
			color:        "red",
			expectedSVG:  `<svg fill="red"></svg>`,
			shouldModify: true,
		},
		{
			name:         "non-data URI logo",
			logoURL:      "https://example.com/logo.svg",
			color:        "blue",
			shouldModify: false,
		},
		{
			name:         "simple icons CDN URL",
			logoURL:      "https://cdn.simpleicons.org/nodejs",
			color:        "green",
			shouldModify: false, // Not implemented yet
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.processLogoColor(tt.logoURL, tt.color)

			if !tt.shouldModify {
				assert.Equal(t, tt.logoURL, result, "Logo should not be modified")
			} else {
				assert.NotEqual(t, tt.logoURL, result, "Logo should be modified")

				// Decode and check the SVG content
				if strings.HasPrefix(result, "data:image/svg+xml;base64,") {
					base64Content := strings.TrimPrefix(result, "data:image/svg+xml;base64,")
					svgBytes, err := base64.StdEncoding.DecodeString(base64Content)
					require.NoError(t, err)

					svgContent := string(svgBytes)
					assert.Equal(t, tt.expectedSVG, svgContent)
					assert.NotContains(t, svgContent, "currentColor", "currentColor should be replaced")
				}
			}
		})
	}
}

// TestApplySVGColor tests the new applySVGColor method that handles various SVG coloring scenarios
func TestApplySVGColor(t *testing.T) {
	generator := New()

	tests := []struct {
		name        string
		svgContent  string
		color       string
		expectedSVG string
	}{
		{
			name:        "replace currentColor",
			svgContent:  `<svg fill="currentColor"><path d="M0 0"/></svg>`,
			color:       "white",
			expectedSVG: `<svg fill="white"><path d="M0 0"/></svg>`,
		},
		{
			name:        "replace 2FAS default red color",
			svgContent:  `<svg fill="#EC1C24"><path d="M0 0"/></svg>`,
			color:       "white",
			expectedSVG: `<svg fill="white"><path d="M0 0"/></svg>`,
		},
		{
			name:        "add fill attribute when missing",
			svgContent:  `<svg role="img" viewBox="0 0 24 24"><path d="M0 0"/></svg>`,
			color:       "white",
			expectedSVG: `<svg role="img" viewBox="0 0 24 24" fill="white"><path d="M0 0"/></svg>`,
		},
		{
			name:        "handle real 2FAS SVG without fill",
			svgContent:  `<svg role="img" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><title>2FAS</title><path d="M12 0c-.918 0"/></svg>`,
			color:       "white",
			expectedSVG: `<svg role="img" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" fill="white"><title>2FAS</title><path d="M12 0c-.918 0"/></svg>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.applySVGColor(tt.svgContent, tt.color)
			assert.Equal(t, tt.expectedSVG, result)
		})
	}
}

// TestApplySVGColorStatic tests the static version used in fetchSimpleIcon
func TestApplySVGColorStatic(t *testing.T) {
	tests := []struct {
		name        string
		svgContent  string
		color       string
		expectedSVG string
	}{
		{
			name:        "replace currentColor",
			svgContent:  `<svg fill="currentColor"><path d="M0 0"/></svg>`,
			color:       "white",
			expectedSVG: `<svg fill="white"><path d="M0 0"/></svg>`,
		},
		{
			name:        "add fill attribute when missing",
			svgContent:  `<svg role="img" viewBox="0 0 24 24"><path d="M0 0"/></svg>`,
			color:       "white",
			expectedSVG: `<svg role="img" viewBox="0 0 24 24" fill="white"><path d="M0 0"/></svg>`,
		},
		{
			name:        "method and static versions are identical",
			svgContent:  `<svg><path fill="currentColor"/></svg>`,
			color:       "red",
			expectedSVG: `<svg><path fill="red"/></svg>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applySVGColorStatic(tt.svgContent, tt.color)
			assert.Equal(t, tt.expectedSVG, result)
		})
	}
}

// TestProcessLogoColorWithNewFunctions tests the updated processLogoColor method
func TestProcessLogoColorWithNewFunctions(t *testing.T) {
	generator := New()

	tests := []struct {
		name         string
		logoURL      string
		color        string
		expectedSVG  string
		shouldModify bool
	}{
		{
			name:         "process SVG without fill attribute",
			logoURL:      "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(`<svg viewBox="0 0 24 24"><path d="M0 0"/></svg>`)),
			color:        "white",
			expectedSVG:  `<svg viewBox="0 0 24 24" fill="white"><path d="M0 0"/></svg>`,
			shouldModify: true,
		},
		{
			name:         "process SVG with 2FAS red fill",
			logoURL:      "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(`<svg fill="#EC1C24"><path d="M0 0"/></svg>`)),
			color:        "white",
			expectedSVG:  `<svg fill="white"><path d="M0 0"/></svg>`,
			shouldModify: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.processLogoColor(tt.logoURL, tt.color)

			if !tt.shouldModify {
				assert.Equal(t, tt.logoURL, result, "Logo should not be modified")
			} else {
				assert.NotEqual(t, tt.logoURL, result, "Logo should be modified")

				// Decode and check the SVG content
				if strings.HasPrefix(result, "data:image/svg+xml;base64,") {
					base64Content := strings.TrimPrefix(result, "data:image/svg+xml;base64,")
					svgBytes, err := base64.StdEncoding.DecodeString(base64Content)
					require.NoError(t, err)

					svgContent := string(svgBytes)
					assert.Equal(t, tt.expectedSVG, svgContent)
				}
			}
		})
	}
}

// mockTransport rewrites Simple Icons CDN URLs to point to mock server
type mockTransport struct {
	mockServerURL string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite cdn.simpleicons.org URLs to mock server
	if strings.Contains(req.URL.Host, "cdn.simpleicons.org") ||
		strings.Contains(req.URL.Host, "raw.githubusercontent.com") {
		// Preserve the path but change the host
		mockURL := m.mockServerURL + req.URL.Path
		newReq, err := http.NewRequestWithContext(req.Context(), req.Method, mockURL, req.Body)
		if err != nil {
			return nil, err
		}
		// Copy headers
		newReq.Header = req.Header
		req = newReq
	}
	return http.DefaultTransport.RoundTrip(req)
}

// createMockSimpleIconsServer creates an httptest server that mocks cdn.simpleicons.org
func createMockSimpleIconsServer(t *testing.T) *httptest.Server {
	t.Helper()

	// Mock SVG content for GitHub icon
	githubSVG := `<svg role="img" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
		<title>GitHub</title>
		<path fill="currentColor" d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/>
	</svg>`

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check User-Agent header
		if r.Header.Get("User-Agent") == "" {
			http.Error(w, "User-Agent required", http.StatusForbidden)
			return
		}

		// Handle different icon requests
		switch {
		case strings.Contains(r.URL.Path, "/github"):
			w.Header().Set("Content-Type", "image/svg+xml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(githubSVG))
		default:
			http.Error(w, "Icon not found", http.StatusNotFound)
		}
	}))
}
