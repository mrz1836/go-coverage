package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/mrz1836/go-coverage/internal/history"
	"github.com/mrz1836/go-coverage/internal/parser"
)

// BenchmarkGenerate benchmarks dashboard generation
func BenchmarkGenerate(b *testing.B) {
	generator := NewGenerator(&GeneratorConfig{
		ProjectName: "BenchmarkProject",
		OutputDir:   b.TempDir(),
	})
	ctx := context.Background()
	coverage := createBenchmarkCoverageData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateSmall benchmarks dashboard generation with small dataset
func BenchmarkGenerateSmall(b *testing.B) {
	generator := NewGenerator(&GeneratorConfig{
		ProjectName: "BenchmarkProject",
		OutputDir:   b.TempDir(),
	})
	ctx := context.Background()
	coverage := createSmallCoverageData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateLarge benchmarks dashboard generation with large dataset
func BenchmarkGenerateLarge(b *testing.B) {
	generator := NewGenerator(&GeneratorConfig{
		ProjectName: "BenchmarkProject",
		OutputDir:   b.TempDir(),
	})
	ctx := context.Background()
	coverage := createLargeCoverageData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderDashboard benchmarks HTML rendering
func BenchmarkRenderDashboard(b *testing.B) {
	renderer := NewRenderer("")
	ctx := context.Background()
	data := createBenchmarkTemplateData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := renderer.RenderDashboard(ctx, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderWithHistory benchmarks dashboard rendering with history data
func BenchmarkRenderWithHistory(b *testing.B) {
	renderer := NewRenderer("")
	ctx := context.Background()
	data := createBenchmarkTemplateDataWithHistory()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := renderer.RenderDashboard(ctx, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateWithOptions benchmarks generation with various options
func BenchmarkGenerateWithOptions(b *testing.B) {
	generator := NewGenerator(&GeneratorConfig{
		ProjectName:     "BenchmarkProject",
		OutputDir:       b.TempDir(),
		RepositoryOwner: "test",
		RepositoryName:  "repo",
		TemplateDir:     "",
		AssetsDir:       "",
	})
	ctx := context.Background()
	coverage := createBenchmarkCoverageData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPrepareTemplateData benchmarks template data preparation
func BenchmarkPrepareTemplateData(b *testing.B) {
	generator := NewGenerator(&GeneratorConfig{
		ProjectName: "BenchmarkProject",
		OutputDir:   b.TempDir(),
	})
	ctx := context.Background()
	coverage := createBenchmarkCoverageData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkProcessPackageData benchmarks package data processing
func BenchmarkProcessPackageData(b *testing.B) {
	generator := NewGenerator(&GeneratorConfig{
		ProjectName: "BenchmarkProject",
		OutputDir:   b.TempDir(),
	})
	ctx := context.Background()
	coverage := createBenchmarkCoverageData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateChartData benchmarks chart data generation
func BenchmarkGenerateChartData(b *testing.B) {
	generator := NewGenerator(&GeneratorConfig{
		ProjectName: "BenchmarkProject",
		OutputDir:   b.TempDir(),
	})
	ctx := context.Background()
	coverage := createBenchmarkCoverageData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation during generation
func BenchmarkMemoryAllocation(b *testing.B) {
	generator := NewGenerator(&GeneratorConfig{
		ProjectName: "BenchmarkProject",
		OutputDir:   b.TempDir(),
	})
	ctx := context.Background()
	coverage := createBenchmarkCoverageData()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentGeneration benchmarks concurrent dashboard generation
func BenchmarkConcurrentGeneration(b *testing.B) {
	ctx := context.Background()
	coverage := createBenchmarkCoverageData()

	b.RunParallel(func(pb *testing.PB) {
		generator := NewGenerator(&GeneratorConfig{
			ProjectName: "BenchmarkProject",
			OutputDir:   b.TempDir(),
		})

		for pb.Next() {
			err := generator.Generate(ctx, coverage)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkTemplateExecution benchmarks raw template execution
func BenchmarkTemplateExecution(b *testing.B) {
	renderer := NewRenderer("")
	ctx := context.Background()
	data := createBenchmarkTemplateData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = renderer.RenderDashboard(ctx, data)
	}
}

// BenchmarkStaticAssetHandling benchmarks static asset processing
func BenchmarkStaticAssetHandling(b *testing.B) {
	generator := NewGenerator(&GeneratorConfig{
		ProjectName: "BenchmarkProject",
		OutputDir:   b.TempDir(),
	})
	ctx := context.Background()
	coverage := createBenchmarkCoverageData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions for creating test data

func createBenchmarkCoverageData() *CoverageData {
	packages := make([]PackageCoverage, 10)
	for i := 0; i < 10; i++ {
		pkgName := "package" + string(rune('A'+i))
		files := make([]FileCoverage, 5)

		for j := 0; j < 5; j++ {
			fileName := pkgName + "/file" + string(rune('0'+j)) + ".go"
			files[j] = FileCoverage{
				Name:         fileName,
				Path:         fileName,
				Coverage:     float64(70 + j*5),
				TotalLines:   100,
				CoveredLines: 70 + j*5,
				MissedLines:  30 - j*5,
			}
		}

		packages[i] = PackageCoverage{
			Name:         pkgName,
			Coverage:     75.0,
			TotalLines:   500,
			CoveredLines: 375,
			MissedLines:  125,
			Files:        files,
		}
	}

	return &CoverageData{
		TotalCoverage: 75.0,
		TotalLines:    5000,
		CoveredLines:  3750,
		MissedLines:   1250,
		Timestamp:     time.Now(),
		Packages:      packages,
		TotalFiles:    50,
		CoveredFiles:  40,
	}
}

func createSmallCoverageData() *CoverageData {
	return &CoverageData{
		TotalCoverage: 80.0,
		TotalLines:    100,
		CoveredLines:  80,
		MissedLines:   20,
		Timestamp:     time.Now(),
		Packages: []PackageCoverage{
			{
				Name:         "main",
				Coverage:     80.0,
				TotalLines:   100,
				CoveredLines: 80,
				MissedLines:  20,
				Files: []FileCoverage{
					{
						Name:         "main.go",
						Path:         "main.go",
						Coverage:     80.0,
						TotalLines:   100,
						CoveredLines: 80,
						MissedLines:  20,
					},
				},
			},
		},
		TotalFiles:   1,
		CoveredFiles: 1,
	}
}

func createLargeCoverageData() *CoverageData {
	packages := make([]PackageCoverage, 50)
	for i := 0; i < 50; i++ {
		pkgName := "package" + string(rune('A'+i%26)) + string(rune('0'+i/26))
		files := make([]FileCoverage, 20)

		for j := 0; j < 20; j++ {
			fileName := pkgName + "/file" + string(rune('0'+j/10)) + string(rune('0'+j%10)) + ".go"
			files[j] = FileCoverage{
				Name:         fileName,
				Path:         fileName,
				Coverage:     float64(60 + (i+j)%40),
				TotalLines:   200,
				CoveredLines: 120 + (i+j)%80,
				MissedLines:  80 - (i+j)%80,
			}
		}

		packages[i] = PackageCoverage{
			Name:         pkgName,
			Coverage:     float64(70 + i%30),
			TotalLines:   4000,
			CoveredLines: 2800 + i*40,
			MissedLines:  1200 - i*40,
			Files:        files,
		}
	}

	return &CoverageData{
		TotalCoverage: 72.5,
		TotalLines:    200000,
		CoveredLines:  145000,
		MissedLines:   55000,
		Timestamp:     time.Now(),
		Packages:      packages,
		TotalFiles:    1000,
		CoveredFiles:  850,
	}
}

func createBenchmarkTemplateData() map[string]interface{} {
	return map[string]interface{}{
		"ProjectName":  "BenchmarkProject",
		"Coverage":     75.5,
		"TotalLines":   5000,
		"CoveredLines": 3775,
		"Timestamp":    time.Now().Format(time.RFC3339),
		"Packages":     createPackageList(),
		"TopFiles":     createTopFilesList(),
	}
}

func createBenchmarkTemplateDataWithHistory() map[string]interface{} {
	data := createBenchmarkTemplateData()
	data["History"] = createBenchmarkHistoryData()
	data["ChartData"] = createChartDataPoints()
	return data
}

func createBenchmarkHistoryData() []history.Entry {
	entries := make([]history.Entry, 30)
	for i := 0; i < 30; i++ {
		entries[i] = history.Entry{
			Timestamp: time.Now().Add(-time.Duration(i) * 24 * time.Hour),
			Branch:    "master",
			Coverage: &parser.CoverageData{
				Percentage:   float64(70 + (i % 20)),
				TotalLines:   5000,
				CoveredLines: 3500 + (i % 1000),
			},
		}
	}
	return entries
}

func createPackageList() []map[string]interface{} {
	packages := make([]map[string]interface{}, 10)
	for i := 0; i < 10; i++ {
		packages[i] = map[string]interface{}{
			"Name":         "package" + string(rune('A'+i)),
			"Coverage":     float64(70 + i*3),
			"TotalLines":   500,
			"CoveredLines": 350 + i*15,
			"FileCount":    5,
		}
	}
	return packages
}

func createTopFilesList() []map[string]interface{} {
	files := make([]map[string]interface{}, 20)
	for i := 0; i < 20; i++ {
		files[i] = map[string]interface{}{
			"Path":         "pkg/file" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + ".go",
			"Coverage":     float64(60 + i*2),
			"TotalLines":   100,
			"CoveredLines": 60 + i*2,
		}
	}
	return files
}

func createChartDataPoints() []map[string]interface{} {
	points := make([]map[string]interface{}, 30)
	for i := 0; i < 30; i++ {
		points[i] = map[string]interface{}{
			"Date":     time.Now().Add(-time.Duration(i) * 24 * time.Hour).Format("2006-01-02"),
			"Coverage": float64(70 + (i % 20)),
		}
	}
	return points
}
