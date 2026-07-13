# ADR-005: Engine simplified to pure algorithm library

**Status:** Accepted

## Context

V1 had a complex "Engine" that managed third-party software connectors, dataset registries, model registries, and execution orchestration. This created a massive surface area for bugs and required constant updates as third-party APIs changed. The Engine was tightly coupled to specific data sources and model formats.

## Decision

The Engine is simplified to a **pure algorithm library** that:

- Executes workflows via registered node factories (in-process execution)
- Has zero knowledge of third-party software
- Has no connectors to external services
- Delegates all external interactions to the Compiler, Runtime, and Plugin system
- Uses Go interfaces for `ExecutableNode` and `NodeRegistry`

```go
type ExecutableNode interface {
    Execute(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error)
}
```

## Consequences

- **Positive:** Smaller, more maintainable codebase
- **Positive:** No dependency on third-party API stability
- **Positive:** Focus on core execution logic and DAG resolution
- **Positive:** Easy to test — pure function execution
- **Negative:** Some functionality moved to other modules (Compiler, Runtime, Plugin)
- **Negative:** Legacy migrations needed for V1 workflows that relied on Engine connectors
- **Negative:** Users must use generated code for third-party interactions
