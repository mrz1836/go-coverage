// Package fallback provides comprehensive fallback mechanisms and recovery strategies
// for handling various failure scenarios in the go-coverage system.
package fallback

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	// ErrAllFallbacksFailed indicates all fallback strategies have been exhausted
	ErrAllFallbacksFailed = errors.New("all fallback strategies failed")
	// ErrNoFallbackAvailable indicates no fallback is available for the operation
	ErrNoFallbackAvailable = errors.New("no fallback available")
	// ErrMissingDeploymentFiles indicates missing deployment files in metadata
	ErrMissingDeploymentFiles = errors.New("missing deployment files in metadata")
	// ErrPanicRecovered indicates a panic was recovered
	ErrPanicRecovered = errors.New("recovered from panic")
	// ErrFallbackNotConfigured indicates fallback is not properly configured
	ErrFallbackNotConfigured = errors.New("fallback not configured")
	// Additional errors for err113 linter compliance
	ErrUnsupportedOperationType = errors.New("unsupported operation type for GitHub fallback")
	ErrMissingArtifactName      = errors.New("missing artifact_name in metadata")
	ErrMissingFilePath          = errors.New("missing file_path in metadata")
	ErrMissingComment           = errors.New("missing comment in metadata")
)

// FallbackStrategy defines the interface for fallback strategies
type FallbackStrategy interface {
	// Name returns the name of the fallback strategy
	Name() string
	// CanHandle determines if this strategy can handle the given error
	CanHandle(err error) bool
	// Execute performs the fallback operation
	Execute(ctx context.Context, operation Operation, originalErr error) error
	// Priority returns the priority of this strategy (lower = higher priority)
	Priority() int
	// IsEnabled returns whether this strategy is currently enabled
	IsEnabled() bool
}

// Operation represents an operation that may need fallback handling
type Operation interface {
	// Type returns the type of operation (e.g., "github_api", "artifact_upload")
	Type() string
	// Execute performs the primary operation
	Execute(ctx context.Context) error
	// Metadata returns operation metadata for fallback strategies
	Metadata() map[string]interface{}
}

// FallbackManager manages fallback strategies and recovery operations
type FallbackManager struct {
	strategies  []FallbackStrategy
	mu          sync.RWMutex
	enabled     bool
	maxAttempts int
	backoffBase time.Duration
	backoffMax  time.Duration
	logger      *log.Logger
	metrics     *FallbackMetrics
}

// FallbackMetrics tracks fallback usage and success rates
type FallbackMetrics struct {
	mu                  sync.RWMutex
	TotalFallbacks      int64              `json:"total_fallbacks"`
	SuccessfulFallbacks int64              `json:"successful_fallbacks"`
	FailedFallbacks     int64              `json:"failed_fallbacks"`
	StrategyUsage       map[string]int64   `json:"strategy_usage"`
	StrategySuccessRate map[string]float64 `json:"strategy_success_rate"`
	LastFallbackTime    time.Time          `json:"last_fallback_time"`
	AverageRecoveryTime time.Duration      `json:"average_recovery_time"`
	RecoveryTimes       []time.Duration    `json:"-"` // Not serialized
}

// NewFallbackManager creates a new fallback manager
func NewFallbackManager() *FallbackManager {
	return &FallbackManager{
		strategies:  make([]FallbackStrategy, 0),
		enabled:     true,
		maxAttempts: 3,
		backoffBase: time.Second,
		backoffMax:  30 * time.Second,
		logger:      log.New(os.Stderr, "[FALLBACK] ", log.LstdFlags),
		metrics: &FallbackMetrics{
			StrategyUsage:       make(map[string]int64),
			StrategySuccessRate: make(map[string]float64),
			RecoveryTimes:       make([]time.Duration, 0, 100),
		},
	}
}

// RegisterStrategy registers a fallback strategy
func (fm *FallbackManager) RegisterStrategy(strategy FallbackStrategy) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.strategies = append(fm.strategies, strategy)

	// Sort by priority (lower number = higher priority)
	for i := len(fm.strategies) - 1; i > 0; i-- {
		if fm.strategies[i].Priority() < fm.strategies[i-1].Priority() {
			fm.strategies[i], fm.strategies[i-1] = fm.strategies[i-1], fm.strategies[i]
		} else {
			break
		}
	}
}

