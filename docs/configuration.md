# ⚙️ Configuration

Complete configuration reference for **go-coverage**, including environment variables, configuration files, and advanced settings.

## 📖 Table of Contents

- [Overview](#-overview)
- [Two-File Configuration System](#-two-file-configuration-system)
- [Environment Variables](#-environment-variables)
- [Configuration File](#-configuration-file)
- [GitHub Integration](#-github-integration)
- [Coverage Settings](#-coverage-settings)
- [Badge Configuration](#-badge-configuration)
- [Report Settings](#-report-settings)
- [History Tracking](#-history-tracking)
- [Advanced Options](#-advanced-options)

## 🎯 Overview

Go-coverage supports configuration through multiple methods:

1. **Command Line Flags** - Override specific options per command (highest priority)
2. **Environment Variables** - Set via modular env files in [`.github/env/`](.github/env/README.md)
3. **Configuration Files** - JSON-based configuration for complex setups
4. **Runtime Detection** - Automatic GitHub context detection (lowest priority)

## 🗂️ Modular Configuration System

The GoFortress workflow system uses a modular, layered configuration approach via [`.github/env/`](.github/env/README.md).

Files are loaded in numeric order, with later files overriding earlier ones:

| File | Purpose |
|------|---------|
| `00-core.env` | Foundation — Go versions, runners, feature flags, timeouts |
| `10-*.env` | Tool configuration — coverage, security, linting, pre-commit |
| `20-*.env` | Service configuration — workflows, Redis, etc. |
| `90-project.env` | **Project-specific overrides** (this is where you customize) |
| `99-local.env` | Local developer overrides (not tracked in git) |

### Usage Example

1. **Edit your project configuration:**
```bash
# In .github/env/90-project.env
GO_COVERAGE_THRESHOLD=85.0
ENABLE_BENCHMARKS=false
PRIMARY_RUNNER=macos-15
```

2. **The system automatically merges all files in order:**
- Core defaults → Tool config → Service config → Project overrides = Final configuration

## 🌍 Environment Variables

### Coverage Analysis

Control how coverage data is processed and analyzed.

```bash
# Core Coverage Settings
export GO_COVERAGE_INPUT_FILE="coverage.txt"          # Input coverage file
export GO_COVERAGE_OUTPUT_DIR="coverage"              # Output directory
export GO_COVERAGE_THRESHOLD=80.0                     # Minimum coverage threshold (0-100)

# Coverage Exclusions
export GO_COVERAGE_EXCLUDE_PATHS="vendor/,test/,testdata/"  # Comma-separated paths to exclude
export GO_COVERAGE_EXCLUDE_FILES="*_test.go,*.pb.go"       # Comma-separated file patterns to exclude
export GO_COVERAGE_EXCLUDE_TESTS=true                      # Exclude test files from coverage
export GO_COVERAGE_EXCLUDE_GENERATED=true                  # Exclude generated files

# Threshold Override (PR Labels)
export GO_COVERAGE_ALLOW_LABEL_OVERRIDE=false         # Allow PR labels to override thresholds
export GO_COVERAGE_MIN_OVERRIDE_THRESHOLD=50.0        # Minimum allowed override threshold
export GO_COVERAGE_MAX_OVERRIDE_THRESHOLD=95.0        # Maximum allowed override threshold
```

### GitHub Integration

Configure GitHub API access and integration features.

```bash
# GitHub API Configuration
export GITHUB_TOKEN="ghp_your_token_here"             # GitHub API token (required)
export GITHUB_REPOSITORY="owner/repo"                 # Repository in owner/repo format
export GITHUB_REPOSITORY_OWNER="owner"                # Repository owner
export GITHUB_SHA="commit_sha"                        # Current commit SHA

# GitHub Actions Context (auto-detected)
export GITHUB_REF_NAME="main"                         # Current branch name
export GITHUB_HEAD_REF="feature-branch"               # PR source branch
export GITHUB_PR_NUMBER="123"                         # Pull request number

# GitHub Integration Features
export GO_COVERAGE_POST_COMMENTS=true                 # Enable PR comments
export GO_COVERAGE_UPDATE_STATUS=true                 # Enable status checks
export GO_COVERAGE_ENABLE_PAGES=true                  # Enable GitHub Pages deployment
```

### Badge Generation

Customize coverage badge appearance and behavior.

```bash
# Badge Styling
export GO_COVERAGE_BADGE_STYLE="flat"                 # Badge style: flat, flat-square, for-the-badge, plastic
export GO_COVERAGE_BADGE_LOGO="go"                 # Logo: nodejs, python, react, docker, etc. (Simple Icons) or custom URL
export GO_COVERAGE_BADGE_LOGO_COLOR="white"           # Logo color: white, red, blue, green, etc.
export GO_COVERAGE_BADGE_LABEL="coverage"             # Badge label text

# Badge Colors
export GO_COVERAGE_BADGE_COLOR_EXCELLENT="brightgreen" # Color for excellent coverage (90%+)
export GO_COVERAGE_BADGE_COLOR_GOOD="yellow"          # Color for good coverage (70-89%)
export GO_COVERAGE_BADGE_COLOR_POOR="red"             # Color for poor coverage (<70%)

# Badge Generation
export GO_COVERAGE_GENERATE_BADGE=true                # Enable badge generation
export GO_COVERAGE_BADGE_FILENAME="coverage.svg"      # Badge filename
```

### Report Generation

Configure HTML report generation and styling.

```bash
# Report Settings
export GO_COVERAGE_REPORT_TITLE="Coverage Report"     # Report title
export GO_COVERAGE_REPORT_THEME="github-light"        # Theme: github-light, github-dark, light
export GO_COVERAGE_SHOW_PACKAGE_LIST=true             # Show package breakdown
export GO_COVERAGE_SHOW_FILE_LIST=true                # Show file-level details

# Report Features
export GO_COVERAGE_ENABLE_SEARCH=true                 # Enable search functionality
export GO_COVERAGE_ENABLE_FILTERS=true                # Enable filtering options
export GO_COVERAGE_RESPONSIVE_DESIGN=true             # Enable responsive design
```

### History Tracking

Configure coverage history and trend analysis.

```bash
# History Settings
export GO_COVERAGE_ENABLE_HISTORY=true                # Enable history tracking
export GO_COVERAGE_HISTORY_RETENTION_DAYS=90          # Days to retain history data
export GO_COVERAGE_HISTORY_MAX_ENTRIES=1000           # Maximum history entries

# Trend Analysis
export GO_COVERAGE_ENABLE_TREND_ANALYSIS=true         # Enable trend calculations
export GO_COVERAGE_TREND_WINDOW_DAYS=30               # Days for trend analysis window
```

### Branch Configuration

Configure branch handling and main branch detection.

```bash
# Branch Settings
export MAIN_BRANCHES="master,main"                    # Comma-separated main branches
export DEFAULT_MAIN_BRANCH="master"                   # Primary main branch
export GO_COVERAGE_AUTO_DETECT_BRANCH=true            # Auto-detect current branch
```

### Advanced Settings

Additional configuration options for fine-tuning behavior.

```bash
# Performance Settings
export GO_COVERAGE_ENABLE_CACHING=true                # Enable result caching
export GO_COVERAGE_CACHE_TTL=3600                     # Cache TTL in seconds
export GO_COVERAGE_PARALLEL_PROCESSING=true           # Enable parallel processing

# Storage Settings
export GO_COVERAGE_STORAGE_TYPE="github-pages"        # Storage backend
export GO_COVERAGE_COMPRESS_REPORTS=true              # Compress generated reports
export GO_COVERAGE_CLEANUP_OLD_REPORTS=true           # Auto-cleanup old reports

# Logging Configuration
export GO_COVERAGE_LOG_LEVEL="info"                   # Log level: debug, info, warn, error
export GO_COVERAGE_LOG_FORMAT="text"                  # Log format: text, json, pretty
export GO_COVERAGE_ENABLE_DEBUG=false                 # Enable debug mode
```

## 📄 Configuration File

Create `.go-coverage.json` in your repository root for complex configurations:

```json
{
  "coverage": {
    "input_file": "coverage.txt",
    "output_dir": "coverage",
    "threshold": 80.0,
    "allow_label_override": false,
    "min_override_threshold": 50.0,
    "max_override_threshold": 95.0,
    "exclude_paths": [
      "vendor/",
      "test/",
      "testdata/",
      "cmd/",
      "internal/generated/"
    ],
    "exclude_files": [
      "*_test.go",
      "*.pb.go",
      "*_gen.go",
      "*_mock.go"
    ],
    "exclude_tests": true,
    "exclude_generated": true
  },
  "github": {
    "post_comments": true,
    "update_status": true,
    "enable_pages": true,
    "comment_template": "default",
    "status_context": "coverage/go-coverage"
  },
  "badge": {
    "style": "flat",
    "logo": "go",
    "label": "coverage",
    "color_excellent": "brightgreen",
    "color_good": "yellow",
    "color_poor": "red",
    "generate_badge": true,
    "filename": "coverage.svg"
  },
  "report": {
    "title": "Go Coverage Report",
    "theme": "github-light",
    "show_package_list": true,
    "show_file_list": true,
    "enable_search": true,
    "enable_filters": true,
    "responsive_design": true
  },
  "history": {
    "enabled": true,
    "retention_days": 90,
    "max_entries": 1000,
    "enable_trend_analysis": true,
    "trend_window_days": 30
  },
  "storage": {
    "type": "github-pages",
    "compress_reports": true,
    "cleanup_old_reports": true
  },
  "log": {
    "level": "info",
    "format": "text",
    "enable_debug": false
  },
  "analytics": {
    "enable_package_breakdown": true,
    "enable_file_analysis": true,
    "enable_predictions": false
  }
}
```

## 🔗 GitHub Integration

### Required Permissions

Your `GITHUB_TOKEN` needs these permissions:

```yaml
permissions:
  contents: read          # Read repository content
  pages: write           # Deploy to GitHub Pages
  id-token: write        # GitHub Pages deployment
  pull-requests: write   # Create/update PR comments
  statuses: write        # Create status checks
```

### GitHub Actions Setup

Environment variables are typically configured in GitHub Actions:

```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  GO_COVERAGE_THRESHOLD: 80
  GO_COVERAGE_EXCLUDE_PATHS: "vendor/,test/"
  GO_COVERAGE_POST_COMMENTS: true
```

### GitHub Pages Configuration

The `setup-pages` command configures these automatically:

- **Environment**: `github-pages`
- **Deployment branches**: `master`, `main`, and all PR branches
- **Custom domain**: Optional custom domain support

## 📊 Coverage Settings

### Threshold Configuration

```bash
# Basic threshold
export GO_COVERAGE_THRESHOLD=80.0

# Color thresholds for badges
export GO_COVERAGE_THRESHOLD_EXCELLENT=90.0    # Green badge
export GO_COVERAGE_THRESHOLD_GOOD=70.0         # Yellow badge
# Below 70% = Red badge
```

### Exclusion Patterns

#### Path Exclusions

```bash
# Exclude specific directories
export GO_COVERAGE_EXCLUDE_PATHS="vendor/,test/,examples/,scripts/"
```

#### File Pattern Exclusions

```bash
# Exclude specific file patterns
export GO_COVERAGE_EXCLUDE_FILES="*_test.go,*.pb.go,*_gen.go,mock_*.go"
```

#### Smart Exclusions

```bash
# Automatically exclude test and generated files
export GO_COVERAGE_EXCLUDE_TESTS=true
export GO_COVERAGE_EXCLUDE_GENERATED=true
```

### Threshold Override via PR Labels

Allow PR labels to temporarily override coverage thresholds:

```bash
export GO_COVERAGE_ALLOW_LABEL_OVERRIDE=true
```

When enabled, PRs with the `coverage-override` label will completely bypass coverage threshold checks. This provides a simple on/off override mechanism for special cases.

## 🏷️ Badge Configuration

### Available Styles

- **flat** (default): Clean, flat design
- **flat-square**: Square corners
- **for-the-badge**: Large, bold style
- **plastic**: 3D appearance

### Logo Options

```bash
# Example logo (for testing)
export GO_COVERAGE_BADGE_LOGO="example"     # Simple star icon for testing/documentation

# Simple Icons (supports 2800+ logos - see https://simpleicons.org)
export GO_COVERAGE_BADGE_LOGO="go"          # Go logo
export GO_COVERAGE_BADGE_LOGO="python"      # Python logo
export GO_COVERAGE_BADGE_LOGO="rust"        # Rust logo
export GO_COVERAGE_BADGE_LOGO="docker"      # Docker logo
export GO_COVERAGE_BADGE_LOGO="kubernetes"  # Kubernetes logo
export GO_COVERAGE_BADGE_LOGO="react"       # React logo
export GO_COVERAGE_BADGE_LOGO="typescript"  # TypeScript logo
export GO_COVERAGE_BADGE_LOGO="postgresql"  # PostgreSQL logo
export GO_COVERAGE_BADGE_LOGO="redis"       # Redis logo
# ... and 2800+ more! See https://simpleicons.org for full list
# Note: Use lowercase names with hyphens (e.g., "visual-studio-code", not "Visual Studio Code")

# Custom logo URL
export GO_COVERAGE_BADGE_LOGO="https://example.com/logo.svg"

# No logo
export GO_COVERAGE_BADGE_LOGO=""

# Logo colorization (applied to compatible SVG logos)
export GO_COVERAGE_BADGE_LOGO_COLOR="white"   # Default
export GO_COVERAGE_BADGE_LOGO_COLOR="red"     # Red logo
export GO_COVERAGE_BADGE_LOGO_COLOR="#3498db" # Custom hex color
```

### Color Customization

```bash
# Predefined colors
export GO_COVERAGE_BADGE_COLOR_EXCELLENT="brightgreen"
export GO_COVERAGE_BADGE_COLOR_GOOD="yellow"
export GO_COVERAGE_BADGE_COLOR_POOR="red"

# Custom hex colors
export GO_COVERAGE_BADGE_COLOR_EXCELLENT="#00ff00"
export GO_COVERAGE_BADGE_COLOR_GOOD="#ffff00"
export GO_COVERAGE_BADGE_COLOR_POOR="#ff0000"
```

## 📋 Report Settings

### Theme Options

- **github-light**: Light theme matching GitHub's design
- **github-dark**: Dark theme matching GitHub's dark mode
- **light**: Clean light theme

### Report Features

```bash
# Content Options
export GO_COVERAGE_SHOW_PACKAGE_LIST=true      # Package breakdown table
export GO_COVERAGE_SHOW_FILE_LIST=true         # File-level details
export GO_COVERAGE_ENABLE_SEARCH=true          # Search functionality
export GO_COVERAGE_ENABLE_FILTERS=true         # Filter controls

# Design Options
export GO_COVERAGE_RESPONSIVE_DESIGN=true      # Mobile-friendly layout
export GO_COVERAGE_ENABLE_DARK_MODE=true       # Dark mode toggle
```

## 📈 History Tracking

### Data Retention

```bash
# Time-based retention
export GO_COVERAGE_HISTORY_RETENTION_DAYS=90   # Keep 90 days of data

# Count-based retention
export GO_COVERAGE_HISTORY_MAX_ENTRIES=1000    # Keep 1000 latest entries

# Trend analysis window
export GO_COVERAGE_TREND_WINDOW_DAYS=30        # 30-day trend analysis
```

### History Features

```bash
# Enable history tracking
export GO_COVERAGE_ENABLE_HISTORY=true

# Enable trend analysis
export GO_COVERAGE_ENABLE_TREND_ANALYSIS=true

# Enable predictions (experimental)
export GO_COVERAGE_ENABLE_PREDICTIONS=false
```

## 🔧 Advanced Options

### Performance Tuning

```bash
# Caching
export GO_COVERAGE_ENABLE_CACHING=true
export GO_COVERAGE_CACHE_TTL=3600              # 1 hour cache

# Parallel Processing
export GO_COVERAGE_PARALLEL_PROCESSING=true
export GO_COVERAGE_MAX_WORKERS=4               # Parallel worker count
```

### Storage Options

```bash
# Storage Backend
export GO_COVERAGE_STORAGE_TYPE="github-pages" # github-pages, local, s3

# Compression
export GO_COVERAGE_COMPRESS_REPORTS=true       # Compress HTML reports
export GO_COVERAGE_COMPRESS_LEVEL=6            # Compression level (1-9)

# Cleanup
export GO_COVERAGE_CLEANUP_OLD_REPORTS=true    # Auto-cleanup old reports
export GO_COVERAGE_CLEANUP_RETENTION_DAYS=30   # Days to keep old reports
```

### Debug and Logging

```bash
# Debug Mode
export GO_COVERAGE_ENABLE_DEBUG=true           # Enable debug logging
export GO_COVERAGE_VERBOSE_OUTPUT=true         # Verbose command output

# Log Configuration
export GO_COVERAGE_LOG_LEVEL="debug"           # debug, info, warn, error
export GO_COVERAGE_LOG_FORMAT="json"           # text, json, pretty
export GO_COVERAGE_LOG_FILE="coverage.log"     # Log to file (optional)
```

## 🎯 Common Configurations

### Minimal Setup

```bash
export GITHUB_TOKEN="${{ secrets.GITHUB_TOKEN }}"
export GO_COVERAGE_THRESHOLD=75
```

### Comprehensive Setup

```bash
# Core settings
export GITHUB_TOKEN="${{ secrets.GITHUB_TOKEN }}"
export GO_COVERAGE_THRESHOLD=80
export GO_COVERAGE_EXCLUDE_PATHS="vendor/,test/,examples/"
export GO_COVERAGE_EXCLUDE_FILES="*_test.go,*.pb.go"

# Badge customization
export GO_COVERAGE_BADGE_STYLE="flat"
export GO_COVERAGE_BADGE_LOGO="go"

# Report features
export GO_COVERAGE_REPORT_THEME="github-light"
export GO_COVERAGE_SHOW_PACKAGE_LIST=true

# History tracking
export GO_COVERAGE_ENABLE_HISTORY=true
export GO_COVERAGE_HISTORY_RETENTION_DAYS=90
```

### High-Security Setup

```bash
# Minimal GitHub integration
export GITHUB_TOKEN="${{ secrets.GITHUB_TOKEN }}"
export GO_COVERAGE_POST_COMMENTS=false         # Disable PR comments
export GO_COVERAGE_UPDATE_STATUS=false         # Disable status checks
export GO_COVERAGE_ENABLE_PAGES=false          # Disable Pages deployment

# Local-only processing
export GO_COVERAGE_STORAGE_TYPE="local"
export GO_COVERAGE_OUTPUT_DIR="./coverage-reports"
```

---

For more information, see:
- [Quickstart Guide](quickstart.md) - Getting started in 5 minutes
- [User Guide](user-guide.md) - Complete usage examples
- [CLI Reference](cli-reference.md) - Command-line options
