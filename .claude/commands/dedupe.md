---
allowed-tools: Task
description: Find and remove duplicate code patterns
model: opus
---

## Context
- Project structure: !`find . -name "*.go" -type f | grep -v vendor | wc -l` Go files
- Similar patterns: !`grep -r "func.*Error" --include="*.go" | cut -d: -f2 | sort | uniq -d | head -10`

## Task

Find and eliminate duplicate code using **code-reviewer** and **performance-optimizer** agents:

1. **Identify Duplicates** (code-reviewer agent):
   - Find copy-pasted code blocks
   - Identify similar function implementations
   - Detect repeated error handling patterns
   - Find duplicate struct definitions

2. **Analyze Patterns** (performance-optimizer agent):
   - Assess performance impact of duplication
   - Identify refactoring opportunities
   - Suggest abstraction patterns

3. **Refactoring Strategy**:
   - Extract common functions to shared packages
   - Create interfaces for similar behaviors
   - Implement DRY principle properly
   - Use generics where appropriate (Go 1.18+)

4. **Implementation**:
   - Create shared utilities
   - Refactor duplicate code
   - Update all references
   - Ensure tests still pass

Focus on:
- Error handling patterns
- Validation logic
- Data transformation functions
- Test helper functions

Provide a report of duplicates found and changes made.
