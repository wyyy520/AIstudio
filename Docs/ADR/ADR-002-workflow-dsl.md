# ADR-002: Workflow DSL Design

**Status:** Accepted
**Date:** 2026-07-12
**Author:** Chief Software Architect

## Context

V1 workflow types mixed declaration data with runtime state. The `Node` struct
contained `Runtime *NodeRuntime` which included `Status`, `Progress`, `Logs`,
and `Error` fields. This violated the principle that Workflow is the Single
Source of Truth for declaration only.

## Decision

Workflow types contain ONLY declaration data:

- No runtime state (status, progress, logs, errors)
- No execution results
- No UI state (positions are kept for UX, not execution)

### Key Changes from V1

| Aspect | V1 | V2 |
|--------|----|----|
| `Node.Runtime` | Included | **Removed** |
| `Node.Parameters` | `map[string]interface{}` | Renamed to `Config` |
| `Workflow.Target` | Not present | **Added** |
| `Workflow.Variables` | Optional | **Standardized** |

### Schema Version

Workflow JSON includes `schema_version: "2.0.0"` for backward compatibility.

## Consequences

Positive:
- workflow.json is a pure declaration
- Can be version-controlled
- Can be validated independently
- Runtime state is managed separately

Negative:
- Migration needed for existing workflows
- UI needs separate state management

## Related Decisions

- ADR-001: Compiler Architecture
- ADR-003: Runtime Architecture