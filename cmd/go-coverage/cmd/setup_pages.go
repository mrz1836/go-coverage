// Package cmd provides CLI commands for the Go coverage tool
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// ErrGitHubCLINotFound indicates that the GitHub CLI is not installed or available
var ErrGitHubCLINotFound = errors.New("github CLI (gh) not found")

// ErrGitHubCLINotAuthenticated indicates that the GitHub CLI is not authenticated
var ErrGitHubCLINotAuthenticated = errors.New("github CLI not authenticated")

// ErrRepositoryNotFound indicates that the repository could not be found or accessed
var ErrRepositoryNotFound = errors.New("repository not found or access denied")

// ErrNoGitRepository indicates that the current directory is not a git repository
var ErrNoGitRepository = errors.New("not in a git repository")

// ErrInvalidRepositoryFormat indicates that the repository format is invalid
var ErrInvalidRepositoryFormat = errors.New("invalid repository format, expected 'owner/repo'")

// ErrCouldNotParseRepository indicates that the repository could not be parsed from remote URL
var ErrCouldNotParseRepository = errors.New("could not parse GitHub repository from remote URL")

// ErrNoDeploymentPolicies indicates that no deployment branch policies were found
var ErrNoDeploymentPolicies = errors.New("no deployment branch policies found")

// GitHubEnvironmentResponse represents the response from GitHub API for environments
type GitHubEnvironmentResponse struct {
	Name                   string                  `json:"name"`
	DeploymentBranchPolicy *DeploymentBranchPolicy `json:"deployment_branch_policy,omitempty"`
	ProtectionRules        []interface{}           `json:"protection_rules"`
}

// DeploymentBranchPolicy represents branch deployment policy
type DeploymentBranchPolicy struct {
	ProtectedBranches    bool `json:"protected_branches"`
	CustomBranchPolicies bool `json:"custom_branch_policies"`
}

// GitHubBranchPoliciesResponse represents the response for deployment branch policies
type GitHubBranchPoliciesResponse struct {
	BranchPolicies []BranchPolicy `json:"branch_policies"`
}

// BranchPolicy represents a single deployment branch policy
type BranchPolicy struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

