# syntax=docker/dockerfile:1

# Stage 1: Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app
COPY ui/package.json ui/package-lock.json ./
RUN npm ci
COPY ui/ .
RUN npm run build

# Stage 2: Build backend (with embedded frontend)
FROM golang:1.26-alpine AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/dist ui/dist
RUN go build -ldflags "-s -w" -o k8svision ./main.go

# Stage 3: Runtime
FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S k8svision && adduser -S k8svision -G k8svision
COPY --from=builder /workspace/k8svision /app/k8svision
EXPOSE 8080
USER k8svision
ENTRYPOINT ["/app/k8svision"]
