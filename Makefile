.PHONY: all build test lint clean dev help

SHELL := /usr/bin/env bash
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# Detect OS and architecture
UNAME_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
UNAME_ARCH := $(shell uname -m)
ifeq ($(UNAME_ARCH), x86_64)
	ARCH := amd64
else ifeq ($(UNAME_ARCH), aarch64)
	ARCH := arm64
else
	ARCH := $(UNAME_ARCH)
endif

BINARY_NAME := aistudio-server-$(UNAME_OS)-$(ARCH)
ifeq ($(UNAME_OS), windows)
	BINARY_NAME := $(BINARY_NAME).exe
endif

BIN_DIR := $(ROOT_DIR)/build/bin

# ==============================================================================
# Default target
# ==============================================================================
help:
	@echo "AIStudio Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build       Build backend binary and verify packages"
	@echo "  make test        Run all Go tests"
	@echo "  make lint        Run linters (go vet)"
	@echo "  make dev         Start development servers"
	@echo "  make clean       Remove build artifacts"
	@echo "  make setup       Install dependencies"

# ==============================================================================
# Build
# ==============================================================================
build: build-packages build-backend build-frontend

build-backend:
	@echo "Building backend..."
	@mkdir -p $(BIN_DIR)
	cd $(ROOT_DIR)/apps/backend && go mod tidy && go build -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/...
	@echo "  -> $(BIN_DIR)/$(BINARY_NAME)"

build-packages:
	@echo "Verifying packages compile..."
	@for pkg in $(ROOT_DIR)/packages/*/; do \
		echo "  -> $$(basename $$pkg)"; \
		cd "$$pkg" && go build ./...; \
	done

build-frontend:
	@echo "Building frontend..."
	@if command -v node > /dev/null 2>&1; then \
		cd $(ROOT_DIR)/apps/desktop && npm ci 2>/dev/null || npm install && npm run build; \
		echo "  -> apps/desktop/dist/ built"; \
	else \
		echo "  -> Node.js not found, skipping frontend build"; \
	fi

# ==============================================================================
# Test
# ==============================================================================
test: test-packages test-backend test-integration

test-packages:
	@echo "Running package tests..."
	@for pkg in $(ROOT_DIR)/packages/*/; do \
		echo "  -> $$(basename $$pkg)"; \
		(cd "$$pkg" && go test ./... -v -count=1); \
	done

test-backend:
	@echo "Running backend tests..."
	cd $(ROOT_DIR)/apps/backend && go test ./... -v -count=1

test-integration:
	@echo "Running integration tests..."
	cd $(ROOT_DIR) && go test ./tests/integration/... -v -count=1

# ==============================================================================
# Lint
# ==============================================================================
lint: lint-packages lint-backend lint-frontend

lint-packages:
	@echo "Running go vet on packages..."
	@for pkg in $(ROOT_DIR)/packages/*/; do \
		echo "  -> $$(basename $$pkg)"; \
		(cd "$$pkg" && go vet ./...); \
	done

lint-backend:
	@echo "Running go vet on backend..."
	cd $(ROOT_DIR)/apps/backend && go vet ./...

lint-frontend:
	@echo "Linting frontend..."
	@if command -v node > /dev/null 2>&1; then \
		cd $(ROOT_DIR)/apps/desktop && npm ci 2>/dev/null || npm install && npm run lint --if-present; \
	else \
		echo "  -> Node.js not found, skipping frontend lint"; \
	fi

# ==============================================================================
# Development
# ==============================================================================
dev:
	@$(ROOT_DIR)/scripts/dev/dev.sh

# ==============================================================================
# Clean
# ==============================================================================
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(ROOT_DIR)/build
	@echo "  -> Done"

# ==============================================================================
# Setup (install dependencies)
# ==============================================================================
setup:
	@echo "Installing dependencies..."
	cd $(ROOT_DIR)/apps/backend && go mod tidy
	@if command -v node > /dev/null 2>&1; then \
		cd $(ROOT_DIR)/apps/desktop && npm install; \
	fi
	@echo "  -> Done"
