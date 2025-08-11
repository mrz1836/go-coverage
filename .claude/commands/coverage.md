---
allowed-tools: Task
description: Analyze coverage and suggest improvements to reach 90%+ target
model: opus
---

## Context
- Overall coverage: !`go test -coverprofile=coverage.txt ./... 2>&1 && go tool cover -func=coverage.txt | grep total`
- Package breakdown: !`go test -cover ./... 2>&1 | grep coverage`
- Uncovered files: !`go tool cover -func=coverage.txt 2>/dev/null | grep -E "0.0%|[0-5][0-9].[0-9]%" | head -20`

## Task

Analyze and improve test coverage using the **coverage-analyzer** and **go-test-runner** agents in parallel:

1. **Coverage Analysis** (coverage-analyzer agent):
   - Parse coverage profile
   - Identify packages below 90% threshold
   - Find critical uncovered code paths
   - Generate coverage report

2. **Test Creation** (go-test-runner agent):
   - Create tests for uncovered functions
   - Focus on high-impact code first
   - Add edge case testing
   - Ensure error paths are covered

3. **Prioritization**:
   - Critical business logic first
   - Public APIs second
   - Error handling paths third
   - Internal helpers last

4. **Validation**:
   - Verify coverage increased to >= 90%
   - Ensure no test quality regression
   - Confirm all tests pass

Generate a summary showing:
- Current coverage vs target
- Files that need attention
- Specific functions lacking tests
- Recommended test additions
