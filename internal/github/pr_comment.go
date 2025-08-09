// Package github provides PR comment management for coverage reporting
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mrz1836/go-coverage/internal/logger"
)

// PRCommentManager handles intelligent PR comment management with anti-spam and lifecycle features
type PRCommentManager struct {
	client *Client
	config *PRCommentConfig
	logger logger.Logger
}

// PRCommentConfig holds configuration for PR comment management
type PRCommentConfig struct {
	// Anti-spam settings
	MinUpdateIntervalMinutes int    // Minimum time between comment updates
	MaxCommentsPerPR         int    // Maximum comments allowed per PR
	CommentSignature         string // Unique signature to identify our comments

	// Template settings
	IncludeTrend           bool // Include trend analysis in comments
	IncludeCoverageDetails bool // Include detailed coverage breakdown
	IncludeFileAnalysis    bool // Include file-level coverage analysis
	ShowCoverageHistory    bool // Show historical coverage data

	// Badge settings
	GeneratePRBadges bool   // Generate PR-specific badges
	BadgeStyle       string // Badge style (flat, flat-square, for-the-badge)

	// Status check settings
	EnableStatusChecks  bool // Enable GitHub status checks
	FailBelowThreshold  bool // Fail status if below threshold
	BlockMergeOnFailure bool // Block PR merge on coverage failure
}

// CoverageComparison represents coverage comparison between base and PR branches
type CoverageComparison struct {
	BaseCoverage     CoverageData    `json:"base_coverage"`
	PRCoverage       CoverageData    `json:"pr_coverage"`
	Difference       float64         `json:"difference"`
	TrendAnalysis    TrendData       `json:"trend_analysis"`
	FileChanges      []FileChange    `json:"file_changes"`
	SignificantFiles []string        `json:"significant_files"`
	PRFileAnalysis   *PRFileAnalysis `json:"pr_file_analysis,omitempty"`
}

// CoverageData represents coverage information for a specific commit
type CoverageData struct {
	Percentage        float64   `json:"percentage"`
	TotalStatements   int       `json:"total_statements"`
	CoveredStatements int       `json:"covered_statements"`
	CommitSHA         string    `json:"commit_sha"`
	Branch            string    `json:"branch"`
	Timestamp         time.Time `json:"timestamp"`
}

// TrendData represents trend analysis information
type TrendData struct {
	Direction        string  `json:"direction"` // "up", "down", "stable"
	Magnitude        string  `json:"magnitude"` // "significant", "moderate", "minor"
	PercentageChange float64 `json:"percentage_change"`
	Momentum         string  `json:"momentum"` // "accelerating", "steady", "decelerating"
}

// FileChange represents coverage change for a specific file
type FileChange struct {
	Filename      string  `json:"filename"`
	BaseCoverage  float64 `json:"base_coverage"`
	PRCoverage    float64 `json:"pr_coverage"`
	Difference    float64 `json:"difference"`
	LinesAdded    int     `json:"lines_added"`
	LinesRemoved  int     `json:"lines_removed"`
	IsSignificant bool    `json:"is_significant"`
}

