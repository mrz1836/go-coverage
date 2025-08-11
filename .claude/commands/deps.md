---
allowed-tools: Task
description: Update dependencies and check for vulnerabilities
model: sonnet
---

## Context
- Current dependencies: !`go list -m all | head -20`
- Outdated packages: !`go list -u -m all | grep '\[' | head -10`
- Module status: !`go mod verify`

## Task

Manage dependencies using **dependency-manager** and **security-scanner** agents in parallel:

1. **Dependency Updates** (dependency-manager agent):
   - Check for available updates
   - Update minor/patch versions safely
   - Document major version changes
   - Run `go mod tidy`
   - Verify module integrity

2. **Security Scanning** (security-scanner agent):
   - Run govulncheck for vulnerabilities
   - Check for known CVEs
   - Scan for leaked secrets
   - Verify license compliance

3. **Update Process**:
   - Update dependencies: `go get -u ./...`
   - Clean up: `go mod tidy`
   - Verify: `go mod verify`
   - Test: `make test`
   - Scan: `make govulncheck`

4. **Risk Assessment**:
   - Breaking changes in major versions
   - Security fixes needed urgently
   - Deprecated packages to replace
   - License compatibility issues

5. **Validation**:
   - All tests pass
   - No new vulnerabilities
   - Build succeeds
   - Performance not degraded

Report:
- Updated packages and versions
- Security issues found/fixed
- Breaking changes requiring attention
- Recommendations for manual review
