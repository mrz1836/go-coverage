package fallback

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// Static test errors
var (
	errMockOperationFailure        = errors.New("mock operation failure")
	errMockStrategyExecutionFailed = errors.New("mock strategy execution failed")
	errGitHubAPIError              = errors.New("github API error")
	errAPIRateLimitExceeded        = errors.New("api rate limit exceeded")
	errHTTP502BadGateway           = errors.New("HTTP 502 Bad Gateway")
	errOperationTimeout            = errors.New("operation timeout")
	errConnectionRefused           = errors.New("connection refused")
	errFileNotFound                = errors.New("file not found")
	errDeploymentFailed            = errors.New("deployment failed")
	errFailedToPushToGHPages       = errors.New("failed to push to gh-pages")
	errGitPushFailed               = errors.New("git push failed")
	errGitHubPagesDeploymentFailed = errors.New("GitHub pages deployment failed")
	errPushToRemoteFailed          = errors.New("push to remote failed")
	errFunctionError               = errors.New("function error")
	errGitHubAPIRateLimitExceeded  = errors.New("github API rate limit exceeded")
)

// Mock operation for testing
type mockOperation struct {
	operationType string
	shouldFail    bool
	failureCount  int
	metadata      map[string]interface{}
}

func newMockOperation() *mockOperation {
	return &mockOperation{
		operationType: "test_operation",
		metadata:      make(map[string]interface{}),
	}
}

func (op *mockOperation) Type() string                        { return op.operationType }
func (op *mockOperation) Metadata() map[string]interface{}    { return op.metadata }
func (op *mockOperation) SetMetadata(k string, v interface{}) { op.metadata[k] = v }
func (op *mockOperation) SetShouldFail(fail bool)             { op.shouldFail = fail }
func (op *mockOperation) SetFailureCount(count int)           { op.failureCount = count }

func (op *mockOperation) Execute(ctx context.Context) error {
	if op.shouldFail {
		if op.failureCount > 0 {
			op.failureCount--
			return errMockOperationFailure
		}
		// If shouldFail is true but failureCount is 0, always fail
		return errMockOperationFailure
	}
	return nil
}

// Mock fallback strategy for testing
type mockFallbackStrategy struct {
	name             string
	priority         int
	enabled          bool
	canHandleErrors  []string
	executeSuccess   bool
	executeCalled    bool
	executeCallCount int
}

func newMockStrategy(name string, priority int) *mockFallbackStrategy {
	return &mockFallbackStrategy{
		name:            name,
		priority:        priority,
		enabled:         true,
		canHandleErrors: make([]string, 0),
		executeSuccess:  true,
	}
}

func (s *mockFallbackStrategy) Name() string                   { return s.name }
func (s *mockFallbackStrategy) Priority() int                  { return s.priority }
func (s *mockFallbackStrategy) IsEnabled() bool                { return s.enabled }
func (s *mockFallbackStrategy) SetEnabled(enabled bool)        { s.enabled = enabled }
func (s *mockFallbackStrategy) SetExecuteSuccess(success bool) { s.executeSuccess = success }

func (s *mockFallbackStrategy) CanHandle(err error) bool {
	if err == nil {
		return false
	}
	for _, errStr := range s.canHandleErrors {
		if strings.Contains(err.Error(), errStr) {
			return true
		}
	}
	return false
}

func (s *mockFallbackStrategy) AddHandleError(errStr string) {
	s.canHandleErrors = append(s.canHandleErrors, errStr)
}

func (s *mockFallbackStrategy) Execute(ctx context.Context, operation Operation, originalErr error) error {
	s.executeCalled = true
	s.executeCallCount++

	if !s.executeSuccess {
		return errMockStrategyExecutionFailed
	}
	return nil
}

func TestNewFallbackManager(t *testing.T) {
	manager := NewFallbackManager()

	if manager == nil {
		t.Fatal("Expected non-nil FallbackManager")
	}

	if !manager.IsEnabled() {
		t.Error("Expected fallback manager to be enabled by default")
	}

	if manager.maxAttempts != 3 {
		t.Errorf("Expected default maxAttempts = 3, got %d", manager.maxAttempts)
	}

	if manager.backoffBase != time.Second {
		t.Errorf("Expected default backoffBase = 1s, got %v", manager.backoffBase)
	}
}

