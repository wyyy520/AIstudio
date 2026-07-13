# AIStudio

**AIStudio** is a visual AI engineering platform for designing, compiling, and executing AI pipelines. Build workflows visually, generate production-ready code, and run them locally or in the cloud.

## Architecture

```
  ┌───────────────────────────────────────────────────────────┐
  │              Frontend (Vue3 + Tauri + Vite)               │
  │  Workflow Editor │ AI Chat │ Plugin Store │ Monitoring    │
  └────────────────────────┬──────────────────────────────────┘
                           │ HTTP / WebSocket / Tauri IPC
  ┌────────────────────────▼──────────────────────────────────┐
  │              Backend (Go + Gin + GORM)                    │
  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐ │
  │  │   API    │ │ Workflow │ │ Plugin   │ │ Environment  │ │
  │  │  Gateway │ │  Engine  │ │  Manager │ │   Service    │ │
  │  └────┬─────┘ └────┬─────┘ └────┬─────┘ └──────┬───────┘ │
  │       │            │            │               │         │
  │  ┌────▼────┐ ┌─────▼─────┐ ┌───▼────┐ ┌────────▼───────┐ │
  │  │  Auth   │ │  Task     │ │  MCP   │ │  Error Analysis│ │
  │  │ Service │ │  Queue    │ │Protocol│ │  & Repair      │ │
  │  └─────────┘ └───────────┘ └────────┘ └────────────────┘ │
  │       │            │            │               │         │
  │       └──────┬─────┴──────┬────┘───────────────┘         │
  │          Skill  │  Agent AI  │  Compiler                  │
  │         Templates│ (LLM)    │  (Python/ROS2/STM32/...)   │
  └──────────────────┼──────────┼────────────────────────────┘
                     │          │
  ┌──────────────────▼──────────▼────────────────────────────┐
  │              Python Engine (FastAPI + PyTorch)            │
  │  Vision │ NLP │ Training │ Inference │ Export │ Dataset  │
  └──────────────────────┬───────────────────────────────────┘
                        │
  ┌──────────────────────▼───────────────────────────────────┐
  │         Storage / Database / Runtime / Plugins            │
  │  PostgreSQL │ Redis │ S3 │ Datasets │ Models │ Logs      │
  └──────────────────────────────────────────────────────────┘
```

## Quick Start

### Dependencies

- **Go** 1.22+
- **Node.js** 20+ (for frontend)
- **Python** 3.10+ (for AI engine)
- **Docker** & Docker Compose (optional, for containerized deployment)

### Setup

```bash
# Clone the repository
git clone https://github.com/aistudio/aistudio.git
cd aistudio

# Configure environment
cp .env.example .env
# Edit .env with your settings (JWT_SECRET, database URL, LLM API key, etc.)

# Install dependencies
make setup

# Start development servers
make dev
```

- Frontend: `http://localhost:5173`
- Backend API: `http://localhost:8081`
- Health check: `http://localhost:8081/api/health`

### Docker Deployment

```bash
# Start all services (backend, engine, frontend, nginx, postgres, redis)
docker compose up -d

# View logs
docker compose logs -f

# Stop all services
docker compose down
```

See [docker-compose.yml](docker-compose.yml) for service configuration.

## Key Features

- **Visual Workflow Editor** — Drag-and-drop AI pipeline design
- **Multi-Target Compilation** — Python, ROS2, STM32, Docker
- **AI Agent** — Natural language to workflow generation (LLM-based)
- **Plugin System V2** — Manifest-based plugin discovery with lifecycle management
- **MCP Protocol** — Model Context Protocol support for tool integration
- **Skill Templates** — Reusable workflow building blocks
- **Real-time Monitoring** — WebSocket-based progress and log streaming
- **Environment Management** — Dependency checking, repair, and installation
- **Error Analysis** — Automated error diagnosis and repair suggestions

## Project Structure

```
AIstudio/
├── apps/
│   ├── backend/          # Go backend server (cmd + internal packages)
│   ├── desktop/          # Tauri desktop app (Vue3 + TypeScript)
│   └── engine/           # Python AI execution engine (FastAPI)
├── packages/             # Shared Go libraries
│   ├── common/           # Common types and utilities
│   ├── config/           # Configuration types
│   ├── environment/      # Environment management
│   ├── generator/        # Code generation framework
│   ├── plugin/           # Plugin system types and registry
│   └── runtime/          # Runtime types and interfaces
├── Docs/
│   ├── api/              # REST API endpoint documentation
│   └── ADR/              # Architecture Decision Records
├── deploy/               # Docker, nginx, and deployment configs
├── scripts/              # Build and development scripts
├── tests/
│   ├── e2e/              # End-to-end tests (standalone Go module)
│   └── integration/      # Integration tests
├── .env.example          # Environment variable template
├── docker-compose.yml    # Multi-service Docker deployment
└── Makefile              # Build and test orchestration
```

