# ADR-003: Runtime Architecture

**Status:** Accepted
**Date:** 2026-07-12
**Author:** Chief Software Architect

## Context

V1 had a Python-specific engine (`engine/runner.go`) that directly executed
Python code. This made it impossible to:
- Support non-Python projects (MATLAB, ROS2, Docker)
- Reuse environment setups across projects
- Have a unified execution lifecycle

## Decision

Runtime is a **unified execution engine** that:

1. Does NOT know about Workflow
2. Does NOT contain business logic
3. Does NOT know about AI algorithms
4. Only executes standard commands

### Key Design Points

1. **Runtime interface** is command-based: `Execute(ctx, project, config)`
2. **BundleManager** manages versioned runtime environments
3. **Executor** handles subprocess lifecycle
4. **Log streaming** is real-time via callbacks

### Runtime States

```
Idle → Detecting → Ready → Installing → Prepared → Executing → Completed/Failed
```

## Consequences

Positive:
- Supports any programming language or tool
- Runtime bundles are cached and shared
- Clean separation from Compiler and Workflow
- Testable with mock commands

Negative:
- Cannot deeply integrate with specific tools
- Need standard command-line interfaces

## Related Decisions

- ADR-001: Compiler Architecture
- ADR-004: Plugin System