func TestRegisterStrategy(t *testing.T) {
	manager := NewFallbackManager()

	strategy1 := newMockStrategy("strategy1", 2)
	strategy2 := newMockStrategy("strategy2", 1) // Higher priority
	strategy3 := newMockStrategy("strategy3", 3) // Lower priority

	manager.RegisterStrategy(strategy1)
	manager.RegisterStrategy(strategy2)
	manager.RegisterStrategy(strategy3)

	// Check that strategies are sorted by priority (lower number = higher priority)
	if len(manager.strategies) != 3 {
		t.Fatalf("Expected 3 strategies, got %d", len(manager.strategies))
	}

	if manager.strategies[0].Name() != "strategy2" { // Priority 1
		t.Errorf("Expected first strategy to be 'strategy2', got '%s'", manager.strategies[0].Name())
	}

	if manager.strategies[1].Name() != "strategy1" { // Priority 2
		t.Errorf("Expected second strategy to be 'strategy1', got '%s'", manager.strategies[1].Name())
	}

	if manager.strategies[2].Name() != "strategy3" { // Priority 3
		t.Errorf("Expected third strategy to be 'strategy3', got '%s'", manager.strategies[2].Name())
	}
}

func TestExecuteWithFallback_PrimarySuccess(t *testing.T) {
	manager := NewFallbackManager()
	operation := newMockOperation()
	operation.SetShouldFail(false)

	ctx := context.Background()
	err := manager.ExecuteWithFallback(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error for successful primary operation, got: %v", err)
	}

	metrics := manager.GetMetrics()
	if metrics.TotalFallbacks != 0 {
		t.Errorf("Expected no fallbacks for successful operation, got %d", metrics.TotalFallbacks)
	}
}

func TestExecuteWithFallback_FallbackSuccess(t *testing.T) {
	manager := NewFallbackManager()

	// Create a strategy that can handle the error
	strategy := newMockStrategy("test_strategy", 1)
	strategy.AddHandleError("mock operation failure")
	manager.RegisterStrategy(strategy)

	operation := newMockOperation()
	operation.SetShouldFail(true)
	operation.SetFailureCount(1) // Fail once, then succeed (but we'll use fallback)

	ctx := context.Background()
	err := manager.ExecuteWithFallback(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error after successful fallback, got: %v", err)
	}

	if !strategy.executeCalled {
		t.Error("Expected fallback strategy to be executed")
	}

	metrics := manager.GetMetrics()
	if metrics.TotalFallbacks != 1 {
		t.Errorf("Expected 1 fallback attempt, got %d", metrics.TotalFallbacks)
	}

	if metrics.SuccessfulFallbacks != 1 {
		t.Errorf("Expected 1 successful fallback, got %d", metrics.SuccessfulFallbacks)
	}
}

func TestExecuteWithFallback_NoApplicableStrategy(t *testing.T) {
	manager := NewFallbackManager()

	// Create a strategy that cannot handle the error
	strategy := newMockStrategy("test_strategy", 1)
	strategy.AddHandleError("different error")
	manager.RegisterStrategy(strategy)

	operation := newMockOperation()
	operation.SetShouldFail(true)

	ctx := context.Background()
	err := manager.ExecuteWithFallback(ctx, operation)

	if err == nil {
		t.Error("Expected error when no applicable fallback strategy")
	}

	if !errors.Is(err, ErrNoFallbackAvailable) {
		t.Errorf("Expected ErrNoFallbackAvailable, got: %v", err)
	}

	if strategy.executeCalled {
		t.Error("Expected strategy not to be called for non-applicable error")
	}
}

func TestExecuteWithFallback_AllStrategiesFail(t *testing.T) {
	manager := NewFallbackManager()

	// Create strategies that can handle the error but will fail
	strategy1 := newMockStrategy("strategy1", 1)
	strategy1.AddHandleError("mock operation failure")
	strategy1.SetExecuteSuccess(false)

	strategy2 := newMockStrategy("strategy2", 2)
	strategy2.AddHandleError("mock operation failure")
	strategy2.SetExecuteSuccess(false)

	manager.RegisterStrategy(strategy1)
	manager.RegisterStrategy(strategy2)

	operation := newMockOperation()
	operation.SetShouldFail(true)

	ctx := context.Background()
	err := manager.ExecuteWithFallback(ctx, operation)

	if err == nil {
		t.Error("Expected error when all fallback strategies fail")
	}

	if !errors.Is(err, ErrAllFallbacksFailed) {
		t.Errorf("Expected ErrAllFallbacksFailed, got: %v", err)
	}

	if !strategy1.executeCalled || !strategy2.executeCalled {
		t.Error("Expected both strategies to be executed")
	}

	metrics := manager.GetMetrics()
	if metrics.FailedFallbacks != 1 {
		t.Errorf("Expected 1 failed fallback, got %d", metrics.FailedFallbacks)
	}
}

