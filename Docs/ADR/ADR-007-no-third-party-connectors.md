# ADR-007: No direct third-party software connectors

**Status:** Accepted

## Context

V1 had direct integrations with third-party tools (Hugging Face, Weights & Biases, MLflow, AWS S3, etc.). Each integration:
- Required its own maintenance
- Broke when third-party APIs changed
- Added code that most users didn't need
- Created security surface area
- Bloated the codebase with ~30% integration code

## Decision

AIStudio V2 has **no direct third-party software connectors**. All integrations go through a clear pipeline:

1. **Compiler + Plugin system** — For generating code that uses third-party tools
2. **Runtime** — For executing that generated code (which talks to third-party APIs)
3. **Environment** — For installing third-party packages via pip/system

Generated code can use any third-party library — but AIStudio itself does not integrate with them directly. For example, instead of a built-in Hugging Face connector, the Python generator produces code that imports and uses the `huggingface_hub` library.

## Consequences

- **Positive:** Smaller, more stable backend
- **Positive:** Users can use any third-party tool via generated code
- **Positive:** No API change maintenance burden
- **Positive:** No security risk from direct integrations
- **Positive:** Community can add support for any tool via plugin manifests
- **Negative:** No built-in UI for third-party service management
- **Negative:** Generated code must include boilerplate for API calls
- **Negative:** Users need basic programming knowledge for custom integrations
