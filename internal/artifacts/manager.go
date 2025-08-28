// Package artifacts provides GitHub artifact-based history management for coverage data.
// This package enables tracking coverage over time using GitHub Actions artifacts
// as storage, eliminating the need for external storage dependencies.
package artifacts

import (
	"context"
	"time"

	"github.com/mrz1836/go-coverage/internal/history"
)

// ArtifactManager defines the interface for managing coverage history via GitHub artifacts
type ArtifactManager interface {
	// DownloadHistory downloads the most recent coverage history from GitHub artifacts
	DownloadHistory(ctx context.Context, opts *DownloadOptions) (*History, error)

	// MergeHistory merges current coverage data with previous history
	MergeHistory(current, previous *History) (*History, error)

	// UploadHistory uploads the updated history as a GitHub artifact
	UploadHistory(ctx context.Context, history *History, opts *UploadOptions) error

	// CleanupOldArtifacts removes expired artifacts based on retention policy
	CleanupOldArtifacts(ctx context.Context, retentionDays int) error

	// ListArtifacts lists available coverage history artifacts
	ListArtifacts(ctx context.Context, opts *ListOptions) ([]*ArtifactInfo, error)
}

// History represents a collection of coverage records with metadata
type History struct {
	Records  []history.CoverageRecord `json:"records"`
	Metadata *HistoryMetadata         `json:"metadata"`
}

// HistoryMetadata contains metadata about the coverage history
type HistoryMetadata struct {
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Repository  string    `json:"repository"`
	TotalSize   int64     `json:"total_size"`
	RecordCount int       `json:"record_count"`
}

// ArtifactInfo contains information about a coverage history artifact
type ArtifactInfo struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Branch      string    `json:"branch,omitempty"`
	CommitSHA   string    `json:"commit_sha,omitempty"`
	PRNumber    string    `json:"pr_number,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Size        int64     `json:"size"`
	DownloadURL string    `json:"download_url,omitempty"`
}

// DownloadOptions configures artifact download behavior
type DownloadOptions struct {
	// Branch specifies the preferred branch for history artifacts
	Branch string

	// MaxRuns limits the number of workflow runs to check for artifacts
	MaxRuns int

	// FallbackToBranch enables fallback to main/master branch if branch history not found
	FallbackToBranch string

	// MaxAge limits the age of artifacts to consider (in hours)
	MaxAge time.Duration
}

// UploadOptions configures artifact upload behavior
type UploadOptions struct {
	// Name specifies the artifact name (auto-generated if empty)
	Name string

	// Branch is the current branch name
	Branch string

	// CommitSHA is the current commit SHA
	CommitSHA string

	// PRNumber is the pull request number (if applicable)
	PRNumber string

	// RetentionDays specifies how long to keep this artifact
	RetentionDays int
}

// ListOptions configures artifact listing behavior
type ListOptions struct {
	// Branch filters artifacts by branch name
	Branch string

	// Limit limits the number of artifacts to return
	Limit int

	// IncludeExpired includes expired artifacts in the results
	IncludeExpired bool
}

// DefaultDownloadOptions returns default download configuration
func DefaultDownloadOptions() *DownloadOptions {
	return &DownloadOptions{
		MaxRuns:          8,
		FallbackToBranch: "main",
		MaxAge:           24 * 7 * time.Hour, // 1 week
	}
}

// DefaultUploadOptions returns default upload configuration
func DefaultUploadOptions() *UploadOptions {
	return &UploadOptions{
		RetentionDays: 30,
	}
}

// DefaultListOptions returns default listing configuration
func DefaultListOptions() *ListOptions {
	return &ListOptions{
		Limit:          50,
		IncludeExpired: false,
	}
}
