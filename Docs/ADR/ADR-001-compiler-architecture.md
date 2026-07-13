# ADR-001: Compiler Architecture

**Status:** Accepted
**Date:** 2026-07-12
**Author:** Chief Software Architect

## Context

AIStudio V1 had no Compiler layer. Workflow nodes directly drove Python execution
through the engine module. This made it impossible to:
- Generate projects that can run independently of AIStudio
- Support multiple target platforms (Python, MATLAB, ROS2, etc.)
- Maintain a clean separation between "what" (workflow) and "how" (execution)

## Decision

We introduce a Compiler layer between Workflow and Runtime:

```
Workflow → Compiler → Generator → Project → Runtime
```

### Key Design Points

1. **Compiler is the single entry point** for all project generation
2. **Generator interface** is the extension point for new platforms
3. **CompileResult** contains all output files and runtime requirements
4. **RuntimeRequirement** declares what the Runtime needs to execute the project

### Generator Interface

```go
type Generator interface {
    ID() workflow.Target
    Generate(ctx, wf, opts) (*GenerateResult, error)
    Validate(wf) error
    RuntimeRequirement(wf) (*RuntimeRequirement, error)
}
```

### Constraints

- Compiler does NOT execute projects
- Compiler does NOT install dependencies
- Generator does NOT know about Runtime
- Generator creates real files on disk

## Consequences

Positive:
- Clean separation of concerns
- Extensible to any target platform
- Generated projects are real and runnable
- Testable in isolation

Negative:
- Added complexity for simple workflows
- Need to maintain multiple generators

## Alternatives Considered

1. **Keep V1 approach** — Rejected: violates all architecture principles
2. **Single Generator** — Rejected: can't support multiple platforms
3. **Template-only approach** — Rejected: needs programmatic generation

## Related Decisions

- ADR-002: Workflow DSL Design
- ADR-003: Runtime Architecture