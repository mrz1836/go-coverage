---
allowed-tools: Task
argument-hint: [feature or file to document]
description: Update documentation for new or modified features
model: opus
---

## Context
- Recent changes: !`git diff --name-only HEAD~5 | grep -E "\.go$"`
- Modified functions: !`git diff HEAD~1 | grep "^+func" | head -10`
- Current docs: !`ls -la *.md docs/*.md 2>/dev/null`

## Task

Update documentation using the **documentation-manager** agent:

1. **Identify Documentation Needs**:
   - New features added: $ARGUMENTS
   - API changes requiring updates
   - README sections needing revision
   - Code comments requiring updates

2. **Update Code Documentation**:
   - Add/update function comments
   - Ensure package documentation is current
   - Add examples for new features
   - Document side effects and gotchas

3. **Update Project Documentation**:
   - README.md for user-facing changes
   - CHANGELOG.md for version history
   - API documentation for new endpoints
   - Configuration docs for new options

4. **Documentation Standards** (from AGENTS.md):
   - Function comments start with function name
   - Document "why" not "what"
   - Include parameter descriptions
   - Note side effects explicitly

5. **Verify Documentation**:
   - Examples compile and run
   - Links are not broken
   - Version numbers are current
   - Instructions actually work

Focus on clarity and completeness. Documentation should enable users to understand and use features without reading code.
