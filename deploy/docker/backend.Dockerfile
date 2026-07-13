# =============================================================================
# AIStudio Backend - Multi-stage Docker Build
# =============================================================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build/apps/backend

COPY apps/backend/go.mod apps/backend/go.sum ./
RUN go mod download

COPY apps/backend/ ./
COPY packages/ /build/packages/

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/aistudio-backend \
    ./cmd/

# -----------------------------------------------------------------------------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /app/aistudio-backend .
COPY Backend/config/ ./config/

RUN mkdir -p /app/storage /app/plugins /app/logs

EXPOSE 8081

ENV AISTUDIO_ENV=production
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8081

CMD ["./aistudio-backend"]