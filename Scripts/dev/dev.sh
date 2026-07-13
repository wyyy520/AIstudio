#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

echo "=== AIStudio Development Server ==="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down services...${NC}"
    if [ -n "${BACKEND_PID:-}" ]; then
        kill "${BACKEND_PID}" 2>/dev/null || true
        wait "${BACKEND_PID}" 2>/dev/null || true
    fi
    if [ -n "${FRONTEND_PID:-}" ]; then
        kill "${FRONTEND_PID}" 2>/dev/null || true
        wait "${FRONTEND_PID}" 2>/dev/null || true
    fi
    echo -e "${GREEN}All services stopped.${NC}"
    exit 0
}
trap cleanup SIGINT SIGTERM

# Start backend
echo -e "${GREEN}Starting backend...${NC}"
cd "${ROOT_DIR}/Backend"
go mod tidy
go run ./cmd/ &
BACKEND_PID=$!
echo -e "  Backend PID: ${BACKEND_PID}"
echo -e "  Listening on: http://localhost:8081"

# Wait for backend to be ready
echo -n "  Waiting for backend..."
for i in $(seq 1 30); do
    if curl -s http://localhost:8081/api/v1/health > /dev/null 2>&1; then
        echo -e " ${GREEN}ready${NC}"
        break
    fi
    if [ "${i}" -eq 30 ]; then
        echo -e " ${RED}timeout${NC}"
    fi
    sleep 1
done

# Start frontend
echo -e "${GREEN}Starting frontend...${NC}"
cd "${ROOT_DIR}/Frontend"
npm install --silent 2>/dev/null
npm run dev &
FRONTEND_PID=$!
echo -e "  Frontend PID: ${FRONTEND_PID}"
echo -e "  Listening on: http://localhost:5173"

echo ""
echo -e "${GREEN}Both services are running.${NC}"
echo -e "  Frontend: ${YELLOW}http://localhost:5173${NC}"
echo -e "  Backend:  ${YELLOW}http://localhost:8081${NC}"
echo -e "  API docs: ${YELLOW}http://localhost:8081/api/v1/health${NC}"
echo ""
echo "Press Ctrl+C to stop all services."

wait
