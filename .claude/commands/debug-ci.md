---
allowed-tools: Task
description: Diagnose and fix CI/GitHub Actions failures
model: opus
---

## Context
- Failed runs: !`gh run list --status failure --limit 3 --json name,conclusion,displayTitle`
- Latest failure: !`gh run list --status failure --limit 1 --json databaseId --jq '.[0].databaseId' | xargs -I {} gh run view {} --log-failed 2>&1 | head -50`
- Workflow files: !`ls -la .github/workflows/*.yml | head -10`

## Task

Diagnose CI issues using **ci-workflow** and **debugger** agents in parallel:

1. **Failure Analysis** (debugger agent):
   - Parse error messages
   - Identify root cause
   - Check for flaky tests
   - Analyze timeout issues
   - Review resource constraints

2. **Workflow Investigation** (ci-workflow agent):
   - Check workflow syntax
   - Verify permissions
   - Review environment variables
   - Validate secrets/tokens
   - Check caching issues

3. **Common CI Issues**:
   - **Test Failures**:
     - Environment differences
     - Race conditions
     - Timing dependencies
     - Missing test data
   
   - **Build Failures**:
     - Dependency issues
     - Version mismatches
     - Cache corruption
     - Permission errors
   
   - **Performance Issues**:
     - Slow tests
     - Resource exhaustion
     - Network timeouts
     - Rate limiting

4. **Fixes to Apply**:
   - Add retry logic for flaky tests
   - Fix environment configuration
   - Update action versions
   - Improve error handling
   - Add debugging output

5. **Validation**:
   - Re-run failed workflow
   - Verify fixes work
   - Document solution

Provide specific fixes for the CI failures found.