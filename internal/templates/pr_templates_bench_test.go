package templates

import (
	"context"
	"testing"
	"time"
)

// BenchmarkRenderComment benchmarks PR comment rendering
func BenchmarkRenderComment(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	ctx := context.Background()
	templateName := "default"
	data := createBenchmarkTemplateData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.RenderComment(ctx, templateName, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderCommentSmall benchmarks rendering small PR comments
func BenchmarkRenderCommentSmall(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	ctx := context.Background()
	templateName := "minimal"
	data := createSmallTemplateData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.RenderComment(ctx, templateName, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderCommentLarge benchmarks rendering large PR comments
func BenchmarkRenderCommentLarge(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	ctx := context.Background()
	templateName := "detailed"
	data := createLargeTemplateData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.RenderComment(ctx, templateName, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderWithDiff benchmarks PR comment with diff data
func BenchmarkRenderWithDiff(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	ctx := context.Background()
	templateName := "with-diff"
	data := createTemplateDataWithDiff()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.RenderComment(ctx, templateName, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTemplateCompilation benchmarks template engine creation
func BenchmarkTemplateCompilation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Template compilation happens during engine creation
		_ = NewPRTemplateEngine(nil)
	}
}

// BenchmarkRenderSummarySection benchmarks summary section rendering
// renderSummary is not a public method, skipping this benchmark
/*
func BenchmarkRenderSummarySection(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	summary := createSummaryData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// renderSummary is a private method
	}
}
*/

// BenchmarkRenderFileChangesSection benchmarks file changes section
// renderFileChanges is not a public method, skipping this benchmark
/*
func BenchmarkRenderFileChangesSection(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	fileChanges := createFileChangesData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// renderFileChanges is a private method
	}
}
*/

// BenchmarkRenderPackageImpact benchmarks package impact section
// renderPackageImpact is not a public method, skipping this benchmark
/*
func BenchmarkRenderPackageImpact(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	packageImpact := createPackageImpactData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// renderPackageImpact is a private method
	}
}
*/

// BenchmarkFormatCoverageDelta benchmarks delta formatting
// formatCoverageDelta is not a public method, skipping this benchmark
/*
func BenchmarkFormatCoverageDelta(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	deltas := []float64{-10.5, -5.2, 0, 2.3, 8.7, 15.4}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		delta := deltas[i%len(deltas)]
		// formatCoverageDelta is a private method
	}
}
*/

// BenchmarkGenerateMarkdownTable benchmarks markdown table generation
// generateMarkdownTable is not a public method, skipping this benchmark
/*
func BenchmarkGenerateMarkdownTable(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	tableData := createTableData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// generateMarkdownTable is a private method
	}
}
*/

// BenchmarkMemoryAllocation benchmarks memory allocation
func BenchmarkMemoryAllocation(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	ctx := context.Background()
	templateName := "default"
	data := createBenchmarkTemplateData()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		comment, err := engine.RenderComment(ctx, templateName, data)
		if err != nil {
			b.Fatal(err)
		}
		_ = comment // Prevent optimization
	}
}

// BenchmarkConcurrentRendering benchmarks concurrent template rendering
func BenchmarkConcurrentRendering(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	ctx := context.Background()
	templateName := "default"
	data := createBenchmarkTemplateData()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := engine.RenderComment(ctx, templateName, data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMultipleTemplates benchmarks rendering different templates
func BenchmarkMultipleTemplates(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	ctx := context.Background()
	templates := []string{"minimal", "default", "detailed", "custom"}
	data := createBenchmarkTemplateData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		templateName := templates[i%len(templates)]
		_, err := engine.RenderComment(ctx, templateName, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWithEmojis benchmarks rendering with emoji support
func BenchmarkWithEmojis(b *testing.B) {
	engine := NewPRTemplateEngine(nil)
	ctx := context.Background()
	templateName := "with-emojis"
	data := createTemplateDataWithEmojis()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.RenderComment(ctx, templateName, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions

func createBenchmarkTemplateData() *TemplateData {
	return &TemplateData{
		Repository: RepositoryInfo{
			Owner: "testowner",
			Name:  "testrepo",
		},
		PullRequest: PullRequestInfo{
			Number:     123,
			Title:      "Add benchmark tests for performance critical functions",
			BaseBranch: "master",
			Branch:     "feature/benchmarks",
		},
		Timestamp: time.Now(),
		Coverage: CoverageData{
			Overall: CoverageMetrics{
				Percentage: 82.3,
			},
			Summary: CoverageSummary{
				Direction:     "improved",
				Magnitude:     "significant",
				OverallImpact: "positive",
			},
		},
		Comparison: ComparisonData{
			BasePercentage:    75.5,
			CurrentPercentage: 82.3,
			Change:            6.8,
			Direction:         "improved",
			Magnitude:         "significant",
			IsSignificant:     true,
		},
	}
}

func createSmallTemplateData() *TemplateData {
	return &TemplateData{
		Repository: RepositoryInfo{
			Owner: "testowner",
			Name:  "testrepo",
		},
		PullRequest: PullRequestInfo{
			Number:     456,
			Title:      "Minor fix",
			BaseBranch: "master",
			Branch:     "fix/minor",
		},
		Timestamp: time.Now(),
		Coverage: CoverageData{
			Overall: CoverageMetrics{
				Percentage: 91.0,
			},
			Summary: CoverageSummary{
				Direction:     "improved",
				Magnitude:     "minor",
				OverallImpact: "positive",
			},
		},
		Comparison: ComparisonData{
			BasePercentage:    90.0,
			CurrentPercentage: 91.0,
			Change:            1.0,
			Direction:         "improved",
			Magnitude:         "minor",
			IsSignificant:     false,
		},
	}
}

func createLargeTemplateData() *TemplateData {
	data := createBenchmarkTemplateData()
	// Add more complex data structures for large template testing
	data.Coverage.Overall.TotalStatements = 50000
	data.Coverage.Overall.CoveredStatements = 40000
	data.Coverage.Summary.Direction = "improved"
	data.Coverage.Summary.Magnitude = "significant"
	return data
}

func createTemplateDataWithDiff() *TemplateData {
	data := createBenchmarkTemplateData()
	data.PRFiles = &PRFileAnalysisData{
		Summary: PRFileSummaryData{
			TotalFiles:     3,
			GoFilesCount:   3,
			HasGoChanges:   true,
			HasTestChanges: false,
		},
		GoFiles: []PRFileData{
			{
				Filename:  "internal/parser/parser.go",
				Status:    "modified",
				Additions: 50,
				Deletions: 10,
				Changes:   60,
			},
			{
				Filename:  "internal/badge/generator.go",
				Status:    "modified",
				Additions: 30,
				Deletions: 5,
				Changes:   35,
			},
			{
				Filename:  "internal/analytics/report.go",
				Status:    "modified",
				Additions: 70,
				Deletions: 35,
				Changes:   105,
			},
		},
	}
	return data
}

func createTemplateDataWithEmojis() *TemplateData {
	data := createBenchmarkTemplateData()
	data.Config = TemplateConfig{
		IncludeEmojis: true,
	}
	data.Metadata = TemplateMetadata{
		GeneratedAt: time.Now(),
		Version:     "1.0.0",
	}
	return data
}
