# syntax=docker/dockerfile:1

# ------------------------------
# Builder stage
# ------------------------------
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

# Install build deps
RUN apk add --no-cache git ca-certificates tzdata build-base

# Enable Go modules and better caching
ENV CGO_ENABLED=0 \
    GO111MODULE=on

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN go build -ldflags "-s -w" -o /workspace/k8svision ./main.go


# ------------------------------
# Runtime stage
# ------------------------------
FROM alpine:3.20 AS runner

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S k8svision && adduser -S k8svision -G k8svision

# Copy binary
COPY --from=builder /workspace/k8svision /app/k8svision

# Copy default config if present (can be overridden by mounting)
COPY --from=builder /workspace/config.yaml /app/config.yaml

# Environment (can be overridden at runtime)
ENV K8SVISION_LOG_LEVEL=info \
    K8SVISION_JWT_SECRET="k8svision-default-jwt-secret-key-32-chars" \
    K8SVISION_AUTH_USERNAME=admin \
    K8SVISION_AUTH_PASSWORD="" \
    K8SVISION_SERVER_HOST=0.0.0.0 \
    K8SVISION_SERVER_PORT=8080

EXPOSE 8080

USER k8svision

# Prefer explicit config path when provided; otherwise app reads defaults/env
ENTRYPOINT ["/app/k8svision"]
CMD ["-config", "/app/config.yaml"]


