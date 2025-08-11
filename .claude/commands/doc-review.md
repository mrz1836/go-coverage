---
allowed-tools: Task
description: Review documentation accuracy against current code
model: sonnet
---

## Context
- Documentation files: !`find . -name "*.md" -type f | grep -v vendor | head -20`
- Recent code changes: !`git log --oneline -10`
- Exported functions: !`grep -r "^func [A-Z]" --include="*.go" | wc -l` public functions

## Task

Review documentation accuracy using **documentation-manager** and **code-reviewer** agents in parallel:

1. **Documentation Audit** (documentation-manager agent):
   - Check if README examples still work
   - Verify CLI commands are current
   - Validate configuration examples
   - Ensure API documentation matches implementation

2. **Code-Doc Alignment** (code-reviewer agent):
   - Verify function documentation matches implementation
   - Check if deprecated features are marked
   - Ensure new features are documented
   - Validate example code compiles

3. **Review Checklist**:
   - [ ] Installation instructions work
   - [ ] Usage examples are accurate
   - [ ] API documentation is complete
   - [ ] Configuration options are documented
   - [ ] Changelog is up to date
   - [ ] Function comments match signatures

4. **Report Issues Found**:
   - Outdated examples
   - Missing documentation
   - Incorrect instructions
   - Broken links
   - Version mismatches

Provide specific recommendations for fixes needed.
