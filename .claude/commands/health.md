---
allowed-tools: Task
description: Comprehensive project health check
model: opus
---

## Context
- CI status: !`gh run list --limit 3 --json status,conclusion | grep -E "status|conclusion"`
- Test status: !`go test ./... 2>&1 | tail -5`
- Coverage: !`go test -cover ./... 2>&1 | grep coverage | tail -5`
- Linting: !`make lint 2>&1 | tail -10`

## Task

Perform comprehensive health check using multiple agents in parallel:

1. **Testing Health** (go-test-runner agent):
   - All tests passing
   - Coverage >= 90%
   - No flaky tests
   - Test execution time reasonable

2. **Code Quality** (go-linter & code-reviewer agents):
   - Linting passes
   - No code smells
   - Consistent formatting
   - Documentation current

3. **Security Status** (security-scanner agent):
   - No vulnerabilities
   - Dependencies up to date
   - No exposed secrets

4. **Performance** (performance-optimizer agent):
   - Benchmarks within targets
   - No performance regression
   - Memory usage acceptable

5. **CI/CD Health** (ci-workflow agent):
   - All workflows passing
   - Build times reasonable
   - No flaky CI issues

6. **Dependencies** (dependency-manager agent):
   - All verified
   - No conflicts
   - Licenses compatible

Health Report Card:
- ✅ Green: Healthy
- ⚠️ Yellow: Needs attention
- ❌ Red: Critical issues

Provide:
- Overall health score
- Issues by priority
- Recommended actions
- Trend analysis if available
