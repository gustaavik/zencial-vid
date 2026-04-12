# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build-time version info (pass via --build-arg)
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
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

EXPOSE 6091

ENTRYPOINT ["/app/api"]
