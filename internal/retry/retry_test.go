package retry

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

// Static test errors
var (
	errNetworkTimeout          = errors.New("network timeout")
	errUnauthorized            = errors.New("401 unauthorized")
	errTimeout                 = errors.New("timeout")
	errNetworkIOTimeout        = errors.New("network i/o timeout")
	errConnectionResetByPeer   = errors.New("connection reset by peer")
	errConnectionRefused       = errors.New("connection refused")
	errTemporaryFailure        = errors.New("temporary failure in name resolution")
	errContextDeadlineExceeded = errors.New("context deadline exceeded")
	errPermanentFailure        = errors.New("permanent failure")
	errNotFound                = errors.New("404 not found")
	errRegularError            = errors.New("regular error")
	errHTTP500                 = errors.New("HTTP 500 Internal Server Error")
	errBadGateway              = errors.New("502 Bad Gateway")
	errServiceUnavailable      = errors.New("503 Service Unavailable")
	errGatewayTimeout          = errors.New("504 Gateway Timeout")
	errTooManyRequests         = errors.New("429 Too Many Requests")
	errNotFoundHTTP            = errors.New("404 Not Found")
	errBadRequest              = errors.New("400 Bad Request")
	errRateLimitExceeded       = errors.New("rate limit exceeded")
	errGitHubServerError       = errors.New("GitHub server error")
	errUnauthorizedHTTP        = errors.New("401 Unauthorized")
	errForbidden               = errors.New("403 Forbidden")
	errAbuseDetection          = errors.New("abuse detection mechanism triggered")
	errResourceUnavailable     = errors.New("resource temporarily unavailable")
	errDeviceBusy              = errors.New("device or resource busy")
	errNoSpaceLeft             = errors.New("no space left on device")
	errPermissionDenied        = errors.New("permission denied")
	errFileNotFound            = errors.New("file not found")
	errTestFailure             = errors.New("test failure")
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts = 3, got %d", config.MaxAttempts)
	}

	if config.InitialDelay != 100*time.Millisecond {
		t.Errorf("Expected InitialDelay = 100ms, got %v", config.InitialDelay)
	}

	if config.MaxDelay != 30*time.Second {
		t.Errorf("Expected MaxDelay = 30s, got %v", config.MaxDelay)
	}

	if config.Multiplier != 2.0 {
		t.Errorf("Expected Multiplier = 2.0, got %f", config.Multiplier)
	}

	if config.JitterFraction != 0.1 {
		t.Errorf("Expected JitterFraction = 0.1, got %f", config.JitterFraction)
	}

	if config.RetryIf == nil {
		t.Error("Expected RetryIf function to be set")
	}
}

func TestNetworkConfig(t *testing.T) {
	config := NetworkConfig()

	if config.MaxAttempts != 5 {
		t.Errorf("Expected MaxAttempts = 5, got %d", config.MaxAttempts)
	}

	if config.InitialDelay != 200*time.Millisecond {
		t.Errorf("Expected InitialDelay = 200ms, got %v", config.InitialDelay)
	}
}

func TestGitHubAPIConfig(t *testing.T) {
	config := GitHubAPIConfig()

	if config.MaxAttempts != 4 {
		t.Errorf("Expected MaxAttempts = 4, got %d", config.MaxAttempts)
	}

	if config.InitialDelay != 500*time.Millisecond {
		t.Errorf("Expected InitialDelay = 500ms, got %v", config.InitialDelay)
	}
}

func TestDoSuccess(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()

	callCount := 0
	err := Do(ctx, config, func() error {
		callCount++
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected function to be called once, got %d calls", callCount)
	}
}

func TestDoRetryableError(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		MaxAttempts:    3,
		InitialDelay:   10 * time.Millisecond,
		MaxDelay:       100 * time.Millisecond,
		Multiplier:     2.0,
		JitterFraction: 0.0, // No jitter for predictable tests
		RetryIf:        IsRetryableError,
	}

	callCount := 0
	testErr := errNetworkTimeout

	err := Do(ctx, config, func() error {
		callCount++
		if callCount < 3 {
			return testErr
		}
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error after retries, got: %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected function to be called 3 times, got %d calls", callCount)
	}
}

func TestDoNonRetryableError(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()

	callCount := 0
	testErr := errUnauthorized

	err := Do(ctx, config, func() error {
		callCount++
		return testErr
	})

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "non-retryable error") {
		t.Errorf("Expected non-retryable error message, got: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected function to be called once, got %d calls", callCount)
	}
}

