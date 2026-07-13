# ADR-008: Monorepo structure

**Status:** Accepted

## Context

V1 had multiple repositories (backend, frontend, generators, plugins, documentation). This created friction:
- Cross-cutting changes required multiple PRs across repos
- Version coordination between repos was manual and error-prone
- New contributors had to find and clone multiple repos
- CI/CD pipelines had to be configured per-repo
- Atomic changes were impossible

## Decision

All AIStudio V2 components live in a **single monorepo** with clear module boundaries:

```
AIStudio/
├── Backend/           ← Go backend server (cmd/main.go + internal/)
├── Frontend/          ← Vue 3 + TypeScript + Vite + Tauri
├── Launcher/          ← Desktop launcher
├── packages/          ← Shared Go packages
│   ├── generators/    ← 8 generator implementations
│   ├── bundles/       ← Runtime bundle specs
│   └── plugins/       ← Plugin schema definitions
├── Plugins/           ← Dynamic plugin discovery (plugin.json files)
├── examples/          ← Example workflow.json files
└── Docs/              ← Documentation (architecture, API, ADRs)
```

Module boundaries are enforced by Go's package system and clear import rules:
- `Backend/internal/` is private to the backend
- `packages/generators/common/` is the public Generator API
- `Plugins/` is for runtime manifests only

## Consequences

- **Positive:** Single `git clone` for everything
- **Positive:** Cross-cutting changes in one PR (atomic commits)
- **Positive:** Consistent CI/CD pipeline for all components
- **Positive:** Easier onboarding for contributors
- **Positive:** Simplified dependency management (one `go.work`, one `package.json`)
- **Positive:** Shared tooling configuration (`.gitignore`, linters)
- **Negative:** Larger clone size (~100MB+)
- **Negative:** Care needed to enforce module boundaries
- **Negative:** Potential for all-or-nothing builds without proper CI gating
- **Negative:** Release tagging requires coordination across modules
