---
allowed-tools: Task
description: Prepare for a new release with all checks
model: opus
---

## Context
- Current version: !`git describe --tags --abbrev=0 2>/dev/null || echo "No tags yet"`
- Pending changes: !`git log $(git describe --tags --abbrev=0 2>/dev/null)..HEAD --oneline | head -20`
- CI status: !`gh run list --branch master --limit 3 --json status,conclusion`

## Task

Prepare for release using **release-manager**, **go-test-runner**, and **coverage-analyzer** agents:

1. **Pre-Release Validation** (go-test-runner agent):
   - Run full test suite
   - Verify coverage >= 90%
   - Run benchmarks
   - Check for flaky tests

2. **Security & Dependencies** (dependency-manager & security-scanner agents):
   - Run security scans
   - Update dependencies
   - Verify no vulnerabilities
   - Check license compliance

3. **Release Preparation** (release-manager agent):
   - Determine version (MAJOR.MINOR.PATCH)
   - Generate changelog
   - Update version in CITATION.cff
   - Prepare release notes
   - Tag strategy

4. **Documentation** (documentation-manager agent):
   - Update README version references
   - Update CHANGELOG.md
   - Verify examples work
   - Update API docs

5. **Release Checklist**:
   - [ ] All CI checks passing
   - [ ] Coverage meets threshold
   - [ ] No security vulnerabilities
   - [ ] Dependencies updated
   - [ ] Documentation current
   - [ ] CHANGELOG updated
   - [ ] Version numbers consistent
   - [ ] Examples tested
   - [ ] Breaking changes documented

6. **Release Commands**:
   ```bash
   make release-snap  # Test snapshot
   make tag version=X.Y.Z  # Create tag
   ```

Provide release readiness assessment and version recommendation.