func TestDoMaxAttemptsReached(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		MaxAttempts:    2,
		InitialDelay:   10 * time.Millisecond,
		MaxDelay:       100 * time.Millisecond,
		Multiplier:     2.0,
		JitterFraction: 0.0,
		RetryIf:        IsRetryableError,
	}

	callCount := 0
	testErr := errNetworkTimeout

	err := Do(ctx, config, func() error {
		callCount++
		return testErr
	})

	if err == nil {
		t.Error("Expected error after max attempts, got nil")
	}

	if !strings.Contains(err.Error(), "operation failed after 2 attempts") {
		t.Errorf("Expected max attempts error message, got: %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected function to be called 2 times, got %d calls", callCount)
	}
}

func TestDoContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &Config{
		MaxAttempts:    5,
		InitialDelay:   100 * time.Millisecond,
		MaxDelay:       1 * time.Second,
		Multiplier:     2.0,
		JitterFraction: 0.0,
		RetryIf:        IsRetryableError,
	}

	callCount := 0
	testErr := errNetworkTimeout

	// Cancel context after first attempt
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Do(ctx, config, func() error {
		callCount++
		return testErr
	})

	if err == nil {
		t.Error("Expected error due to context cancellation, got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected function to be called once before cancellation, got %d calls", callCount)
	}
}

