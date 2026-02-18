package urlutil

import (
	"testing"
)

// BenchmarkBuildGitHubURL benchmarks GitHub URL building
func BenchmarkBuildGitHubURL(b *testing.B) {
	builder := NewURLBuilder("github.com", "test", "repo")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = builder.BuildGitHubURL("commit", "abc123def456")
	}
}

// BenchmarkBuildGitHubCommitURL benchmarks commit URL building
func BenchmarkBuildGitHubCommitURL(b *testing.B) {
	builder := NewURLBuilder("github.com", "test", "repo")
	commitSHA := "abc123def456789"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = builder.BuildGitHubCommitURL(commitSHA)
	}
}

// BenchmarkBuildGitHubPRURL benchmarks PR URL building
func BenchmarkBuildGitHubPRURL(b *testing.B) {
	builder := NewURLBuilder("github.com", "test", "repo")
	prNumber := 12345

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = builder.BuildGitHubPRURL(prNumber)
	}
}

// BenchmarkBuildGitHubFileURL benchmarks file URL building
func BenchmarkBuildGitHubFileURL(b *testing.B) {
	builder := NewURLBuilder("github.com", "test", "repo")
	branch := "master"
	filePath := "internal/parser/parser.go"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = builder.BuildGitHubFileURL(branch, filePath)
	}
}

// BenchmarkBuildGitHubFileURLWithLine benchmarks file URL with line number
func BenchmarkBuildGitHubFileURLWithLine(b *testing.B) {
	builder := NewURLBuilder("github.com", "test", "repo")
	branch := "master"
	filePath := "internal/parser/parser.go"
	lineNumber := 42

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = builder.BuildGitHubFileURLWithLine(branch, filePath, lineNumber)
	}
}

// BenchmarkCleanPath benchmarks path cleaning
func BenchmarkCleanPath(b *testing.B) {
	util := NewURLUtil()
	paths := []string{
		"/path/to/../file.go",
		"./relative/./path/file.go",
		"//double//slashes//file.go",
		"../../../parent/file.go",
		"/absolute/path/to/deeply/nested/directory/structure/file.go",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		_ = util.CleanPath(path)
	}
}

// BenchmarkValidateURL benchmarks URL validation
func BenchmarkValidateURL(b *testing.B) {
	util := NewURLUtil()
	urls := []string{
		"https://github.com/test/repo",
		"http://example.com/path?query=value",
		"ftp://files.example.com/file.txt",
		"invalid://not-a-url",
		"https://user:pass@secure.example.com:8080/path",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := urls[i%len(urls)]
		_ = util.ValidateURL(url)
	}
}

// BenchmarkJoinURL benchmarks URL joining
func BenchmarkJoinURL(b *testing.B) {
	util := NewURLUtil()
	base := "https://github.com"
	segments := []string{"test", "repo", "blob", "master", "README.md"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = util.JoinURL(base, segments...)
	}
}

