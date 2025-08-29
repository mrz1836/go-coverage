package health

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// Mock health checker for testing
type mockHealthChecker struct {
	name   string
	result *CheckResult
}

func (m *mockHealthChecker) Name() string {
	return m.name
}

func (m *mockHealthChecker) Check(ctx context.Context) *CheckResult {
	return m.result
}

type slowMockHealthChecker struct {
	name   string
	delay  time.Duration
	result *CheckResult
}

func (s *slowMockHealthChecker) Name() string {
	return s.name
}

func (s *slowMockHealthChecker) Check(ctx context.Context) *CheckResult {
	time.Sleep(s.delay)
	return s.result
}

type panicMockHealthChecker struct {
	name string
}

func (p *panicMockHealthChecker) Name() string {
	return p.name
}

func (p *panicMockHealthChecker) Check(ctx context.Context) *CheckResult {
	panic("test panic")
}

type nilMockHealthChecker struct {
	name string
}

func (n *nilMockHealthChecker) Name() string {
	return n.name
}

func (n *nilMockHealthChecker) Check(ctx context.Context) *CheckResult {
	return nil
}

func TestManager_AddChecker(t *testing.T) {
	manager := NewManager()
	checker := &mockHealthChecker{name: "test"}

	manager.AddChecker(checker)

	if len(manager.checkers) != 1 {
		t.Errorf("Expected 1 checker, got %d", len(manager.checkers))
	}

	if manager.checkers[0] != checker {
		t.Error("Expected checker to be added correctly")
	}
}

func TestManager_SetTimeout(t *testing.T) {
	manager := NewManager()
	timeout := 45 * time.Second

	manager.SetTimeout(timeout)

	if manager.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, manager.timeout)
	}
}

func TestManager_CheckAll_AllHealthy(t *testing.T) {
	manager := NewManager()

	// Add healthy checkers
	manager.AddChecker(&mockHealthChecker{
		name: "test1",
		result: &CheckResult{
			Name:    "test1",
			Status:  StatusHealthy,
			Message: "All good",
		},
	})

	manager.AddChecker(&mockHealthChecker{
		name: "test2",
		result: &CheckResult{
			Name:    "test2",
			Status:  StatusHealthy,
			Message: "All good too",
		},
	})

	report := manager.CheckAll(context.Background())

	if report.OverallStatus != StatusHealthy {
		t.Errorf("Expected overall status healthy, got %s", report.OverallStatus)
	}

	if report.Summary.Total != 2 {
		t.Errorf("Expected 2 total checks, got %d", report.Summary.Total)
	}

	if report.Summary.Healthy != 2 {
		t.Errorf("Expected 2 healthy checks, got %d", report.Summary.Healthy)
	}

	if len(report.Checks) != 2 {
		t.Errorf("Expected 2 check results, got %d", len(report.Checks))
	}
}

func TestManager_CheckAll_WithWarnings(t *testing.T) {
	manager := NewManager()

	manager.AddChecker(&mockHealthChecker{
		name: "healthy",
		result: &CheckResult{
			Name:    "healthy",
			Status:  StatusHealthy,
			Message: "All good",
		},
	})

	manager.AddChecker(&mockHealthChecker{
		name: "warning",
		result: &CheckResult{
			Name:    "warning",
			Status:  StatusWarning,
			Message: "Something to watch",
		},
	})

	report := manager.CheckAll(context.Background())

	if report.OverallStatus != StatusWarning {
		t.Errorf("Expected overall status warning, got %s", report.OverallStatus)
	}

	if report.Summary.Healthy != 1 {
		t.Errorf("Expected 1 healthy check, got %d", report.Summary.Healthy)
	}

	if report.Summary.Warnings != 1 {
		t.Errorf("Expected 1 warning check, got %d", report.Summary.Warnings)
	}
}

