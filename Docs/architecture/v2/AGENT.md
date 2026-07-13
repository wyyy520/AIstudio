# Agent System — Workflow Builder

## Overview

The Agent is a **Workflow Builder** — it translates natural language requirements into `workflow.json` definitions. The Agent does **no code generation**. It only produces workflow DAGs. The Compiler handles code generation.

```
User: "Train a YOLO model on my dataset"
    │
    ▼
┌──────────────┐
│   Agent      │
│              │
│  ┌────────┐  │
│  │Planner │──┼──► 1. Analyze intent (LLM + Rules)
│  │(LLM +  │  │     2. Plan tool calls
│  │ Rules) │  │
│  └────────┘  │
│       │      │
│       ▼      │
│  ┌────────┐  │
│  │Executor│──┼──► 3. Execute plan sequentially
│  │(tools) │  │
│  └────────┘  │
│       │      │
│       ▼      │
│  ┌────────┐  │
│  │ Tools  │  │
│  │        │  │   create_workflow("YOLO Training", {
│  │• create│  │     nodes: [...],
│  │  _work- │  │     edges: [...]
│  │  flow   │  │   })
│  │• connect│  │
│  │  _nodes │  │
│  │• fill_  │  └──► workflow.json
│  │  config │       (Single Source of Truth)
│  │• vali-  │
│  │  date   │
│  └────────┘  │
└──────────────┘
```

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                        Agent                                  │
│                                                               │
│  ┌─────────┐   ┌──────────┐   ┌──────────┐   ┌───────────┐  │
│  │Planner  │──►│Executor  │──►│ Tools    │   │ Context   │  │
│  │(LLM +   │   │(step-by- │   │ Registry │   │ Manager   │  │
│  │ Rules)  │   │  step)   │   │          │   │           │  │
│  └────┬────┘   └──────────┘   └────┬─────┘   └───────────┘  │
│       │                            │                         │
│  ┌────▼────┐                  ┌────▼─────┐   ┌───────────┐  │
│  │ Memory  │                  │ LLM      │   │ Prompt    │  │
│  │(session │                  │ Provider │   │ Builder   │  │
│  │ history)│                  │(OpenAI,  │   │           │  │
│  └─────────┘                  │ Claude)  │   └───────────┘  │
│                               └──────────┘                  │
└──────────────────────────────────────────────────────────────┘
```

## Agent Interface

```go
type Agent struct {
    planner  *Planner
    executor *Executor
    memory   *Memory
    context  *ContextManager
    tools    *ToolRegistry
    llm      LLMProvider
}
```

### Processing Flow

```
Process(ctx, message, projectID, userID, plugins, envStatus)
    │
    ├── Phase 1: Plan
    │   ├── Set context (project, user)
    │   ├── Save user message to conversation history
    │   ├── planner.Plan() → ActionPlan
    │   │   ├── Try LLM-based planning
    │   │   └── Fallback to rule-based planning
    │   └── Save plan to conversation
    │
    ├── Phase 2: Execute
    │   └── executor.Execute(plan) → []StepResult
    │
    └── Phase 3: Respond
        └── Return AgentResponse with summary
```

## Planner

The Planner analyzes user requirements and generates an action plan.

### LLM-Based Planning

Sends a system prompt with available tools, plugins, and environment status to the LLM:

```go
systemPrompt = `You are AIStudio Workflow Builder...
Available tools: {{.Tools}}
Available plugins: {{.Plugins}}
Environment status: {{.Environment}}
Respond ONLY with valid JSON.`
```

The LLM returns a JSON action plan:

```json
{
  "goal": "Train YOLO model",
  "explanation": "I'll create a workflow with a dataset node and a YOLO trainer node.",
  "steps": [
    {
      "tool": "check_environment",
      "params": {},
      "reason": "Check Python and GPU availability"
    },
    {
      "tool": "list_plugins",
      "params": {},
      "reason": "Find YOLO plugin capabilities"
    },
    {
      "tool": "create_workflow",
      "params": {
        "name": "YOLO Training Pipeline",
        "workflow": {
          "nodes": [...],
          "edges": [...]
        }
      },
      "reason": "Create the training workflow"
    }
  ]
}
```

### Rule-Based Planning (Fallback)

When no LLM is available, rule-based intent detection:

```go
func (p *Planner) planWithRules(userMessage string, plugins []string) *ActionPlan {
    switch {
    case containsAny(msg, "install", "plugin", "add"):
        return p.planInstall(msg, plugins)
    case containsAny(msg, "check", "environment", "env", "status"):
        return p.planEnvironmentCheck(msg)
    case containsAny(msg, "list", "available", "skill", "template"):
        return p.planListSkills(msg)
    default:
        return p.planCreateWorkflow(msg, plugins)
    }
}
```

## Tools

### Tool Interface

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() []ToolParameter
    Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error)
}
```