// CommentMetadata represents metadata stored in comment for tracking
type CommentMetadata struct {
	Signature      string    `json:"signature"`
	CommentVersion string    `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	LastUpdatedAt  time.Time `json:"last_updated_at"`
	UpdateCount    int       `json:"update_count"`
	PRNumber       int       `json:"pr_number"`
	BaseSHA        string    `json:"base_sha"`
	HeadSHA        string    `json:"head_sha"`
}

// PRCommentResponse represents the response from creating/updating a PR comment
type PRCommentResponse struct {
	CommentID      int                `json:"comment_id"`
	Action         string             `json:"action"` // "created", "updated", "skipped"
	Reason         string             `json:"reason"` // Reason for action taken
	Metadata       CommentMetadata    `json:"metadata"`
	CoverageData   CoverageComparison `json:"coverage_data"`
	BadgeURLs      map[string]string  `json:"badge_urls"` // PR-specific badge URLs
	StatusCheckURL string             `json:"status_check_url"`
}

// NewPRCommentManager creates a new PR comment manager with configuration
func NewPRCommentManager(client *Client, config *PRCommentConfig) *PRCommentManager {
	if config == nil {
		config = &PRCommentConfig{
			MinUpdateIntervalMinutes: 5,
			MaxCommentsPerPR:         1,
			CommentSignature:         "gofortress-coverage-v1",
			IncludeTrend:             true,
			IncludeCoverageDetails:   true,
			IncludeFileAnalysis:      false,
			ShowCoverageHistory:      true,
			GeneratePRBadges:         true,
			BadgeStyle:               "flat",
			EnableStatusChecks:       true,
			FailBelowThreshold:       true,
			BlockMergeOnFailure:      false,
		}
	}

	return &PRCommentManager{
		client: client,
		config: config,
		logger: logger.NewFromEnv(),
	}
}

// CreateOrUpdatePRComment creates or updates a PR comment with coverage information
func (m *PRCommentManager) CreateOrUpdatePRComment(ctx context.Context, owner, repo string, prNumber int, commentBody string, comparison *CoverageComparison) (*PRCommentResponse, error) {
	// Get PR information first
	pr, err := m.client.GetPullRequest(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR information: %w", err)
	}

	// Find existing coverage comments
	existingComments, err := m.findExistingCoverageComments(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing comments: %w", err)
	}

	// Determine action based on anti-spam rules
	action, shouldUpdate, reason := m.determineCommentAction(existingComments, comparison)

	if !shouldUpdate {
		return &PRCommentResponse{
			Action:       action,
			Reason:       reason,
			CoverageData: *comparison,
		}, nil
	}

	var comment *Comment
	var commentID int

	if len(existingComments) > 0 {
		// Update existing comment
		comment, err = m.client.updateComment(ctx, owner, repo, existingComments[0].ID, commentBody)
		if err != nil {
			return nil, fmt.Errorf("failed to update comment: %w", err)
		}
		commentID = comment.ID
		action = "updated"
	} else {
		// Create new comment
		comment, err = m.client.createComment(ctx, owner, repo, prNumber, commentBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create comment: %w", err)
		}
		commentID = comment.ID
		action = "created"
	}

	// Badge URLs are now handled by the badge generation system separately
	badgeURLs := make(map[string]string)

	// Create status check if enabled
	statusCheckURL := ""
	if m.config.EnableStatusChecks {
		err = m.createCoverageStatusCheck(ctx, owner, repo, pr.Head.SHA, comparison)
		if err != nil {
			// Don't fail the entire operation if status check fails
			m.logger.WithError(err).WithFields(map[string]interface{}{
				"owner":     owner,
				"repo":      repo,
				"sha":       pr.Head.SHA,
				"operation": "status_check_creation",
			}).Warn("Failed to create GitHub status check")
		} else {
			statusCheckURL = fmt.Sprintf("https://github.com/%s/%s/commit/%s/checks", owner, repo, pr.Head.SHA)
		}
	}

	// Prepare metadata
	metadata := CommentMetadata{
		Signature:      m.config.CommentSignature,
		CommentVersion: "2.0",
		CreatedAt:      time.Now(),
		LastUpdatedAt:  time.Now(),
		UpdateCount:    1,
		PRNumber:       prNumber,
		BaseSHA:        "", // Would need to get base SHA from PR
		HeadSHA:        pr.Head.SHA,
	}

	if len(existingComments) > 0 {
		// Extract existing metadata if available
		if existingMeta := m.extractCommentMetadata(existingComments[0].Body); existingMeta != nil {
			metadata.CreatedAt = existingMeta.CreatedAt
			metadata.UpdateCount = existingMeta.UpdateCount + 1
		}
	}

	return &PRCommentResponse{
		CommentID:      commentID,
		Action:         action,
		Reason:         reason,
		Metadata:       metadata,
		CoverageData:   *comparison,
		BadgeURLs:      badgeURLs,
		StatusCheckURL: statusCheckURL,
	}, nil
}

// findExistingCoverageComments finds existing coverage comments by signature with retry logic
func (m *PRCommentManager) findExistingCoverageComments(ctx context.Context, owner, repo string, prNumber int) ([]Comment, error) {
	m.logger.Debug("Searching for existing coverage comments", map[string]interface{}{
		"owner":     owner,
		"repo":      repo,
		"pr_number": prNumber,
	})

	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", m.client.baseURL, owner, repo, prNumber)

	var allComments []Comment
	var lastErr error

	// Retry logic for robustness
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		m.logger.Debug("Attempting to fetch PR comments", map[string]interface{}{
			"attempt": attempt,
			"url":     url,
		})

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			m.logger.Error("Failed to create request", map[string]interface{}{
				"error":   err,
				"attempt": attempt,
			})
			continue
		}

		req.Header.Set("Authorization", "token "+m.client.token)
		req.Header.Set("User-Agent", m.client.config.UserAgent)

		resp, err := m.client.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to get comments: %w", err)
			m.logger.Error("Failed to execute request", map[string]interface{}{
				"error":   err,
				"attempt": attempt,
			})
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
			continue
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("%w: %d", ErrGitHubAPIError, resp.StatusCode)
			m.logger.Error("GitHub API returned error status", map[string]interface{}{
				"status_code": resp.StatusCode,
				"attempt":     attempt,
			})
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
			continue
		}

		if err := json.NewDecoder(resp.Body).Decode(&allComments); err != nil {
			lastErr = fmt.Errorf("failed to decode comments: %w", err)
			m.logger.Error("Failed to decode response", map[string]interface{}{
				"error":   err,
				"attempt": attempt,
			})
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
			continue
		}

		// Success - break out of retry loop
		lastErr = nil
		break
	}

	if lastErr != nil {
		m.logger.Error("All attempts to fetch comments failed", map[string]interface{}{
			"error":        lastErr,
			"max_attempts": maxRetries,
		})
		return nil, lastErr
	}

	m.logger.Info("Successfully fetched PR comments", map[string]interface{}{
		"total_comments": len(allComments),
	})

	// Filter for our coverage comments with detailed logging
	var coverageComments []Comment
	for i, comment := range allComments {
		isCoverage := m.isCoverageComment(comment.Body)
		m.logger.Debug("Checking comment", map[string]interface{}{
			"comment_id":  comment.ID,
			"comment_idx": i,
			"is_coverage": isCoverage,
			"created_at":  comment.CreatedAt,
			"updated_at":  comment.UpdatedAt,
			"body_preview": func() string {
				if len(comment.Body) > 100 {
					return comment.Body[:100] + "..."
				}
				return comment.Body
			}(),
		})

		if isCoverage {
			coverageComments = append(coverageComments, comment)
		}
	}

	m.logger.Info("Found coverage comments", map[string]interface{}{
		"coverage_comments": len(coverageComments),
		"total_comments":    len(allComments),
	})

	return coverageComments, nil
}

// isCoverageComment checks if a comment is our coverage comment by signature
func (m *PRCommentManager) isCoverageComment(body string) bool {
	signatures := []string{
		m.config.CommentSignature,
		"<!-- gofortress-coverage -->",
		"<!-- coverage-comment -->",
		"[//]: # (gofortress-coverage-v1)", // New Markdown comment format
		"[//]: # (gofortress-coverage)",    // Alternative format
		"<!-- gofortress-coverage-v1 -->",  // Legacy HTML format (may be stripped)
		"## 📊 Coverage Report",
		"Generated by GoFortress Coverage",
		"GoFortress Coverage",          // Match the text in the unified dashboard template
		"*Generated by 🏰 [GoFortress]", // Match the footer from PR comments
		"Generated by 🏰 GoFortress",    // Alternative footer format
		"🏰 [GoFortress]",               // Simplified footer match
		"📊 Coverage Report",            // Header without emoji variant
		"Overall Coverage:",            // Content-based detection
		"Coverage Metrics",             // Table header detection
	}

	// Log which signatures we're checking against for debugging
	m.logger.Debug("Checking comment signatures", map[string]interface{}{
		"signatures_count": len(signatures),
		"comment_length":   len(body),
	})

	for i, signature := range signatures {
		if strings.Contains(body, signature) {
			m.logger.Debug("Comment matched signature", map[string]interface{}{
				"signature_index": i,
				"signature":       signature,
			})
			return true
		}
	}

	m.logger.Debug("Comment did not match any signatures")
	return false
}

// determineCommentAction determines what action to take based on anti-spam rules
func (m *PRCommentManager) determineCommentAction(existingComments []Comment, comparison *CoverageComparison) (string, bool, string) {
	m.logger.Info("Determining comment action", map[string]interface{}{
		"existing_comments": len(existingComments),
		"max_comments":      m.config.MaxCommentsPerPR,
		"min_interval_min":  m.config.MinUpdateIntervalMinutes,
	})

	if len(existingComments) == 0 {
		m.logger.Info("No existing coverage comments found - will create new comment")
		return "create", true, "No existing coverage comment found"
	}

	if len(existingComments) > m.config.MaxCommentsPerPR {
		reason := fmt.Sprintf("Maximum comments per PR (%d) exceeded", m.config.MaxCommentsPerPR)
		m.logger.Warn("Skipping comment creation - too many existing comments", map[string]interface{}{
			"existing_comments": len(existingComments),
			"max_allowed":       m.config.MaxCommentsPerPR,
		})
		return "skipped", false, reason
	}

	// Check time-based anti-spam
	lastComment := existingComments[len(existingComments)-1]
	m.logger.Debug("Checking time-based anti-spam", map[string]interface{}{
		"last_comment_id":         lastComment.ID,
		"last_comment_updated_at": lastComment.UpdatedAt,
	})

	lastUpdateTime, err := time.Parse(time.RFC3339, lastComment.UpdatedAt)
	if err == nil {
		timeSinceUpdate := time.Since(lastUpdateTime)
		minInterval := time.Duration(m.config.MinUpdateIntervalMinutes) * time.Minute

		m.logger.Debug("Time since last update", map[string]interface{}{
			"time_since_update": timeSinceUpdate.String(),
			"min_interval":      minInterval.String(),
			"should_wait":       timeSinceUpdate < minInterval,
		})

		if timeSinceUpdate < minInterval {
			reason := fmt.Sprintf("Minimum update interval (%v) not reached", minInterval)
			m.logger.Info("Skipping comment update - minimum interval not reached", map[string]interface{}{
				"time_since_update": timeSinceUpdate.String(),
				"min_interval":      minInterval.String(),
			})
			return "skipped", false, reason
		}
	} else {
		m.logger.Warn("Failed to parse last comment update time", map[string]interface{}{
			"error":      err,
			"updated_at": lastComment.UpdatedAt,
		})
	}

	// Check for significant changes
	if m.hasSignificantCoverageChange(comparison) {
		m.logger.Info("Significant coverage change detected - will update comment")
		return "update", true, "Significant coverage change detected"
	}

	m.logger.Info("No significant changes but will update comment anyway")
	return "update", true, "Coverage data updated"
}

// hasSignificantCoverageChange determines if the coverage change is significant enough to warrant an update
func (m *PRCommentManager) hasSignificantCoverageChange(comparison *CoverageComparison) bool {
	// Consider changes significant if:
	// 1. Coverage difference > 1%
	// 2. Trend direction changed
	// 3. New files with low coverage

	if comparison.Difference > 1.0 || comparison.Difference < -1.0 {
		return true
	}

	if comparison.TrendAnalysis.Magnitude == "significant" {
		return true
	}

	for _, fileChange := range comparison.FileChanges {
		if fileChange.IsSignificant {
			return true
		}
	}

	return false
}

// createCoverageStatusCheck creates GitHub status check for coverage
func (m *PRCommentManager) createCoverageStatusCheck(ctx context.Context, owner, repo, sha string, comparison *CoverageComparison) error {
	var state string
	var description string

	threshold := 80.0 // Default threshold, should come from config

	if comparison.PRCoverage.Percentage >= threshold {
		state = StatusSuccess
		description = fmt.Sprintf("Coverage: %.1f%% ✅", comparison.PRCoverage.Percentage)
	} else if m.config.FailBelowThreshold {
		state = StatusFailure
		description = fmt.Sprintf("Coverage: %.1f%% (below %.1f%% threshold)",
			comparison.PRCoverage.Percentage, threshold)
	} else {
		state = StatusSuccess
		description = fmt.Sprintf("Coverage: %.1f%% (below threshold but not blocking)",
			comparison.PRCoverage.Percentage)
	}

	statusReq := &StatusRequest{
		State:       state,
		TargetURL:   fmt.Sprintf("https://%s.github.io/%s/coverage/", owner, repo),
		Description: description,
		Context:     "GoFortress/Coverage-PR",
	}

	return m.client.CreateStatus(ctx, owner, repo, sha, statusReq)
}

// extractCommentMetadata extracts metadata from comment body
func (m *PRCommentManager) extractCommentMetadata(body string) *CommentMetadata {
	re := regexp.MustCompile(`<!-- metadata: (.*?) -->`)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return nil
	}

	var metadata CommentMetadata
	if err := json.Unmarshal([]byte(matches[1]), &metadata); err != nil {
		return nil
	}

	return &metadata
}

// DeletePRComments deletes all coverage comments for a PR (cleanup utility)
func (m *PRCommentManager) DeletePRComments(ctx context.Context, owner, repo string, prNumber int) error {
	existingComments, err := m.findExistingCoverageComments(ctx, owner, repo, prNumber)
	if err != nil {
		return fmt.Errorf("failed to find existing comments: %w", err)
	}

	for _, comment := range existingComments {
		url := fmt.Sprintf("%s/repos/%s/%s/issues/comments/%d", m.client.baseURL, owner, repo, comment.ID)

		req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
		if err != nil {
			continue // Skip this comment if request creation fails
		}

		req.Header.Set("Authorization", "token "+m.client.token)
		req.Header.Set("User-Agent", m.client.config.UserAgent)

		resp, err := m.client.httpClient.Do(req)
		if err != nil {
			continue // Skip this comment if deletion fails
		}
		_ = resp.Body.Close()
	}

	return nil
}

// GetPRCommentStats returns statistics about PR comments
func (m *PRCommentManager) GetPRCommentStats(ctx context.Context, owner, repo string, prNumber int) (map[string]interface{}, error) {
	existingComments, err := m.findExistingCoverageComments(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing comments: %w", err)
	}

	stats := map[string]interface{}{
		"total_comments":    len(existingComments),
		"has_comments":      len(existingComments) > 0,
		"last_update_time":  "",
		"comment_signature": m.config.CommentSignature,
	}

	if len(existingComments) > 0 {
		lastComment := existingComments[len(existingComments)-1]
		stats["last_update_time"] = lastComment.UpdatedAt
		stats["last_comment_id"] = lastComment.ID

		if metadata := m.extractCommentMetadata(lastComment.Body); metadata != nil {
			stats["update_count"] = metadata.UpdateCount
			stats["created_at"] = metadata.CreatedAt
		}
	}

	return stats, nil
}
