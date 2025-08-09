---
allowed-tools: Task
description: Run performance benchmarks and analyze results
model: sonnet
---

## Context
- Current benchmarks: !`go test -bench=. -benchmem ./... 2>&1 | grep -E "^Benchmark" | head -15`
- CPU info: !`grep "model name" /proc/cpuinfo 2>/dev/null | head -1 || sysctl -n machdep.cpu.brand_string 2>/dev/null`

## Task

Run and analyze benchmarks using the **performance-optimizer** agent:

1. **Run Benchmarks**:
   - Execute: `make bench`
   - Or: `go test -bench=. -benchmem ./...`
   - Multiple runs for stability: `-count=10`
   - With CPU profile: `-cpuprofile=cpu.prof`
   - With memory profile: `-memprofile=mem.prof`

2. **Key Metrics to Track**:
   - Operations per second
   - Nanoseconds per operation
   - Memory allocations per op
   - Bytes allocated per op

3. **Performance Targets** (from CLAUDE.md):
   - Parse Coverage: ~50ms for 10K lines
   - Generate Badge: ~5ms
   - Build HTML Report: ~200ms
   - Complete Pipeline: ~1-2s
   - Memory Usage: <10MB

4. **Analysis**:
   - Compare against targets
   - Identify slowest operations
   - Find memory allocation hotspots
   - Detect performance regressions

5. **Optimization Opportunities**:
   - Functions with high allocation
   - Inefficient algorithms
   - Missing caching opportunities
   - Unnecessary work in hot paths

6. **Report**:
   - Current performance metrics
   - Comparison with targets
   - Bottlenecks identified
   - Recommendations for improvement

Save benchmark results for future comparison.