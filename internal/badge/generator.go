// Package badge generates SVG coverage badges
package badge

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

// ErrIconFetchFailed is returned when fetching an icon from Simple Icons CDN fails
var ErrIconFetchFailed = errors.New("failed to fetch icon")

// Generator creates professional SVG badges matching GitHub's design language
type Generator struct {
	config *Config
}

// Config holds badge generation configuration
type Config struct {
	Style           string
	Label           string
	Logo            string
	LogoColor       string
	ThresholdConfig ThresholdConfig
}

// ThresholdConfig defines coverage thresholds for color coding
type ThresholdConfig struct {
	Excellent  float64 // 90%+ - green
	Good       float64 // 80%+ - blue
	Acceptable float64 // 60%+ - yellow/warning
	Low        float64 // Below 60% - red
	// Low is not used as a threshold, anything below Acceptable is red
}

// Data represents data needed to generate a badge
type Data struct {
	Label     string
	Message   string
	Color     string
	Style     string
	Logo      string
	LogoColor string
	AriaLabel string
}

// TrendDirection represents coverage trend
type TrendDirection int

const (
	// TrendUp indicates coverage is trending upward
	TrendUp TrendDirection = iota
	// TrendDown indicates coverage is trending downward
	TrendDown
	// TrendStable indicates coverage is stable
	TrendStable
)

// New creates a new badge generator with default configuration
func New() *Generator {
	return &Generator{
		config: &Config{
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
		},
	}
}

// NewWithConfig creates a new badge generator with custom configuration
func NewWithConfig(config *Config) *Generator {
	return &Generator{config: config}
}

// sanitizeUTF8 ensures the string is valid UTF-8, replacing invalid sequences
func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	// Replace invalid UTF-8 sequences with replacement character
	return strings.ToValidUTF8(s, "�")
}

// Generate creates an SVG badge for the given coverage percentage
func (g *Generator) Generate(ctx context.Context, percentage float64, options ...Option) ([]byte, error) {
	opts := &Options{
		Style:     g.config.Style,
		Label:     g.config.Label,
		Logo:      g.config.Logo,
		LogoColor: g.config.LogoColor,
	}

	// Apply options
	for _, opt := range options {
		opt(opts)
	}

	color := g.getColorForPercentage(percentage)
	message := fmt.Sprintf("%.1f%%", percentage)

	badgeData := Data{
		Label:     sanitizeUTF8(opts.Label),
		Message:   message,
		Color:     color,
		Style:     sanitizeUTF8(opts.Style),
		Logo:      g.resolveLogo(ctx, opts.Logo, sanitizeUTF8(opts.LogoColor)),
		LogoColor: sanitizeUTF8(opts.LogoColor),
		AriaLabel: fmt.Sprintf("Code coverage: %.1f percent", percentage),
	}

	return g.renderSVG(ctx, badgeData)
}

// GenerateTrendBadge creates a badge showing coverage trend
func (g *Generator) GenerateTrendBadge(ctx context.Context, current, previous float64, options ...Option) ([]byte, error) {
	diff := current - previous
	var trend string
	var color string

	switch {
	case diff > 0.1:
		trend = fmt.Sprintf("↑ +%.1f%%", diff)
		color = g.getColorByName("excellent")
	case diff < -0.1:
		trend = fmt.Sprintf("↓ %.1f%%", diff)
		color = g.getColorByName("low")
	default:
		trend = "→ stable"
		color = "#8b949e" // neutral gray
	}

	opts := &Options{
		Style: g.config.Style,
		Label: "trend",
	}

	for _, opt := range options {
		opt(opts)
	}

	badgeData := Data{
		Label:     sanitizeUTF8(opts.Label),
		Message:   trend,
		Color:     color,
		Style:     sanitizeUTF8(opts.Style),
		Logo:      g.resolveLogo(ctx, opts.Logo, sanitizeUTF8(opts.LogoColor)),
		LogoColor: sanitizeUTF8(opts.LogoColor),
		AriaLabel: fmt.Sprintf("Coverage trend: %s", trend),
	}

	return g.renderSVG(ctx, badgeData)
}

// getColorForPercentage returns the appropriate color based on coverage percentage
func (g *Generator) getColorForPercentage(percentage float64) string {
	switch {
	case percentage >= g.config.ThresholdConfig.Excellent:
		return "#28a745" // Bright green (excellent coverage 95%+)
	case percentage >= g.config.ThresholdConfig.Good:
		return "#3fb950" // Green (good coverage 85-94%)
	case percentage >= g.config.ThresholdConfig.Acceptable:
		return "#ffc107" // Yellow (acceptable coverage 75-84%)
	case percentage >= g.config.ThresholdConfig.Low:
		return "#fd7e14" // Orange (low coverage 65-74%)
	default:
		return "#dc3545" // Red (poor coverage below 65%)
	}
}

