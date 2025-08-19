# 📚 User Guide

Complete guide to using **go-coverage** for professional code coverage tracking, reporting, and analysis.

## 📖 Table of Contents

- [Overview](#-overview)
- [Installation](#-installation)
- [Basic Workflow](#-basic-workflow)
- [CLI Commands](#-cli-commands)
- [GitHub Integration](#-github-integration)
- [Coverage Reports](#-coverage-reports)
- [Badge Generation](#-badge-generation)
- [History Tracking](#-history-tracking)
- [Configuration](#-configuration)
- [Best Practices](#-best-practices)

## 🎯 Overview

Go Coverage is a self-contained, Go-native coverage system that completely replaces Codecov with zero external dependencies. It provides:

- **Professional Coverage Tracking** - Parse Go coverage profiles with exclusions and thresholds
- **Badge Generation** - Create SVG badges with custom styling and themes
- **HTML Reports** - Build responsive dashboards and detailed coverage reports
- **GitHub Integration** - PR comments, commit statuses, and automated deployments
- **History Tracking** - Monitor coverage trends over time with retention policies

## 📦 Installation

### CLI Tool
```bash
go install github.com/mrz1836/go-coverage/cmd/go-coverage@latest
```

### Library
```bash
go get -u github.com/mrz1836/go-coverage
```

### Verify Installation
```bash
go-coverage --version
```

## 🔄 Basic Workflow

### 1. Generate Coverage Data

```bash
# Run tests with coverage
go test -coverprofile=coverage.txt ./...

# For more detailed coverage
go test -coverprofile=coverage.txt -covermode=atomic ./...
```

### 2. Process Coverage

```bash
# Complete pipeline (recommended)
go-coverage complete -i coverage.txt

# Individual steps
go-coverage parse -i coverage.txt
go-coverage badge -i coverage.txt
go-coverage report -i coverage.txt
```

### 3. Deploy to GitHub Pages

```bash
# Using GitHub Actions (recommended)
- uses: peaceiris/actions-gh-pages@v3
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    publish_dir: ./coverage

# Manual deployment
go-coverage deploy --dir coverage-output
```

## 🛠️ CLI Commands

### `complete` - Full Pipeline

Executes the complete coverage processing pipeline:

```bash
go-coverage complete [flags]

# Basic usage
go-coverage complete -i coverage.txt

# With custom output directory
go-coverage complete -i coverage.txt -o coverage-reports

# For pull requests
go-coverage complete -i coverage.txt --pr 123

# With threshold enforcement
go-coverage complete -i coverage.txt --threshold 80 --fail-under
```

**Flags:**
- `-i, --input string` - Input coverage file (default: "coverage.txt")
- `-o, --output string` - Output directory (default: "coverage")
- `--threshold float` - Coverage threshold percentage
- `--fail-under` - Fail if coverage is below threshold
- `--pr string` - Pull request number for PR-specific reports

### `parse` - Coverage Analysis

Parse and analyze coverage files:

```bash
go-coverage parse [flags]

# Basic parsing
go-coverage parse -i coverage.txt

# With exclusions
go-coverage parse -i coverage.txt --exclude "vendor/,test/"

# Output to file
go-coverage parse -i coverage.txt -o analysis.json

# Different output formats
go-coverage parse -i coverage.txt --format yaml
go-coverage parse -i coverage.txt --format table
```

**Flags:**
- `-i, --input string` - Coverage file to parse
- `-o, --output string` - Output file (default: stdout)
- `--format string` - Output format: json, yaml, table (default: "json")
- `--exclude strings` - Paths to exclude
- `--threshold float` - Coverage threshold for validation

### `comment` - PR Comments

Generate GitHub PR comments with coverage analysis:

```bash
go-coverage comment [flags]

# Basic PR comment
go-coverage comment --pr 123 --coverage coverage.txt

# With base coverage comparison
go-coverage comment --pr 123 \
  --coverage coverage.txt \
  --base-coverage main-coverage.txt

# Custom template
go-coverage comment --pr 123 \
  --coverage coverage.txt \
  --template custom-template.md
```

**Flags:**
- `--pr string` - Pull request number (required)
- `--coverage string` - Current coverage file
- `--base-coverage string` - Base coverage for comparison
- `--template string` - Custom comment template
- `--update` - Update existing comment instead of creating new

### `history` - Coverage History

View and manage coverage history:

```bash
go-coverage history [flags]

# View recent history
go-coverage history --branch main

# Specific time range
go-coverage history --branch main --days 30

# Export history data
go-coverage history --branch main --format json > history.json

# Trend analysis
go-coverage history --branch main --trend
```

**Flags:**
- `--branch string` - Git branch name
- `--days int` - Number of days to include (default: 30)
- `--format string` - Output format: table, json, yaml
- `--trend` - Include trend analysis

### `setup-pages` - GitHub Pages Configuration

Configure GitHub Pages environment:

```bash
go-coverage setup-pages [flags]

# Auto-detect repository
go-coverage setup-pages

# Specify repository
go-coverage setup-pages owner/repo

# Dry run (preview changes)
go-coverage setup-pages --dry-run

# Custom domain
go-coverage setup-pages --custom-domain example.com
```

**Flags:**
- `--dry-run` - Preview changes without applying
- `--custom-domain string` - Custom domain for GitHub Pages
- `--verbose` - Show detailed output

### `upgrade` - Update Tool

Check for updates and upgrade:

```bash
go-coverage upgrade [flags]

# Check for updates
go-coverage upgrade --check

# Upgrade to latest
go-coverage upgrade

# Force reinstall
go-coverage upgrade --force
```

**Flags:**
- `--check` - Check for updates without installing
- `--force` - Force reinstall even if up to date
- `--verbose` - Show detailed upgrade process

## 🔗 GitHub Integration

### GitHub Actions Setup

Create `.github/workflows/coverage.yml`:

```yaml
name: Coverage
on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Required for PR diff analysis

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run Tests with Coverage
        run: go test -coverprofile=coverage.txt -covermode=atomic ./...

      - name: Generate Coverage Reports
        run: go-coverage complete -i coverage.txt
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Deploy to GitHub Pages
        if: github.ref == 'refs/heads/main'
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./coverage
```

### PR Comments

Automatic PR comments include:

- **Coverage Summary** - Overall coverage percentage and change
- **Package Breakdown** - Coverage by package with changes highlighted
- **File Analysis** - Files with significant coverage changes
- **Trend Information** - Historical context and trends

### Status Checks

The system can create GitHub status checks that:

- Pass/fail based on coverage thresholds
- Show coverage percentage in the status
- Link to detailed reports
- Block PRs if coverage drops below threshold

## 📊 Coverage Reports

### Dashboard Features

The interactive HTML dashboard includes:

- **Coverage Overview** - Project-wide metrics and trends
- **Package Explorer** - Drill down into package-level coverage
- **File Browser** - Line-by-line coverage visualization
- **History Charts** - Coverage trends over time
- **Search & Filter** - Find specific files or packages

### Report Types

1. **Dashboard** (`index.html`) - Interactive overview
2. **Detailed Reports** (`coverage.html`) - File-level analysis
3. **JSON API** (`coverage.json`) - Machine-readable data
4. **SVG Badges** (`coverage.svg`) - Embeddable badges

### Responsive Design

All reports are optimized for:
- Desktop browsers
- Mobile devices
- Print output
- Dark/light themes

## 🏷️ Badge Generation

### Badge Styles

```bash
# Flat style (default)
go-coverage badge --style flat

# Flat-square style
go-coverage badge --style flat-square

# For-the-badge style
go-coverage badge --style for-the-badge

# Plastic style
go-coverage badge --style plastic
```

### Custom Colors

```bash
# Custom color
go-coverage badge --color brightgreen

# Color by percentage ranges
go-coverage badge --color-excellent brightgreen --color-good yellow --color-poor red
```

### Logos

```bash
# Built-in example logo
go-coverage badge --logo example

# Custom logo URL
go-coverage badge --logo https://example.com/logo.svg
```

### Usage in README

```markdown
![Coverage](https://yourname.github.io/yourrepo/coverage.svg)

<!-- With link to reports -->
[![Coverage](https://yourname.github.io/yourrepo/coverage.svg)](https://yourname.github.io/yourrepo/)
```

## 📈 History Tracking

### Data Storage

Coverage history is stored in:
- `coverage-history.json` - Time-series data
- `coverage-trends.json` - Trend analysis
- Git commit metadata - Long-term backup

### Trend Analysis

The system tracks:
- **Coverage Percentage** - Over time
- **Total Lines** - Code growth
- **Coverage Delta** - Change between commits
- **Package Changes** - New/removed packages

### Retention Policies

Configure data retention:
```bash
# Keep 90 days of history
export GO_COVERAGE_HISTORY_RETENTION_DAYS=90

# Keep 1000 data points maximum
export GO_COVERAGE_HISTORY_MAX_ENTRIES=1000
```

## ⚙️ Configuration

### Environment Variables

Key configuration options:

```bash
# Coverage thresholds
export GO_COVERAGE_FAIL_UNDER=80
export GO_COVERAGE_THRESHOLD_EXCELLENT=90
export GO_COVERAGE_THRESHOLD_GOOD=70

# Exclusions
export GO_COVERAGE_EXCLUDE_PATHS="vendor/,test/"
export GO_COVERAGE_EXCLUDE_FILES="*.pb.go,*_gen.go"

# GitHub integration
export GO_COVERAGE_PR_COMMENT_ENABLED=true
export GO_COVERAGE_PAGES_AUTO_CREATE=true

# Badge styling
export GO_COVERAGE_BADGE_STYLE=flat
export GO_COVERAGE_BADGE_LOGO=go
```

### Configuration File

Create `.go-coverage.json`:

```json
{
  "coverage": {
    "threshold": 80.0,
    "fail_under": true,
    "exclude_paths": ["vendor/", "test/", "cmd/"],
    "exclude_files": ["*.pb.go", "*_gen.go", "*_mock.go"]
  },
  "badge": {
    "style": "flat",
    "logo": "go",
    "label": "coverage",
    "color_excellent": "brightgreen",
    "color_good": "yellow",
    "color_poor": "red"
  },
  "report": {
    "title": "My Project Coverage",
    "theme": "auto",
    "show_package_list": true,
    "show_file_list": true
  },
  "github": {
    "pr_comment_enabled": true,
    "pr_comment_behavior": "update",
    "status_check_enabled": true,
    "pages_auto_create": true
  },
  "history": {
    "enabled": true,
    "retention_days": 90,
    "max_entries": 1000
  }
}
```

## 🎯 Best Practices

### Testing Setup

1. **Use atomic mode** for accurate coverage:
   ```bash
   go test -covermode=atomic -coverprofile=coverage.txt ./...
   ```

2. **Include all packages**:
   ```bash
   go test -coverprofile=coverage.txt ./...
   ```

3. **Exclude test files** from coverage:
   ```bash
   go test -coverpkg=./... -coverprofile=coverage.txt ./...
   ```

### CI/CD Integration

1. **Run on every PR** to catch coverage regressions
2. **Set appropriate thresholds** - aim for 80%+ overall
3. **Use branch protection** to enforce coverage requirements
4. **Cache dependencies** to speed up builds

### Coverage Targets

- **Excellent**: 90%+ coverage
- **Good**: 70-89% coverage
- **Needs Work**: Below 70% coverage

### Exclusions

Commonly excluded paths:
- `vendor/` - Third-party dependencies
- `test/` - Test utilities
- `cmd/` - CLI entry points
- `internal/generated/` - Generated code

### Monitoring

1. **Track trends** over time to catch gradual degradation
2. **Set up alerts** for significant coverage drops
3. **Review reports** regularly in team meetings
4. **Celebrate improvements** to encourage good practices

---

For more detailed information, see:
- [CLI Reference](cli-reference.md) - Complete command documentation
- [Configuration](configuration.md) - All configuration options
- [Architecture](architecture.md) - Technical implementation details