func TestDoWithResult(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()

	expected := "test result"
	result, err := DoWithResult(ctx, config, func() (string, error) {
		return expected, nil
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result != expected {
		t.Errorf("Expected result %q, got %q", expected, result)
	}
}

func TestDoWithResultRetry(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		MaxAttempts:    3,
		InitialDelay:   10 * time.Millisecond,
		MaxDelay:       100 * time.Millisecond,
		Multiplier:     2.0,
		JitterFraction: 0.0,
		RetryIf:        IsRetryableError,
	}

	callCount := 0
	expected := "success result"
	testErr := errTimeout

	result, err := DoWithResult(ctx, config, func() (string, error) {
		callCount++
		if callCount < 3 {
			return "", testErr
		}
		return expected, nil
	})
	if err != nil {
		t.Errorf("Expected no error after retries, got: %v", err)
	}

	if result != expected {
		t.Errorf("Expected result %q, got %q", expected, result)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestCalculateDelay(t *testing.T) {
	config := &Config{
		InitialDelay:   100 * time.Millisecond,
		MaxDelay:       1 * time.Second,
		Multiplier:     2.0,
		JitterFraction: 0.0, // No jitter for predictable tests
	}

	// Test first attempt (attempt 1)
	delay1 := config.calculateDelay(1)
	expected1 := 100 * time.Millisecond
	if delay1 != expected1 {
		t.Errorf("Expected delay for attempt 1: %v, got %v", expected1, delay1)
	}

	// Test second attempt (attempt 2)
	delay2 := config.calculateDelay(2)
	expected2 := 200 * time.Millisecond
	if delay2 != expected2 {
		t.Errorf("Expected delay for attempt 2: %v, got %v", expected2, delay2)
	}

	// Test third attempt (attempt 3)
	delay3 := config.calculateDelay(3)
	expected3 := 400 * time.Millisecond
	if delay3 != expected3 {
		t.Errorf("Expected delay for attempt 3: %v, got %v", expected3, delay3)
	}

	// Test max delay enforcement
	delay10 := config.calculateDelay(10)
	if delay10 > config.MaxDelay {
		t.Errorf("Expected delay to be capped at MaxDelay (%v), got %v", config.MaxDelay, delay10)
	}
}

func TestCalculateDelayWithJitter(t *testing.T) {
	config := &Config{
		InitialDelay:   100 * time.Millisecond,
		MaxDelay:       1 * time.Second,
		Multiplier:     2.0,
		JitterFraction: 0.1,
	}

	// Test that jitter produces different results
	delays := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		delays[i] = config.calculateDelay(1)
	}

	// Check that not all delays are identical (jitter should vary them)
	allSame := true
	first := delays[0]
	for _, delay := range delays[1:] {
		if delay != first {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Expected jitter to produce different delays, but all delays were identical")
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"timeout error", errTimeout, true},
		{"network timeout", errNetworkIOTimeout, true},
		{"connection reset", errConnectionResetByPeer, true},
		{"connection refused", errConnectionRefused, true},
		{"temporary failure", errTemporaryFailure, true},
		{"context deadline", errContextDeadlineExceeded, true},
		{"permanent error", errPermanentFailure, false},
		{"404 not found", errNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"network timeout", &net.OpError{Op: "dial", Err: errTimeout}, true},
		{"dns error", &net.DNSError{IsTimeout: true, IsTemporary: true}, true},
		{"regular error", errRegularError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNetworkError(tt.err)
			if result != tt.expected {
				t.Errorf("IsNetworkError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsHTTPRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"500 server error", errHTTP500, true},
		{"502 bad gateway", errBadGateway, true},
		{"503 service unavailable", errServiceUnavailable, true},
		{"504 gateway timeout", errGatewayTimeout, true},
		{"429 rate limit", errTooManyRequests, true},
		{"404 not found", errNotFoundHTTP, false},
		{"400 bad request", errBadRequest, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHTTPRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsHTTPRetryableError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsGitHubRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"rate limit", errRateLimitExceeded, true},
		{"server error", errGitHubServerError, true},
		{"502 bad gateway", errBadGateway, true},
		{"503 service unavailable", errServiceUnavailable, true},
		{"401 unauthorized", errUnauthorizedHTTP, false},
		{"403 forbidden", errForbidden, false},
		{"404 not found", errNotFoundHTTP, false},
		{"abuse detection", errAbuseDetection, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitHubRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsGitHubRetryableError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsFileError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"resource unavailable", errResourceUnavailable, true},
		{"device busy", errDeviceBusy, true},
		{"no space left", errNoSpaceLeft, true},
		{"permission denied", errPermissionDenied, true},
		{"file not found", errFileNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFileError(tt.err)
			if result != tt.expected {
				t.Errorf("IsFileError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	// Test initial state
	if cb.State() != "closed" {
		t.Errorf("Expected initial state to be closed, got %s", cb.State())
	}

	// Test failures leading to open state
	testErr := errTestFailure
	for i := 0; i < 3; i++ {
		err := cb.Execute(func() error {
			return testErr
		})
		if err == nil {
			t.Errorf("Expected error on attempt %d", i+1)
		}
	}

	// Circuit should now be open
	if cb.State() != "open" {
		t.Errorf("Expected state to be open after failures, got %s", cb.State())
	}

	// Next call should fail immediately
	err := cb.Execute(func() error {
		return nil // This shouldn't be called
	})
	if err == nil || err.Error() != "circuit breaker is open" {
		t.Errorf("Expected circuit breaker open error, got %v", err)
	}

	// Wait for reset timeout
	time.Sleep(150 * time.Millisecond)

	// Next call should succeed and close circuit
	err = cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected success after reset timeout, got %v", err)
	}

	if cb.State() != "closed" {
		t.Errorf("Expected state to be closed after success, got %s", cb.State())
	}
}

func TestDefaultCircuitBreaker(t *testing.T) {
	cb := DefaultCircuitBreaker()

	if cb.MaxFailures != 5 {
		t.Errorf("Expected MaxFailures = 5, got %d", cb.MaxFailures)
	}

	if cb.ResetTimeout != 60*time.Second {
		t.Errorf("Expected ResetTimeout = 60s, got %v", cb.ResetTimeout)
	}
}

// Benchmark tests for performance
func BenchmarkDoSuccess(b *testing.B) {
	ctx := context.Background()
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Do(ctx, config, func() error {
			return nil
		})
	}
}

func BenchmarkDoWithRetries(b *testing.B) {
	ctx := context.Background()
	config := &Config{
		MaxAttempts:    3,
		InitialDelay:   1 * time.Microsecond,
		MaxDelay:       10 * time.Microsecond,
		Multiplier:     2.0,
		JitterFraction: 0.0,
		RetryIf:        IsRetryableError,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		_ = Do(ctx, config, func() error {
			count++
			if count < 2 {
				return errTimeout
			}
			return nil
		})
	}
}

func BenchmarkCalculateDelay(b *testing.B) {
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.calculateDelay(i%10 + 1)
	}
}

func BenchmarkIsRetryableError(b *testing.B) {
	err := errNetworkTimeout

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsRetryableError(err)
	}
}

func TestExampleUsage(t *testing.T) {
	ctx := context.Background()

	// Example: Retry a network operation
	config := NetworkConfig()

	var result string
	err := Do(ctx, config, func() error {
		// Simulate network operation that succeeds on third try
		switch result {
		case "":
			result = "attempt1"
			return errNetworkTimeout
		case "attempt1":
			result = "attempt2"
			return errConnectionRefused
		}
		result = "success"
		return nil
	})
	if err != nil {
		t.Errorf("Expected success after retries, got: %v", err)
	}

	if result != "success" {
		t.Errorf("Expected result 'success', got %q", result)
	}
}

func TestExampleGitHubAPI(t *testing.T) {
	ctx := context.Background()
	config := GitHubAPIConfig()

	attempts := 0

	// Example: GitHub API call with retries
	response, err := DoWithResult(ctx, config, func() (string, error) {
		attempts++
		switch attempts {
		case 1:
			return "", errTooManyRequests
		case 2:
			return "", errBadGateway
		}
		return "API response data", nil
	})
	if err != nil {
		t.Errorf("Expected success after GitHub API retries, got: %v", err)
	}

	if response != "API response data" {
		t.Errorf("Expected API response, got %q", response)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}
