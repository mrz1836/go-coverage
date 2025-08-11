---
allowed-tools: Task
description: Fix test failures and linter issues automatically
model: opus
---

## Context
- Test status: !`go test ./... 2>&1 | grep -E "FAIL|Error" | head -20`
- Linter issues: !`make lint 2>&1 | head -30`
- Recent changes: !`git diff --name-only HEAD~1`

## Task

You have test failures and/or linter issues that need to be fixed. Use multiple agents in parallel to resolve these issues efficiently:

1. **Use the go-test-runner agent** to identify and fix test failures:
   - Analyze test output for root causes
   - Fix failing assertions
   - Resolve race conditions
   - Ensure test conventions follow AGENTS.md

2. **Use the go-linter agent** to fix formatting and linting issues:
   - Run formatting tools (gofmt, goimports, gofumpt)
   - Fix linting violations
   - Ensure code follows AGENTS.md standards

3. **Use the debugger agent** if there are complex test failures or panics

Work systematically through all issues. For test failures, fix the root cause not just symptoms. For linter issues, apply automatic fixes where possible and manual fixes for complex issues.

After fixing, verify everything passes by running tests and linter again.

$ARGUMENTS
