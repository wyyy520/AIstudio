# ADR-001: Workflow as file-based Single Source of Truth

**Status:** Accepted

## Context

AIStudio V1 stored workflow state in a database with complex serialization. This made it impossible to version-control workflows, edit them manually, or share them between environments. We needed a format that is:

- Human-readable and editable
- Version-controllable (git-friendly)
- Self-describing (contains all compilation info)
- Decoupled from the UI (works without the frontend)

## Decision

The `workflow.json` file at the project root is the Single Source of Truth. It is a declarative DAG (nodes + edges with typed ports) stored as a JSON file. Key rules:

1. **No runtime state** in workflow.json — no status, progress, logs, or errors
2. **Atomic writes** — write to temp file, then rename
3. **Auto-generated** on project creation, but manually editable
4. **Schema versioning** — `schema_version` field for migrations
5. **File-backed** — always on disk at `<project-root>/workflow.json`

Design principles:
- **Pure Declaration** — Workflow describes WHAT, not HOW
- **Target-Aware** — Declares `target` for compilation
- **Self-Describing** — Contains all info needed for compilation

## Consequences

- **Positive:** Git-friendly, portable, manually editable, simple I/O
- **Positive:** UI can be swapped without data loss
- **Positive:** No database corruption risk for core workflow data
- **Negative:** Large workflows produce large JSON files
- **Negative:** No built-in conflict resolution for concurrent edits
- **Negative:** Requires schema migration mechanism for version upgrades
