package analysis

import (
	"context"
	"testing"
	"time"
)

// BenchmarkCompareCoverage benchmarks coverage comparison
func BenchmarkCompareCoverage(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	baseSnapshot := createBenchmarkSnapshot()
	prSnapshot := createModifiedSnapshot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCompareCoverageSmall benchmarks comparison with small datasets
func BenchmarkCompareCoverageSmall(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	baseSnapshot := createSmallSnapshot()
	prSnapshot := createSmallModifiedSnapshot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCompareCoverageLarge benchmarks comparison with large datasets
func BenchmarkCompareCoverageLarge(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	baseSnapshot := createLargeSnapshot()
	prSnapshot := createLargeModifiedSnapshot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCalculateDelta benchmarks delta calculation via CompareCoverage
func BenchmarkCalculateDelta(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	baseSnapshot := createBenchmarkSnapshot()
	prSnapshot := createModifiedSnapshot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Delta calculation happens internally in CompareCoverage
		_, _ = engine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
	}
}

// BenchmarkCompareFiles benchmarks file-level comparison
func BenchmarkCompareFiles(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	baseSnapshot := createBenchmarkSnapshot()
	prSnapshot := createModifiedSnapshot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// File comparison happens internally in CompareCoverage
		_, _ = engine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
	}
}

// BenchmarkComparePackages benchmarks package-level comparison
func BenchmarkComparePackages(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	baseSnapshot := createBenchmarkSnapshot()
	prSnapshot := createModifiedSnapshot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Package comparison happens internally in CompareCoverage
		_, _ = engine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
	}
}

// BenchmarkIdentifyChangedFiles benchmarks changed file identification
func BenchmarkIdentifyChangedFiles(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	baseSnapshot := createBenchmarkSnapshot()
	prSnapshot := createModifiedSnapshot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Changed file identification happens internally in CompareCoverage
		_, _ = engine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
	}
}

// BenchmarkCalculateImpact benchmarks impact calculation via CompareCoverage
func BenchmarkCalculateImpact(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	baseSnapshot := createBenchmarkSnapshot()
	prSnapshot := createModifiedSnapshot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Impact calculation happens internally in CompareCoverage
		_, _ = engine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
	}
}

// BenchmarkGenerateSummary benchmarks summary generation
func BenchmarkGenerateSummary(b *testing.B) {
	engine := NewComparisonEngine(nil)
	result := createComparisonResult()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.generateSummary(result)
	}
}

