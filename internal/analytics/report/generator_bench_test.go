package report

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mrz1836/go-coverage/internal/parser"
)

// BenchmarkGenerate benchmarks report generation
func BenchmarkGenerate(b *testing.B) {
	generator := NewGenerator(&Config{
		OutputDir:      b.TempDir(),
		RepositoryName: "BenchmarkProject",
	})
	ctx := context.Background()
	coverage := createBenchmarkCoverage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateSmall benchmarks report generation with small dataset
func BenchmarkGenerateSmall(b *testing.B) {
	generator := NewGenerator(&Config{
		OutputDir:      b.TempDir(),
		RepositoryName: "BenchmarkProject",
	})
	ctx := context.Background()
	coverage := createSmallCoverage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateLarge benchmarks report generation with large dataset
func BenchmarkGenerateLarge(b *testing.B) {
	generator := NewGenerator(&Config{
		OutputDir:      b.TempDir(),
		RepositoryName: "BenchmarkProject",
	})
	ctx := context.Background()
	coverage := createLargeCoverage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation
func BenchmarkMemoryAllocation(b *testing.B) {
	generator := NewGenerator(&Config{
		OutputDir:      b.TempDir(),
		RepositoryName: "BenchmarkProject",
	})
	ctx := context.Background()
	coverage := createBenchmarkCoverage()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentReportGeneration benchmarks concurrent generation
func BenchmarkConcurrentReportGeneration(b *testing.B) {
	ctx := context.Background()
	coverage := createBenchmarkCoverage()

	b.RunParallel(func(pb *testing.PB) {
		generator := NewGenerator(&Config{
			OutputDir:      b.TempDir(),
			RepositoryName: "BenchmarkProject",
		})

		for pb.Next() {
			err := generator.Generate(ctx, coverage)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkWithMetadata benchmarks generation with extensive metadata
func BenchmarkWithMetadata(b *testing.B) {
	generator := NewGenerator(&Config{
		OutputDir:       b.TempDir(),
		RepositoryName:  "BenchmarkProject",
		RepositoryOwner: "test",
		CommitSHA:       "abc123def456",
		BranchName:      "master",
		PRNumber:        "789",
	})
	ctx := context.Background()
	coverage := createBenchmarkCoverage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJSONGeneration benchmarks JSON report generation
func BenchmarkJSONGeneration(b *testing.B) {
	generator := NewGenerator(&Config{
		OutputDir:      b.TempDir(),
		RepositoryName: "BenchmarkProject",
	})
	ctx := context.Background()
	coverage := createBenchmarkCoverage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.Generate(ctx, coverage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions

func createBenchmarkCoverage() *parser.CoverageData {
	packages := make(map[string]*parser.PackageCoverage)
	for i := 0; i < 20; i++ {
		pkgName := "package" + string(rune('A'+i))
		files := make(map[string]*parser.FileCoverage)

		for j := 0; j < 15; j++ {
			fileName := pkgName + "/file" + string(rune('0'+j/10)) + string(rune('0'+j%10)) + ".go"
			statements := make([]parser.Statement, 20)
			for k := 0; k < 20; k++ {
				statements[k] = parser.Statement{
					StartLine: k * 5,
					EndLine:   k*5 + 3,
					NumStmt:   1,
					Count:     k % 3,
				}
			}

			files[fileName] = &parser.FileCoverage{
				Path:         fileName,
				Percentage:   float64(65 + j*2),
				TotalLines:   200,
				CoveredLines: 130 + j*4,
				Statements:   statements,
			}
		}

		packages[pkgName] = &parser.PackageCoverage{
			Name:         pkgName,
			Percentage:   float64(70 + i%25),
			TotalLines:   3000,
			CoveredLines: 2100 + i*50,
			Files:        files,
		}
	}

	return &parser.CoverageData{
		Mode:         "atomic",
		Percentage:   74.2,
		TotalLines:   60000,
		CoveredLines: 44520,
		Timestamp:    time.Now(),
		Packages:     packages,
	}
}

func createSmallCoverage() *parser.CoverageData {
	return &parser.CoverageData{
		Mode:         "atomic",
		Percentage:   90.0,
		TotalLines:   100,
		CoveredLines: 90,
		Timestamp:    time.Now(),
		Packages: map[string]*parser.PackageCoverage{
			"main": {
				Name:         "main",
				Percentage:   90.0,
				TotalLines:   100,
				CoveredLines: 90,
				Files: map[string]*parser.FileCoverage{
					"main.go": {
						Path:         "main.go",
						Percentage:   90.0,
						TotalLines:   100,
						CoveredLines: 90,
					},
				},
			},
		},
	}
}

func createLargeCoverage() *parser.CoverageData {
	packages := make(map[string]*parser.PackageCoverage)
	for i := 0; i < 200; i++ {
		pkgName := "package" + string(rune('A'+i%26)) + string(rune('0'+i/26))
		files := make(map[string]*parser.FileCoverage)

		for j := 0; j < 100; j++ {
			fileName := pkgName + "/file" + string(rune('0'+j/100)) + string(rune('0'+(j/10)%10)) + string(rune('0'+j%10)) + ".go"
			files[fileName] = &parser.FileCoverage{
				Path:         fileName,
				Percentage:   float64(40 + (i+j)%60),
				TotalLines:   500,
				CoveredLines: 200 + (i+j)%300,
			}
		}

		packages[pkgName] = &parser.PackageCoverage{
			Name:         pkgName,
			Percentage:   float64(60 + i%40),
			TotalLines:   50000,
			CoveredLines: 30000 + i*200,
			Files:        files,
		}
	}

	return &parser.CoverageData{
		Mode:         "atomic",
		Percentage:   68.7,
		TotalLines:   10000000,
		CoveredLines: 6870000,
		Timestamp:    time.Now(),
		Packages:     packages,
	}
}

func createSampleHTMLContent() []byte {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Coverage Report</title>
</head>
<body>
    <h1>Coverage Report</h1>
    <p>Total Coverage: 76.8%</p>
</body>
</html>`
	return []byte(html)
}

// BenchmarkWriteToFile simulates actual file I/O
func BenchmarkWriteToFile(b *testing.B) {
	tempDir := b.TempDir()
	content := createSampleHTMLContent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := tempDir + "/report_" + string(rune('0'+i%10)) + ".html"
		err := os.WriteFile(filename, content, 0o600)
		if err != nil {
			b.Fatal(err)
		}
	}
}
