package main

import (
	"runtime/debug"
	"strings"
	"sync"
)

// VersionInfo holds build-time information
type VersionInfo struct {
	version   string
	commit    string
	buildDate string
}

// versionInstance holds the singleton instance
//
//nolint:gochecknoglobals // Singleton pattern for version info
var versionInstance *VersionInfo

//nolint:gochecknoglobals // Thread-safe initialization
var versionOnce sync.Once

// Build-time variables injected via ldflags
// These are only used during initialization and not exposed as globals
//
//nolint:gochecknoglobals // These are build-time ldflags variables, not application state
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

// getVersionInfo returns the singleton VersionInfo instance
func getVersionInfo() *VersionInfo {
	versionOnce.Do(func() {
		versionInstance = &VersionInfo{
			version:   version,
			commit:    commit,
			buildDate: buildDate,
		}
	})
	return versionInstance
}

// GetVersion returns the version information with fallback to BuildInfo
func GetVersion() string {
	info := getVersionInfo()

	// If version was set via ldflags and it's not a template placeholder, use it
	if info.version != "" && !isTemplateString(info.version) {
		return info.version
	}

	// Try to get version from build info
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		// Check if there's a module version (from go install @version)
		if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
			// Clean up the version string
			version := buildInfo.Main.Version
			// Remove 'v' prefix if present for consistency
			version = strings.TrimPrefix(version, "v")
			return version
		}

		// Try to get VCS revision as fallback
		for _, setting := range buildInfo.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				// Use short commit hash
				if len(setting.Value) > 7 {
					return setting.Value[:7]
				}
				return setting.Value
			}
		}
	}

	// Default to "dev" if nothing else is available
	return "dev"
}

// GetCommit returns the commit hash with fallback to BuildInfo
func GetCommit() string {
	info := getVersionInfo()

	// If commit was set via ldflags and it's not a template placeholder, use it
	if info.commit != "none" && info.commit != "" && !isTemplateString(info.commit) {
		return info.commit
	}

	// Try to get from build info
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range buildInfo.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				return setting.Value
			}
		}
	}

	return "none"
}

// GetBuildDate returns the build date with fallback to BuildInfo
func GetBuildDate() string {
	info := getVersionInfo()

	// If build date was set via ldflags and it's not a template placeholder, use it
	if info.buildDate != "unknown" && info.buildDate != "" && !isTemplateString(info.buildDate) {
		return info.buildDate
	}

	// Try to get from build info
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range buildInfo.Settings {
			if setting.Key == "vcs.time" && setting.Value != "" {
				return setting.Value
			}
		}
	}

	return "unknown"
}

// IsModified returns true if the build has uncommitted changes
func IsModified() bool {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.modified" {
				return setting.Value == "true"
			}
		}
	}
	return false
}

// isTemplateString checks if a string contains unsubstituted template syntax
func isTemplateString(s string) bool {
	return strings.Contains(s, "{{") && strings.Contains(s, "}}")
}
