# 🛠️ CLI Reference

Complete command-line reference for the **go-coverage** tool.

## 📖 Table of Contents

- [Global Options](#-global-options)
- [complete](#complete---full-pipeline)
- [parse](#parse---coverage-analysis)
- [comment](#comment---pr-comments)
- [history](#history---coverage-history)
- [setup-pages](#setup-pages---github-pages-setup)
- [upgrade](#upgrade---tool-updates)
- [Examples](#-examples)

## 🌐 Global Options

All commands support these global flags:

```bash
go-coverage [command] [flags]

Global Flags:
      --debug               Enable debug mode with detailed logging
      --log-format string   Log format: text, json, pretty (default "text")
  -l, --log-level string    Log level: debug, info, warn, error (default "info")
  -h, --help                Show help information
  -v, --version             Show version information
```

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
go-coverage --debug complete -i coverage.txt
```

### Log Formats

- **text**: Human-readable output (default)
- **json**: Structured JSON logs for automation
- **pretty**: Colorized output for terminal use

## `complete` - Full Pipeline

Run the complete coverage processing pipeline in a single command.

### Usage

```bash
go-coverage complete [flags]
```

### Description

Executes the complete coverage pipeline:
1. Parse coverage data
2. Generate badges
3. Create HTML reports
4. Update coverage history
5. Deploy to output directory
6. Create GitHub PR comments (if applicable)
7. Update GitHub status checks

### Flags

```bash
  -i, --input string    Input coverage file path
  -o, --output string   Output directory for generated files
      --dry-run         Preview operations without making changes
      --skip-github     Skip GitHub integration features
      --skip-history    Skip history tracking and trend analysis
  -h, --help            Show help for this command
```

### Examples

```bash
# Basic usage
go-coverage complete -i coverage.txt

# Custom output directory
go-coverage complete -i coverage.txt -o coverage-reports

# Preview without changes
go-coverage complete -i coverage.txt --dry-run

# Skip GitHub features for local use
go-coverage complete -i coverage.txt --skip-github
```

## `parse` - Coverage Analysis

Parse Go coverage profile files and analyze coverage data.

### Usage

```bash
go-coverage parse [flags]
```

### Description

Analyzes Go coverage data and outputs results in various formats. Can validate coverage against thresholds and save results to files.

### Flags

```bash
  -f, --file string       Path to coverage profile file (default "coverage.txt")
      --format string     Output format: text, json (default "text")
  -o, --output string     Output file path (writes to stdout if not specified)
      --threshold float   Coverage threshold percentage (0-100)
  -h, --help              Show help for this command
```

### Examples

```bash
# Parse and display coverage
go-coverage parse -f coverage.txt

# Output as JSON
go-coverage parse -f coverage.txt --format json

# Save results to file
go-coverage parse -f coverage.txt -o coverage-analysis.json

# Check against threshold
go-coverage parse -f coverage.txt --threshold 80

# Combine options
go-coverage parse -f coverage.txt --format json --threshold 85 -o results.json
```

### Output Formats

#### Text Format (Default)

```
Coverage Analysis Results
=========================
Overall Coverage: 87.4% (1,247/1,426 lines)
Packages: 15 analyzed

Package Breakdown:
  github.com/mrz1836/go-coverage/internal/parser: 92.1%
  github.com/mrz1836/go-coverage/internal/badge: 89.3%
  ...
```

#### JSON Format

```json
{
  "mode": "atomic",
  "total_lines": 1426,
  "covered_lines": 1247,
  "percentage": 87.4,
  "packages": {
    "github.com/mrz1836/go-coverage/internal/parser": {
      "name": "parser",
      "total_lines": 156,
      "covered_lines": 144,
      "percentage": 92.1
    }
  }
}
```

## `comment` - PR Comments

Create or update GitHub pull request comments with coverage analysis.

### Usage

```bash
go-coverage comment [flags]
```

### Description

Creates intelligent PR comments with coverage information, including:
- Overall coverage analysis
- Coverage comparison with base branch
- Package-level breakdown
- File-level changes
- PR-specific badges

### Flags

```bash
  -p, --pr int                 Pull request number (required)
  -c, --coverage string        Path to current coverage profile file
      --base-coverage string   Path to base branch coverage for comparison
      --badge-url string       Custom badge URL override
      --report-url string      Custom report URL override
      --anti-spam              Enable anti-spam features (default true)
      --enable-analysis        Enable code quality analysis (default true)
      --generate-badges        Generate PR-specific badges
      --block-merge            Block PR merge on coverage failure
      --status                 Create GitHub commit status (default true)
      --dry-run                Preview comment without posting
  -h, --help                   Show help for this command
```

### Examples

```bash
# Basic PR comment
go-coverage comment -p 123 -c coverage.txt

# With base coverage comparison
go-coverage comment -p 123 -c coverage.txt --base-coverage main-coverage.txt

# Preview comment without posting
go-coverage comment -p 123 -c coverage.txt --dry-run

# Generate PR-specific badge
go-coverage comment -p 123 -c coverage.txt --generate-badges

# Block merge on coverage failure
go-coverage comment -p 123 -c coverage.txt --block-merge
```

### Environment Variables

Required for GitHub integration:

```bash
export GITHUB_TOKEN="your_github_token"
export GITHUB_REPOSITORY="owner/repo"
export GITHUB_REF_NAME="feature-branch"
```

## `history` - Coverage History

Manage and view coverage history and trends.

### Usage

```bash
go-coverage history [flags]
```

### Description

Track coverage changes over time, analyze trends, and manage historical data.

### Flags

```bash
      --branch string     Git branch name for history tracking
      --days int          Number of days to include in history (default 30)
      --format string     Output format: table, json, yaml (default "table")
      --trend             Include trend analysis in output
  -h, --help              Show help for this command
```

### Examples

```bash
# View recent history for current branch
go-coverage history

# Specific branch history
go-coverage history --branch main

# Extended time range
go-coverage history --branch main --days 90

# JSON output for automation
go-coverage history --branch main --format json

# Include trend analysis
go-coverage history --branch main --trend
```

## `setup-pages` - GitHub Pages Setup

Configure GitHub Pages environment for coverage deployment.

### Usage

```bash
go-coverage setup-pages [repository] [flags]
```

### Description

Automatically configures GitHub Pages settings, deployment permissions, and environment protection rules for coverage reports.

### Arguments

```bash
repository    GitHub repository in format owner/repo (auto-detected if not provided)
```

### Flags

```bash
      --custom-domain string   Custom domain for GitHub Pages
      --dry-run                Preview changes without applying them
      --verbose                Show detailed configuration steps
  -h, --help                   Show help for this command
```

### Examples

```bash
# Auto-detect repository and configure
go-coverage setup-pages

# Specify repository explicitly
go-coverage setup-pages owner/repo

# Preview changes without applying
go-coverage setup-pages --dry-run

# Configure with custom domain
go-coverage setup-pages --custom-domain coverage.example.com

# Verbose output for debugging
go-coverage setup-pages --verbose
```

### What It Configures

1. **GitHub Pages Environment** with proper deployment branches
2. **Environment Protection Rules** for secure deployments
3. **Deployment Permissions** for automated workflows
4. **Branch Policies** allowing deployments from main and PR branches

### Requirements

- GitHub CLI (`gh`) must be installed and authenticated
- Repository write permissions
- Admin access to configure environments

## `upgrade` - Tool Updates

Check for updates and upgrade the go-coverage tool.

### Usage

```bash
go-coverage upgrade [flags]
```

### Description

Manages go-coverage tool updates, including checking for new versions and performing upgrades.

### Flags

```bash
      --check     Check for updates without installing
      --force     Force reinstall even if already on latest version
      --verbose   Show detailed upgrade process
  -h, --help      Show help for this command
```

### Examples

```bash
# Check for available updates
go-coverage upgrade --check

# Upgrade to latest version
go-coverage upgrade

# Force reinstall current version
go-coverage upgrade --force

# Verbose upgrade process
go-coverage upgrade --verbose
```

## 📚 Examples

### Complete Workflow

```bash
# 1. Run tests with coverage
go test -coverprofile=coverage.txt -covermode=atomic ./...

# 2. Process coverage with complete pipeline
go-coverage complete -i coverage.txt -o coverage-output

# 3. For pull requests, add comment
go-coverage comment -p 123 -c coverage.txt
```

### CI/CD Integration

```bash
# GitHub Actions workflow
- name: Generate Coverage
  run: |
    go test -coverprofile=coverage.txt ./...
    go-coverage complete -i coverage.txt
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

- name: PR Comment
  if: github.event_name == 'pull_request'
  run: |
    go-coverage comment \
      -p ${{ github.event.number }} \
      -c coverage.txt \
      --generate-badges
```

### Local Development

```bash
# Quick coverage check
go test -coverprofile=coverage.txt ./... && \
go-coverage parse -f coverage.txt --threshold 80

# Generate local reports
go-coverage complete -i coverage.txt --skip-github -o local-coverage
```

### Debugging and Troubleshooting

```bash
# Enable debug mode for detailed logging
go-coverage --debug complete -i coverage.txt

# Preview operations without changes
go-coverage complete -i coverage.txt --dry-run

# Check configuration and setup
go-coverage setup-pages --dry-run --verbose
```

---

For more information, see:
- [User Guide](user-guide.md) - Complete usage examples
- [Configuration](configuration.md) - Environment variables and settings
- [Quickstart](quickstart.md) - Getting started in 5 minutes
