# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build-time version info (pass via --build-arg, auto-detected otherwise).
# SOURCE_COMMIT and COOLIFY_BRANCH are injected automatically by Coolify.
ARG VERSION
ARG GIT_COMMIT
ARG BUILD_TIME
ARG SOURCE_COMMIT
ARG COOLIFY_BRANCH
ARG PORT=6091

# git refuses to operate on a repo it considers "dubious ownership" (e.g. when
# COPY changes the owning UID). Mark the workdir safe so describe/rev-parse work.
RUN git config --global --add safe.directory /app

# Build the binary
RUN COMMIT_FALLBACK="${GIT_COMMIT:-${SOURCE_COMMIT}}" && \
    SHORT_COMMIT_FALLBACK="$(printf '%s' "${COMMIT_FALLBACK}" | cut -c1-7)" && \
    if [ -n "${VERSION}" ]; then \
      RESOLVED_VERSION="${VERSION}"; \
    elif DESCRIBED=$(git describe --tags --always 2>/dev/null) && [ -n "${DESCRIBED}" ] && printf '%s' "${DESCRIBED}" | grep -qE '^v[0-9]'; then \
      RESOLVED_VERSION="${DESCRIBED}"; \
    elif printf '%s' "${COOLIFY_BRANCH}" | grep -qE '^v[0-9]'; then \
      RESOLVED_VERSION="${COOLIFY_BRANCH}"; \
    else \
      SHORT_SHA="$(git rev-parse --short HEAD 2>/dev/null || printf '%s' "${SHORT_COMMIT_FALLBACK}")"; \
      RESOLVED_VERSION="0.0.0-${SHORT_SHA:-unknown}"; \
    fi && \
    RESOLVED_COMMIT="${GIT_COMMIT:-$(git rev-parse HEAD 2>/dev/null || printf '%s' "${SOURCE_COMMIT:-unknown}")}" && \
    RESOLVED_BUILD_TIME="${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" && \
    echo "Building with VERSION=${RESOLVED_VERSION}, GIT_COMMIT=${RESOLVED_COMMIT}, BUILD_TIME=${RESOLVED_BUILD_TIME}" && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s \
      -X 'github.com/zenfulcode/zencial/internal/infrastructure/buildinfo.Version=${RESOLVED_VERSION}' \
      -X 'github.com/zenfulcode/zencial/internal/infrastructure/buildinfo.Commit=${RESOLVED_COMMIT}' \
      -X 'github.com/zenfulcode/zencial/internal/infrastructure/buildinfo.BuildTime=${RESOLVED_BUILD_TIME}'" \
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
