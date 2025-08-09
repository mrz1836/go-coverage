---
allowed-tools: Task
argument-hint: [specific area to optimize, optional]
description: Analyze and optimize performance bottlenecks
model: opus
---

## Context
- Benchmarks: !`go test -bench=. -benchmem ./... 2>&1 | grep -E "^Benchmark" | head -10`
- Package sizes: !`go list -f '{{.ImportPath}} {{len .GoFiles}}' ./...`
- Performance targets: Parse ~50ms, Badge ~5ms, Report ~200ms (from CLAUDE.md)

## Task

Optimize performance using the **performance-optimizer** agent:

Target area: **$ARGUMENTS**

1. **Performance Analysis**:
   - Run benchmarks to establish baseline
   - Profile CPU and memory usage
   - Identify bottlenecks
   - Analyze algorithmic complexity

2. **Optimization Areas**:
   - **Algorithm Optimization**:
     - Replace inefficient algorithms
     - Optimize hot paths
     - Reduce complexity
   
   - **Memory Optimization**:
     - Reduce allocations
     - Use object pooling
     - Optimize data structures
     - Fix memory leaks
   
   - **Concurrency**:
     - Add parallelization where beneficial
     - Optimize goroutine usage
     - Improve synchronization

3. **Specific Targets** (from CLAUDE.md):
   - Coverage parsing: < 50ms for 10K lines
   - Badge generation: < 5ms
   - HTML report: < 200ms
   - Complete pipeline: < 2s
   - Memory usage: < 10MB

4. **Validation**:
   - Run benchmarks to measure improvement
   - Ensure no functional regression
   - Verify memory usage reduced
   - Document optimizations made

Provide before/after metrics and specific improvements achieved.