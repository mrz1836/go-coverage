package dashboard

import (
	"context"
	"testing"
	"time"
)

// BenchmarkGenerateDashboard benchmarks dashboard data generation
func BenchmarkGenerateDashboard(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title: "BenchmarkProject",
	})
	ctx := context.Background()
	request := &Request{
		TimeRange: TimeRange{
			Start: time.Now().Add(-30 * 24 * time.Hour),
			End:   time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dashboard.GenerateDashboard(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateDashboardSmall benchmarks with small dataset
func BenchmarkGenerateDashboardSmall(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title: "BenchmarkProject",
	})
	ctx := context.Background()
	request := &Request{
		TimeRange: TimeRange{
			Start: time.Now().Add(-7 * 24 * time.Hour),
			End:   time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dashboard.GenerateDashboard(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateDashboardLarge benchmarks with large dataset
func BenchmarkGenerateDashboardLarge(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title: "BenchmarkProject",
	})
	ctx := context.Background()
	request := &Request{
		TimeRange: TimeRange{
			Start: time.Now().Add(-90 * 24 * time.Hour),
			End:   time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dashboard.GenerateDashboard(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateHTML benchmarks HTML generation from dashboard data
func BenchmarkGenerateHTML(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title: "BenchmarkProject",
	})
	ctx := context.Background()
	data := createBenchmarkDashboardData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dashboard.GenerateHTML(ctx, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkProcessCoverageData benchmarks coverage data processing
func BenchmarkProcessCoverageData(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title: "BenchmarkProject",
	})
	ctx := context.Background()
	request := &Request{
		TimeRange: TimeRange{
			Start: time.Now().Add(-30 * 24 * time.Hour),
			End:   time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = dashboard.GenerateDashboard(ctx, request)
	}
}

// BenchmarkCalculateMetrics benchmarks metric calculations
func BenchmarkCalculateMetrics(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title: "BenchmarkProject",
	})
	ctx := context.Background()
	request := &Request{
		TimeRange: TimeRange{
			Start: time.Now().Add(-30 * 24 * time.Hour),
			End:   time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = dashboard.GenerateDashboard(ctx, request)
	}
}

// BenchmarkGeneratePackageSummary benchmarks package summary generation
func BenchmarkGeneratePackageSummary(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title: "BenchmarkProject",
	})
	ctx := context.Background()
	request := &Request{
		TimeRange: TimeRange{
			Start: time.Now().Add(-30 * 24 * time.Hour),
			End:   time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = dashboard.GenerateDashboard(ctx, request)
	}
}

// BenchmarkIdentifyLowCoverageFiles benchmarks low coverage file identification
func BenchmarkIdentifyLowCoverageFiles(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title: "BenchmarkProject",
	})
	ctx := context.Background()
	request := &Request{
		TimeRange: TimeRange{
			Start: time.Now().Add(-30 * 24 * time.Hour),
			End:   time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = dashboard.GenerateDashboard(ctx, request)
	}
}

// BenchmarkGenerateInsights benchmarks insight generation
func BenchmarkGenerateInsights(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title: "BenchmarkProject",
	})
	ctx := context.Background()
	request := &Request{
		TimeRange: TimeRange{
			Start: time.Now().Add(-30 * 24 * time.Hour),
			End:   time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = dashboard.GenerateDashboard(ctx, request)
	}
}

// BenchmarkDashboardMemoryAllocation benchmarks memory allocation
func BenchmarkDashboardMemoryAllocation(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title: "BenchmarkProject",
	})
	ctx := context.Background()
	request := &Request{
		TimeRange: TimeRange{
			Start: time.Now().Add(-30 * 24 * time.Hour),
			End:   time.Now(),
		},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := dashboard.GenerateDashboard(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
		_ = data // Prevent optimization
	}
}

// BenchmarkConcurrentDashboardGeneration benchmarks concurrent generation
func BenchmarkConcurrentDashboardGeneration(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title: "BenchmarkProject",
	})
	ctx := context.Background()
	request := &Request{
		TimeRange: TimeRange{
			Start: time.Now().Add(-30 * 24 * time.Hour),
			End:   time.Now(),
		},
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := dashboard.GenerateDashboard(ctx, request)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkWithHistory benchmarks dashboard generation with history
func BenchmarkWithHistory(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title:             "BenchmarkProject",
		EnablePredictions: true,
	})
	ctx := context.Background()
	request := &Request{
		TimeRange: TimeRange{
			Start: time.Now().Add(-30 * 24 * time.Hour),
			End:   time.Now(),
		},
		IncludePredictions: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dashboard.GenerateDashboard(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWithComparison benchmarks dashboard with comparison data
func BenchmarkWithComparison(b *testing.B) {
	dashboard := NewAnalyticsDashboard(&Config{
		Title: "BenchmarkProject",
	})
	ctx := context.Background()
	request := &Request{
		TimeRange: TimeRange{
			Start: time.Now().Add(-30 * 24 * time.Hour),
			End:   time.Now(),
		},
		IncludeTeamData: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dashboard.GenerateDashboard(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions

func createBenchmarkDashboardData() *Data {
	return &Data{
		CurrentMetrics: CurrentMetrics{
			Coverage:       75.5,
			CoverageChange: 2.5,
			TrendDirection: "up",
			TrendStrength:  "strong",
			LastUpdated:    time.Now(),
		},
		GeneratedAt: time.Now(),
	}
}
