---
allowed-tools: Task
description: Comprehensive code review with security and quality checks
model: opus
---

## Context
- Changed files: !`git diff --name-only HEAD~1 | grep -E "\.go$"`
- Recent commits: !`git log --oneline -5`
- Current branch: !`git branch --show-current`

## Task

Perform comprehensive code review using multiple agents in parallel:

1. **Code Quality Review** (code-reviewer agent):
   - Go best practices and idioms
   - AGENTS.md compliance
   - Error handling patterns
   - Interface design
   - Test coverage

2. **Security Review** (security-scanner agent):
   - Vulnerability scanning
   - Secret detection
   - Input validation
   - SQL injection risks
   - Path traversal issues

3. **Performance Review** (performance-optimizer agent):
   - Algorithm efficiency
   - Memory allocations
   - Goroutine usage
   - Resource leaks

4. **Lint & Format** (go-linter agent):
   - Code formatting issues
   - Naming conventions
   - Comment quality
   - Import organization

Review checklist:
- [ ] No security vulnerabilities
- [ ] Proper error handling
- [ ] Adequate test coverage
- [ ] Performance acceptable
- [ ] Code is maintainable
- [ ] Documentation updated
- [ ] No code duplication

Provide feedback organized by priority:
- 🔴 Critical (must fix)
- 🟡 Important (should fix)
- 🟢 Suggestions (consider)
