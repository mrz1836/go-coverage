# CLAUDE.md

## 🤖 Welcome, Claude

You're working with **Go Coverage** - a self-contained, Go-native coverage system that completely replaces Codecov with zero external dependencies. This system provides professional coverage tracking, badge generation, HTML reports, and GitHub Pages deployment.

This repository uses **`AGENTS.md`** as the single source of truth for:

* Go coding conventions (context-first design, interface patterns, error handling)
* Contribution workflows (branch prefixes, commit message style, PR templates)
* Release, CI, and dependency‑management policies
* Security reporting and governance links

> **TL;DR:** **Read `AGENTS.md` first.**
> All technical and procedural questions are answered there.

## 📊 Go Coverage System

This project implements a complete coverage analysis pipeline:

**CLI Tool**: `go-coverage` with 7 core commands:
- `github-actions` - Automated GitHub Actions integration (parse + badge + report + history + deployment)
- `complete` - Full pipeline (parse + badge + report + history + GitHub integration)
- `comment` - Generate PR comments with coverage analysis
- `parse` - Parse coverage data with exclusions and thresholds
- `history` - View coverage trends and historical data
- `setup-pages` - Configure GitHub Pages environment for coverage deployment
- `upgrade` - Check for updates and upgrade to the latest version

**Core Architecture**:
```
cmd/go-coverage/     # CLI application
internal/analytics/          # Dashboard and report generation
internal/badge/              # SVG badge generation
internal/github/             # GitHub API integration
internal/history/            # Coverage history tracking
internal/parser/             # Coverage file parsing
```

## 🎯 Coverage-Specific Guidelines

When working with this coverage system, follow these domain-specific standards:

### **Coverage Analysis Standards**
- **Context-First**: All parsing operations must accept `context.Context` as first parameter
- **Error Handling**: Use wrapped errors with `fmt.Errorf("operation failed: %w", err)`
- **Validation**: Validate coverage thresholds (0-100), file paths, and exclusion patterns
- **Performance**: Keep memory usage under 10MB for parsing large coverage files

### **Badge Generation Requirements**
- **SVG Standards**: Generate valid SVG with proper dimensions and accessibility
- **Styling Consistency**: Support themes (flat, flat-square, for-the-badge, plastic)
- **Performance**: Badge generation must complete in under 100ms
- **Customization**: Support logos, custom colors, and label text

### **GitHub Integration Rules**
- **Rate Limiting**: Respect GitHub API rate limits (5000 requests/hour)
- **Token Security**: Never log or expose GitHub tokens in error messages
- **Context Handling**: Pass contexts through all GitHub API calls for cancellation
- **Error Recovery**: Implement retry logic with exponential backoff

### **Report Generation Standards**
- **Responsive Design**: HTML reports must work on mobile and desktop
- **Asset Management**: CSS/JS assets embedded for offline viewing
- **Performance**: Dashboard generation under 2 seconds for large repositories
- **Accessibility**: Follow WCAG guidelines for coverage visualizations

### **CLI Command Patterns**
- **Flag Consistency**: Use standard flag patterns across all commands
- **Output Formats**: Support multiple output formats (text, json, yaml)
- **Progress Indication**: Show progress for long-running operations
- **Error Messages**: Provide actionable error messages with suggestions

## ✅ Quick Checklist for Claude

1. **Study `AGENTS.md`**
   Understand Go coding standards, context patterns, and interface design
2. **Coverage System Context**
   Remember this replaces Codecov - focus on zero external dependencies
3. **Performance Requirements**
   Keep operations fast for CI/CD environments (see performance table in README)
4. **Follow CLI Patterns**
   Use existing command structure and flag naming conventions
5. **GitHub Integration Security**
   Handle tokens securely, implement proper rate limiting and retry logic
6. **Branch‑prefix and commit‑message standards**
   They drive auto‑labeling and CI gates for the coverage workflow
7. **Never tag releases**
   Only repository code‑owners run `magex version:bump` / `magex release`
8. **Pass CI including coverage**
   Run tests with coverage: `magex test:cover` before opening PR

### **Common Operations**

```bash
# Test the coverage system itself
magex test:cover

# Build CLI tool locally
magex build

# Run the automated GitHub Actions workflow
go-coverage github-actions --input=coverage.txt

# Test GitHub Actions integration
go-coverage github-actions --dry-run --debug

# Run the complete coverage pipeline
go-coverage complete -i coverage.txt -o coverage-output

# Generate PR comment
go-coverage comment --pr 123 --coverage coverage.txt

# Set up GitHub Pages environment
go-coverage setup-pages --dry-run

# Check for updates and upgrade
go-coverage upgrade --check
go-coverage upgrade
```

If you encounter conflicting guidance elsewhere, `AGENTS.md` wins.
Questions about coverage system specifics? Check the CLI help or internal package docs.

Happy coverage hacking! 📊
