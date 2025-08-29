//go:build windows

package health

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// getDiskUsage returns disk usage information for the given path on Windows systems
func (c *DiskSpaceChecker) getDiskUsage(path string) (*DiskUsage, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path: %w", err)
	}

	var freeBytesToCaller, totalBytes, freeBytes uint64

	err = windows.GetDiskFreeSpaceEx(
		pathPtr,
		(*uint64)(unsafe.Pointer(&freeBytesToCaller)),
		(*uint64)(unsafe.Pointer(&totalBytes)),
		(*uint64)(unsafe.Pointer(&freeBytes)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk free space: %w", err)
	}

	return &DiskUsage{
		Total: int64(totalBytes),
		Free:  int64(freeBytesToCaller),
	}, nil
}