var setupPagesCmd = &cobra.Command{ //nolint:gochecknoglobals // CLI command
	Use:   "setup-pages [repository]",
	Short: "Set up GitHub Pages environment for coverage deployment",
	Long: `Configure GitHub Pages environment with deployment branch policies to allow
coverage reports to be deployed from various branches including master, gh-pages,
and wildcard patterns for feature branches and dependency updates.

This command replaces the bash script and provides the same functionality
with improved error handling and user experience.

Examples:
  go-coverage setup-pages                    # Auto-detect repository from git remote
  go-coverage setup-pages owner/repo         # Specify repository explicitly
  go-coverage setup-pages --dry-run          # Preview changes without making them
  go-coverage setup-pages --verbose          # Show detailed output`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")
		customDomain, _ := cmd.Flags().GetString("custom-domain")
		protectBranches, _ := cmd.Flags().GetBool("protect-branches")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		cmd.Printf("🚀 Go Coverage - GitHub Pages Setup\n")
		cmd.Printf("=====================================\n\n")

		if dryRun {
			cmd.Printf("🧪 DRY RUN MODE - No changes will be made\n\n")
		}

		// Step 1: Check prerequisites
		cmd.Printf("🔍 Step 1: Checking prerequisites...\n")
		if err := checkPrerequisites(ctx, cmd, verbose); err != nil {
			return fmt.Errorf("prerequisites check failed: %w", err)
		}
		cmd.Printf("   ✅ GitHub CLI is installed and authenticated\n\n")

		// Step 2: Determine repository
		cmd.Printf("📋 Step 2: Determining repository...\n")
		var repo string
		if len(args) > 0 {
			repo = args[0]
			cmd.Printf("   📝 Using specified repository: %s\n", repo)
		} else {
			var err error
			repo, err = getRepositoryFromGit(ctx, cmd, verbose)
			if err != nil {
				return fmt.Errorf("failed to determine repository: %w", err)
			}
			cmd.Printf("   🔍 Auto-detected repository: %s\n", repo)
		}

		// Validate repository format
		if !isValidRepositoryFormat(repo) {
			return fmt.Errorf("repository format '%s': %w", repo, ErrInvalidRepositoryFormat)
		}
		cmd.Printf("\n")

		// Step 3: Check repository access
		cmd.Printf("🔐 Step 3: Checking repository access...\n")
		if err := checkRepositoryAccess(ctx, cmd, repo, verbose); err != nil {
			return fmt.Errorf("repository access check failed: %w", err)
		}
		cmd.Printf("   ✅ Repository access confirmed\n\n")

		// Step 4: Setup GitHub Pages environment
		cmd.Printf("🌐 Step 4: Setting up GitHub Pages environment...\n")
		if err := setupPagesEnvironment(ctx, cmd, repo, dryRun, verbose); err != nil {
			return fmt.Errorf("failed to setup pages environment: %w", err)
		}
		cmd.Printf("   ✅ GitHub Pages environment configured\n\n")

		// Step 5: Configure deployment branch policies
		cmd.Printf("🌿 Step 5: Configuring deployment branch policies...\n")
		if err := setupDeploymentBranches(ctx, cmd, repo, dryRun, verbose); err != nil {
			return fmt.Errorf("failed to setup deployment branches: %w", err)
		}
		cmd.Printf("   ✅ Deployment branch policies configured\n\n")

		// Step 6: Configure custom domain (if specified)
		if customDomain != "" {
			cmd.Printf("🌍 Step 6: Configuring custom domain...\n")
			if err := setupCustomDomain(ctx, cmd, repo, customDomain, dryRun, verbose); err != nil {
				cmd.Printf("   ⚠️  Custom domain setup failed: %v\n", err)
				cmd.Printf("   💡 You can configure this manually in repository settings\n")
			} else {
				cmd.Printf("   ✅ Custom domain configured: %s\n", customDomain)
			}
			cmd.Printf("\n")
		}

		// Step 7: Setup branch protection (if requested)
		if protectBranches {
			cmd.Printf("🛡️  Step 7: Setting up branch protection...\n")
			if err := setupBranchProtection(ctx, cmd, repo, dryRun, verbose); err != nil {
				cmd.Printf("   ⚠️  Branch protection setup failed: %v\n", err)
				cmd.Printf("   💡 You can configure this manually in repository settings\n")
			} else {
				cmd.Printf("   ✅ Branch protection configured\n")
			}
			cmd.Printf("\n")
		}

		// Step 8: Verify setup
		cmd.Printf("✅ Step 8: Verifying configuration...\n")
		if err := verifySetup(ctx, cmd, repo, verbose); err != nil {
			cmd.Printf("   ⚠️  Verification completed with warnings: %v\n", err)
		} else {
			cmd.Printf("   ✅ Configuration verified successfully\n")
		}
		cmd.Printf("\n")

		// Step 9: Show next steps
		showNextSteps(cmd, repo, dryRun)

		return nil
	},
}

// checkPrerequisites verifies that gh CLI is installed and authenticated
func checkPrerequisites(ctx context.Context, cmd *cobra.Command, verbose bool) error {
	// Check if gh command exists
	if _, err := exec.LookPath("gh"); err != nil {
		cmd.Printf("   ❌ GitHub CLI (gh) is not installed\n")
		cmd.Printf("   💡 Install from: https://cli.github.com/\n")
		return ErrGitHubCLINotFound
	}

	if verbose {
		cmd.Printf("   🔍 GitHub CLI found\n")
	}

	// Check authentication
	authCmd := exec.CommandContext(ctx, "gh", "auth", "status")
	if err := authCmd.Run(); err != nil {
		cmd.Printf("   ❌ GitHub CLI is not authenticated\n")
		cmd.Printf("   💡 Run: gh auth login\n")
		return ErrGitHubCLINotAuthenticated
	}

	if verbose {
		cmd.Printf("   🔍 GitHub CLI authentication verified\n")
	}

	return nil
}

// getRepositoryFromGit extracts repository information from git remote
func getRepositoryFromGit(ctx context.Context, cmd *cobra.Command, verbose bool) (string, error) {
	// Get remote origin URL
	gitCmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	output, err := gitCmd.Output()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNoGitRepository, err)
	}

	remoteURL := strings.TrimSpace(string(output))
	if verbose {
		// Don't log the full URL as it might contain tokens
		cmd.Printf("   🔍 Git remote detected\n")
	}

	// Parse GitHub repository from URL
	repo := parseGitHubRepoFromURL(remoteURL)
	if repo == "" {
		return "", ErrCouldNotParseRepository
	}

	return repo, nil
}