func TestManager_CheckAll_WithUnhealthy(t *testing.T) {
	manager := NewManager()

	manager.AddChecker(&mockHealthChecker{
		name: "healthy",
		result: &CheckResult{
			Name:    "healthy",
			Status:  StatusHealthy,
			Message: "All good",
		},
	})

	manager.AddChecker(&mockHealthChecker{
		name: "unhealthy",
		result: &CheckResult{
			Name:    "unhealthy",
			Status:  StatusUnhealthy,
			Message: "Something is broken",
		},
	})

	report := manager.CheckAll(context.Background())

	if report.OverallStatus != StatusUnhealthy {
		t.Errorf("Expected overall status unhealthy, got %s", report.OverallStatus)
	}

	if report.Summary.Unhealthy != 1 {
		t.Errorf("Expected 1 unhealthy check, got %d", report.Summary.Unhealthy)
	}
}

func TestManager_CheckAll_Timeout(t *testing.T) {
	manager := NewManager()
	manager.SetTimeout(10 * time.Millisecond) // Very short timeout

	// Add a checker that will be slow
	slowChecker := &slowMockHealthChecker{
		name:  "slow",
		delay: 50 * time.Millisecond,
		result: &CheckResult{
			Name:    "slow",
			Status:  StatusHealthy,
			Message: "This should timeout",
		},
	}

	manager.AddChecker(slowChecker)

	report := manager.CheckAll(context.Background())

	if report.OverallStatus != StatusUnhealthy {
		t.Errorf("Expected overall status unhealthy due to timeout, got %s", report.OverallStatus)
	}

	if len(report.Checks) != 1 {
		t.Errorf("Expected 1 check result, got %d", len(report.Checks))
	}

	check := report.Checks[0]
	if check.Status != StatusUnhealthy {
		t.Errorf("Expected check status unhealthy, got %s", check.Status)
	}

	if !strings.Contains(check.Message, "timed out") {
		t.Errorf("Expected timeout message, got: %s", check.Message)
	}
}

func TestManager_CheckAll_PanicRecovery(t *testing.T) {
	manager := NewManager()

	// Add a checker that panics
	panicChecker := &panicMockHealthChecker{name: "panic"}

	manager.AddChecker(panicChecker)

	report := manager.CheckAll(context.Background())

	if report.OverallStatus != StatusUnhealthy {
		t.Errorf("Expected overall status unhealthy due to panic, got %s", report.OverallStatus)
	}

	if len(report.Checks) != 1 {
		t.Errorf("Expected 1 check result, got %d", len(report.Checks))
	}

	check := report.Checks[0]
	if check.Status != StatusUnhealthy {
		t.Errorf("Expected check status unhealthy, got %s", check.Status)
	}

	if !strings.Contains(check.Message, "panicked") {
		t.Errorf("Expected panic message, got: %s", check.Message)
	}
}

func TestManager_CheckAll_NilResult(t *testing.T) {
	manager := NewManager()

	// Add a checker that returns nil
	nilChecker := &nilMockHealthChecker{name: "nil"}

	manager.AddChecker(nilChecker)

	report := manager.CheckAll(context.Background())

	if report.OverallStatus != StatusUnhealthy {
		t.Errorf("Expected overall status unhealthy due to nil result, got %s", report.OverallStatus)
	}

	check := report.Checks[0]
	if !strings.Contains(check.Message, "returned nil result") {
		t.Errorf("Expected nil result message, got: %s", check.Message)
	}
}

func TestGitHubAPIChecker_MissingToken(t *testing.T) {
	checker := NewGitHubAPIChecker("", "owner/repo")
	result := checker.Check(context.Background())

	if result.Status != StatusUnhealthy {
		t.Errorf("Expected unhealthy status for missing token, got %s", result.Status)
	}

	if !strings.Contains(result.Message, "token is not provided") {
		t.Errorf("Expected token missing message, got: %s", result.Message)
	}
}

func TestGitHubAPIChecker_MissingRepository(t *testing.T) {
	checker := NewGitHubAPIChecker("token", "")
	result := checker.Check(context.Background())

	if result.Status != StatusUnhealthy {
		t.Errorf("Expected unhealthy status for missing repository, got %s", result.Status)
	}

	if !strings.Contains(result.Message, "repository is not specified") {
		t.Errorf("Expected repository missing message, got: %s", result.Message)
	}
}

