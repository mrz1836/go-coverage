package templates

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSharedFooter(t *testing.T) {
	tests := []struct {
		name           string
		cssClass       string
		timestampField string
		expectedParts  []string
	}{
		{
			name:           "dashboard style",
			cssClass:       " dashboard",
			timestampField: "Timestamp",
			expectedParts: []string{
				`<footer class="footer">`,
				`<div class="footer-content dashboard">`,
				`data-timestamp="{{.Timestamp.Format "2006-01-02T15:04:05Z07:00"}}"`,
				`Generated {{.Timestamp.Format "2006-01-02 15:04:05 UTC"}}`,
				`<script src="./assets/js/coverage-time.js"></script>`,
				`<script src="./assets/js/theme.js"></script>`,
				`{{.LatestTag}}`,
				`go-coverage-link`,
			},
		},
		{
			name:           "regular style",
			cssClass:       "",
			timestampField: "GeneratedAt",
			expectedParts: []string{
				`<footer class="footer">`,
				`<div class="footer-content">`,
				`data-timestamp="{{.GeneratedAt.Format "2006-01-02T15:04:05Z07:00"}}"`,
				`Generated {{.GeneratedAt.Format "2006-01-02 15:04:05 UTC"}}`,
				`<script src="./assets/js/coverage-time.js"></script>`,
				`<script src="./assets/js/theme.js"></script>`,
				`{{.LatestTag}}`,
				`go-coverage-link`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSharedFooter(tt.cssClass, tt.timestampField)
			require.NotEmpty(t, result)

			// Check that all expected parts are present
			for _, expectedPart := range tt.expectedParts {
				assert.Contains(t, result, expectedPart, "Missing expected part: %s", expectedPart)
			}

			// Verify it's valid HTML structure
			assert.Contains(t, result, "<footer")
			assert.Contains(t, result, "</footer>")
			assert.Contains(t, result, "footer-content")
			assert.Contains(t, result, "footer-info")
			assert.Contains(t, result, "powered-text")
			assert.Contains(t, result, "timestamp-text")
			assert.Contains(t, result, "dynamic-timestamp")

			// Check for proper template variables
			assert.Contains(t, result, "{{.RepositoryOwner}}")
			assert.Contains(t, result, "{{.RepositoryName}}")
			assert.Contains(t, result, fmt.Sprintf("{{.%s.Format", tt.timestampField))
		})
	}
}

func TestGetSharedHead(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
	}{
		{
			name:        "coverage report",
			title:       "{{.RepositoryOwner}}/{{.RepositoryName}} - Coverage Report",
			description: "Code coverage analysis for {{.RepositoryOwner}}/{{.RepositoryName}}",
		},
		{
			name:        "dashboard",
			title:       "{{.ProjectName}} - Coverage Dashboard",
			description: "Real-time code coverage dashboard for {{.ProjectName}}",
		},
		{
			name:        "simple title",
			title:       "Test Coverage",
			description: "Test coverage report",
		},
		{
			name:        "empty values",
			title:       "",
			description: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSharedHead(tt.title, tt.description)
			require.NotEmpty(t, result)

			// Check HTML structure
			assert.Contains(t, result, "<head>")
			assert.Contains(t, result, "</head>")
			assert.Contains(t, result, `<meta charset="UTF-8">`)
			assert.Contains(t, result, `<meta name="viewport"`)

			// Check title and description are included
			if tt.title != "" {
				assert.Contains(t, result, fmt.Sprintf("<title>%s</title>", tt.title))
			} else {
				assert.Contains(t, result, "<title></title>")
			}

			if tt.description != "" {
				assert.Contains(t, result, fmt.Sprintf(`<meta name="description" content="%s">`, tt.description))
			} else {
				assert.Contains(t, result, `<meta name="description" content="">`)
			}

			// Check for required resources
			assert.Contains(t, result, "favicon.ico")
			assert.Contains(t, result, "favicon.svg")
			assert.Contains(t, result, "site.webmanifest")
			assert.Contains(t, result, "fonts.googleapis.com")
			assert.Contains(t, result, "Inter:wght")
			assert.Contains(t, result, "JetBrains+Mono")
			assert.Contains(t, result, "coverage.css")

			// Check for social sharing meta tags
			assert.Contains(t, result, `<meta property="og:title"`)
			assert.Contains(t, result, `<meta property="og:description"`)
			assert.Contains(t, result, `<meta property="og:type" content="website">`)

			// Check for Google Analytics placeholder
			assert.Contains(t, result, "{{.GoogleAnalyticsID}}")
			assert.Contains(t, result, "gtag")

			// Check for template variables
			assert.Contains(t, result, "{{.RepositoryOwner}}")
			assert.Contains(t, result, "{{.RepositoryName}}")
		})
	}
}

func TestComprehensiveTemplate(t *testing.T) {
	t.Run("TemplateContent", func(t *testing.T) {
		assert.NotEmpty(t, comprehensiveTemplate)

		// Check for key template sections
		expectedSections := []string{
			"Coverage Report",
			"Coverage Metrics",
			"Quality Assessment",
			"Recommendations",
			"Resources",
		}

		for _, section := range expectedSections {
			assert.Contains(t, comprehensiveTemplate, section)
		}

		// Check for template functions usage
		expectedFunctions := []string{
			"formatPercent",
			"formatChange",
			"statusEmoji",
			"trendEmoji",
			"gradeEmoji",
			"filterFiles",
			"humanize",
		}

		for _, function := range expectedFunctions {
			assert.Contains(t, comprehensiveTemplate, function)
		}

		// Check for metadata handling
		assert.Contains(t, comprehensiveTemplate, "Metadata.Signature")
		assert.Contains(t, comprehensiveTemplate, "Metadata.Version")
	})
}
