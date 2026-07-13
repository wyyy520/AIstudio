# ADR-003: Generators produce real standard projects

**Status:** Accepted

## Context

V1 generated proprietary project structures that required AIStudio to run. Users couldn't take a generated project and run it independently. This created lock-in and friction. Generated code was often incomplete (stubs, placeholders) and not practically usable.

## Decision

Every Generator produces a **real, standard, independently runnable project**:

- Python projects: standard package with `requirements.txt`, `setup.py`/`pyproject.toml`
- MATLAB projects: standard `.m` files and toolboxes
- ROS2 projects: standard ROS2 packages with `package.xml`, `setup.py`
- C++ projects: `CMakeLists.txt` based build
- All projects include `workflow.json` for reproducibility
- All projects are immediately runnable (e.g., `python train.py`, `colcon build`)

The project directory structure is standardized across all targets:

```
project/
├── workflow.json
├── src/
├── config/
├── data/
├── models/
├── outputs/
└── tests/
```

## Consequences

- **Positive:** Zero lock-in — projects run without AIStudio
- **Positive:** Users can use standard tooling (`pip install`, `colcon build`, etc.)
- **Positive:** Generated projects are educational (real code, not stubs)
- **Positive:** Projects can be shared and collaborated on with non-AIStudio users
- **Negative:** Generators must know standard project structures deeply
- **Negative:** More template code to write per target
- **Negative:** Maintaining target-specific standards adds complexity