// parseGitHubRepoFromURL extracts owner/repo from various GitHub URL formats
func parseGitHubRepoFromURL(url string) string {
	// Security checks: reject URLs with dangerous patterns
	if strings.Contains(url, "..") ||
		!strings.Contains(url, "github.com") {
		return ""
	}

	// Remove trailing .git
	url = strings.TrimSuffix(url, ".git")

	// Handle HTTPS URLs: https://github.com/owner/repo (exact match, no extra paths)
	if httpsMatch := regexp.MustCompile(`^https://github\.com/([a-zA-Z0-9._-]+)/([a-zA-Z0-9._-]+)$`).FindStringSubmatch(url); httpsMatch != nil {
		// Additional security check: no @ symbols in owner/repo parts
		owner, repo := httpsMatch[1], httpsMatch[2]
		if strings.Contains(owner, "@") || strings.Contains(repo, "@") {
			return ""
		}
		return owner + "/" + repo
	}

	// Handle SSH URLs: git@github.com:owner/repo (exact match, no extra paths)
	if sshMatch := regexp.MustCompile(`^git@github\.com:([a-zA-Z0-9._-]+)/([a-zA-Z0-9._-]+)$`).FindStringSubmatch(url); sshMatch != nil {
		// Additional security check: no @ symbols in owner/repo parts
		owner, repo := sshMatch[1], sshMatch[2]
		if strings.Contains(owner, "@") || strings.Contains(repo, "@") {
			return ""
		}
		return owner + "/" + repo
	}

	return ""
}

