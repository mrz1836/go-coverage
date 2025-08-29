package deployment

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Static error definitions for git operations
var (
	ErrDeploymentLockTimeout = errors.New("timeout acquiring deployment lock")
	ErrGitCommandFailed      = errors.New("git command failed")
)

// GitOperations defines the interface for git operations on the gh-pages branch
type GitOperations interface {
	// CloneOrCreateBranch clones the gh-pages branch or creates it if it doesn't exist
	CloneOrCreateBranch(ctx context.Context, workDir string) error

	// CommitChanges commits changes to the gh-pages branch with the specified message
	CommitChanges(ctx context.Context, workDir, message string) (string, error)

	// PushChanges pushes the committed changes to the remote gh-pages branch
	PushChanges(ctx context.Context, workDir string, force bool) error

	// CreateBackup creates a backup reference for rollback purposes
	CreateBackup(ctx context.Context, workDir string) (string, error)

	// Rollback reverts to a previous commit or backup reference
	Rollback(ctx context.Context, workDir, backupRef string) error

	// AcquireLock attempts to acquire a deployment lock
	AcquireLock(ctx context.Context, lockName string, timeout time.Duration) error

	// ReleaseLock releases a deployment lock
	ReleaseLock(ctx context.Context, lockName string) error

	// GetCurrentCommit returns the current commit SHA
	GetCurrentCommit(ctx context.Context, workDir string) (string, error)
}

// GitClient is the concrete implementation of GitOperations
type GitClient struct {
	repository string
	token      string
	userEmail  string
	userName   string
}

// NewGitClient creates a new git client for deployment operations
func NewGitClient(repository, token string) *GitClient {
	return &GitClient{
		repository: repository,
		token:      token,
		userEmail:  "action@github.com",
		userName:   "GitHub Action",
	}
}

// CloneOrCreateBranch clones the gh-pages branch or creates it if it doesn't exist
func (g *GitClient) CloneOrCreateBranch(ctx context.Context, workDir string) error {
	// Ensure work directory exists
	if err := os.MkdirAll(workDir, 0o750); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// Configure git user
	if err := g.configureGit(ctx, workDir); err != nil {
		return fmt.Errorf("failed to configure git: %w", err)
	}

	// Try to clone existing gh-pages branch
	cloneURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", g.token, g.repository)
	// #nosec G204 - cloneURL is constructed from validated inputs (token and repository)
	cloneCmd := exec.CommandContext(ctx, "git", "clone", "--branch", "gh-pages", "--single-branch", cloneURL, ".")
	cloneCmd.Dir = workDir

	if err := cloneCmd.Run(); err != nil {
		// Branch doesn't exist, create it
		return g.createNewBranch(ctx, workDir, cloneURL)
	}

	return nil
}

