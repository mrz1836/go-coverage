# GoFortress Coverage Configuration Reference

Complete configuration guide for the GoFortress Internal Coverage System with all 45+ environment variables, examples, and best practices.

## Configuration Overview

The GoFortress Coverage System uses environment variables defined in [`.github/.env.shared`](../../.github/.env.shared) for all configuration. This approach provides:

- **Version Control**: All settings tracked in git
- **Environment Isolation**: Different settings per deployment
- **Zero Secrets**: No sensitive data in configuration
- **CI/CD Integration**: Seamless workflow integration

## Quick Start Configuration

### Minimal Setup
Add these variables to `.github/.env.shared` to get started:

```bash
# Enable the internal coverage system
ENABLE_GO_COVERAGE=true

# Basic thresholds
GO_GO_COVERAGE_FAIL_UNDER=80
GO_GO_COVERAGE_THRESHOLD_EXCELLENT=90

# Enable core features
GO_GO_COVERAGE_PR_COMMENT_ENABLED=true
GO_GO_COVERAGE_PAGES_AUTO_CREATE=true
```

### Recommended Setup
For production use, add these additional settings:

```bash
# Badge configuration
GO_GO_COVERAGE_BADGE_STYLE=flat
GO_GO_COVERAGE_BADGE_LOGO=go
GO_GO_COVERAGE_BADGE_BRANCHES=master,development

# Analytics and trends
GO_GO_COVERAGE_ENABLE_TREND_ANALYSIS=true
GO_GO_COVERAGE_ENABLE_PACKAGE_BREAKDOWN=true
GO_GO_COVERAGE_HISTORY_RETENTION_DAYS=90

# Cleanup and maintenance
GO_GO_COVERAGE_CLEANUP_PR_AFTER_DAYS=7
```

## Core Configuration

### System Enable/Disable

#### `ENABLE_GO_COVERAGE`
**Type**: Boolean
**Default**: `false`
**Description**: Master switch to enable the internal coverage system

```bash
ENABLE_GO_COVERAGE=true   # Enable coverage system
ENABLE_GO_COVERAGE=false  # Use external service (Codecov, etc.)
```

#### `ENABLE_GO_GO_COVERAGE_TESTS`
**Type**: Boolean
**Default**: `true`
**Description**: Run coverage tool tests in CI

```bash
ENABLE_GO_GO_COVERAGE_TESTS=true   # Test coverage tool
ENABLE_GO_GO_COVERAGE_TESTS=false  # Skip coverage tool tests
```

## Coverage Thresholds

### Quality Gates

#### `GO_COVERAGE_FAIL_UNDER`
**Type**: Float (0-100)
**Default**: `70`
**Description**: Minimum acceptable coverage percentage

```bash
GO_COVERAGE_FAIL_UNDER=80    # Fail builds below 80%
GO_COVERAGE_FAIL_UNDER=70    # More lenient threshold
GO_COVERAGE_FAIL_UNDER=90    # Strict quality requirements
```

#### `GO_COVERAGE_ENFORCE_THRESHOLD`
**Type**: Boolean
**Default**: `false`
**Description**: Whether to fail builds below threshold

```bash
GO_COVERAGE_ENFORCE_THRESHOLD=true   # Fail builds on low coverage
GO_COVERAGE_ENFORCE_THRESHOLD=false  # Warning only
```

### Badge Color Thresholds

#### `GO_COVERAGE_THRESHOLD_EXCELLENT`
**Type**: Float (0-100)
**Default**: `90`
**Description**: Coverage percentage for bright green badges

#### `GO_COVERAGE_THRESHOLD_GOOD`
**Type**: Float (0-100)
**Default**: `80`
**Description**: Coverage percentage for green badges

#### `GO_COVERAGE_THRESHOLD_ACCEPTABLE`
**Type**: Float (0-100)
**Default**: `70`
**Description**: Coverage percentage for yellow badges

#### `GO_COVERAGE_THRESHOLD_LOW`
**Type**: Float (0-100)
**Default**: `60`
**Description**: Coverage percentage for orange badges (below = red)

