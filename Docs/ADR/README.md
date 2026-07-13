# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for AIStudio V2. ADRs document significant architectural decisions, their context, and consequences.

## Index

| ADR | Title | Status |
|-----|-------|--------|
| [ADR-001](ADR-001-workflow-single-source-of-truth.md) | Workflow as file-based Single Source of Truth | Accepted |
| [ADR-002](ADR-002-compiler-first-architecture.md) | Compiler-first architecture | Accepted |
| [ADR-003](ADR-003-generators-produce-real-projects.md) | Generators produce real standard projects | Accepted |
| [ADR-004](ADR-004-plugin-system-v2.md) | Plugin system v2 (pure declarations) | Accepted |
| [ADR-005](ADR-005-engine-simplified.md) | Engine simplified to pure algorithm library | Accepted |
| [ADR-006](ADR-006-agent-as-workflow-builder.md) | Agent as Workflow Builder only | Accepted |
| [ADR-007](ADR-007-no-third-party-connectors.md) | No direct third-party software connectors | Accepted |
| [ADR-008](ADR-008-monorepo-structure.md) | Monorepo structure | Accepted |

---

## What is an ADR?

An Architecture Decision Record (ADR) is a short document that captures an important architectural decision made for the AIStudio project. Each ADR includes:

- **Title** — A clear, descriptive name
- **Status** — Proposed, Accepted, Deprecated, or Superseded
- **Context** — The forces at play and the problem being solved
- **Decision** — The chosen approach
- **Consequences** — The trade-offs, benefits, and drawbacks

## How to Write an ADR

1. Create a new file `ADR-<number>-<title>.md`
2. Use the template below
3. Submit for review

### Template

```markdown
# ADR-<N>: <Title>

**Status:** Proposed | Accepted | Deprecated | Superseded

## Context

What is the issue motivating this decision? What forces are at play?

## Decision

What is the change being proposed? How does it work?

## Consequences

- **Positive:** What benefits does this decision bring?
- **Negative:** What trade-offs or costs are incurred?
- **Risks:** What should be monitored?
```

---

## ADR-001: Workflow as file-based Single Source of Truth

**Status:** Accepted

### Context

AIStudio V1 stored workflow state in a database with complex serialization. This made it impossible to version-control workflows, edit them manually, or share them between environments. We needed a format that is:

- Human-readable and editable
- Version-controllable (git-friendly)
- Self-describing (contains all compilation info)
- Decoupled from the UI (works without the frontend)

### Decision

The `workflow.json` file at the project root is the Single Source of Truth. It is a declarative DAG (nodes + edges with typed ports) stored as a JSON file. Key rules:

1. **No runtime state** in workflow.json — no status, progress, logs, or errors
2. **Atomic writes** — write to temp file, then rename
3. **Auto-generated** on project creation, but manually editable
4. **Schema versioning** — `schema_version` field for migrations
5. **File-backed** — always on disk at `<project-root>/workflow.json`

### Consequences

- **Positive:** Git-friendly, portable, manually editable, simple I/O
- **Positive:** UI can be swapped without data loss
- **Negative:** Large workflows produce large JSON files
- **Negative:** No built-in conflict resolution for concurrent edits

---

## ADR-002: Compiler-first architecture

**Status:** Accepted

### Context

V1 used a monolithic executor that combined compilation, execution, and environment management. This made it hard to add new targets (each target required modifying the core). We needed a pluggable compilation pipeline.

### Decision

The Compiler is a multi-stage pipeline (Plan → Validate → Generate → Verify) that delegates to target-specific Generators. The Compiler:

- Does **not** execute projects (Runtime does)
- Does **not** modify projects (user does)
- Does **not** install dependencies (Environment does)
- Does **not** generate workflow (Agent does)

### Consequences

- **Positive:** Adding a new target = adding a new Generator
- **Positive:** Clear separation of concerns (Compiler vs Runtime vs Environment)
- **Positive:** Dry-run mode for preview without file writes
- **Negative:** More abstraction layers to understand
- **Negative:** Slight overhead for simple compilations

---

## ADR-003: Generators produce real standard projects

**Status:** Accepted

### Context

V1 generated proprietary project structures that required AIStudio to run. Users couldn't take a generated project and run it independently. This created lock-in and friction.

### Decision

Every Generator produces a **real, standard, independently runnable project**:

- Python projects: standard package with `requirements.txt`, `setup.py`
- MATLAB projects: standard `.m` files and toolboxes
- ROS2 projects: standard ROS2 packages with `package.xml`
- All projects include `workflow.json` for reproducibility

### Consequences

- **Positive:** Zero lock-in — projects run without AIStudio
- **Positive:** Users can `pip install`, `pipenv`, or `colcon build`
- **Positive:** Generated projects are educational (real code, not stubs)
- **Negative:** Generators must know standard project structures
- **Negative:** More code to generate per target

---

## ADR-004: Plugin system v2 (pure declarations)

**Status:** Accepted

### Context

