# Simplified single-stage build using golang:alpine
FROM golang:1.22-alpine
WORKDIR /app

# Install build-base and sqlite-dev
RUN apk add --no-cache build-base sqlite-dev
ENV GOTOOLCHAIN=auto

# Install deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source and config
COPY . .

# Ensure storage dir exists for SQLite
RUN mkdir -p /app/storage

# Build the service
RUN CGO_ENABLED=1 go build -o /app/auth ./cmd/auth

# Default config path (override via env if needed)
ENV CONFIG_PATH=/app/config/local.yaml

EXPOSE 44044
CMD ["/app/auth"]