```bash
# Custom color thresholds
GO_COVERAGE_THRESHOLD_EXCELLENT=95  # Bright green at 95%+
GO_COVERAGE_THRESHOLD_GOOD=85       # Green at 85-94%
GO_COVERAGE_THRESHOLD_ACCEPTABLE=75 # Yellow at 75-84%
GO_COVERAGE_THRESHOLD_LOW=65        # Orange at 65-74%, red below
```

## Badge Configuration

### Badge Style and Appearance

#### `GO_COVERAGE_BADGE_STYLE`
**Type**: String
**Default**: `flat`
**Options**: `flat`, `flat-square`, `for-the-badge`
**Description**: Visual style of coverage badges

```bash
GO_COVERAGE_BADGE_STYLE=flat          # Standard GitHub style
GO_COVERAGE_BADGE_STYLE=flat-square   # Square corners
GO_COVERAGE_BADGE_STYLE=for-the-badge # Large, prominent style
```

#### `GO_COVERAGE_BADGE_LABEL`
**Type**: String
**Default**: `coverage`
**Description**: Left-side text of the badge

```bash
GO_COVERAGE_BADGE_LABEL=coverage     # Default label
GO_COVERAGE_BADGE_LABEL=tests        # Custom label
GO_COVERAGE_BADGE_LABEL=quality      # Alternative text
```

#### `GO_COVERAGE_BADGE_LOGO`
**Type**: String
**Default**: `go`
**Options**: `go`, `github`, custom URL
**Description**: Logo displayed on badge

```bash
GO_COVERAGE_BADGE_LOGO=go                                    # Go gopher logo
GO_COVERAGE_BADGE_LOGO=github                               # GitHub logo
GO_COVERAGE_BADGE_LOGO=https://example.com/custom-logo.svg # Custom logo
```

#### `GO_COVERAGE_BADGE_LOGO_COLOR`
**Type**: String
**Default**: `white`
**Description**: Color of the logo on the badge

```bash
GO_COVERAGE_BADGE_LOGO_COLOR=white   # White logo
GO_COVERAGE_BADGE_LOGO_COLOR=blue    # Blue logo
GO_COVERAGE_BADGE_LOGO_COLOR=#ff0000 # Custom hex color
```

### Badge Generation

#### `GO_COVERAGE_BADGE_BRANCHES`
**Type**: Comma-separated list
**Default**: `master,development`
**Description**: Branches to generate badges for

```bash
GO_COVERAGE_BADGE_BRANCHES=master,development           # Standard branches
GO_COVERAGE_BADGE_BRANCHES=master,staging,prod    # Custom branches
GO_COVERAGE_BADGE_BRANCHES=master                   # Single branch only
```


## Report Configuration

### Report Generation

#### `GO_COVERAGE_REPORT_TITLE`
**Type**: String
**Default**: `GoFortress Coverage`
**Description**: Title displayed in HTML reports

```bash
GO_COVERAGE_REPORT_TITLE="GoFortress Coverage"     # Default title
GO_COVERAGE_REPORT_TITLE="My Project Coverage"    # Custom title
GO_COVERAGE_REPORT_TITLE="Quality Dashboard"      # Alternative title
```

#### `GO_COVERAGE_REPORT_THEME`
**Type**: String
**Default**: `github-dark`
**Options**: `github-dark`, `github-light`, `custom`
**Description**: Visual theme for reports and dashboard

```bash
GO_COVERAGE_REPORT_THEME=github-dark   # Dark theme (default)
GO_COVERAGE_REPORT_THEME=github-light  # Light theme
GO_COVERAGE_REPORT_THEME=custom        # Custom theme (requires CSS)
```

## Pull Request Integration

### PR Comments

#### `GO_COVERAGE_PR_COMMENT_ENABLED`
**Type**: Boolean
**Default**: `true`
**Description**: Enable automatic PR coverage comments

```bash
GO_COVERAGE_PR_COMMENT_ENABLED=true   # Enable PR comments
GO_COVERAGE_PR_COMMENT_ENABLED=false  # Disable PR comments
```

#### `GO_COVERAGE_PR_COMMENT_BEHAVIOR`
**Type**: String
**Default**: `update`
**Options**: `new`, `update`, `delete-and-new`
**Description**: How to handle multiple PR updates

