# AIStudio Project Templates

Centralized template index for the AIStudio code generation system.

## Architecture

Templates are embedded in each generator package using Go's `//go:embed` directive:

```
packages/generators/
├── python/templates/     (14 templates) — Python projects (PEP 621)
├── matlab/templates/     (4 templates)  — MATLAB scripts
├── cpp/templates/        (9 templates)  — C++ projects (CMake)
├── java/templates/       (6 templates)  — Java projects (Maven)
├── stm32/templates/      (8 templates)  — STM32 embedded (CubeMX)
├── ros2/templates/       (10 templates) — ROS2 robotics
├── docker/templates/     (8 templates)  — Docker containers
├── unity/templates/      (7 templates)  — Unity C# scripts
└── ansys/templates/      (7 templates)  — ANSYS APDL scripts
```

## Template Engine

All generators use the unified `common.RenderTemplate()` / `common.RenderToFiles()` 
functions from `packages/generators/common/renderer.go` for consistent rendering.

Template syntax: Go `text/template` with `{{.Variable}}` syntax.

Available template functions:
- `lower`, `upper`, `title` — string casing
- `trim` — whitespace trim
- `join`, `contains`, `hasPrefix`, `hasSuffix` — string utilities

## Design Principles (EngStudio.md §4, §16.5)

1. **Template-driven** — No string concatenation for code generation
2. **Domain-specific** — Each language/domain has its own template set
3. **Variable substitution** — Templates receive structured data from ExecutionPlan
4. **Self-contained** — Generated projects run independently of AIStudio
