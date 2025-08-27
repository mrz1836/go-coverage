// Package cmd provides the command-line interface for the Go coverage tool
package cmd

import (
	"github.com/spf13/cobra"
)

// Commands holds all CLI commands and their configuration
type Commands struct {
	Root       *cobra.Command
	Complete   *cobra.Command
	History    *cobra.Command
	Comment    *cobra.Command
	Parse      *cobra.Command
	SetupPages *cobra.Command
	Upgrade    *cobra.Command

	// Version information
	Version VersionInfo
}

// VersionInfo holds version information for the application
type VersionInfo struct {
	Version   string
	Commit    string
	BuildDate string
}

// NewCommands creates and initializes all CLI commands
func NewCommands(version VersionInfo) *Commands {
	cmds := &Commands{
		Version: version,
	}

	// Initialize root command
	cmds.Root = cmds.newRootCmd()

	// Initialize subcommands
	cmds.Complete = cmds.newCompleteCmd()
	cmds.History = cmds.newHistoryCmd()
	cmds.Comment = cmds.newCommentCmd()
	cmds.Parse = cmds.newParseCmd()
	cmds.SetupPages = cmds.newSetupPagesCmd()
	cmds.Upgrade = cmds.newUpgradeCmd()

	// Add subcommands to root
	cmds.Root.AddCommand(
		cmds.Complete,
		cmds.History,
		cmds.Comment,
		cmds.Parse,
		cmds.SetupPages,
		cmds.Upgrade,
	)

	// Set version on root command
	cmds.Root.Version = version.Version

	// Set custom version template to ensure consistent output
	cmds.Root.SetVersionTemplate("{{with .DisplayName}}{{printf \"%s \" .}}{{end}}{{printf \"version %s\" .Version}}\n")

	return cmds
}

// Execute runs the root command
func (c *Commands) Execute() error {
	return c.Root.Execute()
}

// newRootCmd creates the root command
func (c *Commands) newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "go-coverage",
		Short: "Go-native coverage system for CI/CD",
		Long: `Go Coverage is a self-contained, Go-native coverage system that provides
professional coverage tracking, badge generation, and reporting while maintaining
the simplicity and performance that Go developers expect.

Built as a bolt-on solution completely encapsulated within the .github folder,
this tool replaces Codecov with zero external service dependencies.`,
	}

	// Global flags
	cmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
	cmd.PersistentFlags().StringP("log-level", "l", "info", "Log level (debug, info, warn, error)")
	cmd.PersistentFlags().String("log-format", "text", "Log format (text, json, pretty)")

	return cmd
}
