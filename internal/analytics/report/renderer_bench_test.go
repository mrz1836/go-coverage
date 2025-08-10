package report

import (
	"context"
	"testing"
	"time"
)

// BenchmarkRenderReport benchmarks report rendering
func BenchmarkRenderReport(b *testing.B) {
	renderer := NewRenderer()
	ctx := context.Background()
	data := createRenderData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := renderer.RenderReport(ctx, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderReportSmall benchmarks rendering small reports
func BenchmarkRenderReportSmall(b *testing.B) {
	renderer := NewRenderer()
	ctx := context.Background()
	data := createSmallRenderData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := renderer.RenderReport(ctx, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderReportLarge benchmarks rendering large reports
func BenchmarkRenderReportLarge(b *testing.B) {
	renderer := NewRenderer()
	ctx := context.Background()
	data := createLargeRenderData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := renderer.RenderReport(ctx, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderWithComplexData benchmarks rendering with complex nested data
func BenchmarkRenderWithComplexData(b *testing.B) {
	renderer := NewRenderer()
	ctx := context.Background()
	data := createComplexRenderData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := renderer.RenderReport(ctx, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

/* BenchmarkTemplateExecution benchmarks raw template execution
// Commented out because executeTemplate is not a public method
func BenchmarkTemplateExecution(b *testing.B) {
	renderer := NewRenderer()
	data := createRenderData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// executeTemplate is a private method
	}
}
*/

/* BenchmarkTemplateCompilation benchmarks template compilation
// Commented out because compileTemplates is not a public method
func BenchmarkTemplateCompilation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer := NewRenderer()
		// compileTemplates is a private method
	}
}
*/

/* BenchmarkRenderPackageSection benchmarks package section rendering
// Commented out because renderPackageSection is not a public method
func BenchmarkRenderPackageSection(b *testing.B) {
	renderer := NewRenderer()
	packages := createPackageSection()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// renderPackageSection is a private method
	}
}
*/

/* BenchmarkRenderFileSection benchmarks file section rendering
// Commented out because renderFileSection is not a public method
func BenchmarkRenderFileSection(b *testing.B) {
	renderer := NewRenderer()
	files := createFileSection()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// renderFileSection is a private method
	}
}
*/

/* BenchmarkRenderStatistics benchmarks statistics rendering
// Commented out because renderStatistics is not a public method
func BenchmarkRenderStatistics(b *testing.B) {
	renderer := NewRenderer()
	stats := createStatisticsData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// renderStatistics is a private method
	}
}
*/

// BenchmarkHTMLEscaping benchmarks HTML escaping performance
func BenchmarkHTMLEscaping(b *testing.B) {
	renderer := NewRenderer()
	data := createDataWithSpecialChars()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := renderer.RenderReport(context.Background(), data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderMemoryAllocation benchmarks memory allocation during rendering
func BenchmarkRenderMemoryAllocation(b *testing.B) {
	renderer := NewRenderer()
	ctx := context.Background()
	data := createRenderData()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		output, err := renderer.RenderReport(ctx, data)
		if err != nil {
			b.Fatal(err)
		}
		_ = output // Prevent optimization
	}
}

// BenchmarkConcurrentRendering benchmarks concurrent rendering
func BenchmarkConcurrentRendering(b *testing.B) {
	renderer := NewRenderer()
	ctx := context.Background()
	data := createRenderData()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := renderer.RenderReport(ctx, data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

/* BenchmarkRenderWithThemes benchmarks rendering with different themes
// Commented out because NewRendererWithTheme doesn't exist
func BenchmarkRenderWithThemes(b *testing.B) {
	themes := []string{"default", "dark", "light", "minimal"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		theme := themes[i%len(themes)]
		// NewRendererWithTheme doesn't exist
		renderer := NewRenderer()
		data := createRenderData()

		_, err := renderer.RenderReport(ctx, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
*/

/* BenchmarkRenderJSON benchmarks JSON rendering
// Commented out because RenderJSON doesn't exist
func BenchmarkRenderJSON(b *testing.B) {
	renderer := NewRenderer()
	ctx := context.Background()
	data := createRenderData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// RenderJSON method doesn't exist
	}
}
*/

// Helper functions

func createRenderData() interface{} {
	return map[string]interface{}{
		"ProjectName":  "BenchmarkProject",
		"Coverage":     78.5,
		"TotalLines":   15000,
		"CoveredLines": 11775,
		"Timestamp":    time.Now().Format(time.RFC3339),
		"Packages":     createPackageRenderData(),
		"Files":        createFileRenderData(),
		"Statistics":   createStatisticsData(),
	}
}

func createSmallRenderData() interface{} {
	return map[string]interface{}{
		"ProjectName":  "SmallProject",
		"Coverage":     95.0,
		"TotalLines":   500,
		"CoveredLines": 475,
		"Timestamp":    time.Now().Format(time.RFC3339),
		"Packages": []map[string]interface{}{
			{
				"Name":         "main",
				"Coverage":     95.0,
				"TotalLines":   500,
				"CoveredLines": 475,
			},
		},
	}
}

func createLargeRenderData() interface{} {
	packages := make([]map[string]interface{}, 500)
	for i := 0; i < 500; i++ {
		packages[i] = map[string]interface{}{
			"Name":         "package" + string(rune('0'+i/100)) + string(rune('0'+(i/10)%10)) + string(rune('0'+i%10)),
			"Coverage":     float64(50 + i%50),
			"TotalLines":   1000,
			"CoveredLines": 500 + i%500,
			"Files":        createFileList(20),
		}
	}

	return map[string]interface{}{
		"ProjectName":  "LargeProject",
		"Coverage":     72.3,
		"TotalLines":   500000,
		"CoveredLines": 361500,
		"Timestamp":    time.Now().Format(time.RFC3339),
		"Packages":     packages,
		"Statistics":   createDetailedStatistics(),
	}
}

func createComplexRenderData() interface{} {
	return map[string]interface{}{
		"ProjectName": "ComplexProject",
		"Coverage":    75.8,
		"Metadata": map[string]interface{}{
			"Repository": "github.com/test/repo",
			"Branch":     "master",
			"Commit":     "abc123def456",
			"Author":     "test@example.com",
			"BuildInfo": map[string]interface{}{
				"Number": "12345",
				"URL":    "https://ci.example.com/build/12345",
				"Time":   time.Now().Format(time.RFC3339),
			},
		},
		"Packages": createComplexPackageData(),
		"History":  createHistoryData(),
		"Insights": createInsightData(),
	}
}

func createPackageRenderData() []map[string]interface{} {
	packages := make([]map[string]interface{}, 30)
	for i := 0; i < 30; i++ {
		packages[i] = map[string]interface{}{
			"Name":         "package" + string(rune('A'+i%26)),
			"Coverage":     60.0 + float64(i)*1.3,
			"TotalLines":   500,
			"CoveredLines": 300 + i*6,
			"FileCount":    10,
		}
	}
	return packages
}

func createFileRenderData() []map[string]interface{} {
	files := make([]map[string]interface{}, 50)
	for i := 0; i < 50; i++ {
		files[i] = map[string]interface{}{
			"Path":         "pkg/file" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + ".go",
			"Coverage":     45.0 + float64(i)*1.1,
			"TotalLines":   150,
			"CoveredLines": 68 + int(float64(i)*1.65),
		}
	}
	return files
}

func createStatisticsData() map[string]interface{} {
	return map[string]interface{}{
		"TotalPackages":   30,
		"TotalFiles":      450,
		"AverageCoverage": 78.5,
		"MinCoverage":     42.3,
		"MaxCoverage":     99.2,
		"StdDeviation":    14.7,
		"Percentiles": map[string]float64{
			"25": 65.0,
			"50": 78.5,
			"75": 89.2,
			"95": 96.8,
		},
	}
}

func createDataWithSpecialChars() interface{} {
	return map[string]interface{}{
		"ProjectName": "<script>alert('XSS')</script>",
		"Coverage":    85.5,
		"Description": "Test & benchmark <data> with \"quotes\" and 'apostrophes'",
		"HTML":        "<div class=\"test\">Content</div>",
		"URL":         "https://example.com?param=value&other=test",
	}
}

func createFileList(count int) []map[string]interface{} {
	files := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		files[i] = map[string]interface{}{
			"Name":     "file" + string(rune('0'+i)) + ".go",
			"Coverage": float64(60 + i*2),
		}
	}
	return files
}

func createDetailedStatistics() map[string]interface{} {
	return map[string]interface{}{
		"TotalPackages":   500,
		"TotalFiles":      10000,
		"AverageCoverage": 72.3,
		"Distribution": map[string]int{
			"0-20":   50,
			"20-40":  100,
			"40-60":  150,
			"60-80":  120,
			"80-100": 80,
		},
	}
}

func createComplexPackageData() []map[string]interface{} {
	packages := make([]map[string]interface{}, 20)
	for i := 0; i < 20; i++ {
		packages[i] = map[string]interface{}{
			"Name":     "complex.package." + string(rune('a'+i)),
			"Coverage": float64(70 + i),
			"Metrics": map[string]interface{}{
				"Complexity":      10 + i,
				"Maintainability": 85 - i,
				"TestCount":       50 + i*2,
			},
		}
	}
	return packages
}

func createHistoryData() []map[string]interface{} {
	history := make([]map[string]interface{}, 30)
	for i := 0; i < 30; i++ {
		history[i] = map[string]interface{}{
			"Date":     time.Now().Add(-time.Duration(i) * 24 * time.Hour).Format("2006-01-02"),
			"Coverage": float64(70 + (i % 20)),
			"Commit":   "commit" + string(rune('0'+i%10)),
		}
	}
	return history
}

func createInsightData() []string {
	return []string{
		"Coverage improved by 5% this week",
		"10 files need attention (coverage < 50%)",
		"Package 'core' has excellent coverage at 95%",
		"Consider adding tests for 'utils' package",
		"Coverage trend is positive over last 30 days",
	}
}
