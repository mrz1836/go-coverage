---
allowed-tools: Task
argument-hint: [file or package to refactor]
description: Suggest and implement refactoring improvements
model: opus
---

## Context
- Target: $ARGUMENTS
- Complexity: !`gocyclo -over 10 . 2>/dev/null | head -10`
- Long functions: !`grep -n "^func" --include="*.go" -A 50 . | grep -E "^[0-9]+-func|^[0-9]+--$" | awk 'BEGIN{prev=0} /^[0-9]+-func/{start=$1} /^[0-9]+--$/{if($1-start>50) print prev} {prev=$0}' | head -5`

## Task

Refactor code for better maintainability using **code-reviewer** and **go-linter** agents:

1. **Analysis** (code-reviewer agent):
   - Identify code smells
   - Find complex functions (cyclomatic complexity > 10)
   - Detect deep nesting (> 4 levels)
   - Locate large functions (> 50 lines)
   - Find unclear abstractions

2. **Refactoring Patterns**:
   - **Extract Method**: Break large functions
   - **Extract Interface**: Improve testability
   - **Replace Conditionals**: Use polymorphism
   - **Introduce Parameter Object**: Reduce parameter count
   - **Remove Duplication**: DRY principle
   - **Simplify Conditionals**: Guard clauses

3. **Go-Specific Improvements**:
   - Use context properly (first parameter)
   - Apply interface segregation
   - Implement error wrapping correctly
   - Remove global state
   - Apply dependency injection

4. **Validation** (go-linter agent):
   - Ensure tests still pass
   - Verify linting passes
   - Check complexity reduced
   - Confirm better readability

Focus on:
- Making code more testable
- Reducing cognitive complexity
- Improving naming
- Better error handling
- Clearer abstractions
