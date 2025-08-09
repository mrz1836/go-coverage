---
allowed-tools: Task
argument-hint: [feature, module, or file to explain]
description: Explain how a feature or module works with examples
model: opus
---

## Task

Explain how the specified feature/module works: **$ARGUMENTS**

Use the **documentation-manager** and **code-reviewer** agents to provide a comprehensive explanation:

1. **Code Analysis** (code-reviewer agent):
   - Trace through the implementation
   - Identify key components and their roles
   - Map data flow and dependencies
   - Note design patterns used

2. **Documentation** (documentation-manager agent):
   - Create clear explanation with examples
   - Include architecture diagrams if helpful
   - Provide usage examples
   - Document edge cases and limitations

3. **Explanation Structure**:
   - **Overview**: High-level purpose and design
   - **Components**: Key parts and their responsibilities
   - **Data Flow**: How data moves through the system
   - **API**: Public interfaces and how to use them
   - **Examples**: Practical usage scenarios
   - **Internals**: Implementation details if relevant
   - **Edge Cases**: Limitations and special considerations

4. **Make it Clear**:
   - Use analogies where helpful
   - Provide concrete examples
   - Explain the "why" behind design decisions
   - Include code snippets for clarity

If no specific target provided, explain the overall go-coverage system architecture.