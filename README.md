<div align="center">

# 📊&nbsp;&nbsp;go-coverage

**Your Coverage. Your Infrastructure. Pure Go.**

<br/>

<a href="https://github.com/mrz1836/go-coverage/releases"><img src="https://img.shields.io/github/release-pre/mrz1836/go-coverage?include_prereleases&style=flat-square&logo=github&color=black" alt="Release"></a>
<a href="https://golang.org/"><img src="https://img.shields.io/github/go-mod/go-version/mrz1836/go-coverage?style=flat-square&logo=go&color=00ADD8" alt="Go Version"></a>
<a href="https://github.com/mrz1836/go-coverage/blob/master/LICENSE"><img src="https://img.shields.io/github/license/mrz1836/go-coverage?style=flat-square&color=blue" alt="License"></a>

<br/>

<table align="center" border="0">
  <tr>
    <td align="right">
       <code>CI / CD</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://github.com/mrz1836/go-coverage/actions"><img src="https://img.shields.io/github/actions/workflow/status/mrz1836/go-coverage/fortress.yml?branch=master&label=build&logo=github&style=flat-square" alt="Build"></a>
       <a href="https://github.com/mrz1836/go-coverage/actions"><img src="https://img.shields.io/github/last-commit/mrz1836/go-coverage?style=flat-square&logo=git&logoColor=white&label=last%20update" alt="Last Commit"></a>
    </td>
    <td align="right">
       &nbsp;&nbsp;&nbsp;&nbsp; <code>Quality</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://goreportcard.com/report/github.com/mrz1836/go-coverage"><img src="https://goreportcard.com/badge/github.com/mrz1836/go-coverage?style=flat-square" alt="Go Report"></a>
       <a href="https://mrz1836.github.io/go-coverage/" target="_blank"><img src="https://mrz1836.github.io/go-coverage/coverage-flat-square.svg" alt="Code Coverage"></a>
    </td>
  </tr>

  <tr>
    <td align="right">
       <code>Security</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://scorecard.dev/viewer/?uri=github.com/mrz1836/go-coverage"><img src="https://api.scorecard.dev/projects/github.com/mrz1836/go-coverage/badge?style=flat-square" alt="Scorecard"></a>
       <a href=".github/SECURITY.md"><img src="https://img.shields.io/badge/policy-active-success?style=flat-square&logo=security&logoColor=white" alt="Security"></a>
    </td>
    <td align="right">
       &nbsp;&nbsp;&nbsp;&nbsp; <code>Community</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://github.com/mrz1836/go-coverage/graphs/contributors"><img src="https://img.shields.io/github/contributors/mrz1836/go-coverage?style=flat-square&color=orange" alt="Contributors"></a>
       <a href="https://mrz1818.com/"><img src="https://img.shields.io/badge/donate-bitcoin-ff9900?style=flat-square&logo=bitcoin" alt="Bitcoin"></a>
    </td>
  </tr>
</table>

</div>

<br/>
<br/>

<div align="center">

### <code>Project Navigation</code>

</div>

<table align="center">
  <tr>
    <td align="center" width="33%">
       ⚡&nbsp;<a href="#-quickstart"><code>Installation</code></a>
    </td>
    <td align="center" width="33%">
       🧪&nbsp;<a href="#-examples--tests"><code>Examples&nbsp;&&nbsp;Tests</code></a>
    </td>
    <td align="center" width="33%">
       📚&nbsp;<a href="#-documentation"><code>Documentation</code></a>
    </td>
  </tr>
  <tr>
    <td align="center">
       🤝&nbsp;<a href="#-contributing"><code>Contributing</code></a>
    </td>
    <td align="center">
      🛠️&nbsp;<a href="#-code-standards"><code>Code&nbsp;Standards</code></a>
    </td>
    <td align="center">
      ⚡&nbsp;<a href="#-performance"><code>Performance</code></a>
    </td>
  </tr>
  <tr>
    <td align="center">
      🤖&nbsp;<a href="#-ai-usage--assistant-guidelines"><code>AI&nbsp;Usage</code></a>
    </td>
    <td align="center">
       📝&nbsp;<a href="#-license"><code>License</code></a>
    </td>
    <td align="center">
       👥&nbsp;<a href="#-maintainers"><code>Maintainers</code></a>
    </td>
  </tr>