func TestGitHubAPIChecker_APIConnectivity(t *testing.T) {
	// Create a mock GitHub API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "token test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Handle different endpoints
		switch r.URL.Path {
		case "/user":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"login":"testuser"}`))
		case "/repos/owner/repo":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"repo"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	checker := NewGitHubAPIChecker("test-token", "owner/repo")
	checker.BaseURL = server.URL

	result := checker.Check(context.Background())

	if result.Status != StatusHealthy {
		t.Errorf("Expected healthy status for valid API, got %s: %s", result.Status, result.Message)
	}

	if !strings.Contains(result.Message, "accessible") {
		t.Errorf("Expected accessible message, got: %s", result.Message)
	}
}

func TestGitHubAPIChecker_InvalidToken(t *testing.T) {
	// Create a mock server that returns unauthorized
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	checker := NewGitHubAPIChecker("invalid-token", "owner/repo")
	checker.BaseURL = server.URL

	result := checker.Check(context.Background())

	if result.Status != StatusUnhealthy {
		t.Errorf("Expected unhealthy status for invalid token, got %s", result.Status)
	}

	if !strings.Contains(result.Message, "invalid or expired") {
		t.Errorf("Expected invalid token message, got: %s", result.Message)
	}
}

func TestGitHubAPIChecker_RepositoryNotFound(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"login":"testuser"}`))
		case "/repos/owner/repo":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	checker := NewGitHubAPIChecker("valid-token", "owner/repo")
	checker.BaseURL = server.URL

	result := checker.Check(context.Background())

	if result.Status != StatusUnhealthy {
		t.Errorf("Expected unhealthy status for repository not found, got %s", result.Status)
	}

	if !strings.Contains(result.Message, "not found or no access") {
		t.Errorf("Expected repository not found message, got: %s", result.Message)
	}
}

func TestDiskSpaceChecker_SufficientSpace(t *testing.T) {
	tempDir := t.TempDir()

	// Set low requirements so test passes
	checker := NewDiskSpaceChecker(tempDir, 1024, 1.0) // 1KB minimum, 1% minimum
	result := checker.Check(context.Background())

	if result.Status == StatusUnhealthy {
		t.Errorf("Expected healthy or warning status for sufficient space, got %s: %s", result.Status, result.Message)
	}

	if result.Details == nil {
		t.Error("Expected details to be populated")
	}
}

func TestDiskSpaceChecker_NonexistentPath(t *testing.T) {
	checker := NewDiskSpaceChecker("/nonexistent/path/that/should/not/exist", 1024, 1.0)
	result := checker.Check(context.Background())

	// Should still work by checking parent directory or similar
	if result.Status == StatusUnhealthy && !strings.Contains(result.Message, "Failed to get disk usage") {
		t.Errorf("Expected disk usage error or success with fallback, got: %s", result.Message)
	}
}

func TestDiskSpaceChecker_Name(t *testing.T) {
	checker := NewDiskSpaceChecker("/tmp", 1024, 1.0)
	if checker.Name() != "Disk Space" {
		t.Errorf("Expected name 'Disk Space', got %s", checker.Name())
	}
}

func TestNetworkConnectivityChecker_NoEndpoints(t *testing.T) {
	checker := NewNetworkConnectivityChecker([]string{})
	result := checker.Check(context.Background())

	if result.Status != StatusSkipped {
		t.Errorf("Expected skipped status for no endpoints, got %s", result.Status)
	}

	if !strings.Contains(result.Message, "No endpoints configured") {
		t.Errorf("Expected no endpoints message, got: %s", result.Message)
	}
}

func TestNetworkConnectivityChecker_AllReachable(t *testing.T) {
	// Create mock servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	endpoints := []string{server1.URL, server2.URL}
	checker := NewNetworkConnectivityChecker(endpoints)
	result := checker.Check(context.Background())

	if result.Status != StatusHealthy {
		t.Errorf("Expected healthy status for reachable endpoints, got %s: %s", result.Status, result.Message)
	}

	if !strings.Contains(result.Message, "All 2 endpoints are reachable") {
		t.Errorf("Expected all reachable message, got: %s", result.Message)
	}
}

