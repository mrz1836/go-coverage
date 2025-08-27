package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mrz1836/go-coverage/internal/version"
)

var (
	// ErrDevVersionNoForce is returned when trying to upgrade a dev version without --force
	ErrDevVersionNoForce = errors.New("cannot upgrade development build without --force")
	// ErrVersionParseFailed is returned when version cannot be parsed from output
	ErrVersionParseFailed = errors.New("could not parse version from output")
)

// UpgradeConfig holds configuration for the upgrade command
type UpgradeConfig struct {
	Force     bool
	CheckOnly bool
}

// newUpgradeCmd creates the upgrade command
func (c *Commands) newUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade go-coverage to the latest version",
		Long: `Upgrade the Go coverage system to the latest version available.

This command will:
  - Check the latest version available on GitHub
  - Compare with the currently installed version
  - Upgrade if a newer version is available`,
		Example: `  # Check for available updates
  go-coverage upgrade --check

  # Upgrade to latest version
  go-coverage upgrade

  # Force upgrade even if already on latest
  go-coverage upgrade --force`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config := UpgradeConfig{}
			var err error

			config.Force, err = cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}

			config.CheckOnly, err = cmd.Flags().GetBool("check")
			if err != nil {
				return err
			}

			return c.runUpgradeWithConfig(cmd, config)
		},
	}

	// Add flags
	cmd.Flags().BoolP("force", "f", false, "Force upgrade even if already on latest version")
	cmd.Flags().BoolP("check", "c", false, "Check for updates without upgrading")
	cmd.Flags().BoolP("verbose", "v", false, "Show release notes after upgrade")

	return cmd
}

func (c *Commands) runUpgradeWithConfig(cmd *cobra.Command, config UpgradeConfig) error {
	currentVersion := c.Version.Version

	// Handle development version or commit hash
	if isDevelopmentVersion(currentVersion) || currentVersion == "" || isLikelyCommitHash(currentVersion) {
		if !config.Force && !config.CheckOnly {
			cmd.Printf("⚠️  Current version appears to be a development build (%s)\n", currentVersion)
			cmd.Printf("   Use --force to upgrade anyway\n")
			return ErrDevVersionNoForce
		}
	}

	cmd.Printf("Current version: %s\n", formatVersion(currentVersion))

	// Fetch latest release
	cmd.Printf("Checking for updates...\n")
	release, err := version.GetLatestRelease("mrz1836", "go-coverage")
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	cmd.Printf("Latest version: %s\n", formatVersion(latestVersion))

	// Compare versions
	isNewer := version.IsNewerVersion(currentVersion, latestVersion)

	if !isNewer && !config.Force {
		cmd.Printf("✅ You are already on the latest version (%s)\n", formatVersion(currentVersion))
		return nil
	}

	if config.CheckOnly {
		if isNewer {
			cmd.Printf("⚠️  A newer version is available: %s → %s\n", formatVersion(currentVersion), formatVersion(latestVersion))
			cmd.Printf("   Run 'go-coverage upgrade' to upgrade\n")
		} else {
			cmd.Printf("✅ You are on the latest version\n")
		}
		return nil
	}

	// Perform upgrade
	if isNewer {
		cmd.Printf("Upgrading from %s to %s...\n", formatVersion(currentVersion), formatVersion(latestVersion))
	} else if config.Force {
		cmd.Printf("Force reinstalling version %s...\n", formatVersion(latestVersion))
	}

	// Run go install command
	installCmd := fmt.Sprintf("github.com/mrz1836/go-coverage/cmd/go-coverage@v%s", latestVersion)

	cmd.Printf("Running: go install %s\n", installCmd)

	execCmd := exec.CommandContext(context.Background(), "go", "install", installCmd) //nolint:gosec // Command is constructed safely
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("failed to upgrade: %w", err)
	}

	cmd.Printf("✅ Successfully upgraded to version %s\n", formatVersion(latestVersion))

	// Show release notes if available and verbose
	verbose, _ := cmd.Flags().GetBool("verbose")
	if release.Body != "" && verbose {
		cmd.Printf("\nRelease notes for v%s:\n", latestVersion)
		lines := strings.Split(release.Body, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				cmd.Printf("  %s\n", line)
			}
		}
	}

	return nil
}

func formatVersion(v string) string {
	if v == "dev" || v == "" {
		return "dev"
	}
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

// GetInstalledVersion attempts to get the version of the installed binary
func GetInstalledVersion() (string, error) {
	// Try to run the binary with --version flag
	cmd := exec.CommandContext(context.Background(), "go-coverage", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	// Parse the version from output
	// Expected format: "go-coverage version X.Y.Z"
	outputStr := strings.TrimSpace(string(output))
	parts := strings.Fields(outputStr)

	for i, part := range parts {
		if part == "version" && i+1 < len(parts) {
			version := parts[i+1]
			// Clean up version string
			version = strings.TrimPrefix(version, "v")
			return version, nil
		}
	}

	return "", fmt.Errorf("%w: %s", ErrVersionParseFailed, outputStr)
}

// CheckGoInstalled verifies that Go is installed and available
func CheckGoInstalled() error {
	cmd := exec.CommandContext(context.Background(), "go", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go is not installed or not in PATH: %w", err)
	}
	return nil
}

// GetGoPath returns the GOPATH/bin directory where binaries are installed
func GetGoPath() (string, error) {
	cmd := exec.CommandContext(context.Background(), "go", "env", "GOPATH")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GOPATH: %w", err)
	}

	gopath := strings.TrimSpace(string(output))
	if gopath == "" {
		// Use default GOPATH
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		gopath = fmt.Sprintf("%s/go", home)
	}

	return fmt.Sprintf("%s/bin", gopath), nil
}

// IsInPath checks if go-coverage binary is in PATH
func IsInPath() bool {
	_, err := exec.LookPath("go-coverage")
	return err == nil
}

// GetBinaryLocation returns the location of the go-coverage binary
func GetBinaryLocation() (string, error) {
	if runtime.GOOS == "windows" {
		return exec.LookPath("go-coverage.exe")
	}
	return exec.LookPath("go-coverage")
}

// isLikelyCommitHash checks if a version string looks like a commit hash
func isLikelyCommitHash(version string) bool {
	// Remove any -dirty suffix
	version = strings.TrimSuffix(version, "-dirty")

	// Commit hashes are typically 7-40 hex characters
	if len(version) < 7 || len(version) > 40 {
		return false
	}

	// Check if all characters are valid hex
	for _, c := range version {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}

	return true
}

// isDevelopmentVersion checks if a version string represents a development version
func isDevelopmentVersion(version string) bool {
	if version == "dev" {
		return true
	}
	// Check if version starts with "dev" (e.g., "dev-dirty")
	if strings.HasPrefix(version, "dev-") {
		return true
	}
	return false
}