// createNewBranch creates a new gh-pages branch
func (g *GitClient) createNewBranch(ctx context.Context, workDir, cloneURL string) error {
	// Initialize new git repository
	if err := g.runGitCommand(ctx, workDir, "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Add remote origin
	if err := g.runGitCommand(ctx, workDir, "remote", "add", "origin", cloneURL); err != nil {
		return fmt.Errorf("failed to add remote origin: %w", err)
	}

	// Create orphaned gh-pages branch
	if err := g.runGitCommand(ctx, workDir, "checkout", "--orphan", "gh-pages"); err != nil {
		return fmt.Errorf("failed to create orphaned gh-pages branch: %w", err)
	}

	// Create .nojekyll file for GitHub Pages
	nojekyllPath := filepath.Join(workDir, ".nojekyll")
	if err := os.WriteFile(nojekyllPath, []byte{}, 0o600); err != nil {
		return fmt.Errorf("failed to create .nojekyll file: %w", err)
	}

	// Add and commit .nojekyll
	if err := g.runGitCommand(ctx, workDir, "add", ".nojekyll"); err != nil {
		return fmt.Errorf("failed to add .nojekyll: %w", err)
	}

	if _, err := g.CommitChanges(ctx, workDir, "Initialize gh-pages branch"); err != nil {
		return fmt.Errorf("failed to commit initial gh-pages setup: %w", err)
	}

	return nil
}

// CommitChanges commits changes to the gh-pages branch
func (g *GitClient) CommitChanges(ctx context.Context, workDir, message string) (string, error) {
	// Add all changes
	if err := g.runGitCommand(ctx, workDir, "add", "."); err != nil {
		return "", fmt.Errorf("failed to add changes: %w", err)
	}

	// Check if there are changes to commit
	statusCmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	statusCmd.Dir = workDir
	if statusCmd.Run() == nil {
		// No changes to commit
		return g.GetCurrentCommit(ctx, workDir)
	}

	// Commit changes
	if err := g.runGitCommand(ctx, workDir, "commit", "-m", message); err != nil {
		return "", fmt.Errorf("failed to commit changes: %w", err)
	}

	return g.GetCurrentCommit(ctx, workDir)
}

// PushChanges pushes committed changes to the remote gh-pages branch
func (g *GitClient) PushChanges(ctx context.Context, workDir string, force bool) error {
	args := []string{"push", "origin", "gh-pages"}
	if force {
		args = append(args[:2], "--force", "origin", "gh-pages")
	}

	// Retry push operation up to 3 times
	for i := 0; i < 3; i++ {
		if err := g.runGitCommand(ctx, workDir, args...); err != nil {
			if i == 2 { // Last attempt
				return fmt.Errorf("failed to push after %d attempts: %w", i+1, err)
			}
			// Wait before retry
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		return nil
	}

	return nil
}

// CreateBackup creates a backup reference for rollback purposes
func (g *GitClient) CreateBackup(ctx context.Context, workDir string) (string, error) {
	currentCommit, err := g.GetCurrentCommit(ctx, workDir)
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}

	backupRef := fmt.Sprintf("refs/backup/deployment-%d", time.Now().Unix())
	if err := g.runGitCommand(ctx, workDir, "update-ref", backupRef, currentCommit); err != nil {
		return "", fmt.Errorf("failed to create backup reference: %w", err)
	}

	return backupRef, nil
}

// Rollback reverts to a previous commit or backup reference
func (g *GitClient) Rollback(ctx context.Context, workDir, backupRef string) error {
	// Reset to backup reference
	if err := g.runGitCommand(ctx, workDir, "reset", "--hard", backupRef); err != nil {
		return fmt.Errorf("failed to reset to backup: %w", err)
	}

	// Force push the rollback
	if err := g.PushChanges(ctx, workDir, true); err != nil {
		return fmt.Errorf("failed to push rollback: %w", err)
	}

	return nil
}

// AcquireLock attempts to acquire a deployment lock using git references
func (g *GitClient) AcquireLock(ctx context.Context, lockName string, timeout time.Duration) error {
	// For now, implement a simple file-based lock
	// In a production system, this could use git references or external systems
	lockFile := filepath.Join(os.TempDir(), fmt.Sprintf("deployment-lock-%s", lockName))

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(lockFile); os.IsNotExist(err) {
			// Create lock file
			// #nosec G304 - lockFile is constructed from os.TempDir() and validated lockName
			file, err := os.Create(lockFile)
			if err != nil {
				return fmt.Errorf("failed to create lock file: %w", err)
			}
			_ = file.Close()
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			continue
		}
	}

	return ErrDeploymentLockTimeout
}

// ReleaseLock releases a deployment lock
func (g *GitClient) ReleaseLock(ctx context.Context, lockName string) error {
	lockFile := filepath.Join(os.TempDir(), fmt.Sprintf("deployment-lock-%s", lockName))
	if err := os.Remove(lockFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	return nil
}

// GetCurrentCommit returns the current commit SHA
func (g *GitClient) GetCurrentCommit(ctx context.Context, workDir string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = workDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// configureGit configures git user settings
func (g *GitClient) configureGit(ctx context.Context, workDir string) error {
	if err := g.runGitCommand(ctx, workDir, "config", "user.email", g.userEmail); err != nil {
		return err
	}
	if err := g.runGitCommand(ctx, workDir, "config", "user.name", g.userName); err != nil {
		return err
	}
	return nil
}

// runGitCommand runs a git command with error handling
func (g *GitClient) runGitCommand(ctx context.Context, workDir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %w, output: %s", ErrGitCommandFailed, err, string(output))
	}

	return nil
}