// getColorByName returns color by threshold name
func (g *Generator) getColorByName(name string) string {
	switch name {
	case "excellent":
		return "#28a745" // Bright green
	case "good":
		return "#3fb950" // Green
	case "acceptable":
		return "#ffc107" // Yellow
	case "low":
		return "#fd7e14" // Orange
	case "poor":
		return "#dc3545" // Red
	default:
		return "#8b949e" // neutral gray
	}
}

// resolveLogo converts common logo names to SVG data URIs or URLs
func (g *Generator) resolveLogo(ctx context.Context, logo, color string) string {
	switch strings.ToLower(logo) {
	case "example":
		// Example logo - simple star icon as SVG data URI for testing/documentation purposes
		return `data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxNCIgaGVpZ2h0PSIxNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJjdXJyZW50Q29sb3IiPjxwYXRoIGQ9Ik0xMiAyTDE1LjA5IDguMjZMMjIgOS4yN0wxNyAxNC4xNEwxOC4xOCAyMUwxMiAxNy43N0w1LjgyIDIxTDcgMTQuMTRMMiA5LjI3TDguOTEgOC4yNkwxMiAyWiIvPjwvc3ZnPg==`
	case "":
		// Empty string means no logo
		return ""
	default:
		// If it starts with http or data:, assume it's a valid URL/data URI
		if strings.HasPrefix(logo, "http") || strings.HasPrefix(logo, "data:") {
			return logo
		}
		// Check if it's a potentially valid Simple Icons logo name
		// We use conservative validation to avoid obviously invalid names,
		// but trust the Simple Icons CDN to handle requests for non-existent logos gracefully
		logoName := strings.ToLower(logo)
		if isValidSimpleIconName(logoName) {
			// First attempt: Fetch the icon with color (if specified)
			if dataURI, err := fetchSimpleIcon(ctx, logoName, color); err == nil {
				return dataURI
			} else {
				log.Printf("Warning: Failed to fetch logo '%s' with color '%s': %v", logoName, color, err)
			}

			// Fallback attempt: Try fetching without color if the first attempt failed and color was specified
			if color != "" {
				log.Printf("Retrying logo '%s' without color...", logoName)
				if dataURI, err := fetchSimpleIcon(ctx, logoName, ""); err == nil {
					log.Printf("Success: Fetched logo '%s' without color", logoName)
					return dataURI
				} else {
					log.Printf("Error: Failed to fetch logo '%s' even without color: %v", logoName, err)
				}
			}

			// If all attempts fail, log the failure and return empty string
			log.Printf("Error: Unable to fetch logo '%s' from Simple Icons CDN after all attempts", logoName)
			return ""
		}
		// Log invalid logo names for debugging
		log.Printf("Warning: Invalid logo name '%s' - must contain only lowercase letters, numbers, and hyphens", logo)
		return ""
	}
}

// isValidSimpleIconName checks if a logo name is valid for Simple Icons
// Simple Icons uses lowercase letters, numbers, and hyphens only
func isValidSimpleIconName(name string) bool {
	// Must not be empty and should be reasonable length
	if len(name) == 0 || len(name) > 50 {
		return false
	}

	// Simple Icons naming convention: lowercase letters, numbers, hyphens only
	for _, r := range name {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return false
		}
	}

	// Must start with a letter or number (some icons like "2fas" start with numbers)
	firstChar := name[0]
	return (firstChar >= 'a' && firstChar <= 'z') || (firstChar >= '0' && firstChar <= '9')
}

// processLogoColor applies color to logos that use currentColor
func (g *Generator) processLogoColor(logoURL, color string) string {
	// Skip processing if no color specified
	if color == "" {
		return logoURL
	}

	// Only process data URIs with base64 content
	if !strings.HasPrefix(logoURL, "data:image/svg+xml;base64,") {
		// For Simple Icons CDN URLs, we need to fetch and modify
		if strings.Contains(logoURL, "simpleicons.org") {
			// Simple Icons CDN URLs should already have color applied during resolveLogo
			// No further processing needed
			return logoURL
		}
		return logoURL
	}

	// Extract and decode base64 content
	base64Content := strings.TrimPrefix(logoURL, "data:image/svg+xml;base64,")
	svgBytes, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		return logoURL // Return original if decode fails
	}

	svgContent := string(svgBytes)

	// Replace currentColor with the specified color
	modifiedSVG := strings.ReplaceAll(svgContent, "currentColor", color)

	// Re-encode to base64
	newBase64 := base64.StdEncoding.EncodeToString([]byte(modifiedSVG))
	return "data:image/svg+xml;base64," + newBase64
}

