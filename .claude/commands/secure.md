---
allowed-tools: Task
description: Comprehensive security vulnerability scan
model: opus
---

## Context
- Recent changes: !`git diff --name-only HEAD~5 | grep -E "\.go$"`
- Sensitive files: !`find . -name "*.key" -o -name "*.pem" -o -name ".env" 2>/dev/null | head -5`
- Dependencies: !`go list -m all | wc -l` dependencies

## Task

Perform comprehensive security audit using **security-scanner** agent:

1. **Vulnerability Scanning**:
   - Run govulncheck for Go vulnerabilities
   - Scan dependencies for CVEs
   - Check for outdated packages with known issues

2. **Code Security Analysis**:
   - **Secret Detection**:
     - Hardcoded credentials
     - API keys in code
     - Leaked tokens
   
   - **Common Vulnerabilities**:
     - SQL injection risks
     - Path traversal
     - Command injection
     - Weak cryptography
     - Insecure random
   
   - **Input Validation**:
     - Unvalidated user input
     - Missing bounds checks
     - Type confusion

3. **Security Best Practices**:
   - Context handling for cancellation
   - Proper error handling (no info leakage)
   - Resource cleanup (no leaks)
   - Safe concurrency patterns
   - Secure defaults

4. **GitHub Security**:
   - Token handling in github package
   - Rate limiting implementation
   - Webhook validation

5. **Compliance Check**:
   - OpenSSF Scorecard compliance
   - Security policy (SECURITY.md) current
   - Dependency licenses compatible

Priority levels:
- 🔴 Critical (CVSS 9.0+): Immediate fix
- 🟡 High (CVSS 7.0-8.9): Fix within 24h
- 🟢 Medium/Low: Next update cycle

Provide actionable recommendations for each issue found.