func TestNetworkConnectivityChecker_SomeUnreachable(t *testing.T) {
	// Create one working server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Mix of reachable and unreachable endpoints
	endpoints := []string{server.URL, "http://localhost:99999"}
	checker := NewNetworkConnectivityChecker(endpoints)
	result := checker.Check(context.Background())

	if result.Status != StatusWarning {
		t.Errorf("Expected warning status for partially unreachable endpoints, got %s", result.Status)
	}

	if !strings.Contains(result.Message, "1 of 2 endpoints unreachable") {
		t.Errorf("Expected partial unreachable message, got: %s", result.Message)
	}
}

func TestNetworkConnectivityChecker_AllUnreachable(t *testing.T) {
	// Use unreachable endpoints
	endpoints := []string{"http://localhost:99999", "http://localhost:99998"}
	checker := NewNetworkConnectivityChecker(endpoints)
	result := checker.Check(context.Background())

	if result.Status != StatusUnhealthy {
		t.Errorf("Expected unhealthy status for all unreachable endpoints, got %s", result.Status)
	}

	if !strings.Contains(result.Message, "All network endpoints are unreachable") {
		t.Errorf("Expected all unreachable message, got: %s", result.Message)
	}
}

func TestNetworkConnectivityChecker_Name(t *testing.T) {
	checker := NewNetworkConnectivityChecker([]string{})
	if checker.Name() != "Network Connectivity" {
		t.Errorf("Expected name 'Network Connectivity', got %s", checker.Name())
	}
}

func TestDefaultHealthCheckers(t *testing.T) {
	// Save original environment
	originalToken := os.Getenv("GITHUB_TOKEN")
	originalRepo := os.Getenv("GITHUB_REPOSITORY")

	defer func() {
		if originalToken == "" {
			_ = os.Unsetenv("GITHUB_TOKEN")
		} else {
			_ = os.Setenv("GITHUB_TOKEN", originalToken)
		}
		if originalRepo == "" {
			_ = os.Unsetenv("GITHUB_REPOSITORY")
		} else {
			_ = os.Setenv("GITHUB_REPOSITORY", originalRepo)
		}
	}()

	// Test with GitHub environment
	_ = os.Setenv("GITHUB_TOKEN", "test-token")
	_ = os.Setenv("GITHUB_REPOSITORY", "owner/repo")

	checkers := DefaultHealthCheckers()

	if len(checkers) < 3 {
		t.Errorf("Expected at least 3 default checkers, got %d", len(checkers))
	}

	// Check that we have the expected checker types
	var hasGitHub, hasDisk, hasNetwork bool

	for _, checker := range checkers {
		switch checker.Name() {
		case "GitHub API":
			hasGitHub = true
		case "Disk Space":
			hasDisk = true
		case "Network Connectivity":
			hasNetwork = true
		}
	}

	if !hasGitHub {
		t.Error("Expected GitHub API checker in default checkers")
	}
	if !hasDisk {
		t.Error("Expected Disk Space checker in default checkers")
	}
	if !hasNetwork {
		t.Error("Expected Network Connectivity checker in default checkers")
	}
}

