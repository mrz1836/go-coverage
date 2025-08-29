// Package artifacts - Partial upload functionality for handling large operations
// This file provides chunked upload capabilities with resumable upload support
// for handling large coverage history files and artifacts gracefully.
package artifacts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mrz1836/go-coverage/internal/retry"
)

// Error definitions for err113 linter compliance
var (
	ErrChunkMissingOrNotUploaded = errors.New("chunk is missing or not uploaded")
	ErrUploadedSizeMismatch      = errors.New("uploaded size mismatch")
	ErrFileChangedDuringUpload   = errors.New("file has changed since upload started")
)

const (
	// DefaultChunkSize is the default size for upload chunks (1MB)
	DefaultChunkSize = 1024 * 1024
	// MaxChunkSize is the maximum allowed chunk size (10MB)
	MaxChunkSize = 10 * 1024 * 1024
	// MinChunkSize is the minimum chunk size (64KB)
	MinChunkSize = 64 * 1024
	// UploadStateFile is the name of the state file for tracking progress
	UploadStateFile = ".upload_state.json"
	// MaxRetryAttempts for partial uploads
	MaxRetryAttempts = 5
)

var (
	// ErrPartialUploadFailed indicates a partial upload operation failed
	ErrPartialUploadFailed = errors.New("partial upload failed")
	// ErrInvalidChunkSize indicates chunk size is invalid
	ErrInvalidChunkSize = errors.New("invalid chunk size")
	// ErrUploadIncomplete indicates upload is not complete
	ErrUploadIncomplete = errors.New("upload incomplete")
	// ErrCorruptChunk indicates a chunk is corrupted
	ErrCorruptChunk = errors.New("chunk is corrupted")
)

// PartialUploader handles chunked and resumable uploads for large operations
type PartialUploader struct {
	tempDir     string
	chunkSize   int64
	retryConfig *retry.Config
}

// NewPartialUploader creates a new partial uploader
func NewPartialUploader(tempDir string, chunkSize int64) *PartialUploader {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	if chunkSize < MinChunkSize {
		chunkSize = MinChunkSize
	}
	if chunkSize > MaxChunkSize {
		chunkSize = MaxChunkSize
	}

	return &PartialUploader{
		tempDir:     tempDir,
		chunkSize:   chunkSize,
		retryConfig: retry.FileOperationConfig(),
	}
}

// UploadState tracks the state of a partial upload operation
type UploadState struct {
	SessionID       string            `json:"session_id"`
	FileName        string            `json:"file_name"`
	FileSize        int64             `json:"file_size"`
	FileHash        string            `json:"file_hash"`
	ChunkSize       int64             `json:"chunk_size"`
	TotalChunks     int               `json:"total_chunks"`
	CompletedChunks map[int]ChunkInfo `json:"completed_chunks"`
	StartTime       time.Time         `json:"start_time"`
	LastUpdate      time.Time         `json:"last_update"`
	IsCompleted     bool              `json:"is_completed"`
	UploadedBytes   int64             `json:"uploaded_bytes"`
}

// ChunkInfo contains information about an uploaded chunk
type ChunkInfo struct {
	Index     int       `json:"index"`
	Size      int64     `json:"size"`
	Hash      string    `json:"hash"`
	Uploaded  bool      `json:"uploaded"`
	Timestamp time.Time `json:"timestamp"`
	Retries   int       `json:"retries"`
}

// UploadOptions configures partial upload behavior
type PartialUploadOptions struct {
	EnableResume     bool             `json:"enable_resume"`   // Whether to enable resumable uploads
	ChecksumVerify   bool             `json:"checksum_verify"` // Whether to verify chunk checksums
	ParallelChunks   int              `json:"parallel_chunks"` // Number of chunks to upload in parallel
	ProgressCallback ProgressCallback `json:"-"`               // Progress callback function
	MaxRetries       int              `json:"max_retries"`     // Max retries per chunk
	RetryDelay       time.Duration    `json:"retry_delay"`     // Delay between retries
}

// ProgressCallback is called to report upload progress
type ProgressCallback func(uploaded, total int64, chunksCompleted, totalChunks int)

// DefaultPartialUploadOptions returns default options for partial uploads
func DefaultPartialUploadOptions() *PartialUploadOptions {
	return &PartialUploadOptions{
		EnableResume:   true,
		ChecksumVerify: true,
		ParallelChunks: 2,
		MaxRetries:     MaxRetryAttempts,
		RetryDelay:     time.Second,
	}
}