## Development

| Command           | Description                          |
|-------------------|--------------------------------------|
| `make build`      | Build backend binary + frontend      |
| `make test`       | Run all Go tests (packages + backend + integration) |
| `make lint`       | Run linters (go vet, eslint)         |
| `make dev`        | Start development servers            |
| `make clean`      | Remove build artifacts               |
| `make setup`      | Install Go and Node.js dependencies  |

### Running Tests

```bash
# All tests
make test

# Backend unit tests only
cd apps/backend && go test ./... -v -count=1

# Specific package
cd apps/backend && go test ./internal/workflow/... -v -count=1

# E2E tests (standalone module)
cd tests/e2e && go test ./... -v -count=1
```

### Environment Variables

Copy `.env.example` to `.env` and configure:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Backend listen address | `0.0.0.0` |
| `SERVER_PORT` | Backend port | `8081` |
| `DATABASE_TYPE` | Database backend (`postgres` or `sqlite`) | `postgres` |
| `DATABASE_URL` | Database connection string | `postgresql://aistudio:aistudio@db:5432/aistudio` |
| `JWT_SECRET` | JWT signing key (required) | — |
| `JWT_EXPIRE` | JWT token lifetime | `2h` |
| `ENGINE_HOST` | Python engine host | `0.0.0.0` |
| `ENGINE_PORT` | Python engine port | `8082` |
| `LLM_PROVIDER` | AI Agent LLM provider | `openai` |
| `LLM_API_KEY` | LLM API key | — |
| `LLM_MODEL` | LLM model name | `gpt-4o-mini` |
| `REDIS_URL` | Redis connection (optional) | `redis://redis:6379/0` |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins | — |

## API Routes

Base path: `/api`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check |
| POST | `/api/auth/login` | Login |
| POST | `/api/auth/register` | Register |
| POST | `/api/auth/refresh` | Refresh token |
| POST | `/api/auth/logout` | Logout |
| GET | `/api/user/profile` | Get profile |
| PUT | `/api/user/profile` | Update profile |
| GET | `/api/user/sessions` | List sessions |
| GET/POST/DELETE | `/api/user/apikeys` | API key management |
| GET | `/api/providers` | List AI providers |
| GET | `/api/user/quota` | Get quotas |
| GET/POST/PUT/DELETE | `/api/projects` | Project CRUD |
| GET/POST/PUT/DELETE | `/api/workflows` | Workflow CRUD |
| POST | `/api/workflows/:id/run` | Execute workflow |
| GET | `/api/workflows/nodes` | List workflow node types |
| GET/POST/PUT/DELETE | `/api/tasks` | Task management |
| GET/POST/PUT/DELETE | `/api/plugins` | Plugin management |
| POST | `/api/plugins/install` | Install plugin |
| DELETE | `/api/plugins/:name` | Uninstall plugin |
| POST | `/api/plugins/:name/execute` | Execute plugin |
| POST | `/api/agent/chat` | AI Agent chat |
| POST | `/api/agent/generate-workflow` | Generate workflow from prompt |
| GET/POST | `/api/mcp` | MCP tool and server management |
| GET/POST | `/api/environment` | Environment status and repair |
| POST | `/api/error/analyze` | Error analysis |
| GET/PUT | `/api/settings` | Application settings |
| WS | `/api/ws` | WebSocket for real-time updates |

See [Docs/api/](Docs/api/) for detailed endpoint documentation.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Frontend | Vue 3, TypeScript, Tauri, Vite |
| Backend | Go, Gin, GORM, SQLite/PostgreSQL |
| Engine | Python, FastAPI, PyTorch, Ultralytics |
| Database | PostgreSQL 16 (primary), SQLite (dev/test) |
| Cache | Redis 7 (optional) |
| Protocol | REST, WebSocket, MCP JSON-RPC |
| AI Agent | OpenAI / compatible LLM APIs |

## Deployment

- **Desktop**: Tauri package (Frontend + Backend embedded, local Python Engine)
- **Server**: Docker Compose (backend + engine + frontend + nginx + postgres + redis)
- **Standalone**: Backend binary only, with external postgres/redis

## License

MIT — see [LICENSE](LICENSE) for details.