func TestExecuteWithFallback_Disabled(t *testing.T) {
	manager := NewFallbackManager()
	manager.SetEnabled(false)

	operation := newMockOperation()
	operation.SetShouldFail(true)

	ctx := context.Background()
	err := manager.ExecuteWithFallback(ctx, operation)

	if err == nil {
		t.Error("Expected error when fallback is disabled and operation fails")
	}

	// Should return the original operation error, not a fallback error
	if strings.Contains(err.Error(), "fallback") {
		t.Errorf("Expected original operation error, got fallback error: %v", err)
	}
}

func TestExecuteWithFallback_StrategyPriority(t *testing.T) {
	manager := NewFallbackManager()

	// Create strategies with different priorities
	strategy1 := newMockStrategy("low_priority", 3)
	strategy1.AddHandleError("mock operation failure")

	strategy2 := newMockStrategy("high_priority", 1)
	strategy2.AddHandleError("mock operation failure")

	strategy3 := newMockStrategy("medium_priority", 2)
	strategy3.AddHandleError("mock operation failure")

	manager.RegisterStrategy(strategy1)
	manager.RegisterStrategy(strategy2)
	manager.RegisterStrategy(strategy3)

	operation := newMockOperation()
	operation.SetShouldFail(true)

	ctx := context.Background()
	err := manager.ExecuteWithFallback(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error after fallback, got: %v", err)
	}

	// Only the highest priority strategy should be executed
	if !strategy2.executeCalled {
		t.Error("Expected high priority strategy to be executed")
	}

	if strategy1.executeCalled {
		t.Error("Expected low priority strategy NOT to be executed")
	}

	if strategy3.executeCalled {
		t.Error("Expected medium priority strategy NOT to be executed")
	}
}

func TestExecuteWithFallback_DisabledStrategy(t *testing.T) {
	manager := NewFallbackManager()

	strategy1 := newMockStrategy("disabled_strategy", 1)
	strategy1.AddHandleError("mock operation failure")
	strategy1.SetEnabled(false)

	strategy2 := newMockStrategy("enabled_strategy", 2)
	strategy2.AddHandleError("mock operation failure")

	manager.RegisterStrategy(strategy1)
	manager.RegisterStrategy(strategy2)

	operation := newMockOperation()
	operation.SetShouldFail(true)

	ctx := context.Background()
	err := manager.ExecuteWithFallback(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error after fallback, got: %v", err)
	}

	if strategy1.executeCalled {
		t.Error("Expected disabled strategy NOT to be executed")
	}

	if !strategy2.executeCalled {
		t.Error("Expected enabled strategy to be executed")
	}
}

func TestExecuteWithFallback_ContextCancellation(t *testing.T) {
	manager := NewFallbackManager()

	strategy := newMockStrategy("test_strategy", 1)
	strategy.AddHandleError("mock operation failure")
	manager.RegisterStrategy(strategy)

	operation := newMockOperation()
	operation.SetShouldFail(true)

	// Cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := manager.ExecuteWithFallback(ctx, operation)

	if err == nil {
		t.Error("Expected error due to context cancellation")
		return
	}

	if !strings.Contains(err.Error(), "mock operation failure") {
		t.Errorf("Expected original operation error due to immediate cancellation, got: %v", err)
	}
}