// fetchSimpleIcon fetches an SVG icon from Simple Icons CDN with retry logic and returns it as a base64 data URI
func fetchSimpleIcon(ctx context.Context, iconName, color string) (string, error) {
	// Build the URL for Simple Icons CDN
	var url string
	if color != "" {
		// Remove # from hex colors for Simple Icons CDN
		cleanColor := strings.TrimPrefix(color, "#")
		url = fmt.Sprintf("https://cdn.simpleicons.org/%s/%s", iconName, cleanColor)
	} else {
		url = fmt.Sprintf("https://cdn.simpleicons.org/%s", iconName)
	}

	// Retry configuration
	const maxRetries = 3
	const baseDelay = 500 * time.Millisecond

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check if context was canceled
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 15 * time.Second, // Increased timeout for slower networks
		}

		// Create request with context
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request for %s: %w", url, err)
			continue
		}

		// Add User-Agent header (some CDNs require this)
		req.Header.Set("User-Agent", "go-coverage/1.0 (+https://github.com/mrz1836/go-coverage)")

		// Make the request
		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to fetch icon from %s (attempt %d/%d): %w", url, attempt+1, maxRetries, err)
			// Wait before retry with exponential backoff
			if attempt < maxRetries-1 {
				delay := time.Duration(1<<uint(attempt)) * baseDelay
				time.Sleep(delay)
			}
			continue
		}

		// Check response status
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("%w: HTTP %d from %s (attempt %d/%d)", ErrIconFetchFailed, resp.StatusCode, url, attempt+1, maxRetries)
			// Wait before retry with exponential backoff
			if attempt < maxRetries-1 {
				delay := time.Duration(1<<uint(attempt)) * baseDelay
				time.Sleep(delay)
			}
			continue
		}

		// Read the SVG content
		svgContent, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read SVG content from %s (attempt %d/%d): %w", url, attempt+1, maxRetries, err)
			// Wait before retry with exponential backoff
			if attempt < maxRetries-1 {
				delay := time.Duration(1<<uint(attempt)) * baseDelay
				time.Sleep(delay)
			}
			continue
		}

		// Success! Encode as base64 data URI
		base64Content := base64.StdEncoding.EncodeToString(svgContent)
		return "data:image/svg+xml;base64," + base64Content, nil
	}

	// All retries failed
	return "", fmt.Errorf("failed to fetch icon after %d attempts: %w", maxRetries, lastErr)
}

// renderSVG generates the actual SVG content
func (g *Generator) renderSVG(ctx context.Context, data Data) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Calculate dimensions
	labelWidth := g.calculateTextWidth(data.Label)
	messageWidth := g.calculateTextWidth(data.Message)
	logoWidth := 0
	if data.Logo != "" {
		logoWidth = 16 // Standard logo width
	}

	totalWidth := labelWidth + messageWidth + logoWidth + 28 // padding (extra space in percentage section)
	height := 20

	// Generate SVG based on style
	switch data.Style {
	case "flat-square":
		return g.renderFlatSquareBadge(data, totalWidth, height, labelWidth, messageWidth, logoWidth), nil
	case "for-the-badge":
		return g.renderForTheBadge(data, totalWidth, height+8, labelWidth, messageWidth, logoWidth), nil
	default: // flat
		return g.renderFlatBadge(data, totalWidth, labelWidth, messageWidth, logoWidth), nil
	}
}

// renderFlatBadge generates a flat-style badge
func (g *Generator) renderFlatBadge(data Data, width, labelWidth, messageWidth, logoWidth int) []byte {
	height := 20
	template := `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%d" height="%d" role="img" aria-label="%s">
  <title>%s</title>
  <linearGradient id="s" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <clipPath id="r">
    <rect width="%d" height="%d" rx="3" fill="#fff"/>
  </clipPath>
  <g clip-path="url(#r)">
    <rect width="%d" height="%d" fill="#555"/>
    <rect x="%d" width="%d" height="%d" fill="%s"/>
    <rect width="%d" height="%d" fill="url(#s)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="11">
    %s
    <text aria-hidden="true" x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
    <text aria-hidden="true" x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
  </g>
</svg>`

	labelX := logoWidth + labelWidth/2 + 6
	messageX := logoWidth + labelWidth + messageWidth/2 + 16
	logoSvg := ""

	if data.Logo != "" {
		// Process logo to apply color (for currentColor logos)
		processedLogo := g.processLogoColor(data.Logo, data.LogoColor)
		logoSvg = fmt.Sprintf(`<image x="5" y="3" width="14" height="14" xlink:href="%s"/>`, processedLogo)
	}

	return []byte(fmt.Sprintf(template,
		width, height, data.AriaLabel, data.AriaLabel,
		width, height,
		logoWidth+labelWidth+8, height,
		logoWidth+labelWidth+8, messageWidth+20, height, data.Color,
		width, height,
		logoSvg,
		labelX, data.Label,
		labelX, data.Label,
		messageX, data.Message,
		messageX, data.Message,
	))
}

