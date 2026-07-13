# ADR-006: Agent as Workflow Builder only

**Status:** Accepted

## Context

V1 Agent tried to do everything: generate workflows, write code, fix bugs, analyze errors, and answer general questions. This made the Agent unreliable and hard to test — it hallucinated code, produced inconsistent results, and had unclear boundaries. The open-ended scope made it impossible to guarantee quality.

## Decision

The V2 Agent is strictly a **Workflow Builder**. Its sole responsibility is producing `workflow.json`. It:

- Only produces `workflow.json` (the DAG definition with nodes, edges, configs)
- Does **no code generation** (Compiler does that)
- Does **no code modification** (user does that via IDE)
- Does **no error analysis** (Diagnostic engine does that)
- Plans with LLM or rule-based fallback
- Executes discrete tools: `create_workflow`, `connect_nodes`, `fill_config`, `validate`

### Processing Flow

```
User message → Planner (LLM + Rules) → Action Plan → Executor (Tools) → workflow.json
```

## Consequences

- **Positive:** Agent is focused, reliable, and testable
- **Positive:** Clear boundary — Agent produces workflow, Compiler produces code
- **Positive:** Easy to validate output — workflow.json has a defined schema
- **Positive:** Rule-based fallback works without LLM
- **Positive:** Deterministic tool execution
- **Negative:** Can't auto-fix generated code (by design — safety)
- **Negative:** Users must understand workflow concepts (nodes, edges, ports)
- **Negative:** Less "magical" experience for end users