func TestFallbackMetrics(t *testing.T) {
	manager := NewFallbackManager()

	strategy := newMockStrategy("test_strategy", 1)
	strategy.AddHandleError("mock operation failure")
	manager.RegisterStrategy(strategy)

	// Test successful fallback
	operation1 := newMockOperation()
	operation1.SetShouldFail(true)

	ctx := context.Background()
	_ = manager.ExecuteWithFallback(ctx, operation1)

	metrics := manager.GetMetrics()
	if metrics.TotalFallbacks != 1 {
		t.Errorf("Expected 1 total fallback, got %d", metrics.TotalFallbacks)
	}

	if metrics.SuccessfulFallbacks != 1 {
		t.Errorf("Expected 1 successful fallback, got %d", metrics.SuccessfulFallbacks)
	}

	if metrics.FailedFallbacks != 0 {
		t.Errorf("Expected 0 failed fallbacks, got %d", metrics.FailedFallbacks)
	}

	// Check strategy usage
	if usage, exists := metrics.StrategyUsage["test_strategy"]; !exists || usage != 1 {
		t.Errorf("Expected strategy usage = 1, got %d", usage)
	}

	// Check success rate
	if rate, exists := metrics.StrategySuccessRate["test_strategy"]; !exists || rate != 1.0 {
		t.Errorf("Expected strategy success rate = 1.0, got %f", rate)
	}

	// Test failed fallback
	strategy.SetExecuteSuccess(false)
	operation2 := newMockOperation()
	operation2.SetShouldFail(true)

	_ = manager.ExecuteWithFallback(ctx, operation2)

	metrics = manager.GetMetrics()
	if metrics.TotalFallbacks != 2 {
		t.Errorf("Expected 2 total fallbacks, got %d", metrics.TotalFallbacks)
	}

	if metrics.FailedFallbacks != 1 {
		t.Errorf("Expected 1 failed fallback, got %d", metrics.FailedFallbacks)
	}

	// Check updated success rate (1 success out of 2 attempts = 0.5)
	if rate, exists := metrics.StrategySuccessRate["test_strategy"]; !exists || rate != 0.5 {
		t.Errorf("Expected strategy success rate = 0.5, got %f", rate)
	}
}

func TestResetMetrics(t *testing.T) {
	manager := NewFallbackManager()

	strategy := newMockStrategy("test_strategy", 1)
	strategy.AddHandleError("mock operation failure")
	manager.RegisterStrategy(strategy)

	// Generate some metrics
	operation := newMockOperation()
	operation.SetShouldFail(true)

	ctx := context.Background()
	_ = manager.ExecuteWithFallback(ctx, operation)

	// Verify metrics exist
	metrics := manager.GetMetrics()
	if metrics.TotalFallbacks == 0 {
		t.Error("Expected some metrics before reset")
	}

	// Reset metrics
	manager.ResetMetrics()

	// Verify metrics are reset
	metrics = manager.GetMetrics()
	if metrics.TotalFallbacks != 0 {
		t.Errorf("Expected TotalFallbacks = 0 after reset, got %d", metrics.TotalFallbacks)
	}

	if metrics.SuccessfulFallbacks != 0 {
		t.Errorf("Expected SuccessfulFallbacks = 0 after reset, got %d", metrics.SuccessfulFallbacks)
	}

	if len(metrics.StrategyUsage) != 0 {
		t.Errorf("Expected empty StrategyUsage after reset, got %v", metrics.StrategyUsage)
	}
}

func TestGitHubAPIFallbackStrategy(t *testing.T) {
	strategy := NewGitHubAPIFallbackStrategy()

	if strategy.Name() != "github_api_fallback" {
		t.Errorf("Expected name 'github_api_fallback', got '%s'", strategy.Name())
	}

	if strategy.Priority() != 1 {
		t.Errorf("Expected priority = 1, got %d", strategy.Priority())
	}

	if !strategy.IsEnabled() {
		t.Error("Expected strategy to be enabled")
	}
}

func TestGitHubAPIFallbackStrategy_CanHandle(t *testing.T) {
	strategy := NewGitHubAPIFallbackStrategy()

	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "github error",
			err:      errGitHubAPIError,
			expected: true,
		},
		{
			name:     "rate limit error",
			err:      errAPIRateLimitExceeded,
			expected: true,
		},
		{
			name:     "502 error",
			err:      errHTTP502BadGateway,
			expected: true,
		},
		{
			name:     "timeout error",
			err:      errOperationTimeout,
			expected: true,
		},
		{
			name:     "connection refused",
			err:      errConnectionRefused,
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      errFileNotFound,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := strategy.CanHandle(tc.err)
			if result != tc.expected {
				t.Errorf("Expected %t for error '%v', got %t", tc.expected, tc.err, result)
			}
		})
	}
}

func TestDeploymentFallbackStrategy(t *testing.T) {
	strategy := NewDeploymentFallbackStrategy()

	if strategy.Name() != "deployment_fallback" {
		t.Errorf("Expected name 'deployment_fallback', got '%s'", strategy.Name())
	}

	if strategy.Priority() != 2 {
		t.Errorf("Expected priority = 2, got %d", strategy.Priority())
	}

	if !strategy.IsEnabled() {
		t.Error("Expected strategy to be enabled")
	}
}

