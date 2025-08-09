---
allowed-tools: Task
description: Clean up code formatting, imports, and style issues
model: haiku
---

## Context
- Modified files: !`git diff --name-only | grep -E "\.go$"`
- Formatting check: !`gofmt -l . | head -10`
- Import issues: !`goimports -l . | head -10`

## Task

Clean up code using the **go-linter** agent:

1. **Auto-Format Code**:
   - Run `go fmt ./...`
   - Run `goimports -w .`
   - Run `gofumpt -w .`

2. **Fix Common Issues**:
   - Remove unused imports
   - Organize import groups
   - Fix whitespace issues
   - Correct indentation
   - Remove trailing spaces

3. **Style Improvements**:
   - Consistent naming conventions
   - Proper comment formatting
   - Line length compliance
   - File organization

4. **YAML Files**:
   - Format with prettier
   - Fix indentation
   - Validate syntax

5. **Validation**:
   - No formatting changes with `go fmt`
   - No import issues with `goimports`
   - Linter passes

Quick operation - should complete in seconds. No functional changes, only formatting and style cleanup.