func TestDefaultHealthCheckers_NoGitHub(t *testing.T) {
	// Save original environment
	originalToken := os.Getenv("GITHUB_TOKEN")
	originalRepo := os.Getenv("GITHUB_REPOSITORY")

	defer func() {
		if originalToken == "" {
			_ = os.Unsetenv("GITHUB_TOKEN")
		} else {
			_ = os.Setenv("GITHUB_TOKEN", originalToken)
		}
		if originalRepo == "" {
			_ = os.Unsetenv("GITHUB_REPOSITORY")
		} else {
			_ = os.Setenv("GITHUB_REPOSITORY", originalRepo)
		}
	}()

	// Clear GitHub environment
	_ = os.Unsetenv("GITHUB_TOKEN")
	_ = os.Unsetenv("GITHUB_REPOSITORY")

	checkers := DefaultHealthCheckers()

	// Should still have disk and network checkers, but no GitHub checker
	if len(checkers) < 2 {
		t.Errorf("Expected at least 2 default checkers without GitHub, got %d", len(checkers))
	}

	// Check that we don't have GitHub checker
	for _, checker := range checkers {
		if checker.Name() == "GitHub API" {
			t.Error("Should not have GitHub API checker without token/repo")
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, test := range tests {
		result := formatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", test.bytes, result, test.expected)
		}
	}
}

func TestHealthReport_String(t *testing.T) {
	report := &HealthReport{
		OverallStatus: StatusWarning,
		Timestamp:     time.Now(),
		Checks: []CheckResult{
			{
				Name:     "test1",
				Status:   StatusHealthy,
				Message:  "All good",
				Duration: 10 * time.Millisecond,
			},
			{
				Name:     "test2",
				Status:   StatusUnhealthy,
				Message:  "Something wrong",
				Duration: 5 * time.Millisecond,
				Suggestions: []string{
					"Fix the issue",
					"Try again",
				},
			},
		},
		Summary: HealthSummary{
			Total:     2,
			Healthy:   1,
			Warnings:  0,
			Unhealthy: 1,
			Skipped:   0,
		},
	}

	str := report.String()

	if !strings.Contains(str, "Health Status: WARNING") {
		t.Error("Expected overall status in string output")
	}

	if !strings.Contains(str, "test1 [HEALTHY]") {
		t.Error("Expected test1 healthy status in string output")
	}

	if !strings.Contains(str, "test2 [UNHEALTHY]") {
		t.Error("Expected test2 unhealthy status in string output")
	}

	if !strings.Contains(str, "→ Fix the issue") {
		t.Error("Expected suggestions in string output")
	}

	if !strings.Contains(str, "2 total checks") {
		t.Error("Expected summary in string output")
	}
}

func TestHealthReport_ToJSON(t *testing.T) {
	report := &HealthReport{
		OverallStatus: StatusHealthy,
		Timestamp:     time.Now(),
		Checks: []CheckResult{
			{
				Name:     "test",
				Status:   StatusHealthy,
				Message:  "All good",
				Duration: 10 * time.Millisecond,
			},
		},
		Summary: HealthSummary{
			Total:     1,
			Healthy:   1,
			Warnings:  0,
			Unhealthy: 0,
			Skipped:   0,
		},
	}

	jsonStr, err := report.ToJSON()
	if err != nil {
		t.Errorf("Failed to convert to JSON: %v", err)
	}

	if !strings.Contains(jsonStr, `"overall_status": "healthy"`) {
		t.Error("Expected overall status in JSON output")
	}

	if !strings.Contains(jsonStr, `"name": "test"`) {
		t.Error("Expected check name in JSON output")
	}
}

func BenchmarkManager_CheckAll(b *testing.B) {
	manager := NewManager()

	// Add several fast checkers
	for i := 0; i < 5; i++ {
		checker := &mockHealthChecker{
			name: fmt.Sprintf("test%d", i),
			result: &CheckResult{
				Name:    fmt.Sprintf("test%d", i),
				Status:  StatusHealthy,
				Message: "All good",
			},
		}
		manager.AddChecker(checker)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.CheckAll(ctx)
	}
}

func BenchmarkGitHubAPIChecker(b *testing.B) {
	// Create a fast mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"login":"test"}`))
	}))
	defer server.Close()

	checker := NewGitHubAPIChecker("test-token", "owner/repo")
	checker.BaseURL = server.URL

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checker.Check(ctx)
	}
}

func BenchmarkDiskSpaceChecker(b *testing.B) {
	tempDir := b.TempDir()
	checker := NewDiskSpaceChecker(tempDir, 1024, 1.0)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checker.Check(ctx)
	}
}

func BenchmarkNetworkConnectivityChecker(b *testing.B) {
	// Create a fast mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewNetworkConnectivityChecker([]string{server.URL})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checker.Check(ctx)
	}
}
