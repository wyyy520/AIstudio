# API Documentation

## Overview

AIStudio V2 exposes a RESTful API (built with Gin) and WebSocket endpoints. All routes are grouped under `/api/v1/` (new) with legacy `/api/` routes for backward compatibility.

## Authentication

### Auth Routes (Public)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/auth/register` | User registration |
| POST | `/api/v1/auth/refresh` | Refresh access token |

### Auth Routes (Authenticated)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/logout` | User logout |

### Request/Response

**Login:**
```json
// POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "securepassword"
}
// Response 200
{
  "token": "eyJhbG...",
  "refreshToken": "eyJhbG...",
  "user": { "id": "usr-1", "name": "User", "email": "user@example.com" }
}
```

### User Profile

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/user/profile` | Get user profile |
| PUT | `/api/user/profile` | Update user profile |
| GET | `/api/user/sessions` | List active sessions |

### API Keys

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/user/apikeys` | List API keys |
| POST | `/api/user/apikeys` | Create API key |
| DELETE | `/api/user/apikeys/:id` | Delete API key |
| GET | `/api/providers` | List LLM providers |

### Quota

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/user/quota` | Get user quotas |
| GET | `/api/user/quota/check` | Check quota availability |
| POST | `/api/admin/quota` | Update quota (admin) |

## Middleware

- **Auth Middleware** — JWT token validation
- **CORS Middleware** — Configurable CORS origins
- **Rate Limit Middleware** — Configurable rate limiting
- **Logger Middleware** — Request logging
- **Recovery Middleware** — Panic recovery

## Projects

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects` | List projects |
| POST | `/api/v1/projects` | Create project |
| GET | `/api/v1/projects/:id` | Get project details |
| DELETE | `/api/v1/projects/:id` | Delete project |

**Create Project:**
```json
// POST /api/v1/projects
{
  "name": "My Project",
  "description": "YOLO training pipeline",
  "target": "python"
}
// Response 201
{
  "id": "prj-uuid",
  "name": "My Project",
  "rootPath": "/data/projects/my_project",
  "target": "python",
  "status": "active"
}
```

## Workflows

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/workflows` | List workflows |
| GET | `/api/v1/workflows/:id` | Get workflow metadata |
| POST | `/api/v1/workflows` | Create workflow metadata |
| DELETE | `/api/v1/workflows/:id` | Delete workflow metadata |
| POST | `/api/v1/workflows/:id/validate` | Validate workflow |

### Project-Based Workflow (file-backed)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/projects/:id/workflow` | Get workflow from project |
| PUT | `/api/v1/projects/:id/workflow` | Update workflow.json |
| POST | `/api/v1/projects/:id/workflow/validate` | Validate workflow.json |
| POST | `/api/v1/projects/:id/workflow/save` | Save workflow.json |
| GET | `/api/v1/projects/:id/workflow/json` | Get raw workflow JSON |

**Update workflow.json:**
```json
// PUT /api/v1/projects/:id/workflow
{
  "workflow": {
    "schema_version": "2.0.0",
    "name": "YOLO Training",
    "target": "python",
    "nodes": [...],
    "edges": [...]
  }
}
// Response 200
{ "status": "saved", "path": "/data/projects/my_project/workflow.json" }
```

## Compiler

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/compiler/targets` | List compilation targets |
| POST | `/api/v1/compiler/compile` | Compile workflow to project |
| POST | `/api/v1/compiler/validate` | Validate workflow |

**Compile:**
```json
// POST /api/v1/compiler/compile
{
  "workflow_id": "wf-uuid",
  "project_id": "prj-uuid",
  "options": {
    "force": false,
    "dry_run": false,
    "variables": { "epochs": "50" }
  }
}
// Response 200
{
  "target": "python",
  "projectRoot": "/data/projects/my_project",
  "entryPoints": ["src/train.py"],
  "files": [
    { "path": "src/train.py", "mode": 420 },
    { "path": "requirements.txt", "mode": 420 }
  ],
  "duration": "1.234s",
  "warnings": []
}
```

## Runtime

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/runtime/execute` | Execute a project |
| POST | `/api/v1/runtime/stop` | Stop running execution |
| GET | `/api/v1/runtime/status/:runId` | Get execution status |
| GET | `/api/v1/runtime/running` | List running executions |

**Execute:**
```json
// POST /api/v1/runtime/execute
{
  "projectId": "prj-uuid",
  "entryPoint": "train.py",
  "args": ["--epochs", "100"],
  "timeout": 3600
}
// Response 200
{
  "runId": "run-uuid",
  "status": "running",
  "startedAt": "2026-07-12T00:00:00Z"
}
```

**Status:**
```json
// GET /api/v1/runtime/status/:runId
{
  "runId": "run-uuid",
  "status": "completed",
  "exitCode": 0,
  "duration": "3542.123s"
}
```

## Plugins

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/plugins` | List all plugins |
| GET | `/api/v1/plugins/:name` | Get plugin details |
| PUT | `/api/v1/plugins/:name/status` | Update plugin status |
| GET | `/api/v1/plugin-nodes` | Get all plugin node types |

## Agent

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/agent/chat` | Chat with agent |
| POST | `/api/v1/agent/generate-workflow` | Generate workflow from NL |

