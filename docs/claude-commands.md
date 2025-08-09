# Claude Code Commands Reference

This document provides a comprehensive guide to the custom Claude Code commands available for the go-coverage project. These commands leverage our specialized sub-agents to efficiently manage development, testing, and deployment workflows.

## Table of Contents

- [Quick Reference](#quick-reference)
- [Quality & Testing Commands](#quality--testing-commands)
- [Documentation Commands](#documentation-commands)
- [Development Commands](#development-commands)
- [Maintenance Commands](#maintenance-commands)
- [Workflow Commands](#workflow-commands)
- [Best Practices](#best-practices)
- [Agent Coordination](#agent-coordination)

## Quick Reference

| Command | Purpose | Model | Agents Used |
|---------|---------|-------|-------------|
| `/fix` | Fix test failures and linter issues | opus | go-test-runner, go-linter, debugger |
| `/test [path]` | Create comprehensive tests | opus | go-test-runner |
| `/coverage` | Analyze and improve coverage | opus | coverage-analyzer, go-test-runner |
| `/dedupe` | Remove duplicate code | opus | code-reviewer, performance-optimizer |
| `/doc-update [feature]` | Update documentation | opus | documentation-manager |
| `/doc-review` | Review documentation accuracy | sonnet | documentation-manager, code-reviewer |
| `/explain [feature]` | Explain how something works | opus | documentation-manager, code-reviewer |
| `/prd [feature]` | Design PRD for feature | opus | documentation-manager |
| `/review` | Comprehensive code review | opus | code-reviewer, security-scanner, performance-optimizer, go-linter |
| `/optimize [area]` | Performance optimization | opus | performance-optimizer |
| `/refactor [target]` | Suggest refactoring | opus | code-reviewer, go-linter |
| `/deps` | Update and audit dependencies | sonnet | dependency-manager, security-scanner |
| `/secure` | Security vulnerability scan | opus | security-scanner |
| `/health` | Project health check | opus | Multiple agents |
| `/clean` | Clean up code formatting | haiku | go-linter |
| `/pr-ready` | Make code PR-ready | opus | Multiple agents in sequence |
| `/debug-ci` | Diagnose CI issues | opus | ci-workflow, debugger |
| `/release-prep` | Prepare for release | opus | release-manager, go-test-runner, coverage-analyzer |
| `/benchmark` | Run performance benchmarks | sonnet | performance-optimizer |
| `/commit [message]` | Smart git commit | haiku | Direct git operations |

## Quality & Testing Commands

### `/fix` - Fix Test Failures and Linter Issues

**Purpose**: Automatically fix test failures and linting issues in your codebase.

**Usage**:
```
/fix
```

**What it does**:
- Runs tests to identify failures
- Analyzes linter output for issues
- Uses parallel agents to fix problems
- Verifies fixes work correctly

**Example workflow**:
```bash
# After making changes that break tests
/fix
# Agents will automatically fix test failures and linting issues
```

### `/test` - Create Comprehensive Tests

**Purpose**: Generate comprehensive test suites for your code.

**Usage**:
```
/test internal/parser
/test parser.go
/test  # Tests recently modified files
```

**What it does**:
- Analyzes code to identify untested functions
- Creates table-driven tests following AGENTS.md standards
- Includes edge cases and error conditions
- Ensures >= 90% coverage target

**Example**:
```bash
# Create tests for a new feature
/test internal/badge/generator.go
```

### `/coverage` - Analyze and Improve Coverage

**Purpose**: Analyze test coverage and create tests to reach 90%+ target.

**Usage**:
```
/coverage
```

**What it does**:
- Generates coverage profile
- Identifies packages below threshold
- Creates tests for uncovered code
- Prioritizes critical business logic

**Coverage targets**:
- Project: >= 90%
- Critical packages: 100%
- New code: >= 95%

## Documentation Commands

### `/doc-update` - Update Documentation

**Purpose**: Update documentation for new or modified features.

**Usage**:
```
/doc-update badge generation
/doc-update  # Updates based on recent changes
```

**What it does**:
- Updates function and package comments
- Modifies README sections
- Updates API documentation
- Ensures examples are current

### `/doc-review` - Review Documentation Accuracy

**Purpose**: Verify documentation matches current implementation.

**Usage**:
```
/doc-review
```

**What it does**:
- Checks if examples still work
- Verifies CLI commands are current
- Validates configuration documentation
- Reports outdated content

### `/explain` - Explain How Something Works

**Purpose**: Get detailed explanation of a feature or module.

**Usage**:
```
/explain coverage parsing
/explain internal/badge
/explain  # Explains overall system
```

**What it does**:
- Analyzes code implementation
- Creates clear explanations with examples
- Documents architecture and data flow
- Provides usage scenarios

## Development Commands

### `/prd` - Design Product Requirements Document

**Purpose**: Create comprehensive PRD for new features.

**Usage**:
```
/prd real-time coverage tracking dashboard
/prd API for coverage data export
```

**What it does**:
- Defines problem and solution
- Specifies functional requirements
- Documents technical design
- Creates implementation plan

### `/review` - Comprehensive Code Review

**Purpose**: Perform thorough code review with multiple quality checks.

**Usage**:
```
/review
```

**What it does**:
- Reviews for Go best practices
- Performs security analysis
- Checks performance implications
- Validates test coverage
- Ensures AGENTS.md compliance

**Review output**:
- 🔴 Critical issues (must fix)
- 🟡 Important issues (should fix)
- 🟢 Suggestions (consider)

### `/optimize` - Performance Optimization

**Purpose**: Analyze and optimize performance bottlenecks.

**Usage**:
```
/optimize
/optimize parser
/optimize badge generation
```

**Performance targets**:
- Parse coverage: ~50ms for 10K lines
- Generate badge: ~5ms
- Build HTML report: ~200ms
- Complete pipeline: ~1-2s
- Memory usage: <10MB

### `/refactor` - Suggest Refactoring

**Purpose**: Identify and implement refactoring improvements.

**Usage**:
```
/refactor internal/parser
/refactor complex_function.go
```

**What it does**:
- Identifies code smells
- Reduces complexity
- Improves testability
- Applies Go best practices

## Maintenance Commands

### `/deps` - Update Dependencies

**Purpose**: Update dependencies and check for vulnerabilities.

**Usage**:
```
/deps
```

**What it does**:
- Checks for updates
- Runs security scans
- Updates safely
- Validates compatibility

### `/secure` - Security Scan

**Purpose**: Comprehensive security vulnerability scanning.

**Usage**:
```
/secure
```

**Security checks**:
- Go vulnerability database
- Secret detection
- Common vulnerabilities (SQL injection, path traversal)
- Dependency CVEs
- License compliance

### `/health` - Project Health Check

**Purpose**: Comprehensive project health assessment.

**Usage**:
```
/health
```

**Health metrics**:
- ✅ Tests passing
- ✅ Coverage >= 90%
- ✅ No vulnerabilities
- ✅ Dependencies current
- ✅ CI/CD working
- ✅ Performance on target

### `/clean` - Clean Up Code

**Purpose**: Quick code formatting and style cleanup.

**Usage**:
```
/clean
```

**What it does**:
- Runs `go fmt`, `goimports`, `gofumpt`
- Fixes whitespace issues
- Organizes imports
- Formats YAML files

## Workflow Commands

### `/pr-ready` - Make PR Ready

**Purpose**: Prepare code for pull request submission.

**Usage**:
```
/pr-ready
```

**PR checklist**:
- ✅ Tests pass
- ✅ Coverage >= 90%
- ✅ Linting passes
- ✅ No security issues
- ✅ Documentation updated
- ✅ Commits follow convention

### `/debug-ci` - Debug CI Issues

**Purpose**: Diagnose and fix CI/GitHub Actions failures.

**Usage**:
```
/debug-ci
```

**Common fixes**:
- Environment configuration
- Flaky test handling
- Permission issues
- Cache problems
- Timeout adjustments

### `/release-prep` - Prepare Release

**Purpose**: Prepare for a new version release.

**Usage**:
```
/release-prep
```

**Release checklist**:
- Version determination (MAJOR.MINOR.PATCH)
- Changelog generation
- Security validation
- Documentation updates
- Tag preparation

### `/benchmark` - Run Benchmarks

**Purpose**: Execute and analyze performance benchmarks.

**Usage**:
```
/benchmark
```

**Metrics tracked**:
- Operations per second
- Nanoseconds per operation
- Memory allocations
- Bytes allocated

### `/commit` - Smart Git Commit

**Purpose**: Create properly formatted git commits.

**Usage**:
```
/commit
/commit "feat: add new badge styles"
```

**Commit format**:
```
<type>(<scope>): <description>

<body>
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `build`, `ci`

## Best Practices

### Efficient Command Usage

1. **Use commands in sequence for complex workflows**:
   ```bash
   /clean      # Clean up formatting
   /test       # Create missing tests
   /coverage   # Verify coverage
   /pr-ready   # Final validation
   ```

2. **Leverage parallel agent execution**:
   - Commands like `/review` and `/health` use multiple agents in parallel
   - This significantly reduces execution time

3. **Provide specific targets when possible**:
   ```bash
   /test internal/parser     # More efficient than testing everything
   /optimize badge           # Focused optimization
   ```

### Common Workflows

#### New Feature Development
```bash
/prd feature description    # Design the feature
# Implement the feature
/test                       # Create tests
/doc-update                 # Update documentation
/review                     # Review code
/pr-ready                   # Prepare for PR
```

#### Bug Fixing
```bash
/debug-ci                   # If CI is failing
/fix                        # Fix test/lint issues
/test                       # Ensure tests pass
/commit "fix: description"  # Commit with proper message
```

#### Performance Improvement
```bash
/benchmark                  # Baseline metrics
/optimize target_area       # Optimize specific area
/benchmark                  # Verify improvement
```

#### Release Preparation
```bash
/deps                       # Update dependencies
/secure                     # Security scan
/health                     # Overall health check
/release-prep              # Prepare release
```

## Agent Coordination

### How Commands Use Agents

Commands orchestrate multiple specialized agents for maximum efficiency:

1. **Parallel Execution**: Commands like `/review` run multiple agents simultaneously:
   - code-reviewer: Quality checks
   - security-scanner: Vulnerability scanning
   - performance-optimizer: Performance analysis
   - go-linter: Style checks

2. **Sequential Workflows**: Commands like `/pr-ready` run agents in order:
   - go-linter → go-test-runner → security-scanner → code-reviewer → documentation-manager

3. **Smart Agent Selection**: Each command uses only the agents it needs:
   - `/clean` uses only go-linter (fast)
   - `/health` uses all agents (comprehensive)

### Performance Considerations

- **Haiku model** for simple tasks (`/clean`, `/commit`): Fast, efficient
- **Sonnet model** for moderate complexity (`/deps`, `/benchmark`): Balanced
- **Opus model** for complex analysis (`/review`, `/optimize`): Thorough

### Context Gathering

Commands use bash execution (`!` prefix) to gather context:
- Git status and diffs
- Test results
- Coverage metrics
- CI status

This context helps agents make informed decisions without manual input.

## Tips and Tricks

1. **Quick fixes**: Use `/fix` after any failed tests or lint issues
2. **Regular health checks**: Run `/health` weekly to catch issues early
3. **Before PRs**: Always run `/pr-ready` for comprehensive validation
4. **Documentation**: Run `/doc-review` after major changes
5. **Security**: Run `/secure` before releases and after dependency updates

## Troubleshooting

### Command Not Working?

1. Check if command file exists:
   ```bash
   ls .claude/commands/
   ```

2. Verify agents are configured:
   ```
   /agents
   ```

3. Check command syntax:
   ```
   /help
   ```

### Performance Issues?

- Use specific targets instead of scanning entire codebase
- Run heavy commands (`/health`, `/review`) during off-peak times
- Use `/benchmark` to identify bottlenecks

### Need Help?

- Run `/help` to see all available commands
- Check individual command files in `.claude/commands/`
- Review agent configurations in `.claude/agents/`