V1 plugins were Go `.so` files loaded at runtime. This was:
- Platform-specific (only Linux/macOS)
- Hard to debug (crashes in plugins crash the host)
- Impossible to validate before loading
- No schema for configuration

### Decision

V2 plugins are **pure declarative JSON manifests** with **zero executable code**:

- `plugin.json` describes nodes, ports, config schemas (JSON Schema)
- The Generator reads manifests to know what code to generate
- No runtime loading of plugin code
- Auto-discovery from `Plugins/` directory
- Schema validation of node configurations

### Consequences

- **Positive:** Safe — plugins can't crash the host
- **Positive:** Cross-platform — manifests are just JSON
- **Positive:** Validatable — schemas are JSON Schema
- **Positive:** Easy to create — just write a JSON file
- **Negative:** Can't run plugin code directly (must be generated)
- **Negative:** Static analysis only — no dynamic behavior

---

## ADR-005: Engine simplified to pure algorithm library

**Status:** Accepted

### Context

V1 had a complex "Engine" that managed third-party software connectors, datasets, model registries, and execution orchestration. This created a massive surface area for bugs and required constant updates as third-party APIs changed.

### Decision

The Engine is simplified to a **pure algorithm library** that:

- Executes workflows via registered node factories (in-process)
- Has zero knowledge of third-party software
- Has no connectors to external services
- Delegates all external interactions to the Compiler, Runtime, and Plugin system

### Consequences

- **Positive:** Smaller, more maintainable codebase
- **Positive:** No dependency on third-party API stability
- **Positive:** Focus on core execution logic
- **Negative:** Some functionality moved to other modules
- **Negative:** Legacy migrations needed for V1 workflows

---

## ADR-006: Agent as Workflow Builder only

**Status:** Accepted

### Context

V1 Agent tried to do everything: generate workflows, write code, fix bugs, analyze errors, and answer questions. This made the Agent unreliable and hard to test — it hallucinated code and produced inconsistent results.

### Decision

The V2 Agent is strictly a **Workflow Builder**. It:

- Only produces `workflow.json` (the DAG definition)
- Does **no code generation** (Compiler does that)
- Does **no code modification** (user does that)
- Plans with LLM or rule-based fallback
- Executes tools to create nodes, connect edges, fill configs, validate

### Consequences

- **Positive:** Agent is focused and reliable
- **Positive:** Clear boundary — Agent produces workflow, Compiler produces code
- **Positive:** Easier to test — deterministic output (workflow.json)
- **Negative:** Can't auto-fix generated code (by design)
- **Negative:** Users must understand workflow concepts

---

## ADR-007: No direct third-party software connectors

**Status:** Accepted

### Context

V1 had direct integrations with third-party tools (Hugging Face, Weights & Biases, MLflow, etc.). Each integration required maintenance, broke when APIs changed, and bloated the codebase.

### Decision

AIStudio V2 has **no direct third-party software connectors**. All integrations go through:

1. **Compiler + Plugin system** — For generating code that uses third-party tools
2. **Runtime** — For executing that code (which talks to third-party APIs)
3. **Environment** — For installing third-party packages

Generated code can use any third-party library — but AIStudio itself does not integrate with them directly.

### Consequences

- **Positive:** Smaller, more stable backend
- **Positive:** Users can use any third-party tool via generated code
- **Positive:** No API change maintenance burden
- **Negative:** No built-in UI for third-party service management
- **Negative:** Generated code must include boilerplate for API calls

---

## ADR-008: Monorepo structure

**Status:** Accepted

### Context

V1 had multiple repositories (backend, frontend, generators, plugins). This made cross-cutting changes painful (multiple PRs, version coordination). New contributors had to set up multiple repos.

### Decision

All AIStudio V2 components live in a **single monorepo**:

```
AIStudio/
├── apps/
│   ├── backend/    ← Go backend (Gin + GORM)
│   ├── desktop/    ← Vue 3 + Tauri frontend
│   └── engine/     ← Python AI execution engine (FastAPI)
├── packages/       ← Shared Go packages
├── Docs/           ← Documentation and ADRs
├── deploy/         ← Docker, nginx, deployment configs
├── scripts/        ← Build and development scripts
├── tests/          ← E2E and integration tests
└── examples/       ← Example workflows
```

### Consequences

- **Positive:** Single `git clone` for everything
- **Positive:** Cross-cutting changes in one PR
- **Positive:** Consistent CI/CD pipeline
- **Positive:** Easier onboarding for contributors
- **Negative:** Larger clone size
- **Negative:** Care needed to enforce module boundaries
- **Negative:** All-or-nothing builds if not properly gated

---

## How ADRs are Managed

1. **Propose** — Create a new ADR file with status "Proposed"
2. **Review** — Discuss with the team (PR review, async comments)
3. **Accept** — Change status to "Accepted"
4. **Deprecate** — If superseded, mark as "Deprecated" and reference the new ADR

ADR files follow the naming convention: `ADR-<number>-<kebab-case-title>.md`