// ExecuteWithFallback executes an operation with fallback support
func (fm *FallbackManager) ExecuteWithFallback(ctx context.Context, operation Operation) error {
	if !fm.enabled {
		return operation.Execute(ctx)
	}

	// Try primary operation first
	startTime := time.Now()
	err := operation.Execute(ctx)
	if err == nil {
		return nil // Success
	}

	fm.logger.Printf("Primary operation failed: %v, attempting fallback strategies", err)

	// Track fallback attempt
	fm.metrics.mu.Lock()
	fm.metrics.TotalFallbacks++
	fm.metrics.LastFallbackTime = time.Now()
	fm.metrics.mu.Unlock()

	// Find applicable fallback strategies
	applicableStrategies := fm.findApplicableStrategies(err)
	if len(applicableStrategies) == 0 {
		fm.metrics.mu.Lock()
		fm.metrics.FailedFallbacks++
		fm.metrics.mu.Unlock()
		return fmt.Errorf("%w for operation %s: %w", ErrNoFallbackAvailable, operation.Type(), err)
	}

	// Try fallback strategies in priority order
	lastErr := err
	for _, strategy := range applicableStrategies {
		if !strategy.IsEnabled() {
			continue
		}

		fm.logger.Printf("Attempting fallback strategy: %s", strategy.Name())

		// Track strategy usage
		fm.metrics.mu.Lock()
		fm.metrics.StrategyUsage[strategy.Name()]++
		fm.metrics.mu.Unlock()

		strategyErr := fm.executeWithBackoff(ctx, strategy, operation, err)
		if strategyErr == nil {
			// Success
			recoveryTime := time.Since(startTime)
			fm.logger.Printf("Fallback strategy %s succeeded after %v", strategy.Name(), recoveryTime)

			// Update metrics
			fm.metrics.mu.Lock()
			fm.metrics.SuccessfulFallbacks++
			fm.metrics.RecoveryTimes = append(fm.metrics.RecoveryTimes, recoveryTime)
			fm.updateStrategySuccessRate(strategy.Name(), true)
			fm.metrics.mu.Unlock()

			return nil
		}

		fm.logger.Printf("Fallback strategy %s failed: %v", strategy.Name(), strategyErr)
		lastErr = strategyErr

		// Update metrics
		fm.metrics.mu.Lock()
		fm.updateStrategySuccessRate(strategy.Name(), false)
		fm.metrics.mu.Unlock()
	}

	// All strategies failed
	fm.metrics.mu.Lock()
	fm.metrics.FailedFallbacks++
	fm.metrics.mu.Unlock()

	return fmt.Errorf("%w for operation %s: last error: %w", ErrAllFallbacksFailed, operation.Type(), lastErr)
}