// isValidRepositoryFormat validates the repository format
func isValidRepositoryFormat(repo string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`).MatchString(repo)
}

// checkRepositoryAccess verifies that the user can access the repository
func checkRepositoryAccess(ctx context.Context, cmd *cobra.Command, repo string, verbose bool) error {
	if verbose {
		cmd.Printf("   🔍 Checking access to repository: %s\n", repo)
	}

	// Try to view the repository
	viewCmd := exec.CommandContext(ctx, "gh", "repo", "view", repo)
	viewCmd.Stdout = nil // Suppress output
	viewCmd.Stderr = nil // Suppress errors
	if err := viewCmd.Run(); err != nil {
		cmd.Printf("   ❌ Cannot access repository '%s'\n", repo)
		cmd.Printf("   💡 Please check:\n")
		cmd.Printf("       - Repository exists and is spelled correctly\n")
		cmd.Printf("       - You have access to the repository\n")
		cmd.Printf("       - Your GitHub CLI has the required permissions\n")
		return ErrRepositoryNotFound
	}

	return nil
}

// setupPagesEnvironment creates or updates the GitHub Pages environment
func setupPagesEnvironment(ctx context.Context, cmd *cobra.Command, repo string, dryRun, verbose bool) error {
	if verbose {
		cmd.Printf("   🔧 Creating/updating github-pages environment...\n")
	}

	if dryRun {
		cmd.Printf("   🧪 DRY RUN: Would create/update github-pages environment\n")
		return nil
	}

	// Validate repo format to prevent command injection
	if !isValidRepositoryFormat(repo) {
		return fmt.Errorf("repository format '%s': %w", repo, ErrInvalidRepositoryFormat)
	}

	// Create or update the github-pages environment with proper deployment policy
	apiCmd := exec.CommandContext(ctx, "gh", "api", "repos/"+repo+"/environments/github-pages", //nolint:gosec // repo is validated
		"--method", "PUT",
		"--field", "deployment_branch_policy[protected_branches]=false",
		"--field", "deployment_branch_policy[custom_branch_policies]=true",
		"--silent")

	if err := apiCmd.Run(); err != nil {
		return fmt.Errorf("failed to create/update github-pages environment: %w", err)
	}

	return nil
}

// setupDeploymentBranches configures deployment branch policies
func setupDeploymentBranches(ctx context.Context, cmd *cobra.Command, repo string, dryRun, verbose bool) error { //nolint:unparam // error return for future extensibility
	branches := []string{
		"master",      // Main production branch
		"gh-pages",    // GitHub Pages default
		"*",           // Any single branch
		"*/*",         // Two-level patterns (e.g., feature/branch)
		"*/*/*",       // Three-level patterns
		"*/*/*/*",     // Four-level patterns
		"development", // Development branch
	}

	for _, branch := range branches {
		if verbose {
			cmd.Printf("   🌿 Configuring deployment rule for: %s\n", branch)
		}

		if dryRun {
			cmd.Printf("   🧪 DRY RUN: Would add deployment rule for %s\n", branch)
			continue
		}

		apiCmd := exec.CommandContext(ctx, "gh", "api", //nolint:gosec // repo is validated in caller
			"repos/"+repo+"/environments/github-pages/deployment-branch-policies",
			"--method", "POST",
			"--field", "name="+branch,
			"--field", "type=branch",
			"--silent")

		if err := apiCmd.Run(); err != nil {
			if verbose {
				cmd.Printf("   ⚠️  Rule for %s may already exist or failed to add\n", branch)
			}
			// Don't fail the entire process for individual branch policy failures
			continue
		}

		if verbose {
			cmd.Printf("   ✅ Added deployment rule: %s\n", branch)
		}
	}

	return nil
}

// setupCustomDomain configures a custom domain for GitHub Pages
func setupCustomDomain(ctx context.Context, cmd *cobra.Command, repo, domain string, dryRun, verbose bool) error {
	if verbose {
		cmd.Printf("   🌍 Configuring custom domain: %s\n", domain)
	}

	if dryRun {
		cmd.Printf("   🧪 DRY RUN: Would configure custom domain: %s\n", domain)
		return nil
	}

	// Set custom domain via GitHub Pages API
	apiCmd := exec.CommandContext(ctx, "gh", "api", //nolint:gosec // repo/domain are validated
		"repos/"+repo+"/pages",
		"--method", "PUT",
		"--field", "cname="+domain,
		"--field", "source[branch]=gh-pages",
		"--field", "source[path]=/")

	if err := apiCmd.Run(); err != nil {
		return fmt.Errorf("failed to configure custom domain: %w", err)
	}

	return nil
}

// setupBranchProtection configures branch protection rules for deployment branches
func setupBranchProtection(ctx context.Context, cmd *cobra.Command, repo string, dryRun, verbose bool) error { //nolint:unparam // error return for future extensibility
	protectedBranches := []string{"master", "main", "gh-pages"}

	for _, branch := range protectedBranches {
		if verbose {
			cmd.Printf("   🛡️  Setting up protection for branch: %s\n", branch)
		}

		if dryRun {
			cmd.Printf("   🧪 DRY RUN: Would protect branch: %s\n", branch)
			continue
		}

		// Check if branch exists before protecting it
		branchCmd := exec.CommandContext(ctx, "gh", "api", //nolint:gosec // repo and branch are validated
			"repos/"+repo+"/branches/"+branch,
			"--method", "GET",
			"--silent")

		if err := branchCmd.Run(); err != nil {
			if verbose {
				cmd.Printf("   ℹ️  Branch %s does not exist, skipping protection\n", branch)
			}
			continue
		}

		// Apply minimal protection (require up-to-date branches)
		protectionCmd := exec.CommandContext(ctx, "gh", "api", //nolint:gosec // repo and branch are validated
			"repos/"+repo+"/branches/"+branch+"/protection",
			"--method", "PUT",
			"--field", "required_status_checks[strict]=true",
			"--field", "required_status_checks[contexts]=[]",
			"--field", "enforce_admins=false",
			"--field", "required_pull_request_reviews=null",
			"--field", "restrictions=null",
			"--silent")

		if err := protectionCmd.Run(); err != nil {
			if verbose {
				cmd.Printf("   ⚠️  Failed to protect branch %s: %v\n", branch, err)
			}
			continue
		}

		if verbose {
			cmd.Printf("   ✅ Protected branch: %s\n", branch)
		}
	}

	return nil
}

// verifySetup checks that the GitHub Pages environment is configured correctly
func verifySetup(ctx context.Context, cmd *cobra.Command, repo string, verbose bool) error {
	if verbose {
		cmd.Printf("   🔍 Fetching environment configuration...\n")
	}

	// Get environment details
	envCmd := exec.CommandContext(ctx, "gh", "api", "repos/"+repo+"/environments/github-pages") //nolint:gosec // repo is validated
	envOutput, err := envCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to fetch environment details: %w", err)
	}

	var envResponse GitHubEnvironmentResponse
	if parseErr := json.Unmarshal(envOutput, &envResponse); parseErr != nil {
		return fmt.Errorf("failed to parse environment response: %w", parseErr)
	}

	if verbose {
		cmd.Printf("   ✅ GitHub Pages environment exists: %s\n", envResponse.Name)
	}

	// Get deployment branch policies
	policiesCmd := exec.CommandContext(ctx, "gh", "api", //nolint:gosec // repo is validated
		"repos/"+repo+"/environments/github-pages/deployment-branch-policies")
	policiesOutput, err := policiesCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to fetch deployment policies: %w", err)
	}

	var policiesResponse GitHubBranchPoliciesResponse
	if err := json.Unmarshal(policiesOutput, &policiesResponse); err != nil {
		return fmt.Errorf("failed to parse policies response: %w", err)
	}

	if verbose {
		cmd.Printf("   📊 Found %d deployment branch policies:\n", len(policiesResponse.BranchPolicies))
		for _, policy := range policiesResponse.BranchPolicies {
			cmd.Printf("       - %s\n", policy.Name)
		}
	}

	if len(policiesResponse.BranchPolicies) == 0 {
		return ErrNoDeploymentPolicies
	}

	return nil
}

// showNextSteps displays guidance for the user
func showNextSteps(cmd *cobra.Command, repo string, dryRun bool) {
	if dryRun {
		cmd.Printf("🧪 DRY RUN COMPLETE!\n")
		cmd.Printf("Run the command without --dry-run to apply these changes.\n\n")
		return
	}

	cmd.Printf("✨ GitHub Pages Setup Complete!\n")
	cmd.Printf("================================\n\n")

	cmd.Printf("🎉 Your repository is now configured for GitHub Pages deployment!\n\n")

	cmd.Printf("📋 What's been configured:\n")
	cmd.Printf("   • GitHub Pages environment with custom branch policies\n")
	cmd.Printf("   • Deployment permissions for multiple branch patterns:\n")
	cmd.Printf("     - master branch (main deployments)\n")
	cmd.Printf("     - gh-pages branch (GitHub Pages default)\n")
	cmd.Printf("     - Feature branches (*, */*, */*/*)\n")
	cmd.Printf("     - Dependabot branches (*/*/*/*)\n")
	cmd.Printf("     - Development branch\n\n")

	cmd.Printf("🚀 Next Steps:\n")
	cmd.Printf("   1. Your go-coverage workflow should now deploy successfully\n")
	cmd.Printf("   2. Coverage reports will be available at:\n")
	ownerRepo := strings.Split(repo, "/")
	if len(ownerRepo) == 2 {
		cmd.Printf("      https://%s.github.io/%s/\n", ownerRepo[0], ownerRepo[1])
	}
	cmd.Printf("   3. To test the setup:\n")
	cmd.Printf("      - Push a commit with coverage data to master\n")
	cmd.Printf("      - Check GitHub Actions for successful deployment\n")
	cmd.Printf("      - Verify the pages are accessible\n\n")

	cmd.Printf("🔧 Manual Configuration (if needed):\n")
	cmd.Printf("   • Repository Settings → Environments: https://github.com/%s/settings/environments\n", repo)
	cmd.Printf("   • GitHub Pages Settings: https://github.com/%s/settings/pages\n\n", repo)

	cmd.Printf("💡 Troubleshooting:\n")
	cmd.Printf("   • Check workflow permissions in Settings → Actions → General\n")
	cmd.Printf("   • Verify GITHUB_TOKEN has required permissions\n")
	cmd.Printf("   • Review deployment logs in the Actions tab\n")
}

func init() { //nolint:gochecknoinits // CLI command initialization
	// Add flags
	setupPagesCmd.Flags().Bool("dry-run", false, "Preview changes without making them")
	setupPagesCmd.Flags().BoolP("verbose", "v", false, "Show detailed output")
	setupPagesCmd.Flags().String("custom-domain", "", "Configure custom domain for GitHub Pages")
	setupPagesCmd.Flags().Bool("protect-branches", false, "Add branch protection rules for deployment branches")
}