// BenchmarkWithMetadata benchmarks comparison with test metadata
func BenchmarkWithMetadata(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	baseSnapshot := createBenchmarkSnapshot()
	prSnapshot := createModifiedSnapshot()
	// Add test metadata to simulate real-world usage
	prSnapshot.TestMetadata = TestMetadata{
		TestCount:    250,
		FailedTests:  5,
		SkippedTests: 0,
		TestDuration: time.Duration(30) * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation
func BenchmarkMemoryAllocation(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	baseSnapshot := createBenchmarkSnapshot()
	prSnapshot := createModifiedSnapshot()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := engine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
		if err != nil {
			b.Fatal(err)
		}
		_ = result // Prevent optimization
	}
}

// BenchmarkConcurrentComparison benchmarks concurrent comparisons
func BenchmarkConcurrentComparison(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	baseSnapshot := createBenchmarkSnapshot()
	prSnapshot := createModifiedSnapshot()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := engine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkIdenticalComparison benchmarks comparison of identical snapshots
func BenchmarkIdenticalComparison(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	snapshot := createBenchmarkSnapshot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.CompareCoverage(ctx, snapshot, snapshot)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCompletelyDifferent benchmarks comparison of completely different snapshots
func BenchmarkCompletelyDifferent(b *testing.B) {
	engine := NewComparisonEngine(nil)
	ctx := context.Background()
	baseSnapshot := createBenchmarkSnapshot()
	prSnapshot := createCompletelyDifferentSnapshot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.CompareCoverage(ctx, baseSnapshot, prSnapshot)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions

func createBenchmarkSnapshot() *CoverageSnapshot {
	return &CoverageSnapshot{
		OverallCoverage: CoverageMetrics{
			Percentage:        75.5,
			TotalLines:        10000,
			CoveredLines:      7550,
			TotalStatements:   12000,
			CoveredStatements: 9060,
			TotalFunctions:    500,
			CoveredFunctions:  400,
		},
		FileCoverage:    createBenchmarkFileMetrics(),
		PackageCoverage: createBenchmarkPackageMetrics(),
		Timestamp:       time.Now(),
		Branch:          "master",
		CommitSHA:       "abc123",
	}
}

func createModifiedSnapshot() *CoverageSnapshot {
	snapshot := createBenchmarkSnapshot()
	snapshot.OverallCoverage.Percentage = 82.3
	snapshot.OverallCoverage.CoveredLines = 8230
	snapshot.OverallCoverage.CoveredStatements = 9876
	snapshot.CommitSHA = "def456"

	// Modify some package metrics
	for name, pkg := range snapshot.PackageCoverage {
		if name == "packageA" || name == "packageB" {
			pkg.Percentage += 10
			pkg.CoveredStatements = int(float64(pkg.TotalStatements) * (pkg.Percentage / 100))
			snapshot.PackageCoverage[name] = pkg
		}
	}

	return snapshot
}

func createSmallSnapshot() *CoverageSnapshot {
	return &CoverageSnapshot{
		OverallCoverage: CoverageMetrics{
			Percentage:        90.0,
			TotalLines:        500,
			CoveredLines:      450,
			TotalStatements:   600,
			CoveredStatements: 540,
			TotalFunctions:    25,
			CoveredFunctions:  23,
		},
		FileCoverage: map[string]FileMetrics{
			"main.go": {
				Filename:          "main.go",
				Package:           "main",
				Percentage:        90.0,
				TotalStatements:   500,
				CoveredStatements: 450,
			},
		},
		PackageCoverage: map[string]PackageMetrics{
			"main": {
				Package:           "main",
				Percentage:        90.0,
				TotalStatements:   500,
				CoveredStatements: 450,
			},
		},
		Timestamp: time.Now(),
		Branch:    "master",
		CommitSHA: "small123",
	}
}

func createSmallModifiedSnapshot() *CoverageSnapshot {
	snapshot := createSmallSnapshot()
	snapshot.OverallCoverage.Percentage = 95.0
	snapshot.OverallCoverage.CoveredLines = 475
	snapshot.OverallCoverage.CoveredStatements = 570
	snapshot.CommitSHA = "small456"
	return snapshot
}

func createLargeSnapshot() *CoverageSnapshot {
	packageCoverage := make(map[string]PackageMetrics)
	fileCoverage := make(map[string]FileMetrics)

	for i := 0; i < 200; i++ {
		pkgName := "package" + string(rune('A'+i%26)) + string(rune('0'+i/26))
		packageCoverage[pkgName] = PackageMetrics{
			Package:           pkgName,
			Percentage:        float64(60 + i%40),
			TotalStatements:   5000,
			CoveredStatements: 3000 + i*10,
		}
		// Add some files for each package
		for j := 0; j < 5; j++ {
			fileName := pkgName + "/file" + string(rune('0'+j)) + ".go"
			fileCoverage[fileName] = FileMetrics{
				Filename:          fileName,
				Package:           pkgName,
				Percentage:        float64(60 + (i+j)%40),
				TotalStatements:   1000,
				CoveredStatements: 600 + (i+j)*2,
			}
		}
	}

	return &CoverageSnapshot{
		OverallCoverage: CoverageMetrics{
			Percentage:        70.5,
			TotalLines:        1000000,
			CoveredLines:      705000,
			TotalStatements:   1200000,
			CoveredStatements: 846000,
			TotalFunctions:    10000,
			CoveredFunctions:  7050,
		},
		FileCoverage:    fileCoverage,
		PackageCoverage: packageCoverage,
		Timestamp:       time.Now(),
		Branch:          "master",
		CommitSHA:       "large123",
	}
}

func createLargeModifiedSnapshot() *CoverageSnapshot {
	snapshot := createLargeSnapshot()
	snapshot.OverallCoverage.Percentage = 73.8
	snapshot.OverallCoverage.CoveredLines = 738000
	snapshot.OverallCoverage.CoveredStatements = 885600
	snapshot.CommitSHA = "large456"

	// Modify 30% of packages
	count := 0
	for name, pkg := range snapshot.PackageCoverage {
		if count%3 == 0 {
			pkg.Percentage += 5
			pkg.CoveredStatements = int(float64(pkg.TotalStatements) * (pkg.Percentage / 100))
			snapshot.PackageCoverage[name] = pkg
		}
		count++
	}

	return snapshot
}

func createCompletelyDifferentSnapshot() *CoverageSnapshot {
	packageCoverage := make(map[string]PackageMetrics)
	fileCoverage := make(map[string]FileMetrics)

	for i := 0; i < 30; i++ {
		pkgName := "newpackage" + string(rune('X'+i%3)) + string(rune('0'+i))
		packageCoverage[pkgName] = PackageMetrics{
			Package:           pkgName,
			Percentage:        float64(40 + i*2),
			TotalStatements:   1000,
			CoveredStatements: 400 + i*20,
			FileCount:         3,
		}
		// Add some files for each package
		for j := 0; j < 3; j++ {
			fileName := pkgName + "/file" + string(rune('0'+j)) + ".go"
			fileCoverage[fileName] = FileMetrics{
				Filename:          fileName,
				Package:           pkgName,
				Percentage:        float64(40 + (i+j)*2),
				TotalStatements:   333,
				CoveredStatements: 133 + (i+j)*7,
			}
		}
	}

	return &CoverageSnapshot{
		OverallCoverage: CoverageMetrics{
			Percentage:        55.0,
			TotalLines:        30000,
			CoveredLines:      16500,
			TotalStatements:   36000,
			CoveredStatements: 19800,
			TotalFunctions:    1500,
			CoveredFunctions:  825,
		},
		FileCoverage:    fileCoverage,
		PackageCoverage: packageCoverage,
		Timestamp:       time.Now(),
		Branch:          "feature",
		CommitSHA:       "different123",
	}
}

func createComparisonResult() *ComparisonResult {
	return &ComparisonResult{
		BaseSnapshot: *createBenchmarkSnapshot(),
		PRSnapshot:   *createModifiedSnapshot(),
		OverallChange: OverallChangeAnalysis{
			PercentageChange:       6.8,
			StatementChange:        230,
			CoveredStatementChange: 680,
			Direction:              "improved",
			Magnitude:              "significant",
			IsSignificant:          true,
		},
		FileChanges: []FileChangeAnalysis{
			{
				Filename:         "main.go",
				BasePercentage:   70.0,
				PRPercentage:     75.5,
				PercentageChange: 5.5,
				Direction:        "improved",
				Magnitude:        "moderate",
				IsSignificant:    true,
			},
		},
		PackageChanges: []PackageChangeAnalysis{
			{
				Package:          "main",
				PercentageChange: 6.8,
				Direction:        "improved",
				IsSignificant:    true,
			},
		},
		TrendAnalysis: TrendAnalysis{
			Direction:  "upward",
			Momentum:   "steady",
			Volatility: 2.5,
			Prediction: Prediction{
				NextCoverage: 85.0,
			},
			HistoricalContext: HistoricalContext{},
		},
		QualityAssessment: QualityAssessment{
			OverallGrade:  "A",
			CoverageGrade: "A",
			TrendGrade:    "B+",
			RiskLevel:     "low",
			QualityScore:  85.0,
			Strengths:     []string{"Good coverage", "Improving trend"},
			Weaknesses:    []string{},
		},
		Summary: ComparisonSummary{
			OverallImpact:  "positive",
			KeyChanges:     []string{"Coverage improved by 6.8%", "25 files changed"},
			CriticalIssues: []string{},
			Highlights:     []string{"20 files improved", "8 packages improved"},
			NextSteps:      []string{"Maintain coverage above 80%"},
		},
	}
}

// Removed createDiffData - DiffData type doesn't exist in the package

func createBenchmarkFileMetrics() map[string]FileMetrics {
	fileMetrics := make(map[string]FileMetrics)
	for i := 0; i < 50; i++ {
		fileName := "pkg/file" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + ".go"
		fileMetrics[fileName] = FileMetrics{
			Filename:          fileName,
			Package:           "package" + string(rune('A'+i/10)),
			Percentage:        float64(70 + i%30),
			TotalStatements:   200,
			CoveredStatements: 140 + i%60,
		}
	}
	return fileMetrics
}

func createBenchmarkPackageMetrics() map[string]PackageMetrics {
	packageMetrics := make(map[string]PackageMetrics)
	for i := 0; i < 10; i++ {
		pkgName := "package" + string(rune('A'+i))
		packageMetrics[pkgName] = PackageMetrics{
			Package:           pkgName,
			Percentage:        float64(75 + i*2),
			TotalStatements:   1000,
			CoveredStatements: 750 + i*20,
			FileCount:         5,
		}
	}
	return packageMetrics
}