// StartPartialUpload initiates a partial upload operation
func (pu *PartialUploader) StartPartialUpload(ctx context.Context, filePath string, opts *PartialUploadOptions) (*UploadState, error) {
	if opts == nil {
		opts = DefaultPartialUploadOptions()
	}

	// Validate file exists and get info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Calculate file hash for integrity checking
	fileHash, err := pu.calculateFileHash(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}

	// Create upload session
	sessionID := pu.generateSessionID(filePath, fileHash)
	totalChunks := int((fileInfo.Size() + pu.chunkSize - 1) / pu.chunkSize)

	state := &UploadState{
		SessionID:       sessionID,
		FileName:        filepath.Base(filePath),
		FileSize:        fileInfo.Size(),
		FileHash:        fileHash,
		ChunkSize:       pu.chunkSize,
		TotalChunks:     totalChunks,
		CompletedChunks: make(map[int]ChunkInfo),
		StartTime:       time.Now(),
		LastUpdate:      time.Now(),
		IsCompleted:     false,
		UploadedBytes:   0,
	}

	// Check for existing state if resume is enabled
	if opts.EnableResume {
		if existingState, err := pu.loadUploadState(sessionID); err == nil {
			// Verify the existing state matches current file
			if existingState.FileHash == fileHash && existingState.FileSize == fileInfo.Size() {
				state = existingState
				state.LastUpdate = time.Now()
			}
		}
	}

	// Save initial state
	if err := pu.saveUploadState(state); err != nil {
		return nil, fmt.Errorf("failed to save upload state: %w", err)
	}

	return state, nil
}

// UploadChunks uploads the file in chunks according to the upload state
func (pu *PartialUploader) UploadChunks(ctx context.Context, filePath string, state *UploadState, opts *PartialUploadOptions, uploadFunc func(ctx context.Context, chunkData []byte, chunkIndex int, state *UploadState) error) error {
	if opts == nil {
		opts = DefaultPartialUploadOptions()
	}

	// Open file for reading
	file, err := os.Open(filePath) //nolint:gosec // path validated in StartPartialUpload
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Create semaphore for controlling parallel uploads
	semaphore := make(chan struct{}, opts.ParallelChunks)

	// Channel for collecting upload results
	results := make(chan uploadResult, state.TotalChunks)

	// Upload chunks
	for chunkIndex := 0; chunkIndex < state.TotalChunks; chunkIndex++ {
		// Skip already completed chunks
		if chunkInfo, exists := state.CompletedChunks[chunkIndex]; exists && chunkInfo.Uploaded {
			results <- uploadResult{index: chunkIndex, err: nil}
			continue
		}

		// Acquire semaphore
		semaphore <- struct{}{}

		go func(index int) {
			defer func() { <-semaphore }()

			result := uploadResult{index: index}
			result.err = pu.uploadSingleChunk(ctx, file, index, state, opts, uploadFunc)
			results <- result
		}(chunkIndex)
	}

	// Wait for all uploads to complete
	var uploadErrors []error
	completedChunks := 0

	for i := 0; i < state.TotalChunks; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case result := <-results:
			if result.err != nil {
				uploadErrors = append(uploadErrors, fmt.Errorf("chunk %d: %w", result.index, result.err))
			} else {
				completedChunks++

				// Update progress
				if opts.ProgressCallback != nil {
					opts.ProgressCallback(state.UploadedBytes, state.FileSize, completedChunks, state.TotalChunks)
				}
			}
		}
	}

	// Check for errors
	if len(uploadErrors) > 0 {
		// Save current state before returning error
		_ = pu.saveUploadState(state)
		return fmt.Errorf("%w: %d chunks failed: %v", ErrPartialUploadFailed, len(uploadErrors), uploadErrors)
	}

	// Mark as completed
	state.IsCompleted = true
	state.LastUpdate = time.Now()

	// Save final state
	if err := pu.saveUploadState(state); err != nil {
		return fmt.Errorf("failed to save final state: %w", err)
	}

	return nil
}

// uploadResult represents the result of uploading a single chunk
type uploadResult struct {
	index int
	err   error
}