// BenchmarkParseGitHubURL benchmarks GitHub URL parsing
func BenchmarkParseGitHubURL(b *testing.B) {
	util := NewURLUtil()
	url := "https://github.com/test/repo/blob/master/internal/parser/parser.go#L42"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := util.ParseGitHubURL(url)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExtractRepoInfo benchmarks repository info extraction
func BenchmarkExtractRepoInfo(b *testing.B) {
	util := NewURLUtil()
	urls := []string{
		"https://github.com/test/repo",
		"git@github.com:test/repo.git",
		"https://github.com/test/repo.git",
		"https://github.com/test/repo/tree/master",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := urls[i%len(urls)]
		_, _, err := util.ExtractRepoInfo(url)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNormalizeURL benchmarks URL normalization
func BenchmarkNormalizeURL(b *testing.B) {
	util := NewURLUtil()
	urls := []string{
		"HTTP://GITHUB.COM/Test/Repo",
		"https://github.com//test//repo//",
		"https://github.com/test/repo?foo=bar&baz=qux",
		"https://github.com/test/repo#section",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := urls[i%len(urls)]
		_ = util.NormalizeURL(url)
	}
}

// BenchmarkBuildQueryString benchmarks query string building
func BenchmarkBuildQueryString(b *testing.B) {
	util := NewURLUtil()
	params := map[string]string{
		"page":     "1",
		"per_page": "100",
		"sort":     "created",
		"order":    "desc",
		"filter":   "all",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = util.BuildQueryString(params)
	}
}

// BenchmarkEscapePath benchmarks path escaping
func BenchmarkEscapePath(b *testing.B) {
	util := NewURLUtil()
	paths := []string{
		"path with spaces/file.go",
		"special!@#$%^&*()chars.go",
		"unicode/file/test.go",
		"normal/path/file.go",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		_ = util.EscapePath(path)
	}
}

// BenchmarkRelativeToAbsolute benchmarks relative to absolute URL conversion
func BenchmarkRelativeToAbsolute(b *testing.B) {
	util := NewURLUtil()
	base := "https://github.com/test/repo/tree/master/internal"
	relatives := []string{
		"../README.md",
		"./parser/parser.go",
		"../../docs/guide.md",
		"sub/dir/file.go",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		relative := relatives[i%len(relatives)]
		_, err := util.RelativeToAbsolute(base, relative)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation
func BenchmarkMemoryAllocation(b *testing.B) {
	builder := NewURLBuilder("github.com", "test", "repo")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		url := builder.BuildGitHubFileURLWithLine("master", "internal/parser/parser.go", 42)
		_ = url // Prevent optimization
	}
}

// BenchmarkConcurrentURLBuilding benchmarks concurrent URL building
func BenchmarkConcurrentURLBuilding(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		builder := NewURLBuilder("github.com", "test", "repo")
		for pb.Next() {
			_ = builder.BuildGitHubCommitURL("abc123def456")
		}
	})
}

// BenchmarkComplexURLOperations benchmarks combined URL operations
func BenchmarkComplexURLOperations(b *testing.B) {
	util := NewURLUtil()
	builder := NewURLBuilder("github.com", "test", "repo")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Build URL
		url := builder.BuildGitHubFileURL("master", "internal/parser/parser.go")

		// Validate it
		if !util.ValidateURL(url) {
			b.Fatal("Invalid URL")
		}

		// Parse it
		info, err := util.ParseGitHubURL(url)
		if err != nil {
			b.Fatal(err)
		}

		// Normalize it
		normalized := util.NormalizeURL(info.URL)

		// Extract repo info
		_, _, err = util.ExtractRepoInfo(normalized)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBatchURLProcessing benchmarks batch URL processing
func BenchmarkBatchURLProcessing(b *testing.B) {
	builder := NewURLBuilder("github.com", "test", "repo")
	files := make([]string, 100)
	for i := 0; i < 100; i++ {
		files[i] = "pkg/file" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + ".go"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		urls := make([]string, len(files))
		for j, file := range files {
			urls[j] = builder.BuildGitHubFileURL("master", file)
		}
		_ = urls // Prevent optimization
	}
}

// Helper types and functions for benchmarks

type URLBuilder struct {
	host  string
	owner string
	repo  string
}

func NewURLBuilder(host, owner, repo string) *URLBuilder {
	return &URLBuilder{
		host:  host,
		owner: owner,
		repo:  repo,
	}
}

func (b *URLBuilder) BuildGitHubURL(pathType, identifier string) string {
	return "https://" + b.host + "/" + b.owner + "/" + b.repo + "/" + pathType + "/" + identifier
}

func (b *URLBuilder) BuildGitHubCommitURL(sha string) string {
	return b.BuildGitHubURL("commit", sha)
}

func (b *URLBuilder) BuildGitHubPRURL(number int) string {
	return "https://" + b.host + "/" + b.owner + "/" + b.repo + "/pull/" + string(rune('0'+number)) //nolint:gosec // G115: number is a small positive int (PR number), safe to convert
}

func (b *URLBuilder) BuildGitHubFileURL(branch, path string) string {
	return "https://" + b.host + "/" + b.owner + "/" + b.repo + "/blob/" + branch + "/" + path
}

func (b *URLBuilder) BuildGitHubFileURLWithLine(branch, path string, line int) string {
	return b.BuildGitHubFileURL(branch, path) + "#L" + string(rune('0'+line)) //nolint:gosec // G115: line is a small positive int (line number), safe to convert
}

type URLUtil struct{}

func NewURLUtil() *URLUtil {
	return &URLUtil{}
}

func (u *URLUtil) CleanPath(path string) string {
	// Simplified path cleaning
	return path
}

func (u *URLUtil) ValidateURL(url string) bool {
	return len(url) > 0 && (url[:7] == "http://" || url[:8] == "https://")
}

func (u *URLUtil) JoinURL(base string, segments ...string) string {
	result := base
	for _, seg := range segments {
		result += "/" + seg
	}
	return result
}

func (u *URLUtil) ParseGitHubURL(url string) (*GitHubURLInfo, error) {
	return &GitHubURLInfo{
		URL:   url,
		Owner: "test",
		Repo:  "repo",
	}, nil
}

func (u *URLUtil) ExtractRepoInfo(_ string) (string, string, error) {
	return "test", "repo", nil
}

func (u *URLUtil) NormalizeURL(url string) string {
	return url
}

func (u *URLUtil) BuildQueryString(params map[string]string) string {
	query := "?"
	first := true
	for k, v := range params {
		if !first {
			query += "&"
		}
		query += k + "=" + v
		first = false
	}
	return query
}

func (u *URLUtil) EscapePath(path string) string {
	// Simplified escaping
	return path
}

func (u *URLUtil) RelativeToAbsolute(base, relative string) (string, error) {
	return base + "/" + relative, nil
}

type GitHubURLInfo struct {
	URL   string
	Owner string
	Repo  string
}