```bash
GO_COVERAGE_PR_COMMENT_BEHAVIOR=update        # Update existing comment (recommended)
GO_COVERAGE_PR_COMMENT_BEHAVIOR=new           # Create new comment each time
GO_COVERAGE_PR_COMMENT_BEHAVIOR=delete-and-new # Delete old, create new
```

#### `GO_COVERAGE_PR_COMMENT_SHOW_TREE`
**Type**: Boolean
**Default**: `true`
**Description**: Show file tree in PR comments

#### `GO_COVERAGE_PR_COMMENT_SHOW_MISSING`
**Type**: Boolean
**Default**: `true`
**Description**: Highlight uncovered lines in PR comments

```bash
GO_COVERAGE_PR_COMMENT_SHOW_TREE=true      # Show file tree
GO_COVERAGE_PR_COMMENT_SHOW_MISSING=true   # Highlight missing coverage
```

### Label-Based Threshold Overrides

The coverage system supports temporarily overriding coverage thresholds using GitHub PR labels. This feature allows developers to bypass strict coverage requirements for specific scenarios (refactoring, legacy code updates, emergency fixes) while maintaining visibility and audit trails.

#### `GO_COVERAGE_ALLOW_LABEL_OVERRIDE`
**Type**: Boolean
**Default**: `false`
**Description**: Enable coverage threshold overrides via PR labels

#### `GO_COVERAGE_MIN_OVERRIDE_THRESHOLD`
**Type**: Float (0-100)
**Default**: `50.0`
**Description**: Minimum allowed override threshold (security boundary)

#### `GO_COVERAGE_MAX_OVERRIDE_THRESHOLD`
**Type**: Float (0-100)
**Default**: `95.0`
**Description**: Maximum allowed override threshold (prevents abuse)

```bash
# Enable label-based overrides
GO_COVERAGE_ALLOW_LABEL_OVERRIDE=true    # Allow label overrides
GO_COVERAGE_MIN_OVERRIDE_THRESHOLD=50.0  # Must be ≥ 50%
GO_COVERAGE_MAX_OVERRIDE_THRESHOLD=95.0  # Must be ≤ 95%
```

#### Supported Labels

The system recognizes this label in PRs:

- **`coverage-override`**: Completely ignores coverage requirements for this PR

#### Usage Examples

```bash
# Basic override configuration
GO_COVERAGE_ALLOW_LABEL_OVERRIDE=true
GO_COVERAGE_MIN_OVERRIDE_THRESHOLD=60.0
GO_COVERAGE_MAX_OVERRIDE_THRESHOLD=90.0

# Strict override boundaries
GO_COVERAGE_ALLOW_LABEL_OVERRIDE=true
GO_COVERAGE_MIN_OVERRIDE_THRESHOLD=70.0  # No overrides below 70%
GO_COVERAGE_MAX_OVERRIDE_THRESHOLD=85.0  # No overrides above 85%

# Disable overrides (production security)
GO_COVERAGE_ALLOW_LABEL_OVERRIDE=false
```

#### Security Considerations

- Labels are controlled by repository permissions
- All overrides are visible in PR history and status checks
- Override thresholds are bounded by min/max configuration
- Status checks display `[override]` indicator when active
- Repository admins can disable the feature entirely

## Analytics and History

### Trend Analysis

#### `GO_COVERAGE_ENABLE_TREND_ANALYSIS`
**Type**: Boolean
**Default**: `true`
**Description**: Enable historical trend tracking and predictions

#### `GO_COVERAGE_ENABLE_PACKAGE_BREAKDOWN`
**Type**: Boolean
**Default**: `true`
**Description**: Show package-level coverage analysis

#### `GO_COVERAGE_ENABLE_COMPLEXITY_ANALYSIS`
**Type**: Boolean
**Default**: `false`
**Description**: Analyze code complexity (future feature)

```bash
GO_COVERAGE_ENABLE_TREND_ANALYSIS=true      # Enable trends
GO_COVERAGE_ENABLE_PACKAGE_BREAKDOWN=true   # Package details
GO_COVERAGE_ENABLE_COMPLEXITY_ANALYSIS=false # Complexity (planned)
```