**Chat:**
```json
// POST /api/v1/agent/chat
{
  "message": "Create a YOLO training workflow with 100 epochs",
  "project_id": "prj-uuid"
}
// Response 200
{
  "goal": "create_yolo_training_workflow",
  "explanation": "I'll create a workflow with dataset loading and YOLO training nodes.",
  "plan": [...],
  "steps": [...],
  "status": "completed",
  "summary": "Successfully created YOLO training workflow with 2 nodes."
}
```

## Skills

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/skills` | List available skills |
| POST | `/api/v1/skills/apply` | Apply a skill template |

## Environment

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/environment/status` | Get environment status |
| POST | `/api/v1/environment/detect` | Run environment detection |
| GET | `/api/environment/repair-plan` | Get repair plan |
| POST | `/api/environment/repair` | Execute repairs |
| POST | `/api/environment/install` | Install dependency |
| GET | `/api/environment/logs` | Get logs |
| DELETE | `/api/environment/logs` | Clear logs |

## Logs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/logs` | Query logs |
| GET | `/api/v1/logs/task/:taskId` | Get logs by task |
| GET | `/api/v1/logs/workflow/:workflowId` | Get logs by workflow |
| GET | `/api/v1/logs/stream` | Stream logs |

## MCP

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/mcp/tools` | List MCP tools |
| GET | `/api/mcp/servers` | List MCP servers |
| GET | `/api/mcp/status` | MCP connection status |
| GET | `/api/mcp/config` | Export MCP config |
| POST | `/api/mcp/connect` | Connect to MCP server |
| POST | `/api/mcp/disconnect` | Disconnect MCP server |
| POST | `/api/mcp/call` | Call MCP tool |
| POST | `/api/mcp/servers` | Add MCP server |
| DELETE | `/api/mcp/servers/:name` | Remove MCP server |

## Error Analysis

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/error/analyze` | Analyze error |
| POST | `/api/error/repair` | Repair error |
| GET | `/api/error/analysis/:taskId` | Get error analysis |
| GET | `/api/error/fix/:fixId/status` | Get fix status |

## Settings

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/settings` | Get settings |
| PUT | `/api/settings` | Update settings |
| GET | `/api/settings/engine` | Get engine config |
| PUT | `/api/settings/engine` | Update engine config |
| POST | `/api/settings/engine/test` | Test engine connection |

## Users (Admin)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/users` | List users |
| GET | `/api/users/:id` | Get user |
| POST | `/api/users` | Create user |
| PUT | `/api/users/:id` | Update user |
| DELETE | `/api/users/:id` | Delete user |
| PUT | `/api/users/:id/password` | Change password |

## Tasks

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/tasks` | List tasks |
| GET | `/api/tasks/:id` | Get task |
| POST | `/api/tasks` | Create task |
| PUT | `/api/tasks/:id/cancel` | Cancel task |
| PUT | `/api/tasks/:id/status` | Update task status |
| DELETE | `/api/tasks/:id` | Delete task |
| GET | `/api/tasks/:taskId/logs` | Get task logs |

## WebSocket Events

### Connection

```
GET /api/ws
```

### Server → Client Events

| Event | Description | Data |
|-------|-------------|------|
| `compile.started` | Compilation started | `{workflowId, target, progress: 0}` |
| `compile.progress` | Compilation progress | `{workflowId, target, progress, message}` |
| `compile.completed` | Compilation completed | `{workflowId, target, outputDir, duration}` |
| `compile.failed` | Compilation failed | `{workflowId, target, error}` |
| `runtime.started` | Execution started | `{runId, status, taskId}` |
| `runtime.log` | Log output | `{runId, level, message, source}` |
| `runtime.completed` | Execution completed | `{runId, status, duration}` |
| `runtime.failed` | Execution failed | `{runId, error}` |
| `runtime.progress` | Execution progress | `{runId, progress, message}` |
| `task.created` | Task created | `{taskId, type, status}` |
| `task.completed` | Task completed | `{taskId, status}` |
| `task.progress` | Task progress | `{taskId, progress, message}` |
| `plugin.installed` | Plugin installed | `{pluginId, name, version}` |
| `plugin.error` | Plugin error | `{pluginId, name, error}` |
| `environment.bundle.ready` | Bundle installed | `{bundleName, status}` |

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_REQUEST` | 400 | Malformed request body |
| `VALIDATION_ERROR` | 400 | Workflow validation failed |
| `NOT_FOUND` | 404 | Resource not found |
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `CONFLICT` | 409 | Resource already exists |
| `RATE_LIMITED` | 429 | Too many requests |
| `COMPILATION_FAILED` | 500 | Compilation error |
| `GENERATOR_NOT_FOUND` | 404 | No generator for target |
| `BUNDLE_NOT_FOUND` | 404 | Runtime bundle not found |
| `INSTALLATION_FAILED` | 500 | Bundle installation failed |
| `EXECUTION_FAILED` | 500 | Runtime execution error |
| `ENV_NOT_READY` | 412 | Environment requirements not met |
| `PLUGIN_ERROR` | 500 | Plugin operation failed |

### Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Workflow validation failed for target python",
    "details": {
      "node": "node-train",
      "field": "config.epochs",
      "reason": "must be positive integer"
    }
  }
}
```

## Rate Limiting

- Default: 100 requests/minute per user
- Configurable via `config.Server.RateLimit`
- Headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`
- Response on limit: `429 Too Many Requests`

## Health Check

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Health check |

```json
// Response 200
{
  "status": "ok",
  "version": "2.0.0",
  "uptime": "12345s"
}
```
