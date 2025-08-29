// Package retry provides configurable retry logic with exponential backoff
// for network operations and other potentially transient failures.
package retry

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"strings"
	"time"
)

// ErrRetryStop is returned to indicate that retries should stop
var (
	ErrRetryStop          = errors.New("stop retrying")
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)

// Config defines the retry configuration with exponential backoff
type Config struct {
	MaxAttempts    int              // Maximum number of retry attempts (including first attempt)
	InitialDelay   time.Duration    // Initial delay between retries
	MaxDelay       time.Duration    // Maximum delay between retries
	Multiplier     float64          // Exponential backoff multiplier
	JitterFraction float64          // Amount of jitter to add (0.0-1.0)
	RetryIf        func(error) bool // Function to determine if error should trigger retry
}

// DefaultConfig returns a sensible default retry configuration
func DefaultConfig() *Config {
	return &Config{
		MaxAttempts:    3,
		InitialDelay:   100 * time.Millisecond,
		MaxDelay:       30 * time.Second,
		Multiplier:     2.0,
		JitterFraction: 0.1,
		RetryIf:        IsRetryableError,
	}
}

// NetworkConfig returns retry configuration optimized for network operations
func NetworkConfig() *Config {
	return &Config{
		MaxAttempts:    5,
		InitialDelay:   200 * time.Millisecond,
		MaxDelay:       10 * time.Second,
		Multiplier:     1.5,
		JitterFraction: 0.2,
		RetryIf:        IsNetworkError,
	}
}

// GitHubAPIConfig returns retry configuration optimized for GitHub API calls
func GitHubAPIConfig() *Config {
	return &Config{
		MaxAttempts:    4,
		InitialDelay:   500 * time.Millisecond,
		MaxDelay:       15 * time.Second,
		Multiplier:     2.0,
		JitterFraction: 0.15,
		RetryIf:        IsGitHubRetryableError,
	}
}

// FileOperationConfig returns retry configuration for file operations
func FileOperationConfig() *Config {
	return &Config{
		MaxAttempts:    3,
		InitialDelay:   50 * time.Millisecond,
		MaxDelay:       2 * time.Second,
		Multiplier:     1.5,
		JitterFraction: 0.05,
		RetryIf:        IsFileError,
	}
}

// Do executes the given function with retry logic according to the configuration
func Do(ctx context.Context, config *Config, fn func() error) error {
	if config == nil {
		config = DefaultConfig()
	}

	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if we should retry this error
		if config.RetryIf != nil && !config.RetryIf(err) {
			return fmt.Errorf("non-retryable error on attempt %d: %w", attempt, err)
		}

		// Don't wait after the last attempt
		if attempt == config.MaxAttempts {
			break
		}

		// Calculate delay with exponential backoff and jitter
		delay := config.calculateDelay(attempt)

		// Check if context is canceled before waiting
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled after %d attempts: %w", attempt, ctx.Err())
		default:
		}

		// Wait for the calculated delay
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled during retry delay after %d attempts: %w", attempt, ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// DoWithResult executes a function that returns a result and error with retry logic
func DoWithResult[T any](ctx context.Context, config *Config, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error

	if config == nil {
		config = DefaultConfig()
	}

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Execute the function
		res, err := fn()
		if err == nil {
			return res, nil // Success
		}

		lastErr = err

		// Check if we should retry this error
		if config.RetryIf != nil && !config.RetryIf(err) {
			return result, fmt.Errorf("non-retryable error on attempt %d: %w", attempt, err)
		}

		// Don't wait after the last attempt
		if attempt == config.MaxAttempts {
			break
		}

		// Calculate delay with exponential backoff and jitter
		delay := config.calculateDelay(attempt)

		// Check if context is canceled before waiting
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("context canceled after %d attempts: %w", attempt, ctx.Err())
		default:
		}

		// Wait for the calculated delay
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("context canceled during retry delay after %d attempts: %w", attempt, ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return result, fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// calculateDelay calculates the delay for the given attempt with exponential backoff and jitter
func (c *Config) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff delay
	delay := float64(c.InitialDelay) * math.Pow(c.Multiplier, float64(attempt-1))

	// Apply maximum delay limit
	if delay > float64(c.MaxDelay) {
		delay = float64(c.MaxDelay)
	}

	// Add jitter to prevent thundering herd
	if c.JitterFraction > 0 {
		// Use crypto/rand for secure random number generation
		var b [8]byte
		if _, err := rand.Read(b[:]); err == nil {
			// Convert random bytes to float64 in range [-1, 1]
			randVal := float64(binary.BigEndian.Uint64(b[:]))/float64(^uint64(0))*2 - 1
			jitter := delay * c.JitterFraction * randVal
			delay += jitter
		}

		// Ensure delay is not negative
		if delay < 0 {
			delay = float64(c.InitialDelay)
		}
	}

	return time.Duration(delay)
}