### Available Tools

| Tool | Name | Description | Parameters |
|------|------|-------------|------------|
| CheckEnvironment | `check_environment` | Check AI development environment | none |
| ListPlugins | `list_plugins` | List available plugins | none |
| CreateWorkflow | `create_workflow` | Create a workflow from JSON | `name` (string), `workflow` (object) |
| RunWorkflow | `run_workflow` | Execute a workflow by ID | `workflow_id` (string), `parameters` (object) |
| GetTaskStatus | `get_task_status` | Check task status | `task_id` (string) |
| InstallPlugin | `install_plugin` | Install a plugin | `plugin_name` (string) |
| ListSkills | `list_skills` | List workflow templates | none |
| ApplySkill | `apply_skill` | Apply a skill template | `skill_id` (string), `parameters` (object) |

### Tool Registration

Tools are injected externally via function callbacks:

```go
agent := NewAgent(llm, memory)

// Register tools with external implementations
agent.ToolRegistry().Register(&CheckEnvironmentTool{
    CheckFn: func(ctx context.Context) (map[string]interface{}, error) {
        return envManager.Check(), nil
    },
})

agent.ToolRegistry().Register(&CreateWorkflowTool{
    CreateFn: func(ctx context.Context, name string, wfJSON json.RawMessage) (string, error) {
        return workflowService.Create(name, wfJSON)
    },
})
```

## Executor

The Executor runs plan steps sequentially:

```go
func (e *Executor) Execute(ctx context.Context, plan *ActionPlan, dryRun bool) []StepResult {
    for _, step := range plan.Steps {
        tool, ok := e.tools.Get(step.Tool)
        if !ok {
            return error result
        }
        result, err := tool.Execute(ctx, step.Params)
        // Record step result
    }
    return stepResults
}
```

## Memory

The Memory component stores conversation history:

```go
type Memory struct {
    conversations map[string][]ConversationEntry
}

type ConversationEntry struct {
    ProjectID string
    UserID    string
    Role      string    // "user" or "agent"
    Content   string
    Goal      string
    CreatedAt time.Time
}
```

## Context Manager

```go
type ContextManager struct {
    ProjectID  string
    UserID     string
    Goal       string
    Plan       *ActionPlan
    History    []Message
}
```

## Prompt Templates

### System Prompt

```go
const systemPrompt = `You are AIStudio Workflow Builder, an AI assistant that helps users design and configure AI workflows.
You translate natural language requirements into structured workflow.json definitions.

Your capabilities include:
1. Creating workflows with nodes and connections
2. Adding, removing, and connecting workflow nodes
3. Filling in node configuration parameters
4. Validating workflow structure and connectivity
5. Listing available plugins and skills for workflow building
6. Checking the environment for compatibility

Available tools:
{{.Tools}}

Current context:
- Project ID: {{.ProjectID}}
- Available plugins: {{.Plugins}}
- Environment status: {{.Environment}}

Respond ONLY with valid JSON, no other text.`
```

### Planner Prompt

```go
const plannerPrompt = `You are an AI workflow planner. Given a user's requirement, determine:
1. What type of workflow this is (vision, NLP, timeseries, simulation, etc.)
2. What plugins are needed for each stage
3. What workflow structure should be generated
4. What parameters need to be configured per node

Response JSON:
{
  "task_type": "vision|nlp|timeseries|simulation|system",
  "required_plugins": ["plugin_name"],
  "workflow_name": "suggested name",
  "workflow_goal": "what the workflow should accomplish",
  "nodes": [...],
  "edges": [...]
}`
```

## How Agent Generates workflow.json

1. User sends natural language request (e.g., "Train a YOLO model on my dataset with 100 epochs")
2. Agent analyzes intent via LLM or rules
3. Agent determines: workflow type, nodes needed, connections, parameters
4. Agent calls `create_workflow` tool with the workflow JSON
5. The tool writes `workflow.json` to the project directory
6. Result is returned to user with workflow ID

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/agent/chat` | Chat with the agent (Plan → Execute → Respond) |
| POST | `/api/agent/plan` | Plan only (no execution) |
| POST | `/api/agent/generate-workflow` | Generate workflow.json from description |
| POST | `/api/agent/generate-and-run` | Generate and immediately compile |