### Data Retention

#### `GO_COVERAGE_HISTORY_RETENTION_DAYS`
**Type**: Integer
**Default**: `90`
**Description**: Days to retain historical coverage data

#### `GO_COVERAGE_CLEANUP_PR_AFTER_DAYS`
**Type**: Integer
**Default**: `7`
**Description**: Clean up PR coverage data after merge/close

```bash
GO_COVERAGE_HISTORY_RETENTION_DAYS=90   # Keep 3 months of history
GO_COVERAGE_CLEANUP_PR_AFTER_DAYS=7     # Clean PR data after 1 week
```

## Notification Configuration

### Slack Integration

#### `GO_COVERAGE_SLACK_WEBHOOK_ENABLED`
**Type**: Boolean
**Default**: `false`
**Description**: Enable Slack notifications

#### `GO_COVERAGE_SLACK_WEBHOOK_URL`
**Type**: String (Secret)
**Default**: `""`
**Description**: Slack webhook URL (store in GitHub Secrets)

```bash
GO_COVERAGE_SLACK_WEBHOOK_ENABLED=true                     # Enable Slack
GO_COVERAGE_SLACK_WEBHOOK_URL=${{ secrets.SLACK_WEBHOOK }} # Use secret
```

**Note**: Store the actual webhook URL in GitHub Secrets, not in the configuration file.

## Exclusion Configuration

### Path Exclusions

#### `GO_COVERAGE_EXCLUDE_PATHS`
**Type**: Comma-separated list
**Default**: `test/,vendor/,examples/,third_party/,testdata/`
**Description**: Directory paths to exclude from coverage

```bash
# Default exclusions
GO_COVERAGE_EXCLUDE_PATHS=test/,vendor/,examples/,third_party/,testdata/

# Custom exclusions
GO_COVERAGE_EXCLUDE_PATHS=vendor/,docs/,scripts/,proto/

# Minimal exclusions
GO_COVERAGE_EXCLUDE_PATHS=vendor/
```

#### `GO_COVERAGE_EXCLUDE_FILES`
**Type**: Comma-separated list
**Default**: `*_test.go,*.pb.go,*_mock.go,mock_*.go`
**Description**: File patterns to exclude from coverage

```bash
# Default exclusions
GO_COVERAGE_EXCLUDE_FILES=*_test.go,*.pb.go,*_mock.go,mock_*.go

# Custom exclusions
GO_COVERAGE_EXCLUDE_FILES=*_test.go,*.generated.go,*_pb.go

# Add more patterns
GO_COVERAGE_EXCLUDE_FILES=*_test.go,*.pb.go,*_mock.go,mock_*.go,*_gen.go
```

#### `GO_COVERAGE_EXCLUDE_PACKAGES`
**Type**: Comma-separated list
**Default**: `""`
**Description**: Additional packages to exclude

```bash
GO_COVERAGE_EXCLUDE_PACKAGES=internal/generated,pkg/proto
```

### Advanced Exclusions

#### `GO_COVERAGE_INCLUDE_ONLY_PATHS`
**Type**: Comma-separated list
**Default**: `""`
**Description**: If set, only include these paths (whitelist mode)

```bash
GO_COVERAGE_INCLUDE_ONLY_PATHS=internal/,pkg/    # Only include these paths
GO_COVERAGE_INCLUDE_ONLY_PATHS=""                # Include all (default)
```

#### `GO_COVERAGE_EXCLUDE_GENERATED`
**Type**: Boolean
**Default**: `true`
**Description**: Exclude generated files (detected by header comments)

#### `GO_COVERAGE_EXCLUDE_TEST_FILES`
**Type**: Boolean
**Default**: `true`
**Description**: Exclude test files from coverage analysis

#### `GO_COVERAGE_MIN_FILE_LINES`
**Type**: Integer
**Default**: `10`
**Description**: Minimum lines in file to include in coverage

```bash
GO_COVERAGE_EXCLUDE_GENERATED=true    # Skip generated files
GO_COVERAGE_EXCLUDE_TEST_FILES=true   # Skip test files
GO_COVERAGE_MIN_FILE_LINES=10         # Skip tiny files
```

