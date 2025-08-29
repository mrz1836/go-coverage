// Package fallback - Operation implementations for use with fallback manager
package fallback

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	// ErrUploadNotConfigured indicates upload function is not configured
	ErrUploadNotConfigured = errors.New("upload function not configured")
	// ErrDeploymentNotConfigured indicates deployment function is not configured
	ErrDeploymentNotConfigured = errors.New("deployment function not configured")
	// ErrUnsupportedFileOperation indicates unsupported file operation
	ErrUnsupportedFileOperation = errors.New("unsupported file operation")
	// ErrNoDataToWrite indicates no data to write
	ErrNoDataToWrite = errors.New("no data to write")
	// ErrTargetPathNotSpecified indicates target path not specified
	ErrTargetPathNotSpecified = errors.New("target path not specified")
	// ErrGitHubAPIError indicates GitHub API returned an error
	ErrGitHubAPIError = errors.New("github api returned error")
)

// GitHubAPIOperation represents a GitHub API operation
type GitHubAPIOperation struct {
	operationType string
	url           string
	method        string
	token         string
	body          io.Reader
	headers       map[string]string
	metadata      map[string]interface{}
	client        *http.Client
}

// NewGitHubAPIOperation creates a new GitHub API operation
func NewGitHubAPIOperation(method, url, token string) *GitHubAPIOperation {
	return &GitHubAPIOperation{
		operationType: "github_api_request",
		method:        method,
		url:           url,
		token:         token,
		headers:       make(map[string]string),
		metadata:      make(map[string]interface{}),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (op *GitHubAPIOperation) Type() string {
	return op.operationType
}

func (op *GitHubAPIOperation) Metadata() map[string]interface{} {
	// Add request details to metadata
	op.metadata["method"] = op.method
	op.metadata["url"] = op.url
	op.metadata["has_token"] = op.token != ""
	return op.metadata
}

func (op *GitHubAPIOperation) SetBody(body io.Reader) *GitHubAPIOperation {
	op.body = body
	return op
}

func (op *GitHubAPIOperation) SetHeader(key, value string) *GitHubAPIOperation {
	op.headers[key] = value
	return op
}

func (op *GitHubAPIOperation) SetMetadata(key string, value interface{}) *GitHubAPIOperation {
	op.metadata[key] = value
	return op
}

func (op *GitHubAPIOperation) Execute(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, op.method, op.url, op.body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header
	if op.token != "" {
		req.Header.Set("Authorization", "token "+op.token)
	}

	// Set custom headers
	for k, v := range op.headers {
		req.Header.Set(k, v)
	}

	// Set standard headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "go-coverage/1.0")

	resp, err := op.client.Do(req)
	if err != nil {
		return fmt.Errorf("github api request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: %d %s: %s", ErrGitHubAPIError, resp.StatusCode, resp.Status, string(body))
	}

	return nil
}

// ArtifactUploadOperation represents an artifact upload operation
type ArtifactUploadOperation struct {
	operationType string
	artifactName  string
	filePath      string
	retentionDays int
	metadata      map[string]interface{}
	uploadFunc    func(ctx context.Context, name, path string, retention int) error
}

// NewArtifactUploadOperation creates a new artifact upload operation
func NewArtifactUploadOperation(artifactName, filePath string, retentionDays int, uploadFunc func(ctx context.Context, name, path string, retention int) error) *ArtifactUploadOperation {
	return &ArtifactUploadOperation{
		operationType: "artifact_upload",
		artifactName:  artifactName,
		filePath:      filePath,
		retentionDays: retentionDays,
		metadata:      make(map[string]interface{}),
		uploadFunc:    uploadFunc,
	}
}

func (op *ArtifactUploadOperation) Type() string {
	return op.operationType
}

func (op *ArtifactUploadOperation) Metadata() map[string]interface{} {
	op.metadata["artifact_name"] = op.artifactName
	op.metadata["file_path"] = op.filePath
	op.metadata["retention_days"] = op.retentionDays

	// Add file information
	if info, err := os.Stat(op.filePath); err == nil {
		op.metadata["file_size"] = info.Size()
		op.metadata["file_mode"] = info.Mode().String()
		op.metadata["file_modified"] = info.ModTime()
	}

	return op.metadata
}

func (op *ArtifactUploadOperation) SetMetadata(key string, value interface{}) *ArtifactUploadOperation {
	op.metadata[key] = value
	return op
}

func (op *ArtifactUploadOperation) Execute(ctx context.Context) error {
	if op.uploadFunc == nil {
		return ErrUploadNotConfigured
	}

	// Validate file exists
	if _, err := os.Stat(op.filePath); err != nil {
		return fmt.Errorf("artifact file not found: %w", err)
	}

	return op.uploadFunc(ctx, op.artifactName, op.filePath, op.retentionDays)
}

// PRCommentOperation represents a PR comment operation
type PRCommentOperation struct {
	operationType string
	owner         string
	repo          string
	prNumber      int
	comment       string
	token         string
	metadata      map[string]interface{}
	client        *http.Client
}

// NewPRCommentOperation creates a new PR comment operation
func NewPRCommentOperation(owner, repo string, prNumber int, comment, token string) *PRCommentOperation {
	return &PRCommentOperation{
		operationType: "pr_comment",
		owner:         owner,
		repo:          repo,
		prNumber:      prNumber,
		comment:       comment,
		token:         token,
		metadata:      make(map[string]interface{}),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (op *PRCommentOperation) Type() string {
	return op.operationType
}

func (op *PRCommentOperation) Metadata() map[string]interface{} {
	op.metadata["owner"] = op.owner
	op.metadata["repo"] = op.repo
	op.metadata["pr_number"] = op.prNumber
	op.metadata["comment"] = op.comment
	op.metadata["comment_length"] = len(op.comment)
	return op.metadata
}

func (op *PRCommentOperation) SetMetadata(key string, value interface{}) *PRCommentOperation {
	op.metadata[key] = value
	return op
}

func (op *PRCommentOperation) Execute(ctx context.Context) error {
	// This would normally use the GitHub API client
	// For demonstration, we'll simulate the API call
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", op.owner, op.repo, op.prNumber)

	apiOp := NewGitHubAPIOperation("POST", url, op.token)
	apiOp.SetHeader("Content-Type", "application/json")
	apiOp.SetMetadata("comment", op.comment)

	return apiOp.Execute(ctx)
}

// DeploymentOperation represents a deployment operation
type DeploymentOperation struct {
	operationType string
	files         []string
	targetBranch  string
	commitMessage string
	workDir       string
	metadata      map[string]interface{}
	deployFunc    func(ctx context.Context, files []string, branch, message, workDir string) error
}

// NewDeploymentOperation creates a new deployment operation
func NewDeploymentOperation(files []string, targetBranch, commitMessage, workDir string, deployFunc func(ctx context.Context, files []string, branch, message, workDir string) error) *DeploymentOperation {
	return &DeploymentOperation{
		operationType: "deployment",
		files:         files,
		targetBranch:  targetBranch,
		commitMessage: commitMessage,
		workDir:       workDir,
		metadata:      make(map[string]interface{}),
		deployFunc:    deployFunc,
	}
}

func (op *DeploymentOperation) Type() string {
	return op.operationType
}

func (op *DeploymentOperation) Metadata() map[string]interface{} {
	op.metadata["files"] = op.files
	op.metadata["target_branch"] = op.targetBranch
	op.metadata["commit_message"] = op.commitMessage
	op.metadata["work_dir"] = op.workDir
	op.metadata["file_count"] = len(op.files)

	// Add file size information
	var totalSize int64
	for _, file := range op.files {
		if info, err := os.Stat(file); err == nil {
			totalSize += info.Size()
		}
	}
	op.metadata["total_size"] = totalSize

	return op.metadata
}

func (op *DeploymentOperation) SetMetadata(key string, value interface{}) *DeploymentOperation {
	op.metadata[key] = value
	return op
}

func (op *DeploymentOperation) Execute(ctx context.Context) error {
	if op.deployFunc == nil {
		return ErrDeploymentNotConfigured
	}

	// Validate files exist
	for _, file := range op.files {
		if _, err := os.Stat(file); err != nil {
			return fmt.Errorf("deployment file not found: %s: %w", file, err)
		}
	}

	// Validate work directory
	if op.workDir != "" {
		if _, err := os.Stat(op.workDir); err != nil {
			return fmt.Errorf("work directory not found: %w", err)
		}
	}

	return op.deployFunc(ctx, op.files, op.targetBranch, op.commitMessage, op.workDir)
}

// FileOperation represents a generic file operation
type FileOperation struct {
	operationType string
	filePath      string
	operation     string // "read", "write", "delete", "copy", "move"
	data          []byte
	targetPath    string
	metadata      map[string]interface{}
}

// NewFileOperation creates a new file operation
func NewFileOperation(operation, filePath string) *FileOperation {
	return &FileOperation{
		operationType: "file_operation",
		operation:     operation,
		filePath:      filePath,
		metadata:      make(map[string]interface{}),
	}
}

func (op *FileOperation) Type() string {
	return op.operationType
}

func (op *FileOperation) Metadata() map[string]interface{} {
	op.metadata["operation"] = op.operation
	op.metadata["file_path"] = op.filePath
	if op.targetPath != "" {
		op.metadata["target_path"] = op.targetPath
	}
	if op.data != nil {
		op.metadata["data_size"] = len(op.data)
	}
	return op.metadata
}

func (op *FileOperation) SetData(data []byte) *FileOperation {
	op.data = data
	return op
}

func (op *FileOperation) SetTargetPath(path string) *FileOperation {
	op.targetPath = path
	return op
}

func (op *FileOperation) SetMetadata(key string, value interface{}) *FileOperation {
	op.metadata[key] = value
	return op
}

func (op *FileOperation) Execute(ctx context.Context) error {
	switch op.operation {
	case "read":
		return op.executeRead()
	case "write":
		return op.executeWrite()
	case "delete":
		return op.executeDelete()
	case "copy":
		return op.executeCopy()
	case "move":
		return op.executeMove()
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedFileOperation, op.operation)
	}
}

func (op *FileOperation) executeRead() error {
	data, err := os.ReadFile(op.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	op.data = data
	return nil
}

func (op *FileOperation) executeWrite() error {
	if op.data == nil {
		return ErrNoDataToWrite
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(op.filePath), 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(op.filePath, op.data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func (op *FileOperation) executeDelete() error {
	if err := os.Remove(op.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (op *FileOperation) executeCopy() error {
	if op.targetPath == "" {
		return fmt.Errorf("%w for copy operation", ErrTargetPathNotSpecified)
	}

	// Read source file
	data, err := os.ReadFile(op.filePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(op.targetPath), 0o750); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Write to target
	if err := os.WriteFile(op.targetPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write target file: %w", err)
	}

	return nil
}

func (op *FileOperation) executeMove() error {
	if op.targetPath == "" {
		return fmt.Errorf("%w for move operation", ErrTargetPathNotSpecified)
	}

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(op.targetPath), 0o750); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	if err := os.Rename(op.filePath, op.targetPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// OperationBuilder helps build complex operations
type OperationBuilder struct {
	operation Operation
}

// NewOperationBuilder creates a new operation builder
func NewOperationBuilder(operation Operation) *OperationBuilder {
	return &OperationBuilder{operation: operation}
}

// WithMetadata adds metadata to the operation
func (b *OperationBuilder) WithMetadata(key string, value interface{}) *OperationBuilder {
	switch op := b.operation.(type) {
	case *GitHubAPIOperation:
		op.SetMetadata(key, value)
	case *ArtifactUploadOperation:
		op.SetMetadata(key, value)
	case *PRCommentOperation:
		op.SetMetadata(key, value)
	case *DeploymentOperation:
		op.SetMetadata(key, value)
	case *FileOperation:
		op.SetMetadata(key, value)
	}
	return b
}

// Build returns the built operation
func (b *OperationBuilder) Build() Operation {
	return b.operation
}

// Helper functions for creating common operations

// CreateGitHubAPIOperation creates a GitHub API operation with common settings
func CreateGitHubAPIOperation(method, url, token string) *GitHubAPIOperation {
	return NewGitHubAPIOperation(method, url, token).
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("User-Agent", "go-coverage/1.0")
}

// CreateArtifactUploadWithFallback creates an artifact upload operation with fallback support
func CreateArtifactUploadWithFallback(artifactName, filePath string, retentionDays int, uploadFunc func(ctx context.Context, name, path string, retention int) error) *ArtifactUploadOperation {
	return NewArtifactUploadOperation(artifactName, filePath, retentionDays, uploadFunc).
		SetMetadata("fallback_enabled", true).
		SetMetadata("created_at", time.Now())
}

// CreateDeploymentWithFallback creates a deployment operation with fallback support
func CreateDeploymentWithFallback(files []string, targetBranch, commitMessage, workDir string, deployFunc func(ctx context.Context, files []string, branch, message, workDir string) error) *DeploymentOperation {
	return NewDeploymentOperation(files, targetBranch, commitMessage, workDir, deployFunc).
		SetMetadata("fallback_enabled", true).
		SetMetadata("created_at", time.Now())
}
