# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build-time version info (pass via --build-arg, auto-detected otherwise)
ARG VERSION
ARG GIT_COMMIT
ARG BUILD_TIME
ARG PORT=6091

# Build the binary
RUN VERSION="${VERSION:-$(git describe --tags --abbrev=0 2>/dev/null || echo '0.0.0')}" && \
    GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}" && \
    BUILD_TIME="${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" && \
    echo "Building with VERSION=${VERSION}, GIT_COMMIT=${GIT_COMMIT}, BUILD_TIME=${BUILD_TIME}" && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s \
      -X 'github.com/zenfulcode/zencial/internal/infrastructure/buildinfo.Version=${VERSION}' \
      -X 'github.com/zenfulcode/zencial/internal/infrastructure/buildinfo.Commit=${GIT_COMMIT}' \
      -X 'github.com/zenfulcode/zencial/internal/infrastructure/buildinfo.BuildTime=${BUILD_TIME}'" \
    -o /app/bin/api \
    ./cmd/api

# Build the migrate binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/migrate \
    ./cmd/migrate

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata ffmpeg

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/bin/migrate /app/migrate

# Copy migrations
COPY --from=builder /app/migrations /app/migrations

# Non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE ${PORT}

ENTRYPOINT ["/app/api"]