// IsRetryableError determines if an error is potentially retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common retryable error patterns
	errStr := err.Error()
	retryablePatterns := []string{
		"timeout",
		"temporary failure",
		"connection reset",
		"connection refused",
		"no such host",
		"network is unreachable",
		"i/o timeout",
		"context deadline exceeded",
	}

	errStrLower := strings.ToLower(errStr)
	for _, pattern := range retryablePatterns {
		if strings.Contains(errStrLower, pattern) {
			return true
		}
	}

	// Check for specific error types
	return IsNetworkError(err) || IsHTTPRetryableError(err)
}

// IsNetworkError checks if an error is a network-related error
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific network error types first
	var netOpErr *net.OpError
	if errors.As(err, &netOpErr) {
		return true
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.Temporary() || dnsErr.Timeout()
	}

	// Check for net.Error interface (includes timeout failures)
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	return false
}

// IsHTTPRetryableError checks if an HTTP error is retryable based on status code
func IsHTTPRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Extract HTTP status code from error message
	errStr := err.Error()

	// Common retryable HTTP status patterns
	retryableHTTPPatterns := []string{
		"429", // Too Many Requests
		"500", // Internal Server Error
		"502", // Bad Gateway
		"503", // Service Unavailable
		"504", // Gateway Timeout
		"408", // Request Timeout
	}

	for _, pattern := range retryableHTTPPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// IsGitHubRetryableError checks if an error from GitHub API is retryable
func IsGitHubRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// First check generic HTTP and network errors
	if IsNetworkError(err) || IsHTTPRetryableError(err) {
		return true
	}

	errStr := strings.ToLower(err.Error())

	// GitHub-specific retryable error patterns
	githubRetryablePatterns := []string{
		"rate limit",
		"server error",
		"abuse detection",
		"temporarily unavailable",
		"service unavailable",
		"502", "503", "504", // Server errors
		"422", // Sometimes retryable for GitHub (rate limits)
	}

	for _, pattern := range githubRetryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// Don't retry client errors (4xx except specific ones)
	clientErrorPatterns := []string{
		"401", "403", "404", "400",
	}

	for _, pattern := range clientErrorPatterns {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	return false
}

// IsFileError checks if an error is a retryable file operation error
func IsFileError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// File operation errors that might be temporary
	retryableFilePatterns := []string{
		"resource temporarily unavailable",
		"file exists",       // Sometimes temporary for concurrent operations
		"permission denied", // Might be temporary
		"device or resource busy",
		"no space left on device",
		"disk full",
	}

	for _, pattern := range retryableFilePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// Check for temporary errors
	return IsNetworkError(err) // File operations over network
}

// CircuitBreaker provides circuit breaker functionality to prevent cascading failures
type CircuitBreaker struct {
	MaxFailures     int           // Maximum consecutive failures before opening circuit
	ResetTimeout    time.Duration // Time to wait before attempting to close circuit
	failureCount    int
	lastFailureTime time.Time
	state           circuitBreakerState
}

type circuitBreakerState int

const (
	circuitClosed circuitBreakerState = iota
	circuitOpen
	circuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		MaxFailures:  maxFailures,
		ResetTimeout: resetTimeout,
		state:        circuitClosed,
	}
}

// DefaultCircuitBreaker returns a circuit breaker with sensible defaults
func DefaultCircuitBreaker() *CircuitBreaker {
	return NewCircuitBreaker(5, 60*time.Second)
}

// Execute runs the given function through the circuit breaker
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if cb.state == circuitOpen {
		// Check if we should attempt to close the circuit
		if time.Since(cb.lastFailureTime) > cb.ResetTimeout {
			cb.state = circuitHalfOpen
		} else {
			return ErrCircuitBreakerOpen
		}
	}

	err := fn()
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= cb.MaxFailures {
		cb.state = circuitOpen
	}
}

// recordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) recordSuccess() {
	cb.failureCount = 0
	cb.state = circuitClosed
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() string {
	switch cb.state {
	case circuitClosed:
		return "closed"
	case circuitOpen:
		return "open"
	case circuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
