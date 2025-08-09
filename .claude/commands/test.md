---
allowed-tools: Task
argument-hint: [file or package path]
description: Create comprehensive tests for code
model: opus
---

## Context
- Target: $ARGUMENTS
- Current coverage: !`go test -cover ./$ARGUMENTS 2>/dev/null | grep coverage || echo "No tests found"`
- Package structure: !`ls -la $ARGUMENTS 2>/dev/null || find . -name "*.go" -path "*$ARGUMENTS*" | head -10`

## Task

Create comprehensive tests for the specified code using the **go-test-runner agent**:

1. **Analyze existing code** to understand:
   - Public functions and methods that need testing
   - Edge cases and error conditions
   - Dependencies that need mocking

2. **Create test files** following AGENTS.md standards:
   - Use TestFunctionNameScenarioDescription naming
   - Use testify/assert and testify/require appropriately
   - Create table-driven tests with named cases
   - Mock external dependencies
   - Test error cases and edge conditions

3. **Ensure coverage targets**:
   - Aim for >= 90% coverage
   - Cover all public functions
   - Test both success and failure paths
   - Include boundary conditions

4. **Validate test quality**:
   - Tests are deterministic (no flaky tests)
   - Tests are isolated (no side effects)
   - Tests follow project conventions

If no specific target is provided, analyze recently modified files and create tests for uncovered code.