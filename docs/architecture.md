# 🏗️ Architecture Overview

Technical architecture and design decisions for the **go-coverage** system.

## 📖 Table of Contents

- [System Overview](#-system-overview)
- [Core Components](#-core-components)
- [Data Flow](#-data-flow)
- [Package Architecture](#-package-architecture)
- [CLI Design](#-cli-design)
- [GitHub Integration](#-github-integration)
- [Performance Characteristics](#-performance-characteristics)
- [Security Model](#-security-model)
- [Design Decisions](#-design-decisions)

## 🎯 System Overview

Go-coverage is designed as a **self-contained, zero-dependency** coverage system that replaces external services like Codecov. The architecture prioritizes:

- **Simplicity**: Single binary with minimal configuration
- **Performance**: Sub-second processing for typical Go projects
- **Security**: Zero external data transmission
- **Reliability**: No external service dependencies
- **Portability**: Pure Go implementation, cross-platform

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Go Tests      │    │   go-coverage   │    │ GitHub Pages    │
│                 │───▶│      CLI        │───▶│   Deployment    │
│ coverage.txt    │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │ GitHub API      │
                       │ (PR Comments,   │
                       │  Status Checks) │
                       └─────────────────┘
```

## 🧩 Core Components

### 1. Coverage Parser (`internal/parser`)

**Purpose**: Parse Go coverage profiles and extract coverage metrics.

**Key Features**:
- Supports all Go coverage modes (set, count, atomic)
- Path and file pattern exclusions
- Package-level and file-level analysis
- Statement-level coverage tracking

**Design**:
- Context-aware parsing with cancellation support
- Stream-based processing for large files
- Comprehensive error handling with detailed context

```go
// Core parsing function
func ParseStatements(ctx context.Context, input string, opts ParseOptions) (*CoverageData, error)
```

### 2. Badge Generator (`internal/badge`)

**Purpose**: Generate SVG coverage badges with customizable styling.

**Key Features**:
- Multiple badge styles (flat, flat-square, for-the-badge, plastic)
- Custom colors and logos
- Responsive design for different display contexts
- Template-based generation for consistency

**Design**:
- Template-driven SVG generation
- Color calculation based on coverage percentages
- Logo embedding support (built-in and custom URLs)

### 3. Analytics Engine (`internal/analytics`)

**Purpose**: Generate interactive HTML reports and dashboards.

**Components**:
- **Dashboard** (`dashboard/`): Main coverage overview with charts
- **Report** (`report/`): Detailed file-level coverage reports

**Key Features**:
- Responsive HTML/CSS/JavaScript
- Interactive package and file navigation
- Search and filtering capabilities
- Multiple themes (light, dark, GitHub-style)

### 4. GitHub Integration (`internal/github`)

**Purpose**: Integrate with GitHub API for PR comments, status checks, and deployments.

**Key Features**:
- PR comment management with anti-spam features
- GitHub status check creation
- Rate limiting and retry logic
- Context-aware API calls

**Design**:
- Client abstraction for testability
- Exponential backoff for retries
- Proper error context propagation

### 5. History Tracker (`internal/history`)

**Purpose**: Track coverage changes over time and analyze trends.

**Key Features**:
- Time-series coverage data storage
- Trend analysis and predictions
- Data retention policies
- Branch-specific history tracking

**Design**:
- JSON-based storage for simplicity
- Automatic cleanup based on retention policies
- Efficient querying for trend analysis

### 6. Configuration System (`internal/config`)

**Purpose**: Centralized configuration management with environment variable support.

**Key Features**:
- Environment variable-based configuration
- JSON file support for complex setups
- Validation and default value handling
- GitHub context auto-detection

## 🔄 Data Flow

### Complete Pipeline Flow

```
1. Input: coverage.txt (Go coverage profile)
   │
   ▼
2. Parser: Extract coverage metrics
   │ ├─ Apply exclusions (paths, files, patterns)
   │ ├─ Calculate package-level coverage
   │ └─ Generate coverage summary
   ▼
3. Badge Generator: Create SVG badge
   │ ├─ Determine color based on threshold
   │ ├─ Apply styling and logos
   │ └─ Generate SVG output
   ▼
4. Report Generator: Create HTML reports
   │ ├─ Dashboard with overview charts
   │ ├─ Detailed file-level reports
   │ └─ Interactive navigation
   ▼
5. History Tracker: Update coverage history
   │ ├─ Store current coverage data
   │ ├─ Calculate trends
   │ └─ Clean up old data
   ▼
6. GitHub Integration: Update PR/status
   │ ├─ Create/update PR comments
   │ ├─ Set commit status checks
   │ └─ Handle rate limiting
   ▼
7. Output: Coverage reports ready for deployment
```

### CLI Command Flow

```go
// Complete command execution flow
func (c *CompleteCmd) Execute(ctx context.Context) error {
    // 1. Load configuration
    cfg := config.Load()

    // 2. Parse coverage data
    coverage, err := parser.ParseFile(ctx, cfg.Coverage.InputFile)

    // 3. Generate badge
    badge, err := badge.Generate(ctx, coverage, cfg.Badge)

    // 4. Generate reports
    reports, err := analytics.Generate(ctx, coverage, cfg.Report)

    // 5. Update history
    err = history.Update(ctx, coverage, cfg.History)

    // 6. GitHub integration (if enabled)
    if cfg.GitHub.Enabled {
        err = github.UpdatePR(ctx, coverage, cfg.GitHub)
    }

    // 7. Save outputs
    return saveOutputs(ctx, badge, reports, cfg.Coverage.OutputDir)
}
```

## 📦 Package Architecture

### Dependency Hierarchy

```
cmd/go-coverage (CLI entry point)
├── internal/config (configuration management)
├── internal/parser (coverage parsing)
├── internal/badge (SVG generation)
├── internal/analytics
│   ├── dashboard (interactive dashboard)
│   └── report (detailed reports)
├── internal/github (GitHub API integration)
├── internal/history (coverage history)
├── internal/templates (template rendering)
└── internal/urlutil (URL utilities)
```

### Package Responsibilities

| Package | Responsibility | External Dependencies |
|---------|---------------|----------------------|
| `config` | Configuration management | None |
| `parser` | Coverage file parsing | None |
| `badge` | SVG badge generation | None |
| `analytics` | Report generation | None |
| `github` | GitHub API integration | HTTP client only |
| `history` | Coverage history tracking | JSON encoding only |
| `templates` | Template rendering | `text/template` |

### Design Principles

1. **No Global State**: All packages use dependency injection
2. **Context Propagation**: All operations accept `context.Context`
3. **Error Wrapping**: Comprehensive error context with `fmt.Errorf`
4. **Interface-Based**: Accept interfaces, return concrete types
5. **Testing First**: All packages designed for testability

## 🖥️ CLI Design

### Command Structure

The CLI uses Cobra for command organization:

```go
type Commands struct {
    Root       *cobra.Command  // Root command with global flags
    Complete   *cobra.Command  // Full pipeline execution
    Parse      *cobra.Command  // Coverage parsing only
    Comment    *cobra.Command  // PR comment generation
    History    *cobra.Command  // Coverage history management
    SetupPages *cobra.Command  // GitHub Pages setup
    Upgrade    *cobra.Command  // Tool upgrade management
}
```

### Command Design Patterns

1. **Consistent Flag Naming**: Similar flags across commands
2. **Context Handling**: All commands support cancellation
3. **Error Handling**: Comprehensive error messages with suggestions
4. **Dry Run Support**: Preview operations without changes
5. **Verbose Logging**: Debug information when requested

### Configuration Loading

```go
// Configuration precedence (highest to lowest):
// 1. Command line flags
// 2. Environment variables
// 3. Configuration file (.go-coverage.json)
// 4. Default values

func LoadConfig() *Config {
    cfg := loadDefaults()
    cfg.applyConfigFile()
    cfg.applyEnvironmentVars()
    cfg.applyCommandFlags()
    return cfg
}
```

## 🔗 GitHub Integration

### API Client Design

```go
type Client interface {
    CreateComment(ctx context.Context, pr int, body string) error
    UpdateComment(ctx context.Context, commentID int, body string) error
    CreateStatus(ctx context.Context, sha string, status Status) error
    GetPullRequest(ctx context.Context, pr int) (*PullRequest, error)
}

type GitHubClient struct {
    client     *http.Client
    token      string
    owner      string
    repository string

    // Rate limiting
    rateLimiter *rate.Limiter

    // Retry logic
    retryBackoff time.Duration
    maxRetries   int
}
```

### Rate Limiting Strategy

- **Token bucket** algorithm for API rate limiting
- **Exponential backoff** for retry logic
- **Circuit breaker** pattern for handling API failures
- **Request batching** where possible

### Security Considerations

1. **Token Security**: Never log or expose GitHub tokens
2. **Permission Validation**: Verify required permissions before operations
3. **Input Sanitization**: Sanitize all user inputs in PR comments
4. **Error Information**: Avoid leaking sensitive data in error messages

## ⚡ Performance Characteristics

### Benchmark Results

Based on testing with real Go projects:

| Operation | Small Project (<100 files) | Large Project (>1000 files) |
|-----------|---------------------------|----------------------------|
| Parse Coverage | <50ms | <500ms |
| Generate Badge | <10ms | <10ms |
| Generate Reports | <200ms | <2s |
| Update History | <20ms | <100ms |
| Complete Pipeline | <500ms | <5s |

### Memory Usage

- **Peak Memory**: <50MB for large projects
- **Memory Efficiency**: Streaming parsers for large files
- **Garbage Collection**: Minimal allocations in hot paths

### Optimization Strategies

1. **Concurrent Processing**: Parallel package analysis
2. **Template Caching**: Pre-compiled templates
3. **Asset Embedding**: Embedded CSS/JS to reduce I/O
4. **Efficient Data Structures**: Optimized for access patterns

## 🔒 Security Model

### Threat Model

1. **Code Injection**: Malicious coverage data
2. **Path Traversal**: Malicious file paths
3. **Token Exposure**: GitHub token leakage
4. **Resource Exhaustion**: Large input files

### Security Controls

1. **Input Validation**: All inputs validated and sanitized
2. **Path Sanitization**: Safe path handling for file operations
3. **Token Handling**: Secure token management practices
4. **Resource Limits**: Timeouts and size limits
5. **Error Handling**: No sensitive data in error messages

### Secure Defaults

- **Minimal Permissions**: Request only necessary GitHub permissions
- **Safe File Paths**: Prevent directory traversal attacks
- **Input Sanitization**: Clean all user-provided data
- **Timeout Protection**: All operations have reasonable timeouts

## 🧠 Design Decisions

### Why Go?

1. **Performance**: Compiled binary with fast execution
2. **Simplicity**: Single binary deployment
3. **Ecosystem**: Natural fit for Go project coverage
4. **Concurrency**: Built-in support for parallel processing
5. **Standard Library**: Rich standard library reduces dependencies

### Why Self-Contained?

1. **Privacy**: No external data transmission
2. **Reliability**: No external service dependencies
3. **Cost**: Zero ongoing operational costs
4. **Control**: Complete control over data and processing
5. **Security**: Reduced attack surface

### Why GitHub Pages?

1. **Availability**: High availability and CDN
2. **Integration**: Native GitHub integration
3. **Cost**: Free for public repositories
4. **Simplicity**: No additional infrastructure needed
5. **Performance**: Fast global delivery

### Technology Choices

| Component | Technology | Rationale |
|-----------|------------|-----------|
| CLI Framework | Cobra | Industry standard, powerful features |
| Template Engine | text/template | Standard library, secure |
| HTTP Client | net/http | Standard library, sufficient features |
| JSON Handling | encoding/json | Standard library, performance |
| Testing | testify | De facto standard, rich assertions |
| Linting | golangci-lint | Comprehensive linting suite |

### Alternative Approaches Considered

1. **External Database**: Rejected for complexity and dependencies
2. **External Storage**: Rejected for privacy and cost concerns
3. **WebAssembly**: Considered for client-side processing but rejected for complexity
4. **Docker Container**: Rejected for deployment complexity
5. **Separate Microservices**: Rejected for operational overhead

## 🔮 Future Considerations

### Scalability

- **Multi-language Support**: Extend beyond Go coverage
- **Parallel Processing**: Enhanced concurrency for large projects
- **Incremental Processing**: Process only changed files

### Features

- **Advanced Analytics**: Machine learning for coverage predictions
- **Integration APIs**: Webhooks and external integrations
- **Custom Dashboards**: User-customizable report layouts

### Performance

- **Caching Layer**: Intelligent caching for faster processing
- **Streaming Processing**: Memory-efficient processing of large files
- **Compression**: Compressed storage for reports and history

---

This architecture provides a solid foundation for a reliable, performant, and secure coverage system while maintaining simplicity and ease of use.
