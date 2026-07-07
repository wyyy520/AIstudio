# Database Models

Core data models for AIStudio Backend.

## Models

| Model | Table | Description |
|-------|-------|-------------|
| User | `users` | Platform user accounts |
| Project | `projects` | AI projects owned by users |
| Task | `tasks` | Async task records (workflow runs, training, etc.) |
| Plugin | `plugins` | Registered plugins |
| Workflow | `workflows` | Workflow definitions (DAG in JSON) |

## Conventions

- All models use `uint` auto-increment primary keys
- Timestamps (`CreatedAt`, `UpdatedAt`) are auto-managed by GORM
- JSON fields use `snake_case` via GORM serializer or explicit mapping
- Passwords are never exposed in JSON output (`json:"-"`)