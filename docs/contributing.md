# 🤝 Contributing Guide

Thanks for your interest in contributing to **go-coverage**! This guide will help you get started with contributing to our self-contained, Go-native coverage system.

## 📖 Table of Contents

- [Getting Started](#-getting-started)
- [Development Setup](#-development-setup)
- [Code Standards](#-code-standards)
- [Testing Guidelines](#-testing-guidelines)
- [Pull Request Process](#-pull-request-process)
- [Project Structure](#-project-structure)
- [Release Process](#-release-process)

## 🚀 Getting Started

### Prerequisites

- **Go 1.24+** - Latest stable Go version
- **Git** - For version control
- **GitHub CLI** (optional) - For easier GitHub integration testing
- **Make** - For running build tasks

### Quick Start

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/go-coverage.git
   cd go-coverage
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Verify Setup**
   ```bash
   magex test
   go run ./cmd/go-coverage --version
   ```

4. **Run the Complete Test Suite**
   ```bash
   magex test:coverrace     # Full test suite with race detection
   magex lint               # Code quality checks
   magex test:cover         # Generate coverage report
   ```

## 🛠️ Development Setup

### Build the CLI Tool

```bash
# Build locally
magex build

# Install globally
go install ./cmd/go-coverage

# Run without installing
go run ./cmd/go-coverage [command]
```

### Development Workflow

```bash
# 1. Create a feature branch
git checkout -b feat/your-feature-name

# 2. Make your changes and test
magex test
magex lint

# 3. Generate coverage for the coverage system itself
go test -coverprofile=coverage.txt ./...
go run ./cmd/go-coverage complete -i coverage.txt --skip-github

# 4. Commit following conventional commit format
git commit -m "feat: add new feature description"

# 5. Push and create PR
git push origin feat/your-feature-name
```

### Environment Setup

For local development and testing:

```bash
# Optional: Set up test environment
export GO_COVERAGE_ENABLE_DEBUG=true
export GO_COVERAGE_LOG_LEVEL=debug
export GO_COVERAGE_DRY_RUN=true  # Preview changes without applying
```

## 📋 Code Standards

We follow the standards defined in [`.github/AGENTS.md`](../.github/AGENTS.md). Key points:

### Go Code Conventions

- **Context-First Design**: Always pass `context.Context` as the first parameter
- **No Global State**: Use dependency injection instead of package-level variables
- **No `init()` Functions**: Use explicit constructors (`NewXxx()` functions)
- **Interface Design**: Accept interfaces, return concrete types
- **Error Handling**: Always check errors, use wrapped errors with context

### Naming Conventions

- **Packages**: Short, lowercase, one-word (e.g., `parser`, `badge`)
- **Files**: Snake_case (e.g., `badge_generator.go`, `pr_comment_test.go`)
- **Functions**: VerbNoun naming (e.g., `ParseCoverage`, `GenerateBadge`)
- **Variables**: CamelCase for exported, camelCase for internal

### Code Quality

```bash
# Format code
magex format:fix         # Enhanced Go formatting

# Lint code
magex lint               # Run golangci-lint

# Vet code
magex vet                # Run go vet

# All quality checks
magex test               # Includes formatting, linting, and testing
```

## 🧪 Testing Guidelines

### Test Structure

We use **testify** for all tests:

```go
func TestParseStatementCoverage(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        input    string
        expected CoverageData
        wantErr  bool
    }{
        {
            name: "ValidCoverageData",
            input: "mode: atomic\nfile.go:1.1,2.2 1 1\n",
            expected: CoverageData{
                Mode: "atomic",
                // ... expected data
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            result, err := ParseStatement(context.Background(), tt.input)

            if tt.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.expected.Mode, result.Mode)
        })
    }
}
```

### Testing Best Practices

1. **Use `testify/require`** for error checks and critical assertions
2. **Use `testify/assert`** for general comparisons
3. **Table-driven tests** for multiple scenarios
4. **Parallel tests** with `t.Parallel()` when safe
5. **Descriptive test names** that explain the scenario
6. **Test error cases** - ensure error handling works correctly

### Coverage Requirements

- **Minimum 90% coverage** for all packages
- **100% coverage** for critical paths (parsing, badge generation)
- **Test all error conditions** and edge cases
- **Benchmark performance-critical code**

### Running Tests

```bash
# Run all tests
magex test

# Run tests with race detection
magex test:race

# Run tests with coverage
magex test:cover

# Run specific package tests
go test ./internal/parser/...

# Run specific test
go test -run TestParseStatement ./internal/parser/

# Run benchmarks
magex bench
```

## 🔄 Pull Request Process

### Before Submitting

1. **Read the standards**: Review [`.github/AGENTS.md`](../.github/AGENTS.md)
2. **Run all checks**: `magex test` must pass
3. **Update documentation**: Add/update docs for new features
4. **Test locally**: Verify your changes work end-to-end

### PR Requirements

1. **Clear Description**: Use the PR template with:
   - What changed
   - Why it was necessary
   - Testing performed
   - Impact/risk assessment

2. **Conventional Commits**: Follow the commit message format:
   ```
   type(scope): imperative description

   feat(parser): add support for new coverage format
   fix(badge): handle edge case in color calculation
   docs(readme): update installation instructions
   ```

3. **Branch Naming**: Use prefixes:
   - `feat/` - New features
   - `fix/` - Bug fixes
   - `docs/` - Documentation updates
   - `refactor/` - Code refactoring
   - `test/` - Test improvements

### Review Process

1. **Automated Checks**: CI must pass (tests, linting, coverage)
2. **Code Review**: At least one approval required
3. **Documentation**: Ensure docs are updated for user-facing changes
4. **No Breaking Changes**: Unless it's a major version bump

### Common Review Feedback

- **Missing tests**: Ensure new code is tested
- **Error handling**: Check all error paths
- **Context usage**: Verify context is properly passed
- **Performance**: Consider performance implications
- **Documentation**: Update relevant documentation

## 🏗️ Project Structure

Understanding the codebase organization:

```
go-coverage/
├── cmd/go-coverage/          # CLI application entry point
│   ├── cmd/                  # Cobra commands implementation
│   ├── main.go              # Application main
│   └── version.go           # Version information
├── internal/                # Internal packages (not importable)
│   ├── analytics/           # Dashboard and report generation
│   ├── badge/               # SVG badge generation
│   ├── config/              # Configuration management
│   ├── github/              # GitHub API integration
│   ├── history/             # Coverage history tracking
│   ├── parser/              # Coverage file parsing
│   ├── templates/           # Template rendering
│   ├── types/               # Core data types
│   └── urlutil/             # URL utilities
├── docs/                    # Documentation
├── .github/                 # GitHub workflows and templates
├── .mage.yml                 # Build and development tasks
└── go.mod                   # Go module definition
```

### Key Packages

- **`parser`**: Core coverage parsing logic
- **`badge`**: SVG badge generation with themes
- **`analytics`**: HTML report and dashboard generation
- **`github`**: GitHub API integration and PR management
- **`config`**: Configuration loading and validation
- **`history`**: Coverage history and trend analysis

## 📦 Release Process

**Note**: Only maintainers can create releases.

### Version Bumping

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Steps (Maintainers Only)

1. **Prepare Release**
   ```bash
   # Update version and run tests
   magex test:coverrace
   ```

2. **Create Tag**
   ```bash
   magex version:bump push=true bump=patch
   ```

3. **Publish Release**
   - GitHub Actions automatically builds and publishes
   - Release notes are generated from commits
   - Binaries are built for multiple platforms

### Release Artifacts

Each release includes:
- **Source code** (tar.gz and zip)
- **Binaries** for multiple platforms
- **Checksums** for verification
- **Release notes** with changes

## 💡 Development Tips

### Local Testing

```bash
# Test the full pipeline locally
go test -coverprofile=coverage.txt ./...
go run ./cmd/go-coverage complete -i coverage.txt --skip-github

# Test specific commands
go run ./cmd/go-coverage parse -f coverage.txt --format json
go run ./cmd/go-coverage --debug complete -i coverage.txt --dry-run
```

### Debugging

```bash
# Enable debug logging
export GO_COVERAGE_LOG_LEVEL=debug
go run ./cmd/go-coverage --debug [command]

# Use dry-run mode to preview changes
go run ./cmd/go-coverage complete -i coverage.txt --dry-run
```

### Performance Testing

```bash
# Run benchmarks
magex bench

# Profile specific operations
go test -bench=BenchmarkParseStatement -cpuprofile=cpu.prof ./internal/parser/
go tool pprof cpu.prof
```

## 🆘 Getting Help

- **Documentation**: Check the [docs/](../docs) directory first
- **Issues**: Search existing [GitHub Issues](https://github.com/mrz1836/go-coverage/issues)
- **Discussions**: Use [GitHub Discussions](https://github.com/mrz1836/go-coverage/discussions) for questions
- **Code Standards**: Review [`.github/AGENTS.md`](../.github/AGENTS.md) for detailed guidelines

## 🙏 Recognition

Contributors are recognized in:
- **GitHub Contributors** graph
- **Release notes** for significant contributions
- **Documentation** for major feature additions

Thank you for contributing to go-coverage! 🚀

---

For more information:
- [Code Standards](../.github/AGENTS.md) - Detailed coding guidelines
- [Security Policy](../.github/SECURITY.md) - Security reporting process
- [Support](../.github/SUPPORT.md) - Getting help and support