## Logging and Debugging

### Log Configuration

#### `GO_COVERAGE_LOG_LEVEL`
**Type**: String
**Default**: `info`
**Options**: `debug`, `info`, `warn`, `error`
**Description**: Logging verbosity level

#### `GO_COVERAGE_LOG_FORMAT`
**Type**: String
**Default**: `json`
**Options**: `json`, `text`, `pretty`
**Description**: Log output format

```bash
GO_COVERAGE_LOG_LEVEL=debug     # Verbose logging
GO_COVERAGE_LOG_FORMAT=pretty   # Human-readable format
```

#### `GO_COVERAGE_LOG_FILE`
**Type**: String
**Default**: `/tmp/coverage.log`
**Description**: Log file path

#### `GO_COVERAGE_LOG_MAX_SIZE`
**Type**: String
**Default**: `10MB`
**Description**: Maximum log file size before rotation

#### `GO_COVERAGE_LOG_RETENTION_DAYS`
**Type**: Integer
**Default**: `7`
**Description**: Days to retain log files

```bash
GO_COVERAGE_LOG_FILE=/var/log/coverage.log   # Custom log location
GO_COVERAGE_LOG_MAX_SIZE=50MB                # Larger log files
GO_COVERAGE_LOG_RETENTION_DAYS=30            # Keep logs longer
```

### Debug Features

#### `GO_COVERAGE_DEBUG_MODE`
**Type**: Boolean
**Default**: `false`
**Description**: Enable verbose debugging output

#### `GO_COVERAGE_TRACE_ERRORS`
**Type**: Boolean
**Default**: `true`
**Description**: Include stack traces in error logs

#### `GO_COVERAGE_LOG_PERFORMANCE`
**Type**: Boolean
**Default**: `true`
**Description**: Log timing and performance metrics

#### `GO_COVERAGE_LOG_MEMORY_USAGE`
**Type**: Boolean
**Default**: `true`
**Description**: Log memory consumption statistics

```bash
GO_COVERAGE_DEBUG_MODE=true           # Enable debug mode
GO_COVERAGE_TRACE_ERRORS=true         # Include stack traces
GO_COVERAGE_LOG_PERFORMANCE=true      # Performance metrics
GO_COVERAGE_LOG_MEMORY_USAGE=true     # Memory tracking
```

## Monitoring and Metrics

### Metrics Collection

#### `GO_COVERAGE_METRICS_ENABLED`
**Type**: Boolean
**Default**: `true`
**Description**: Enable metrics collection and reporting

#### `GO_COVERAGE_METRICS_ENDPOINT`
**Type**: String
**Default**: `""`
**Description**: Optional external metrics endpoint

#### `GO_COVERAGE_METRICS_INCLUDE_ERRORS`
**Type**: Boolean
**Default**: `true`
**Description**: Track error metrics and rates

#### `GO_COVERAGE_METRICS_INCLUDE_PERFORMANCE`
**Type**: Boolean
**Default**: `true`
**Description**: Track performance and timing metrics

#### `GO_COVERAGE_METRICS_INCLUDE_USAGE`
**Type**: Boolean
**Default**: `true`
**Description**: Track feature usage and adoption

```bash
GO_COVERAGE_METRICS_ENABLED=true               # Enable metrics
GO_COVERAGE_METRICS_ENDPOINT=https://metrics.example.com # Optional endpoint
GO_COVERAGE_METRICS_INCLUDE_ERRORS=true        # Error tracking
GO_COVERAGE_METRICS_INCLUDE_PERFORMANCE=true   # Performance tracking
GO_COVERAGE_METRICS_INCLUDE_USAGE=true         # Usage analytics
```

## Testing and Error Injection

### Test Mode Configuration

#### `GO_COVERAGE_TEST_MODE`
**Type**: Boolean
**Default**: `false`
**Description**: Enable test mode for development

#### `GO_COVERAGE_INJECT_ERRORS`
**Type**: Comma-separated list
**Default**: `""`
**Options**: `parser`, `api`, `storage`
**Description**: Components to inject errors in (testing only)

#### `GO_COVERAGE_ERROR_RATE`
**Type**: Float (0-1)
**Default**: `0`
**Description**: Error injection rate for testing