func TestDeploymentFallbackStrategy_CanHandle(t *testing.T) {
	strategy := NewDeploymentFallbackStrategy()

	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "deployment error",
			err:      errDeploymentFailed,
			expected: true,
		},
		{
			name:     "gh-pages error",
			err:      errFailedToPushToGHPages,
			expected: true,
		},
		{
			name:     "git error",
			err:      errGitPushFailed,
			expected: true,
		},
		{
			name:     "pages error",
			err:      errGitHubPagesDeploymentFailed,
			expected: true,
		},
		{
			name:     "push error",
			err:      errPushToRemoteFailed,
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      errFileNotFound,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := strategy.CanHandle(tc.err)
			if result != tc.expected {
				t.Errorf("Expected %t for error '%v', got %t", tc.expected, tc.err, result)
			}
		})
	}
}

func TestGetDefaultFallbackManager(t *testing.T) {
	manager := GetDefaultFallbackManager()

	if manager == nil {
		t.Fatal("Expected non-nil FallbackManager")
	}

	if len(manager.strategies) < 2 {
		t.Errorf("Expected at least 2 default strategies, got %d", len(manager.strategies))
	}

	// Check that GitHub API strategy is registered
	hasGitHubStrategy := false
	hasDeploymentStrategy := false

	for _, strategy := range manager.strategies {
		switch strategy.Name() {
		case "github_api_fallback":
			hasGitHubStrategy = true
		case "deployment_fallback":
			hasDeploymentStrategy = true
		}
	}

	if !hasGitHubStrategy {
		t.Error("Expected GitHub API fallback strategy to be registered")
	}

	if !hasDeploymentStrategy {
		t.Error("Expected deployment fallback strategy to be registered")
	}
}

func TestRecoverFromPanic(t *testing.T) {
	// Test normal execution (no panic)
	err := func() (err error) {
		defer func() {
			if r := RecoverFromPanic(); r != nil {
				err = r
			}
		}()
		return nil
	}()
	if err != nil {
		t.Errorf("Expected no error for normal execution, got: %v", err)
	}

	// Test panic recovery
	err = ExecuteWithRecovery(func() error {
		panic("test panic")
	})

	if err == nil {
		t.Error("Expected error from panic recovery")
	}

	if !strings.Contains(err.Error(), "recovered from panic") {
		t.Errorf("Expected panic recovery message, got: %v", err)
	}

	if !strings.Contains(err.Error(), "test panic") {
		t.Errorf("Expected panic message to be included, got: %v", err)
	}
}

func TestExecuteWithRecovery(t *testing.T) {
	// Test normal execution
	err := ExecuteWithRecovery(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error for normal execution, got: %v", err)
	}

	// Test function returning error
	testErr := errFunctionError
	err = ExecuteWithRecovery(func() error {
		return testErr
	})

	if !errors.Is(err, testErr) {
		t.Errorf("Expected function error, got: %v", err)
	}

	// Test panic recovery
	err = ExecuteWithRecovery(func() error {
		panic("test panic")
	})

	if err == nil {
		t.Error("Expected error from panic recovery")
	}

	if !strings.Contains(err.Error(), "recovered from panic") {
		t.Errorf("Expected panic recovery message, got: %v", err)
	}
}

// Benchmark tests
func BenchmarkExecuteWithFallback_Success(b *testing.B) {
	manager := NewFallbackManager()
	operation := newMockOperation()
	operation.SetShouldFail(false)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.ExecuteWithFallback(ctx, operation)
	}
}

func BenchmarkExecuteWithFallback_WithFallback(b *testing.B) {
	manager := NewFallbackManager()
	strategy := newMockStrategy("test_strategy", 1)
	strategy.AddHandleError("mock operation failure")
	manager.RegisterStrategy(strategy)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		operation := newMockOperation()
		operation.SetShouldFail(true)
		operation.SetFailureCount(1)

		_ = manager.ExecuteWithFallback(ctx, operation)
	}
}

func BenchmarkCanHandle(b *testing.B) {
	strategy := NewGitHubAPIFallbackStrategy()
	err := errGitHubAPIRateLimitExceeded

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strategy.CanHandle(err)
	}
}
