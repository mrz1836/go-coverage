// Package main provides the Go coverage CLI tool
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mrz1836/go-coverage/cmd/go-coverage/cmd"
)

// BuildInfo holds build-time information that gets injected via ldflags
type BuildInfo struct {
	Version   string
	Commit    string
	BuildDate string
}

// getBuildInfo returns build information from version constants
func getBuildInfo() BuildInfo {
	version := GetVersion()

	// Add modified suffix if there are uncommitted changes
	if IsModified() && !strings.HasSuffix(version, "-dirty") {
		version += "-dirty"
	}

	return BuildInfo{
		Version:   version,
		Commit:    GetCommit(),
		BuildDate: GetBuildDate(),
	}
}

func main() {
	os.Exit(run())
}

// run executes the main application logic and returns the exit code.
// This function is separated from main() to enable testing.
func run() int {
	// Get build information
	buildInfo := getBuildInfo()

	// Create version info struct
	versionInfo := cmd.VersionInfo{
		Version:   buildInfo.Version,
		Commit:    buildInfo.Commit,
		BuildDate: buildInfo.BuildDate,
	}

	// Initialize CLI with new Commands structure
	commands := cmd.NewCommands(versionInfo)

	// Execute the root command
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}
