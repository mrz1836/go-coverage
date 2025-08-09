---
allowed-tools: Bash(git add:*), Bash(git status:*), Bash(git commit:*), Bash(git diff:*)
argument-hint: [optional commit message]
description: Create smart git commit with proper message format
model: haiku
---

## Context
- Git status: !`git status --short`
- Staged changes: !`git diff --staged --stat`
- Recent commits: !`git log --oneline -5`
- Branch: !`git branch --show-current`

## Task

Create a well-formatted git commit following AGENTS.md conventions.

## Commit Message Format

Follow the conventional format:
```
<type>(<scope>): <short description>

<body>
```

### Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions/changes
- `refactor`: Code refactoring
- `chore`: Maintenance tasks
- `build`: Build system changes
- `ci`: CI/CD changes

### Rules:
- Subject line ≤ 50 characters
- Use imperative mood ("add" not "added")
- Body wrapped at 72 characters
- Reference issues with "Closes #123"

## Process

1. Review the changes to understand what was modified
2. Determine the appropriate type and scope
3. Write a clear, concise description
4. Add body if needed for context
5. Stage appropriate files
6. Create the commit

User input: $ARGUMENTS

If no message provided, analyze changes and generate appropriate commit message based on the diff.