# AIStudio Backend Configuration

## Overview

The AIStudio Backend uses a layered configuration system with three priority levels:

| Priority | Source | Example |
|----------|--------|---------|
| Highest  | Environment Variables | `SERVER_PORT=9090` |
| Medium   | Configuration Files  | `config/development.yaml` |
| Lowest   | Default Values       | Hardcoded in `internal/config/config.go` |

## Configuration Files

All configuration files are located in the `Backend/config/` directory:

| File | Purpose |
|------|---------|
| `default.yaml` | Base configuration for all environments |
| `development.yaml` | Development-specific overrides (loaded when `AISTUDIO_ENV=development`) |
| `production.yaml` | Production-specific overrides (loaded when `AISTUDIO_ENV=production`) |

The active environment is selected via the `AISTUDIO_ENV` environment variable:

```bash
# Development (default)
export AISTUDIO_ENV=development

# Production
export AISTUDIO_ENV=production
```

## Configuration Structure

```yaml
server:
  port: "8081"        # HTTP server port
  host: "0.0.0.0"     # HTTP server host

database:
  type: "sqlite"       # sqlite | postgres
  url: "aistudio.db"   # connection string

jwt:
  secret: "..."        # JWT signing key
  expire: "24h"        # token TTL

engine:
  address: "localhost"  # AI Engine host
  grpc_port: 50051      # AI Engine gRPC port

plugin:
  directory: "../Plugins"  # plugins directory path

log:
  level: "info"         # debug | info | warn | error

websocket:
  port: "8082"          # WebSocket server port

task:
  num_workers: 4        # concurrent task workers
```

## Environment Variables

Every configuration field can be overridden by an environment variable:

| Variable | Config Field | Example |
|----------|-------------|---------|
| `SERVER_PORT` | `server.port` | `8081` |
| `SERVER_HOST` | `server.host` | `0.0.0.0` |
| `DATABASE_TYPE` | `database.type` | `postgres` |
| `DATABASE_URL` | `database.url` | `postgresql://user:pass@localhost:5432/aistudio` |
| `JWT_SECRET` | `jwt.secret` | `your-256-bit-secret` |
| `JWT_EXPIRE` | `jwt.expire` | `2h` |
| `ENGINE_ADDRESS` | `engine.address` | `10.0.0.1` |
| `ENGINE_GRPC_PORT` | `engine.grpc_port` | `50051` |
| `PLUGIN_DIRECTORY` | `plugin.directory` | `/opt/aistudio/plugins` |
| `LOG_LEVEL` | `log.level` | `debug` |
| `WEBSOCKET_PORT` | `websocket.port` | `8082` |
| `TASK_NUM_WORKERS` | `task.num_workers` | `4` |
| `AISTUDIO_ENV` | Selects environment profile | `production` |

## Usage

### In Go code

```go
import "github.com/aistudio/backend/internal/config"

// Initialize at startup
if err := config.Load(); err != nil {
    log.Fatal(err)
}

// Read configuration
cfg := config.Get()
port := cfg.Server.Port
dbURL := cfg.Database.URL

// Use helper methods
addr := cfg.Server.Addr()        // "0.0.0.0:8081"
dsn := cfg.Database.DSN()        // "aistudio.db"
grpcAddr := cfg.Engine.GrpcAddr() // "localhost:50051"

// Hot-reload at runtime
if err := config.Reload(); err != nil {
    log.Printf("config reload failed: %v", err)
}
```

### Running with specific configuration

```bash
# Default (development with SQLite)
go run cmd/main.go

# Production with Postgres
AISTUDIO_ENV=production DATABASE_URL=postgresql://... go run cmd/main.go

# Custom config file
AISTUDIO_CONFIG=/path/to/custom.yaml go run cmd/main.go
```

## Security Notes

- **Never commit secrets** to version control. Use environment variables for sensitive values.
- In production, always set `JWT_SECRET` via environment variable.
- Use a strong, randomly-generated JWT secret (at least 256 bits).
- Database credentials should be injected via `DATABASE_URL` environment variable, not stored in config files.