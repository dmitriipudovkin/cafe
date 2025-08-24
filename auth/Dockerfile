# Build
FROM golang:1.24.5-alpine as builder
WORKDIR /app

# Install deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source and config
COPY . .

# Ensure storage dir exists for SQLite
RUN mkdir -p /app/storage

# Build the service
RUN go build -o /app/auth ./cmd/auth

FROM alpine:latest

# Default config path (override via env if needed)
ENV CONFIG_PATH=/app/config/local.yaml

WORKDIR /app

COPY --from=builder /app/auth .
COPY --from=builder /app/config/ ./config/

EXPOSE 44044
CMD ["/app/auth"]