---
allowed-tools: Task
description: Make code PR-ready with tests, linting, and review
model: opus
---

## Context
- Current branch: !`git branch --show-current`
- Changed files: !`git diff --name-only origin/master | head -20`
- CI status: !`gh run list --limit 1 --json status,conclusion`

## Task

Prepare code for PR using multiple agents in sequence:

1. **Code Cleanup** (go-linter agent):
   - Format all code
   - Fix linting issues
   - Organize imports
   - Clean up whitespace

2. **Test Suite** (go-test-runner agent):
   - Run all tests
   - Ensure coverage >= 90%
   - Fix any test failures
   - Add tests for new code

3. **Security Check** (security-scanner agent):
   - Scan for vulnerabilities
   - Check for secrets
   - Validate dependencies

4. **Code Review** (code-reviewer agent):
   - Review for best practices
   - Check AGENTS.md compliance
   - Identify potential issues
   - Suggest improvements

5. **Documentation** (documentation-manager agent):
   - Update function comments
   - Update README if needed
   - Add changelog entry
   - Verify examples work

6. **Final Validation**:
   - [ ] All tests pass
   - [ ] Coverage >= 90%
   - [ ] Linting passes
   - [ ] No security issues
   - [ ] Documentation updated
   - [ ] Commit messages follow convention
   - [ ] Branch up to date with master

Provide a PR readiness report with any remaining issues to address.