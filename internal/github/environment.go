package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Static error definitions
var (
	ErrNotGitHubActions       = errors.New("not running in GitHub Actions environment")
	ErrEventPathNotSet        = errors.New("GITHUB_EVENT_PATH not set")
	ErrNoPullRequestNumber    = errors.New("no pull request number found in event payload")
	ErrMissingEnvironmentVars = errors.New("missing required environment variables")
)

// GitHubContext holds information about the GitHub Actions environment
type GitHubContext struct {
	IsGitHubActions bool   `json:"is_github_actions"`
	Repository      string `json:"repository"`
	Branch          string `json:"branch"`
	CommitSHA       string `json:"commit_sha"`
	PRNumber        string `json:"pr_number"`
	EventName       string `json:"event_name"`
	RunID           string `json:"run_id"`
	Token           string `json:"token"`
}

// GitHubEvent represents the structure of GitHub webhook event payloads
type GitHubEvent struct {
	PullRequest *struct {
		Number int `json:"number"`
	} `json:"pull_request"`
}

// DetectEnvironment detects and returns information about the GitHub Actions environment
func DetectEnvironment() (*GitHubContext, error) {
	ctx := &GitHubContext{}

	// Check if we're in GitHub Actions
	ctx.IsGitHubActions = os.Getenv("GITHUB_ACTIONS") == "true"

	// Get basic GitHub environment information
	ctx.Repository = os.Getenv("GITHUB_REPOSITORY")
	ctx.Branch = getBranch()
	ctx.CommitSHA = os.Getenv("GITHUB_SHA")
	ctx.EventName = os.Getenv("GITHUB_EVENT_NAME")
	ctx.RunID = os.Getenv("GITHUB_RUN_ID")
	ctx.Token = os.Getenv("GITHUB_TOKEN")

	// Extract PR number if in pull request context
	if ctx.EventName == "pull_request" || ctx.EventName == "pull_request_target" {
		prNumber, err := extractPRNumber()
		if err != nil {
			// Log warning but don't fail - PR number might be available via other means
			fmt.Fprintf(os.Stderr, "Warning: Failed to extract PR number: %v\n", err)
		}
		ctx.PRNumber = prNumber
	}

	return ctx, nil
}

// getBranch gets the current branch name from environment variables
func getBranch() string {
	// GITHUB_REF_NAME contains the branch name directly for push events
	if branch := os.Getenv("GITHUB_REF_NAME"); branch != "" {
		return branch
	}

	// Fallback: extract branch from GITHUB_REF (refs/heads/branch-name)
	if ref := os.Getenv("GITHUB_REF"); ref != "" {
		if len(ref) > 11 && ref[:11] == "refs/heads/" {
			return ref[11:]
		}
		if len(ref) > 10 && ref[:10] == "refs/tags/" {
			return ref[10:]
		}
	}

	return ""
}

// extractPRNumber extracts the pull request number from the GitHub event payload
func extractPRNumber() (string, error) {
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return "", ErrEventPathNotSet
	}

	// Read the event payload file
	// #nosec G304 - eventPath is controlled by GitHub Actions and comes from GITHUB_EVENT_PATH env var
	eventData, err := os.ReadFile(eventPath)
	if err != nil {
		return "", fmt.Errorf("failed to read event payload: %w", err)
	}

	// Parse the JSON payload
	var event GitHubEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return "", fmt.Errorf("failed to parse event payload: %w", err)
	}

	// Extract PR number
	if event.PullRequest != nil && event.PullRequest.Number > 0 {
		return strconv.Itoa(event.PullRequest.Number), nil
	}

	return "", ErrNoPullRequestNumber
}

// IsGitHubActions returns true if running in GitHub Actions environment
func IsGitHubActions() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

// IsPullRequest returns true if the current event is a pull request
func IsPullRequest() bool {
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	return eventName == "pull_request" || eventName == "pull_request_target"
}

// GetRepository returns the repository in owner/name format
func GetRepository() string {
	return os.Getenv("GITHUB_REPOSITORY")
}

// GetCommitSHA returns the current commit SHA
func GetCommitSHA() string {
	return os.Getenv("GITHUB_SHA")
}

// GetToken returns the GitHub token from environment
func GetToken() string {
	return os.Getenv("GITHUB_TOKEN")
}

// ValidateEnvironment checks if the required GitHub Actions environment variables are present
func ValidateEnvironment() error {
	if !IsGitHubActions() {
		return ErrNotGitHubActions
	}

	required := map[string]string{
		"GITHUB_REPOSITORY": GetRepository(),
		"GITHUB_SHA":        GetCommitSHA(),
		"GITHUB_TOKEN":      GetToken(),
	}

	var missing []string
	for name, value := range required {
		if value == "" {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: %v", ErrMissingEnvironmentVars, missing)
	}

	return nil
}
