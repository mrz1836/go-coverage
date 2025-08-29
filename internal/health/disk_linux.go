//go:build linux

package health

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// getDiskUsage returns disk usage information for the given path on Linux systems
func (c *DiskSpaceChecker) getDiskUsage(path string) (*DiskUsage, error) {
	var stat unix.Statfs_t
	err := unix.Statfs(path, &stat)
	if err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	// Calculate total and free space
	// On Linux, stat.Bsize is already int64, no conversion needed
	blockSize := stat.Bsize
	// Use safe calculation with overflow protection
	var total, free int64
	const maxInt64 = int64(^uint64(0) >> 1)

	// Calculate total space with overflow check
	if stat.Blocks > uint64(maxInt64/blockSize) { //nolint:gosec // safe conversion within checked bounds
		total = maxInt64 // Use max int64 if overflow would occur
	} else {
		total = int64(stat.Blocks) * blockSize //nolint:gosec // overflow checked above
	}

	// Calculate free space with overflow check
	if stat.Bavail > uint64(maxInt64/blockSize) { //nolint:gosec // safe conversion within checked bounds
		free = maxInt64 // Use max int64 if overflow would occur
	} else {
		free = int64(stat.Bavail) * blockSize //nolint:gosec // overflow checked above
	}

	return &DiskUsage{
		Total: total,
		Free:  free,
	}, nil
}