// uploadSingleChunk uploads a single chunk with retry logic
func (pu *PartialUploader) uploadSingleChunk(ctx context.Context, file *os.File, chunkIndex int, state *UploadState, opts *PartialUploadOptions, uploadFunc func(ctx context.Context, chunkData []byte, chunkIndex int, state *UploadState) error) error {
	// Calculate chunk offset and size
	offset := int64(chunkIndex) * state.ChunkSize
	size := state.ChunkSize
	if offset+size > state.FileSize {
		size = state.FileSize - offset
	}

	// Read chunk data
	chunkData := make([]byte, size)
	if _, err := file.ReadAt(chunkData, offset); err != nil && err != io.EOF {
		return fmt.Errorf("failed to read chunk data: %w", err)
	}

	// Calculate chunk hash for verification
	var chunkHash string
	if opts.ChecksumVerify {
		hash := sha256.Sum256(chunkData)
		chunkHash = hex.EncodeToString(hash[:])
	}

	// Upload chunk with retry logic
	err := retry.Do(ctx, pu.retryConfig, func() error {
		return uploadFunc(ctx, chunkData, chunkIndex, state)
	})
	if err != nil {
		// Update retry count for this chunk
		if chunkInfo, exists := state.CompletedChunks[chunkIndex]; exists {
			chunkInfo.Retries++
			state.CompletedChunks[chunkIndex] = chunkInfo
		} else {
			state.CompletedChunks[chunkIndex] = ChunkInfo{
				Index:     chunkIndex,
				Size:      size,
				Hash:      chunkHash,
				Uploaded:  false,
				Timestamp: time.Now(),
				Retries:   1,
			}
		}

		// Save state with retry info
		_ = pu.saveUploadState(state)

		return fmt.Errorf("failed to upload chunk after retries: %w", err)
	}

	// Mark chunk as completed
	state.CompletedChunks[chunkIndex] = ChunkInfo{
		Index:     chunkIndex,
		Size:      size,
		Hash:      chunkHash,
		Uploaded:  true,
		Timestamp: time.Now(),
		Retries:   0,
	}

	state.UploadedBytes += size
	state.LastUpdate = time.Now()

	// Save state periodically
	if chunkIndex%10 == 0 { // Save every 10 chunks
		_ = pu.saveUploadState(state)
	}

	return nil
}

// VerifyUpload verifies that an upload is complete and all chunks are intact
func (pu *PartialUploader) VerifyUpload(state *UploadState) error {
	if !state.IsCompleted {
		return ErrUploadIncomplete
	}

	// Check all chunks are present
	for i := 0; i < state.TotalChunks; i++ {
		chunkInfo, exists := state.CompletedChunks[i]
		if !exists || !chunkInfo.Uploaded {
			return fmt.Errorf("%w: chunk %d", ErrChunkMissingOrNotUploaded, i)
		}
	}

	// Verify total uploaded bytes
	var totalUploaded int64
	for _, chunkInfo := range state.CompletedChunks {
		totalUploaded += chunkInfo.Size
	}

	if totalUploaded != state.FileSize {
		return fmt.Errorf("%w: expected %d, got %d", ErrUploadedSizeMismatch, state.FileSize, totalUploaded)
	}

	return nil
}

// CleanupUploadState removes upload state files after successful completion
func (pu *PartialUploader) CleanupUploadState(sessionID string) error {
	stateFile := filepath.Join(pu.tempDir, sessionID+"-"+UploadStateFile)
	if err := os.Remove(stateFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to cleanup upload state: %w", err)
	}
	return nil
}

// ListIncompleteUploads returns a list of incomplete upload sessions
func (pu *PartialUploader) ListIncompleteUploads() ([]*UploadState, error) {
	if err := os.MkdirAll(pu.tempDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	files, err := os.ReadDir(pu.tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read temp directory: %w", err)
	}

	var incompleteUploads []*UploadState

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), UploadStateFile) {
			continue
		}

		stateFile := filepath.Join(pu.tempDir, file.Name())
		state, err := pu.loadUploadStateFromFile(stateFile)
		if err != nil {
			continue // Skip corrupted state files
		}

		if !state.IsCompleted {
			incompleteUploads = append(incompleteUploads, state)
		}
	}

	// Sort by start time (oldest first)
	sort.Slice(incompleteUploads, func(i, j int) bool {
		return incompleteUploads[i].StartTime.Before(incompleteUploads[j].StartTime)
	})

	return incompleteUploads, nil
}

// ResumeUpload resumes a previously started upload
func (pu *PartialUploader) ResumeUpload(ctx context.Context, sessionID, filePath string, opts *PartialUploadOptions, uploadFunc func(ctx context.Context, chunkData []byte, chunkIndex int, state *UploadState) error) error {
	// Load existing state
	state, err := pu.loadUploadState(sessionID)
	if err != nil {
		return fmt.Errorf("failed to load upload state: %w", err)
	}

	// Verify file still matches
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	fileHash, err := pu.calculateFileHash(filePath)
	if err != nil {
		return fmt.Errorf("failed to calculate file hash: %w", err)
	}

	if state.FileHash != fileHash || state.FileSize != fileInfo.Size() {
		return ErrFileChangedDuringUpload
	}

	// Resume upload
	return pu.UploadChunks(ctx, filePath, state, opts, uploadFunc)
}