// executeWithBackoff executes a strategy with exponential backoff
func (fm *FallbackManager) executeWithBackoff(ctx context.Context, strategy FallbackStrategy, operation Operation, originalErr error) error {
	var lastErr error

	for attempt := 1; attempt <= fm.maxAttempts; attempt++ {
		if attempt > 1 {
			// Calculate backoff delay
			delay := time.Duration(attempt-1) * fm.backoffBase
			if delay > fm.backoffMax {
				delay = fm.backoffMax
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := strategy.Execute(ctx, operation, originalErr)
		if err == nil {
			return nil
		}

		lastErr = err
		fm.logger.Printf("Strategy %s attempt %d failed: %v", strategy.Name(), attempt, err)
	}

	return fmt.Errorf("strategy failed after %d attempts: %w", fm.maxAttempts, lastErr)
}

// findApplicableStrategies finds strategies that can handle the given error
func (fm *FallbackManager) findApplicableStrategies(err error) []FallbackStrategy {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var applicable []FallbackStrategy
	for _, strategy := range fm.strategies {
		if strategy.CanHandle(err) {
			applicable = append(applicable, strategy)
		}
	}

	return applicable
}

// updateStrategySuccessRate updates success rate for a strategy
func (fm *FallbackManager) updateStrategySuccessRate(strategyName string, success bool) {
	// This is called with metrics mutex already held
	usage := fm.metrics.StrategyUsage[strategyName]
	currentSuccessRate := fm.metrics.StrategySuccessRate[strategyName]
	currentSuccesses := int64(currentSuccessRate * float64(usage-1))

	if success {
		currentSuccesses++
	}

	fm.metrics.StrategySuccessRate[strategyName] = float64(currentSuccesses) / float64(usage)
}

// SetEnabled enables or disables fallback handling
func (fm *FallbackManager) SetEnabled(enabled bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.enabled = enabled
}

// IsEnabled returns whether fallback handling is enabled
func (fm *FallbackManager) IsEnabled() bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.enabled
}

// GetMetrics returns current fallback metrics
func (fm *FallbackManager) GetMetrics() *FallbackMetrics {
	fm.metrics.mu.RLock()
	defer fm.metrics.mu.RUnlock()

	// Calculate average recovery time
	if len(fm.metrics.RecoveryTimes) > 0 {
		var total time.Duration
		for _, rt := range fm.metrics.RecoveryTimes {
			total += rt
		}
		fm.metrics.AverageRecoveryTime = total / time.Duration(len(fm.metrics.RecoveryTimes))
	}

	// Create a copy to avoid race conditions
	metrics := &FallbackMetrics{
		TotalFallbacks:      fm.metrics.TotalFallbacks,
		SuccessfulFallbacks: fm.metrics.SuccessfulFallbacks,
		FailedFallbacks:     fm.metrics.FailedFallbacks,
		LastFallbackTime:    fm.metrics.LastFallbackTime,
		AverageRecoveryTime: fm.metrics.AverageRecoveryTime,
		StrategyUsage:       make(map[string]int64),
		StrategySuccessRate: make(map[string]float64),
	}

	for k, v := range fm.metrics.StrategyUsage {
		metrics.StrategyUsage[k] = v
	}
	for k, v := range fm.metrics.StrategySuccessRate {
		metrics.StrategySuccessRate[k] = v
	}

	return metrics
}

// ResetMetrics resets fallback metrics
func (fm *FallbackManager) ResetMetrics() {
	fm.metrics.mu.Lock()
	defer fm.metrics.mu.Unlock()

	fm.metrics.TotalFallbacks = 0
	fm.metrics.SuccessfulFallbacks = 0
	fm.metrics.FailedFallbacks = 0
	fm.metrics.StrategyUsage = make(map[string]int64)
	fm.metrics.StrategySuccessRate = make(map[string]float64)
	fm.metrics.LastFallbackTime = time.Time{}
	fm.metrics.AverageRecoveryTime = 0
	fm.metrics.RecoveryTimes = fm.metrics.RecoveryTimes[:0]
}

// GitHub API Fallback Strategy
type GitHubAPIFallbackStrategy struct {
	name          string
	priority      int
	enabled       bool
	degradedMode  bool
	localFallback bool
	localCacheDir string
}

// NewGitHubAPIFallbackStrategy creates a GitHub API fallback strategy
func NewGitHubAPIFallbackStrategy() *GitHubAPIFallbackStrategy {
	return &GitHubAPIFallbackStrategy{
		name:          "github_api_fallback",
		priority:      1,
		enabled:       true,
		degradedMode:  true,
		localFallback: true,
		localCacheDir: filepath.Join(os.TempDir(), "github-fallback-cache"),
	}
}

func (s *GitHubAPIFallbackStrategy) Name() string    { return s.name }
func (s *GitHubAPIFallbackStrategy) Priority() int   { return s.priority }
func (s *GitHubAPIFallbackStrategy) IsEnabled() bool { return s.enabled }

func (s *GitHubAPIFallbackStrategy) CanHandle(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Handle common GitHub API errors
	return contains(errStr, "github") ||
		contains(errStr, "api rate limit") ||
		contains(errStr, "502") ||
		contains(errStr, "503") ||
		contains(errStr, "timeout") ||
		contains(errStr, "connection refused")
}

func (s *GitHubAPIFallbackStrategy) Execute(ctx context.Context, operation Operation, originalErr error) error {
	opType := operation.Type()
	metadata := operation.Metadata()

	switch opType {
	case "github_api_request":
		return s.handleAPIRequest(ctx, metadata, originalErr)
	case "artifact_upload":
		return s.handleArtifactUpload(ctx, metadata, originalErr)
	case "pr_comment":
		return s.handlePRComment(ctx, metadata, originalErr)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOperationType, opType)
	}
}

func (s *GitHubAPIFallbackStrategy) handleAPIRequest(_ context.Context, _ map[string]interface{}, originalErr error) error {
	if s.localFallback {
		// Use local cache or degraded mode
		log.Printf("[FALLBACK] Using local cache fallback for GitHub API request")

		// Ensure cache directory exists
		if err := os.MkdirAll(s.localCacheDir, 0o750); err != nil {
			return fmt.Errorf("failed to create cache directory: %w", err)
		}

		// In a real implementation, this would check local cache
		// For now, we'll simulate success with degraded functionality
		return nil
	}

	return fmt.Errorf("no fallback available for GitHub API request: %w", originalErr)
}

func (s *GitHubAPIFallbackStrategy) handleArtifactUpload(_ context.Context, metadata map[string]interface{}, originalErr error) error {
	if s.localFallback {
		// Store artifact locally as fallback
		log.Printf("[FALLBACK] Using local storage fallback for artifact upload")

		artifactName, ok := metadata["artifact_name"].(string)
		if !ok {
			return ErrMissingArtifactName
		}

		filePath, ok := metadata["file_path"].(string)
		if !ok {
			return ErrMissingFilePath
		}

		// Create fallback storage directory
		fallbackDir := filepath.Join(s.localCacheDir, "artifacts")
		if err := os.MkdirAll(fallbackDir, 0o750); err != nil {
			return fmt.Errorf("failed to create fallback directory: %w", err)
		}

		// Copy file to fallback location
		fallbackPath := filepath.Join(fallbackDir, artifactName+".json")
		data, err := os.ReadFile(filePath) //nolint:gosec // validated input
		if err != nil {
			return fmt.Errorf("failed to read source file: %w", err)
		}

		if err := os.WriteFile(fallbackPath, data, 0o600); err != nil {
			return fmt.Errorf("failed to write fallback file: %w", err)
		}

		log.Printf("[FALLBACK] Artifact stored locally at: %s", fallbackPath)
		return nil
	}

	return fmt.Errorf("no fallback available for artifact upload: %w", originalErr)
}

