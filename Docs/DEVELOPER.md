# Developer Guide

## Prerequisites

| Tool      | Version   | Notes                              |
|-----------|-----------|------------------------------------|
| Go        | 1.25+     | Backend and packages               |
| Node.js   | 20+       | Frontend (Vue 3 + TypeScript)      |
| Python    | 3.9+      | AI Engine                          |
| Make      | any       | Build orchestration                |

Optional:
- CUDA 11.8+ (GPU acceleration)
- Tauri CLI (desktop packaging)

## Project Structure

```
AIstudio/
├── apps/
│   ├── backend/          # Go backend (restructured)
│   └── desktop/          # Vue3 + Tauri desktop app
├── packages/             # Shared Go packages
├── Backend/              # Go backend server (legacy)
├── Frontend/             # Vue3 frontend
├── Engine/               # Python AI execution engine
├── docs/                 # Developer documentation
├── Docs/                 # Architecture & design docs
├── Scripts/              # Build and CI scripts
├── Tests/                # Integration and E2E tests
└── Config/               # Global configuration
```

## Build Instructions

### Quick Build

```bash
make build
```

This builds:
- Backend binary (`build/bin/aistudio-server-<os>-<arch>`)
- Frontend static assets (`Frontend/dist/`)
- Verifies all Go packages compile

### Backend Only

```bash
cd Backend
go mod tidy
go build -o aistudio-server ./cmd/...
```

### Frontend Only

```bash
cd Frontend
npm install
npm run build
```

## Running Locally

### Development Servers

```bash
make dev
```

Or manually:

```bash
# Terminal 1: Backend
cd Backend
go run ./cmd/

# Terminal 2: Frontend
cd Frontend
npm run dev
```

### Configuration

The backend loads configuration from `Config/` directory:

| File               | Purpose                          |
|--------------------|----------------------------------|
| `app.yaml`         | Application-level settings       |
| `backend.yaml`     | Server, database, JWT settings   |
| `engine.yaml`      | Python engine configuration      |
| `plugin.yaml`      | Plugin system settings           |

Environment variables override config values:

| Variable          | Default      | Description               |
|-------------------|--------------|---------------------------|
| `SERVER_PORT`     | `8081`       | HTTP server port          |
| `DATABASE_TYPE`   | `sqlite`     | Database backend          |
| `AISTUDIO_ENV`    | `production` | Environment mode          |

## Testing

### Run All Tests

```bash
make test
```

### Run Package Tests

```bash
make test-packages
```

### Run Backend Tests

```bash
make test-backend
```

### Run Integration Tests

```bash
make test-integration
```

### Test Requirements

- Tests must be **deterministic**
- Use `t.TempDir()` for temporary directories
- Avoid external dependencies in unit tests
- Integration tests may require Python installed

## Code Style

### Go

- Follow standard `gofmt` formatting
- Use `go vet` before committing
- Package names: lowercase, single word
- Error handling: always check and wrap errors with `%w`
- Use `context.Context` as first parameter in public APIs

### TypeScript / Vue

- Follow ESLint and Prettier config in `Frontend/`
- Use Composition API and `<script setup>` for Vue components
- Pinia for state management
- TypeScript strict mode

### Python

- Follow PEP 8
- Type hints required for public APIs
- Google-style docstrings

## Pull Request Guidelines

1. **Branch naming**: `feature/description`, `fix/description`, `docs/description`
2. **Commit messages**: Conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`)
3. **Before requesting review**:
   - Run `make lint` and fix all warnings
   - Run `make test` and ensure all tests pass
   - Update documentation if adding/changing features
4. **Review requirements**:
   - At least one approval from team member
   - All CI checks must pass
   - No merge conflicts
5. **Merge strategy**: Squash merge to `develop`

## Release Process

1. Create a release branch from `develop`
2. Update version in `Backend/internal/config/config.go` and `Frontend/package.json`
3. Update `CHANGELOG.md`
4. Create PR from release branch to `main`
5. Tag with `v<major>.<minor>.<patch>` after merge
6. CI will build binaries and create GitHub Release