</table>
<br/>

## ⚡ Quickstart

**Go Coverage** is a complete replacement for Codecov that runs entirely in your CI/CD pipeline with zero external dependencies. Get coverage reports, badges, and dashboards deployed to GitHub Pages automatically.

<br>

### Installation

**Go Coverage** requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy).

**Install CLI Tool** (recommended):
```bash
go install github.com/mrz1836/go-coverage/cmd/go-coverage@latest
```

**Or install go-coverage as a Library**:
```bash
go get -u github.com/mrz1836/go-coverage
```

**Verify Installation**:
```bash
go-coverage --version
# go-coverage version v1.0...
```

**Upgrade to Latest Version**:
```bash
# Check for available updates
go-coverage upgrade --check

# Upgrade to the latest version
go-coverage upgrade

# Force reinstall even if already on latest
go-coverage upgrade --force
```

<br>

### Development Setup

Install [MAGE-X](https://github.com/mrz1836/mage-x) build tool for development:

```bash
# Install MAGE-X for development and building
go install github.com/mrz1836/mage-x/cmd/magex@latest
magex update:install
```

<br>

### Choose Your Coverage System

**Option 1: Internal Coverage System (GitHub Pages)** - Recommended

Set up GitHub Pages environment for coverage deployment:
```bash
go-coverage setup-pages
```

**Option 2: External Coverage System (Codecov)**

Create `.github/.env.custom` with:
```bash
GO_COVERAGE_PROVIDER=codecov
CODECOV_TOKEN_REQUIRED=true
```

**Next, deploy to your main branch and generate coverage reports!**

<br>

### Core Features

- 🏷️ **SVG Badge Generation** – Custom badges with themes and logos
- 📊 **HTML Reports & Dashboards** – Beautiful, responsive coverage visualizations
- 📈 **History & Trends** – Track coverage changes over time
- 🤖 **GitHub Integration** – PR comments, commit statuses, automated deployments
- 🚀 **GitHub Pages** – Automated deployment with zero configuration
- 🔧 **Highly Configurable** – Thresholds, exclusions, templates, and more
- ⬆️ **Auto-Upgrade** – Built-in upgrade command for easy updates

<br/>


## 🚀 GitHub Pages Setup

**Automatic Deployment**: Go Coverage automatically deploys coverage reports, badges, and dashboards to GitHub Pages with zero configuration.

### Quick Setup

Set up GitHub Pages environment using the integrated CLI command:

```bash
# Auto-detect repository from git remote
go-coverage setup-pages

# Or specify repository explicitly
go-coverage setup-pages owner/repo

# Preview changes without making them
go-coverage setup-pages --dry-run

# Use a custom domain for GitHub Pages
go-coverage setup-pages --custom-domain mysite.com
```


This configures:
- ✅ **GitHub Pages Environment** with proper branch policies
- ✅ **Deployment Permissions** for `master`, `gh-pages`, and any `*/*/*/*/*/*` branches
- ✅ **Environment Protection** rules for secure deployments

### What Gets Deployed

Your coverage system automatically creates:

```
https://yourname.github.io/yourrepo/
├── coverage.svg              # Live coverage badge
├── index.html                # Coverage dashboard
├── coverage.html             # Detailed coverage report
└── reports/branch/master/    # Branch-specific reports
```

<details>
<summary><strong>Manual GitHub Pages Configuration</strong></summary>

If the setup command fails, manually configure:

1. Go to **Settings** → **Environments** → **github-pages**
2. Under **Deployment branches**, select "Selected branches and tags"
3. Add these deployment branch rules:
   - `master` (main deployments)
   - `gh-pages` (GitHub Pages default)
   - `*`, `*/*`, `*/*/*`, `*/*/*/*`, `*/*/*/*/*`, `*/*/*/*/*/*` (all branches for PR-specific reports)
4. Save changes and verify in workflow runs

</details>

### Integration with CI/CD

The coverage system integrates with your existing GitHub Actions:

```yaml
# In your .github/workflows/ci.yml
- name: Generate Coverage Report
  run: |
    go test -coverprofile=coverage.txt ./...
    go-coverage complete -i coverage.txt
```

<br/>

## 🎯 Starting a New Project

### 1. Install the CLI Tool

```bash
go install github.com/mrz1836/go-coverage/cmd/go-coverage@latest
```

### 2. Configure GitHub Pages

```bash
# Use the integrated command (requires gh CLI)
go-coverage setup-pages
```

### 3. Add to GitHub Actions

Add coverage generation to your workflow:

```yaml
name: Coverage
on: [push, pull_request]

jobs:
  coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run Tests with Coverage
        run: go test -coverprofile=coverage.txt ./...

      - name: Generate Coverage Reports
        run: go-coverage complete -i coverage.txt -o coverage
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./coverage
```

### 4. First Run

Commit and push - your coverage reports will be available at:
- **Reports**: `https://yourname.github.io/yourrepo/`
- **Badge**: `https://yourname.github.io/yourrepo/coverage.svg`

<details>
<summary><strong>Advanced Configuration</strong></summary>

Create a `.go-coverage.json` config file:

```json
{
  "coverage": {
    "threshold": 80.0,
    "exclude_paths": ["vendor/", "test/"],
    "exclude_files": ["*.pb.go", "*_gen.go"]
  },
  "badge": {
    "style": "flat",
    "logo": "go"
  },
  "report": {
    "title": "My Project Coverage",
    "theme": "dark"
  },
  "history": {
    "enabled": true,
    "retention_days": 90
  }
}
```

</details>

<br/>

## 📚 Documentation

### Quick Start Guides
- **[⚡ Quickstart](docs/quickstart.md)** – Get started in 5 minutes with installation and basic setup
- **[📚 User Guide](docs/user-guide.md)** – Complete usage guide with examples and workflows

### Reference Documentation
- **[🛠️ CLI Reference](docs/cli-reference.md)** – Detailed command-line reference and options
- **[⚙️ Configuration](docs/configuration.md)** – Environment variables and configuration options
- **CLI Reference** – Complete command documentation at [pkg.go.dev/github.com/mrz1836/go-coverage](https://pkg.go.dev/github.com/mrz1836/go-coverage)

### Developer Resources
- **[🤝 Contributing](docs/contributing.md)** – How to contribute code, tests, and documentation
- **[🏗️ Architecture](docs/architecture.md)** – Technical architecture and design decisions

### Features Overview
- **Coverage Analysis** – Parse Go coverage profiles with exclusions and thresholds
- **Badge Generation** – Create SVG badges with custom styling and themes
- **Report Generation** – Build HTML dashboards and detailed coverage reports
- **History Tracking** – Monitor coverage trends over time with retention policies
- **GitHub Integration** – PR comments, commit statuses, and automated deployments

<br/>

<details>
<summary><strong><code>Go Coverage Features</code></strong></summary>
<br/>

* **Zero External Dependencies** – Complete coverage system that runs entirely in your CI/CD pipeline with no third-party services required.
* **GitHub Pages Integration** – Automatic deployment of coverage reports, badges, and dashboards with branch-specific and PR-specific deployments.
* **Advanced Coverage Analysis** – Parse Go coverage profiles with support for path exclusions, file pattern exclusions, and threshold enforcement.
* **Professional Badge Generation** – SVG coverage badges with customizable styles, colors, logos, and themes that update automatically.
* **Rich HTML Reports** – Beautiful, responsive coverage dashboards with detailed file-level analysis and interactive visualizations.
* **Coverage History & Trends** – Track coverage changes over time with retention policies, trend analysis, and historical comparisons.
* **Smart GitHub Integration** – Automated PR comments with coverage analysis, commit status checks, and diff-based coverage reporting.
* **Multi-Branch Support** – Separate coverage tracking for different branches with automatic main branch detection and PR context handling.
* **Comprehensive CLI Tool** – Six powerful commands (`complete`, `comment`, `parse`, `history`, `setup-pages`, `upgrade`) for all coverage operations.
* **Highly Configurable** – JSON-based configuration for thresholds, exclusions, badge styling, report themes, and integration settings.
* **Enterprise Ready** – Built with security, performance, and scalability in mind for production environments.
* **Self-Contained Deployment** – Everything runs in your repository's `.github` folder with no external service dependencies or accounts required.

</details>

<details>
<summary><strong><code>Library Deployment</code></strong></summary>
<br/>

This project uses [goreleaser](https://github.com/goreleaser/goreleaser) for streamlined binary and library deployment to GitHub. To get started, install it via:

```bash
brew install goreleaser
```

The release process is defined in the [.goreleaser.yml](.goreleaser.yml) configuration file.

Then create and push a new Git tag using:

```bash
magex version:bump push=true bump=patch
```

This process ensures consistent, repeatable releases with properly versioned artifacts and citation metadata.

</details>

<details>
<summary><strong><code>Build Commands</code></strong></summary>
<br/>

View all build commands

```bash script
magex help
```

</details>

<details>
<summary><strong><code>GitHub Workflows</code></strong></summary>
<br/>


### 🎛️ The Workflow Control Center

All GitHub Actions workflows in this repository are powered by a flexible two-file configuration system – your one-stop shop for tweaking CI/CD behavior without touching a single YAML file! 🎯

**Configuration Files:**
- **[.env.base](.github/.env.base)** – Default configuration that works for most Go projects
- **[.env.custom](.github/.env.custom)** – Optional project-specific overrides

This powerful system controls everything from:
- **🚀 Go version matrix** (test on multiple versions or just one)
- **🏃 Runner selection** (Ubuntu or macOS, your wallet decides)
- **🔬 Feature toggles** (coverage, fuzzing, linting, race detection, benchmarks)
- **🛡️ Security tool versions** (gitleaks, nancy, govulncheck)
- **🤖 Auto-merge behaviors** (how aggressive should the bots be?)
- **🏷️ PR management rules** (size labels, auto-assignment, welcome messages)

> **Pro tip:** Want to disable code coverage? Create `.env.custom` with `ENABLE_CODE_COVERAGE=false` and push. No YAML archaeology required!

<br/>

| Workflow Name                                                                | Description                                                                                                            |
|------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------|
| [auto-merge-on-approval.yml](.github/workflows/auto-merge-on-approval.yml)   | Automatically merges PRs after approval and all required checks, following strict rules.                               |
| [codeql-analysis.yml](.github/workflows/codeql-analysis.yml)                 | Analyzes code for security vulnerabilities using [GitHub CodeQL](https://codeql.github.com/).                          |
| [dependabot-auto-merge.yml](.github/workflows/dependabot-auto-merge.yml)     | Automatically merges [Dependabot](https://github.com/dependabot) PRs that meet all requirements.                       |
| [fortress.yml](.github/workflows/fortress.yml)                               | Runs the Go Coverage security and testing workflow, including linting, testing, releasing, and vulnerability checks.   |
| [pull-request-management.yml](.github/workflows/pull-request-management.yml) | Labels PRs by branch prefix, assigns a default user if none is assigned, and welcomes new contributors with a comment. |
| [scorecard.yml](.github/workflows/scorecard.yml)                             | Runs [OpenSSF](https://openssf.org/) Scorecard to assess supply chain security.                                        |
| [stale.yml](.github/workflows/stale-check.yml)                               | Warns about (and optionally closes) inactive issues and PRs on a schedule or manual trigger.                           |
| [sync-labels.yml](.github/workflows/sync-labels.yml)                         | Keeps GitHub labels in sync with the declarative manifest at [`.github/labels.yml`](./.github/labels.yml).             |

</details>

<details>
<summary><strong><code>Updating Dependencies</code></strong></summary>
<br/>

To update all dependencies (Go modules, linters, and related tools), run:

```bash
magex deps:update
```

This command ensures all dependencies are brought up to date in a single step, including Go modules and any tools managed by MAGE-X. It is the recommended way to keep your development environment and CI in sync with the latest versions.

</details>

<details>
<summary><strong>🔧 Pre-commit Hooks</strong></summary>
<br/>

Set up the Go-Pre-commit System to run the same formatting, linting, and tests defined in [AGENTS.md](.github/AGENTS.md) before every commit:

```bash
go install github.com/mrz1836/go-pre-commit/cmd/go-pre-commit@latest
go-pre-commit install
```

The system is configured via [.env.base](.github/.env.base) and can be customized using also using [.env.custom](.github/.env.custom) and provides 17x faster execution than traditional Python-based pre-commit hooks. See the [complete documentation](http://github.com/mrz1836/go-pre-commit) for details.

</details>

<br/>

## 🧪 Examples & Tests

The **Go Coverage** system is thoroughly tested via [GitHub Actions](https://github.com/mrz1836/go-coverage/actions) and uses [Go version 1.24.x](https://go.dev/doc/go1.24). View the [configuration file](.github/workflows/fortress.yml).

### CLI Command Examples

```bash
# Complete coverage pipeline (parse + badge + report + history + GitHub)
go-coverage complete -i coverage.txt -o coverage-reports

# Generate PR comment with coverage analysis
go-coverage comment --pr 123 --coverage coverage.txt --base-coverage base-coverage.txt

# Parse coverage data with exclusions
go-coverage parse -i coverage.txt --exclude-paths "vendor/,test/" --threshold 80

# View coverage history and trends
go-coverage history --branch master --days 30 --format json

# Set up GitHub Pages environment for coverage deployment
go-coverage setup-pages --verbose --dry-run
go-coverage setup-pages owner/repo --custom-domain example.com

# Upgrade to the latest version
go-coverage upgrade --check
go-coverage upgrade --force --verbose
```

### Testing the Coverage System

Run all tests (fast):

```bash script
magex test
```

Run all tests with race detector (slower):
```bash script
magex test:race
```

<details>
<summary><strong>🔬 Fuzz Testing</strong></summary>

The coverage system includes comprehensive fuzz tests for critical functions to ensure robustness and security:

#### Available Fuzz Tests

| Package                                            | Functions Tested                  | Coverage | Security Focus                 |
|----------------------------------------------------|-----------------------------------|----------|--------------------------------|
| **[urlutil](internal/urlutil/urls_fuzz_test.go)**  | URL building, path cleaning       | 100%     | Path traversal, XSS prevention |
| **[badge](internal/badge/generator_fuzz_test.go)** | Badge generation, color selection | 97.7%    | SVG injection, encoding        |
| **[parser](internal/parser/parser_fuzz_test.go)**  | Coverage parsing, file exclusion  | 84.1%    | Malformed input handling       |

#### Running Fuzz Tests

```bash script
# Run specific fuzz tests
go test -fuzz=FuzzBuildGitHubCommitURL -fuzztime=10s ./internal/urlutil/
go test -fuzz=FuzzGetColorForPercentage -fuzztime=10s ./internal/badge/
go test -fuzz=FuzzParseStatementSimple -fuzztime=10s ./internal/parser/

# Run all fuzz tests (via MAGE-X)
magex test-fuzz
```

#### Fuzz Test Features

- ✅ **Comprehensive Input Coverage**: Valid inputs, edge cases, malformed data
- ✅ **Security Testing**: Path traversal, XSS, injection attempts, null bytes
- ✅ **Panic Prevention**: Never panics on any input
- ✅ **Unicode Support**: Proper UTF-8 handling and validation
- ✅ **Performance Testing**: Long inputs, memory efficiency

</details>

### Example Output

The system generates comprehensive coverage reports:

```
📊 Coverage Dashboard: https://yourname.github.io/yourrepo/
🏷️ Coverage Badge: https://yourname.github.io/yourrepo/coverage.svg
📈 Coverage: 87.4% (1,247/1,426 lines)
📦 Packages: 15 analyzed
🔍 Trend: UP (+2.3% from last run)
```

<br/>

## ⚡ Performance

The **Go Coverage** system is optimized for speed and efficiency in CI/CD environments.

```bash script
magex bench
```

<details>
<summary><strong>⚡ Benchmark Results & Performance Metrics</strong></summary>

### Benchmark Results

| Component     | Operation           | Time/op | Memory/op | Allocs/op | Description                        |
|---------------|---------------------|---------|-----------|-----------|------------------------------------|
| **Parser**    | Parse (100 files)   | 105.9ns | 8B        | 0         | Parse coverage data with 100 files |
| **Parser**    | Parse (1000 files)  | 14.4ms  | 8.0MB     | 106,870   | Large coverage files               |
| **Badge**     | Generate SVG        | 1.76µs  | 2.5KB     | 14        | Badge generation                   |
| **Badge**     | Generate with Logo  | 1.82µs  | 2.7KB     | 15        | Badge with custom logo             |
| **Dashboard** | Generate HTML       | 12.3ms  | 1.4MB     | 10,645    | Full dashboard generation          |
| **Report**    | Generate Report     | 8.17ms  | 1.1MB     | 7,890     | Coverage report generation         |
| **History**   | Record Entry        | 240µs   | 9.2KB     | 68        | Store coverage entry               |
| **History**   | Get Trend (30 days) | 1.7ms   | 255KB     | 1,254     | Trend analysis                     |
| **Analysis**  | Compare Coverage    | 20.4µs  | 42KB      | 146       | Coverage comparison                |
| **Templates** | Render PR Comment   | 38.9µs  | 11KB      | 377       | Comment generation                 |
| **URL**       | Build GitHub URL    | 50.1ns  | 48B       | 1         | URL construction                   |

### Performance Characteristics

- **Concurrent Operations**: All critical paths support concurrent execution
- **Memory Efficiency**: Streaming parsers for large files
- **Caching**: Template compilation and static asset caching
- **Optimization**: Profile-guided optimizations for hot paths

### Running Benchmarks

```bash
# Run all benchmarks
magex bench

# Run specific component benchmarks
go test -bench=. ./internal/parser/...
go test -bench=. ./internal/badge/...
go test -bench=. ./internal/analytics/...
go test -bench=. ./internal/history/...
go test -bench=. ./internal/analysis/...

# Run with memory profiling
go test -bench=. -benchmem ./...

# Generate benchmark comparison
go test -bench=. -count=5 ./... | tee new.txt
benchstat old.txt new.txt
```

### Real-World Metrics

- ⚡ **CI/CD Integration**: Adds < 2 seconds to your workflow
- 📊 **Memory Efficient**: Peak usage under 10MB for large repositories
- 🚀 **GitHub Pages**: Deploy coverage reports in under 30 seconds
- 📈 **Scalable**: Tested with repositories containing 100,000+ lines of code

> Performance benchmarks measured on GitHub Actions runners (10-core CPU) with production Go projects.

</details>

<br/>

## 🛠️ Code Standards
Read more about this Go project's [code standards](.github/CODE_STANDARDS.md).

<br/>

## 🤖 AI Usage & Assistant Guidelines
Read the [AI Usage & Assistant Guidelines](.github/tech-conventions/ai-compliance.md) for details on how AI is used in this project and how to interact with the AI assistants.

<br/>

<details>
<summary><strong>🤖 Claude Code Sub-Agents & Commands</strong></summary>

### Sub-Agents Overview

This project leverages a comprehensive team of specialized Claude Code sub-agents to manage development, testing, and deployment workflows. Each agent has specific expertise and can work independently or collaboratively to maintain the go-coverage system.

### Available Sub-Agents

| Agent                                                                | Specialization                                        | Primary Tools          | Proactive Triggers              |
|----------------------------------------------------------------------|-------------------------------------------------------|------------------------|---------------------------------|
| **[go-test-runner](.claude/agents/go-test-runner.md)**               | Test execution, coverage analysis, failure resolution | Bash, Read, Edit, Task | After code changes, before PRs  |
| **[go-linter](.claude/agents/go-linter.md)**                         | Code formatting, linting, standards enforcement       | Bash, Edit, Glob       | After any Go file modification  |
| **[coverage-analyzer](.claude/agents/coverage-analyzer.md)**         | Coverage reports, badges, GitHub Pages deployment     | Bash, Write, WebFetch  | After successful test runs      |
| **[github-integration](.claude/agents/github-integration.md)**       | PR management, status checks, API operations          | Bash, WebFetch         | PR events, deployments          |
| **[dependency-manager](.claude/agents/dependency-manager.md)**       | Module updates, vulnerability scanning                | Bash, Edit, WebFetch   | go.mod changes, weekly scans    |
| **[ci-workflow](.claude/agents/ci-workflow.md)**                     | GitHub Actions, pipeline optimization                 | Read, Edit, Bash       | Workflow failures, CI updates   |
| **[code-reviewer](.claude/agents/code-reviewer.md)**                 | Code quality, security review, best practices         | Read, Grep, Glob       | After code writing/modification |
| **[documentation-manager](.claude/agents/documentation-manager.md)** | README, API docs, changelog maintenance               | Read, Edit, WebFetch   | API changes, new features       |
| **[performance-optimizer](.claude/agents/performance-optimizer.md)** | Benchmarking, profiling, optimization                 | Bash, Edit, Grep       | Performance issues, benchmarks  |
| **[security-scanner](.claude/agents/security-scanner.md)**           | Vulnerability detection, compliance checks            | Bash, Grep, WebFetch   | Security advisories, scans      |
| **[debugger](.claude/agents/debugger.md)**                           | Error analysis, test debugging, issue resolution      | Read, Edit, Bash       | Test failures, errors, panics   |

### Using Sub-Agents

Sub-agents can be invoked in two ways:

1. **Automatic Delegation**: Claude Code automatically delegates tasks based on context and the agent's specialization
2. **Explicit Invocation**: Request a specific agent by name:
   ```
   > Use the code-reviewer agent to review my recent changes
   > Have the debugger investigate this test failure
   > Ask the coverage-analyzer to generate a new report
   ```

### Agent Coordination

Sub-agents work together cohesively:
- **go-test-runner** → triggers **coverage-analyzer** after successful tests
- **code-reviewer** → invokes **go-linter** for style issues
- **dependency-manager** → calls **security-scanner** for vulnerability checks

### Configuration

Sub-agent configurations are stored in `.claude/agents/` and can be customized:
- Edit agent prompts to adjust behavior
- Modify tool access for security constraints
- Add project-specific instructions

### Benefits

- **Specialized Expertise**: Each agent excels in its domain
- **Parallel Processing**: Multiple agents can work simultaneously
- **Isolated Contexts**: Agents maintain separate contexts to prevent pollution
- **Consistent Workflows**: Standardized approaches across the team
- **Improved Efficiency**: Faster task completion with focused agents

For detailed information about each sub-agent's capabilities and configuration, see the individual agent files in `.claude/agents/`.

### Claude Code Commands

The project includes **20 powerful slash commands** that orchestrate our sub-agents for common development tasks. These commands provide quick access to complex workflows:

#### Quick Examples

```bash
/fix              # Automatically fix test failures and linter issues
/test parser.go   # Create comprehensive tests for a file
/coverage         # Analyze and improve test coverage to 90%+
/pr-ready        # Make your code PR-ready with all checks
/review          # Get comprehensive code review
/secure          # Run security vulnerability scan
/health          # Complete project health check
```

#### Command Categories

- **Quality & Testing**: `/fix`, `/test`, `/coverage`, `/dedupe`
- **Documentation**: `/doc-update`, `/doc-review`, `/explain`, `/prd`
- **Development**: `/review`, `/optimize`, `/refactor`
- **Maintenance**: `/deps`, `/secure`, `/health`, `/clean`
- **Workflow**: `/pr-ready`, `/debug-ci`, `/release-prep`, `/benchmark`, `/commit`

See the complete [**Claude Code Commands Reference**](docs/claude-commands.md) for detailed usage, examples, and best practices.

</details>

<br/>

## 👥 Maintainers
| [<img src="https://github.com/mrz1836.png" height="50" width="50" alt="MrZ" />](https://github.com/mrz1836) |
|:-----------------------------------------------------------------------------------------------------------:|
|                                      [MrZ](https://github.com/mrz1836)                                      |

<br/>

## 🤝 Contributing
View the [contributing guidelines](.github/CONTRIBUTING.md) and please follow the [code of conduct](.github/CODE_OF_CONDUCT.md).

### How can I help?
All kinds of contributions are welcome :raised_hands:!
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:.
You can also support this project by [becoming a sponsor on GitHub](https://github.com/sponsors/mrz1836) :clap:
or by making a [**bitcoin donation**](https://mrz1818.com/?tab=tips&utm_source=github&utm_medium=sponsor-link&utm_campaign=go-coverage&utm_term=go-coverage&utm_content=go-coverage) to ensure this journey continues indefinitely! :rocket:

[![Stars](https://img.shields.io/github/stars/mrz1836/go-coverage?label=Please%20like%20us&style=social&v=1)](https://github.com/mrz1836/go-coverage/stargazers)

<br/>

## 📝 License

[![License](https://img.shields.io/github/license/mrz1836/go-coverage.svg?style=flat&v=1)](LICENSE)