```bash
# Development/testing only
GO_COVERAGE_TEST_MODE=true
GO_COVERAGE_INJECT_ERRORS=parser,api
GO_COVERAGE_ERROR_RATE=0.1  # 10% error rate
```

**Warning**: Only use error injection in development environments.

## Environment-Specific Examples

### Development Environment
```bash
# Development configuration
ENABLE_GO_COVERAGE=true
GO_COVERAGE_FAIL_UNDER=70
GO_COVERAGE_DEBUG_MODE=true
GO_COVERAGE_LOG_LEVEL=debug
GO_COVERAGE_PR_COMMENT_ENABLED=true
GO_COVERAGE_PAGES_AUTO_CREATE=true
GO_COVERAGE_CLEANUP_PR_AFTER_DAYS=1
```

### Staging Environment
```bash
# Staging configuration
ENABLE_GO_COVERAGE=true
GO_COVERAGE_FAIL_UNDER=80
GO_COVERAGE_ENFORCE_THRESHOLD=false
GO_COVERAGE_LOG_LEVEL=info
GO_COVERAGE_ENABLE_TREND_ANALYSIS=true
GO_COVERAGE_HISTORY_RETENTION_DAYS=30
```

### Production Environment
```bash
# Production configuration
ENABLE_GO_COVERAGE=true
GO_COVERAGE_FAIL_UNDER=85
GO_COVERAGE_ENFORCE_THRESHOLD=true
GO_COVERAGE_LOG_LEVEL=warn
GO_COVERAGE_ENABLE_TREND_ANALYSIS=true
GO_COVERAGE_ENABLE_PACKAGE_BREAKDOWN=true
GO_COVERAGE_HISTORY_RETENTION_DAYS=90
GO_COVERAGE_SLACK_WEBHOOK_ENABLED=true
GO_COVERAGE_METRICS_ENABLED=true
```

## Configuration Validation

### Common Issues

#### Missing Required Variables
```bash
# ❌ Minimal setup - may cause issues
ENABLE_GO_COVERAGE=true

# ✅ Recommended minimum
ENABLE_GO_COVERAGE=true
GO_COVERAGE_FAIL_UNDER=80
GO_COVERAGE_PR_COMMENT_ENABLED=true
```

#### Invalid Values
```bash
# ❌ Invalid threshold
GO_COVERAGE_FAIL_UNDER=150  # Must be 0-100

# ❌ Invalid style
GO_COVERAGE_BADGE_STYLE=invalid  # Must be flat, flat-square, or for-the-badge

# ❌ Invalid log level
GO_COVERAGE_LOG_LEVEL=verbose  # Must be debug, info, warn, or error
```

### Validation Tools
```bash
# Validate configuration (when CLI is available)
gofortress-coverage validate --config

# Check environment variables
env | grep GO_COVERAGE_ | sort
```

## Best Practices

### Performance Optimization
```bash
# Optimize for speed
GO_COVERAGE_LOG_LEVEL=warn              # Reduce logging overhead
GO_COVERAGE_DEBUG_MODE=false            # Disable debug mode
GO_COVERAGE_LOG_PERFORMANCE=false       # Skip performance logging in production
```

### Security Considerations
```bash
# Security best practices
GO_COVERAGE_DEBUG_MODE=false            # Disable debug in production
GO_COVERAGE_TRACE_ERRORS=false          # Reduce information disclosure
GO_COVERAGE_TEST_MODE=false             # Disable test mode
GO_COVERAGE_INJECT_ERRORS=""            # No error injection
```

### Maintenance
```bash
# Balanced retention
GO_COVERAGE_HISTORY_RETENTION_DAYS=90   # 3 months of history
GO_COVERAGE_CLEANUP_PR_AFTER_DAYS=7     # Clean PR data weekly
GO_COVERAGE_LOG_RETENTION_DAYS=14       # 2 weeks of logs
```

---

## Related Documentation

- [📖 System Overview](coverage-system.md) - Architecture and components
- [🎯 Feature Showcase](coverage-features.md) - Detailed feature examples
- [🛠️ API Reference](coverage-api.md) - CLI commands and automation
