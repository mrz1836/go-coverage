package github

// These benchmarks are commented out because they were using
// non-existent APIs (NewDiffParser, ParseDiff, etc.) that don't match
// the actual implementation. The actual implementation uses Client.GetPRDiff
// and AnalyzePRFiles functions. These benchmarks need to be rewritten
// to match the actual API.

/*
import (
	"testing"
)

// BenchmarkParseDiff benchmarks diff parsing performance
func BenchmarkParseDiff(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createBenchmarkDiff()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseDiff(diffContent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseDiffSmall benchmarks parsing small diffs
func BenchmarkParseDiffSmall(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createSmallDiff()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseDiff(diffContent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseDiffLarge benchmarks parsing large diffs
func BenchmarkParseDiffLarge(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createLargeDiff()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseDiff(diffContent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExtractChangedFiles benchmarks changed file extraction
func BenchmarkExtractChangedFiles(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createBenchmarkDiff()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diff, _ := parser.ParseDiff(diffContent)
		_ = diff.GetChangedFiles()
	}
}

// BenchmarkIdentifyGoFiles benchmarks Go file identification
func BenchmarkIdentifyGoFiles(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createBenchmarkDiff()
	diff, _ := parser.ParseDiff(diffContent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = diff.GetGoFiles()
	}
}

// BenchmarkCountAdditionsAndDeletions benchmarks line count calculation
func BenchmarkCountAdditionsAndDeletions(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createBenchmarkDiff()
	diff, _ := parser.ParseDiff(diffContent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = diff.GetAdditions(), diff.GetDeletions()
	}
}

// BenchmarkAnalyzeCoverageImpact benchmarks coverage impact analysis
func BenchmarkAnalyzeCoverageImpact(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createBenchmarkDiff()
	diff, _ := parser.ParseDiff(diffContent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = diff.AnalyzeCoverageImpact()
	}
}

// BenchmarkFilterTestFiles benchmarks test file filtering
func BenchmarkFilterTestFiles(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createBenchmarkDiff()
	diff, _ := parser.ParseDiff(diffContent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = diff.GetTestFiles()
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation
func BenchmarkMemoryAllocation(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createLargeDiff()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		diff, err := parser.ParseDiff(diffContent)
		if err != nil {
			b.Fatal(err)
		}
		_ = diff // Prevent optimization
	}
}

// BenchmarkConcurrentParsing benchmarks concurrent diff parsing
func BenchmarkConcurrentParsing(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createBenchmarkDiff()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := parser.ParseDiff(diffContent)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkLargeMergeConflict benchmarks parsing diffs with merge conflicts
func BenchmarkLargeMergeConflict(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createMergeConflictDiff()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseDiff(diffContent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenameDetection benchmarks file rename detection
func BenchmarkRenameDetection(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createRenamedFilesDiff()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diff, _ := parser.ParseDiff(diffContent)
		_ = diff.GetRenamedFiles()
	}
}

// BenchmarkBinaryFileHandling benchmarks binary file handling
func BenchmarkBinaryFileHandling(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createBinaryFilesDiff()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseDiff(diffContent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPermissionChanges benchmarks permission change detection
func BenchmarkPermissionChanges(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createPermissionChangesDiff()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diff, _ := parser.ParseDiff(diffContent)
		_ = diff.GetPermissionChanges()
	}
}

// BenchmarkSubmoduleChanges benchmarks submodule change handling
func BenchmarkSubmoduleChanges(b *testing.B) {
	parser := NewDiffParser()
	diffContent := createSubmoduleDiff()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseDiff(diffContent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions

func createBenchmarkDiff() string {
	return `diff --git a/internal/parser/parser.go b/internal/parser/parser.go
index abc123..def456 100644
--- a/internal/parser/parser.go
+++ b/internal/parser/parser.go
@@ -10,7 +10,9 @@ func Parse(data string) (*Result, error) {
+    // Add validation
+    if err := validate(data); err != nil {
+        return nil, err
+    }
     result := &Result{}
-    // Old parsing logic
+    // New improved parsing logic
     return result, nil
 }

diff --git a/internal/badge/generator.go b/internal/badge/generator.go
index def456..ghi789 100644
--- a/internal/badge/generator.go
+++ b/internal/badge/generator.go
@@ -5,8 +5,10 @@ import (

 func Generate(coverage float64) string {
+    // Enhanced badge generation
+    color := getColor(coverage)
     badge := fmt.Sprintf("Coverage: %.1f%%", coverage)
-    return badge
+    return applyColor(badge, color)
 }
`
}

func createSmallDiff() string {
	return `diff --git a/main.go b/main.go
index abc123..def456 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+// Simple change
 func main() {}
`
}

func createLargeDiff() string {
	// Create a large diff with many files
	diff := ""
	for i := 0; i < 100; i++ {
		diff += fmt.Sprintf(`diff --git a/file%d.go b/file%d.go
index abc123..def456 100644
--- a/file%d.go
+++ b/file%d.go
@@ -1,10 +1,15 @@
 package pkg%d

 import (
+    "fmt"
     "testing"
 )

 func Function%d() {
-    // Old implementation
+    // New implementation
+    fmt.Println("Updated")
+    for i := 0; i < 10; i++ {
+        process(i)
+    }
 }

`, i, i, i, i, i, i)
	}
	return diff
}

func createMergeConflictDiff() string {
	return `diff --git a/conflict.go b/conflict.go
index abc123..def456 100644
--- a/conflict.go
+++ b/conflict.go
@@ -1,10 +1,14 @@
 package main

 func Conflict() {
<<<<<<< HEAD
     return "version1"
=======
     return "version2"
>>>>>>> feature-branch
 }
`
}

func createRenamedFilesDiff() string {
	return `diff --git a/old_name.go b/new_name.go
similarity index 95%
rename from old_name.go
rename to new_name.go
index abc123..def456 100644
--- a/old_name.go
+++ b/new_name.go
@@ -1,4 +1,4 @@
-// Package old provides old functionality
+// Package new provides new functionality
 package new

 func Function() {}
`
}

func createBinaryFilesDiff() string {
	return `diff --git a/image.png b/image.png
index abc123..def456 100644
Binary files a/image.png and b/image.png differ
diff --git a/data.bin b/data.bin
new file mode 100644
index 0000000..def456
Binary files /dev/null and b/data.bin differ
`
}

func createPermissionChangesDiff() string {
	return `diff --git a/script.sh b/script.sh
old mode 100644
new mode 100755
index abc123..def456 100644
--- a/script.sh
+++ b/script.sh
@@ -1,2 +1,3 @@
 #!/bin/bash
+echo "Now executable"
 echo "Script content"
`
}

func createSubmoduleDiff() string {
	return `diff --git a/.gitmodules b/.gitmodules
index abc123..def456 100644
--- a/.gitmodules
+++ b/.gitmodules
@@ -1,3 +1,6 @@
 [submodule "vendor/lib"]
     path = vendor/lib
     url = https://github.com/example/lib.git
+[submodule "vendor/new-lib"]
+    path = vendor/new-lib
+    url = https://github.com/example/new-lib.git
Subproject commit abc123def456789
`
}

func createDiffResult() *DiffResult {
	return &DiffResult{
		Files: []FileDiff{
			{
				Path:      "main.go",
				Additions: 10,
				Deletions: 5,
				IsNew:     false,
				IsDeleted: false,
			},
		},
		TotalAdditions: 100,
		TotalDeletions: 50,
	}
}

func createHunks() []Hunk {
	hunks := make([]Hunk, 10)
	for i := 0; i < 10; i++ {
		hunks[i] = Hunk{
			OldStart: i * 10,
			OldLines: 10,
			NewStart: i * 11,
			NewLines: 11,
			Content:  fmt.Sprintf("Hunk %d content", i),
		}
	}
	return hunks
}

func createChangedFiles() []string {
	files := make([]string, 50)
	for i := 0; i < 50; i++ {
		files[i] = fmt.Sprintf("file%d.go", i)
	}
	return files
}

func createGoFiles() []string {
	files := make([]string, 30)
	for i := 0; i < 30; i++ {
		files[i] = fmt.Sprintf("pkg/file%d.go", i)
	}
	return files
}

func createTestFiles() []string {
	files := make([]string, 20)
	for i := 0; i < 20; i++ {
		files[i] = fmt.Sprintf("pkg/file%d_test.go", i)
	}
	return files
}
*/