// renderFlatSquareBadge generates a flat-square style badge
func (g *Generator) renderFlatSquareBadge(data Data, width, height, labelWidth, messageWidth, logoWidth int) []byte {
	template := `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%d" height="%d" role="img" aria-label="%s">
  <title>%s</title>
  <g shape-rendering="crispEdges">
    <rect width="%d" height="%d" fill="#555"/>
    <rect x="%d" width="%d" height="%d" fill="%s"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="11">
    %s
    <text x="%d" y="15">%s</text>
    <text x="%d" y="15">%s</text>
  </g>
</svg>`

	labelX := logoWidth + labelWidth/2 + 6
	messageX := logoWidth + labelWidth + messageWidth/2 + 16
	logoSvg := ""

	if data.Logo != "" {
		// Process logo to apply color (for currentColor logos)
		processedLogo := g.processLogoColor(data.Logo, data.LogoColor)
		logoSvg = fmt.Sprintf(`<image x="5" y="3" width="14" height="14" xlink:href="%s"/>`, processedLogo)
	}

	return []byte(fmt.Sprintf(template,
		width, height, data.AriaLabel, data.AriaLabel,
		logoWidth+labelWidth+8, height,
		logoWidth+labelWidth+8, messageWidth+20, height, data.Color,
		logoSvg,
		labelX, data.Label,
		messageX, data.Message,
	))
}

// renderForTheBadge generates a "for-the-badge" style badge
func (g *Generator) renderForTheBadge(data Data, width, height, labelWidth, messageWidth, logoWidth int) []byte {
	template := `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%d" height="%d" role="img" aria-label="%s">
  <title>%s</title>
  <g shape-rendering="crispEdges">
    <rect width="%d" height="%d" fill="#555"/>
    <rect x="%d" width="%d" height="%d" fill="%s"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="11" font-weight="bold">
    %s
    <text x="%d" y="19">%s</text>
    <text x="%d" y="19">%s</text>
  </g>
</svg>`

	labelX := logoWidth + labelWidth/2 + 6
	messageX := logoWidth + labelWidth + messageWidth/2 + 16
	logoSvg := ""

	if data.Logo != "" {
		// Process logo to apply color (for currentColor logos)
		processedLogo := g.processLogoColor(data.Logo, data.LogoColor)
		logoSvg = fmt.Sprintf(`<image x="5" y="6" width="16" height="16" xlink:href="%s"/>`, processedLogo)
	}

	// Convert to uppercase for "for-the-badge" style
	label := strings.ToUpper(data.Label)
	message := strings.ToUpper(data.Message)

	return []byte(fmt.Sprintf(template,
		width, height, data.AriaLabel, data.AriaLabel,
		logoWidth+labelWidth+8, height,
		logoWidth+labelWidth+8, messageWidth+20, height, data.Color,
		logoSvg,
		labelX, label,
		messageX, message,
	))
}

// calculateTextWidth estimates text width (simplified calculation)
func (g *Generator) calculateTextWidth(text string) int {
	// Rough estimation: average character width ~6.5px for Verdana 11px
	return int(math.Ceil(float64(len(text)) * 6.5))
}

// Options represents options for badge generation
type Options struct {
	Style     string
	Label     string
	Logo      string
	LogoColor string
}

// Option is a function type for configuring badge options
type Option func(*Options)

// WithStyle sets the badge style
func WithStyle(style string) Option {
	return func(opts *Options) {
		opts.Style = style
	}
}

// WithLabel sets the badge label
func WithLabel(label string) Option {
	return func(opts *Options) {
		opts.Label = label
	}
}

// WithLogo sets the badge logo
func WithLogo(logo string) Option {
	return func(opts *Options) {
		opts.Logo = logo
	}
}

// WithLogoColor sets the logo color
func WithLogoColor(color string) Option {
	return func(opts *Options) {
		opts.LogoColor = color
	}
}
