---
allowed-tools: Task
argument-hint: [feature name and brief description]
description: Design a Product Requirements Document for a new feature
model: opus
---

## Context
- Project goals: Go coverage system with zero external dependencies
- Current features: !`grep "^##" README.md | grep -i feature | head -10`
- Architecture: !`ls -la internal/ cmd/`

## Task

Design a comprehensive PRD for: **$ARGUMENTS**

Use the **documentation-manager** agent to create a detailed product requirements document:

1. **Feature Overview**:
   - Problem statement
   - Proposed solution
   - Value proposition
   - Success metrics

2. **Requirements**:
   - **Functional Requirements**:
     - Core functionality
     - User workflows
     - API specifications
     - Configuration options

   - **Non-Functional Requirements**:
     - Performance targets (based on CLAUDE.md metrics)
     - Security requirements
     - Compatibility needs
     - Scalability considerations

3. **Technical Design**:
   - Architecture overview
   - Component design
   - Data structures
   - Interface definitions
   - Integration points

4. **Implementation Plan**:
   - Development phases
   - Testing strategy
   - Rollout plan
   - Documentation needs

5. **Considerations**:
   - Dependencies
   - Risks and mitigations
   - Alternatives considered
   - Future extensibility

Format as a structured document suitable for team review and implementation planning.