// calculateFileHash calculates SHA256 hash of a file
func (pu *PartialUploader) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath) //nolint:gosec // validated caller
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// generateSessionID generates a unique session ID
func (pu *PartialUploader) generateSessionID(filePath, fileHash string) string {
	data := fmt.Sprintf("%s-%s-%d", filepath.Base(filePath), fileHash, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes for shorter ID
}

// saveUploadState saves upload state to disk
func (pu *PartialUploader) saveUploadState(state *UploadState) error {
	if err := os.MkdirAll(pu.tempDir, 0o750); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	stateFile := filepath.Join(pu.tempDir, state.SessionID+"-"+UploadStateFile)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// loadUploadState loads upload state from disk
func (pu *PartialUploader) loadUploadState(sessionID string) (*UploadState, error) {
	stateFile := filepath.Join(pu.tempDir, sessionID+"-"+UploadStateFile)
	return pu.loadUploadStateFromFile(stateFile)
}

// loadUploadStateFromFile loads upload state from a specific file
func (pu *PartialUploader) loadUploadStateFromFile(stateFile string) (*UploadState, error) {
	data, err := os.ReadFile(stateFile) //nolint:gosec // internal temp file
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state UploadState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// GetUploadProgress returns progress information for an upload session
func (pu *PartialUploader) GetUploadProgress(sessionID string) (*UploadProgress, error) {
	state, err := pu.loadUploadState(sessionID)
	if err != nil {
		return nil, err
	}

	progress := &UploadProgress{
		SessionID:       state.SessionID,
		FileName:        state.FileName,
		TotalBytes:      state.FileSize,
		UploadedBytes:   state.UploadedBytes,
		TotalChunks:     state.TotalChunks,
		CompletedChunks: len(state.CompletedChunks),
		IsCompleted:     state.IsCompleted,
		StartTime:       state.StartTime,
		LastUpdate:      state.LastUpdate,
		PercentComplete: float64(state.UploadedBytes) / float64(state.FileSize) * 100,
	}

	if !state.StartTime.IsZero() {
		progress.Duration = time.Since(state.StartTime)
		if progress.UploadedBytes > 0 {
			progress.AverageSpeed = float64(progress.UploadedBytes) / progress.Duration.Seconds()
		}
	}

	// Calculate ETA if not completed
	if !progress.IsCompleted && progress.AverageSpeed > 0 {
		remainingBytes := progress.TotalBytes - progress.UploadedBytes
		etaSeconds := float64(remainingBytes) / progress.AverageSpeed
		progress.EstimatedTimeRemaining = time.Duration(etaSeconds) * time.Second
	}

	return progress, nil
}

// UploadProgress provides progress information for uploads
type UploadProgress struct {
	SessionID              string        `json:"session_id"`
	FileName               string        `json:"file_name"`
	TotalBytes             int64         `json:"total_bytes"`
	UploadedBytes          int64         `json:"uploaded_bytes"`
	TotalChunks            int           `json:"total_chunks"`
	CompletedChunks        int           `json:"completed_chunks"`
	IsCompleted            bool          `json:"is_completed"`
	StartTime              time.Time     `json:"start_time"`
	LastUpdate             time.Time     `json:"last_update"`
	Duration               time.Duration `json:"duration"`
	PercentComplete        float64       `json:"percent_complete"`
	AverageSpeed           float64       `json:"average_speed"` // bytes per second
	EstimatedTimeRemaining time.Duration `json:"estimated_time_remaining"`
}

// String returns a human-readable progress string
func (p *UploadProgress) String() string {
	if p.IsCompleted {
		return fmt.Sprintf("Upload completed: %s (%s in %s)",
			p.FileName, formatBytes(p.TotalBytes), p.Duration.Round(time.Second))
	}

	speed := ""
	if p.AverageSpeed > 0 {
		speed = fmt.Sprintf(" at %s/s", formatBytes(int64(p.AverageSpeed)))
	}

	eta := ""
	if p.EstimatedTimeRemaining > 0 {
		eta = fmt.Sprintf(", ETA: %s", p.EstimatedTimeRemaining.Round(time.Second))
	}

	return fmt.Sprintf("Uploading %s: %.1f%% (%s/%s)%s%s",
		p.FileName, p.PercentComplete,
		formatBytes(p.UploadedBytes), formatBytes(p.TotalBytes),
		speed, eta)
}

// formatBytes formats bytes in human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
