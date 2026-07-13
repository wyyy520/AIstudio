# ADR-002: Compiler-first architecture

**Status:** Accepted

## Context

V1 used a monolithic executor that combined compilation, execution, and environment management. This made it hard to add new targets (each target required modifying the core). The system was not extensible and tightly coupled all concerns together.

## Decision

The Compiler is a multi-stage pipeline (Plan → Validate → Generate → Verify) that delegates to target-specific Generators. The Compiler:

- Does **not** execute projects (Runtime does)
- Does **not** modify projects (user does)
- Does **not** install dependencies (Environment does)
- Does **not** generate workflow (Agent does)

```
Workflow → Compiler → Generator → Project
```

The Compiler uses a `GeneratorRegistry` that maps targets to generators. New targets are added by registering a new generator — no core code changes needed.

## Consequences

- **Positive:** Adding a new target = adding a new Generator
- **Positive:** Clear separation of concerns (Compiler vs Runtime vs Environment)
- **Positive:** Dry-run mode for preview without file writes
- **Positive:** Event-based progress reporting for UI
- **Negative:** More abstraction layers to understand
- **Negative:** Slight overhead for simple compilations
- **Negative:** Requires adapter pattern between public and internal Generator interfaces