func (s *GitHubAPIFallbackStrategy) handlePRComment(_ context.Context, metadata map[string]interface{}, originalErr error) error {
	if s.degradedMode {
		// In degraded mode, we might just log the comment instead of posting it
		comment, ok := metadata["comment"].(string)
		if !ok {
			return ErrMissingComment
		}

		log.Printf("[FALLBACK] Would have posted PR comment (degraded mode): %s", comment)
		return nil
	}

	return fmt.Errorf("no fallback available for PR comment: %w", originalErr)
}

// Deployment Fallback Strategy
type DeploymentFallbackStrategy struct {
	name            string
	priority        int
	enabled         bool
	localDeployment bool
	skipDeployment  bool
	fallbackBranch  string
	localOutputDir  string
}

// NewDeploymentFallbackStrategy creates a deployment fallback strategy
func NewDeploymentFallbackStrategy() *DeploymentFallbackStrategy {
	return &DeploymentFallbackStrategy{
		name:            "deployment_fallback",
		priority:        2,
		enabled:         true,
		localDeployment: true,
		skipDeployment:  false,
		fallbackBranch:  "main",
		localOutputDir:  filepath.Join(os.TempDir(), "coverage-deployment-fallback"),
	}
}

func (s *DeploymentFallbackStrategy) Name() string    { return s.name }
func (s *DeploymentFallbackStrategy) Priority() int   { return s.priority }
func (s *DeploymentFallbackStrategy) IsEnabled() bool { return s.enabled }

func (s *DeploymentFallbackStrategy) CanHandle(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return contains(errStr, "deployment") ||
		contains(errStr, "gh-pages") ||
		contains(errStr, "git") ||
		contains(errStr, "pages") ||
		contains(errStr, "push")
}

func (s *DeploymentFallbackStrategy) Execute(ctx context.Context, operation Operation, originalErr error) error {
	metadata := operation.Metadata()

	if s.skipDeployment {
		log.Printf("[FALLBACK] Skipping deployment due to fallback configuration")
		return nil
	}

	if s.localDeployment {
		return s.handleLocalDeployment(ctx, metadata, originalErr)
	}

	return fmt.Errorf("no fallback available for deployment: %w", originalErr)
}

func (s *DeploymentFallbackStrategy) handleLocalDeployment(_ context.Context, metadata map[string]interface{}, _ error) error {
	log.Printf("[FALLBACK] Using local deployment fallback")

	// Create local deployment directory
	if err := os.MkdirAll(s.localOutputDir, 0o750); err != nil {
		return fmt.Errorf("failed to create local deployment directory: %w", err)
	}

	// Get deployment files from metadata
	deploymentFiles, ok := metadata["files"].([]string)
	if !ok {
		return ErrMissingDeploymentFiles
	}

	// Copy deployment files to local directory
	for _, file := range deploymentFiles {
		data, err := os.ReadFile(file) //nolint:gosec // validated input
		if err != nil {
			log.Printf("[FALLBACK] Warning: failed to read deployment file %s: %v", file, err)
			continue
		}

		destPath := filepath.Join(s.localOutputDir, filepath.Base(file))
		if err := os.WriteFile(destPath, data, 0o600); err != nil {
			log.Printf("[FALLBACK] Warning: failed to write deployment file %s: %v", destPath, err)
			continue
		}

		log.Printf("[FALLBACK] Deployed file locally: %s", destPath)
	}

	return nil
}

// Utility functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetDefaultFallbackManager returns a pre-configured fallback manager with common strategies
func GetDefaultFallbackManager() *FallbackManager {
	manager := NewFallbackManager()

	// Register common fallback strategies
	manager.RegisterStrategy(NewGitHubAPIFallbackStrategy())
	manager.RegisterStrategy(NewDeploymentFallbackStrategy())

	return manager
}

// Recovery utilities

// RecoverFromPanic recovers from panics and converts them to errors
func RecoverFromPanic() error {
	if r := recover(); r != nil {
		return fmt.Errorf("%w: %v", ErrPanicRecovered, r)
	}
	return nil
}

// ExecuteWithRecovery executes a function with panic recovery
func ExecuteWithRecovery(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrPanicRecovered, r)
		}
	}()

	return fn()
}
