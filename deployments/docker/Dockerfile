# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build metadata. SOURCE_COMMIT is auto-injected by Coolify; COMMIT is passed
# by the GitHub Release workflow. Version itself is baked into the source by
# the Release workflow (see .github/workflows/release.yml).
ARG COMMIT
ARG SOURCE_COMMIT
ARG BUILD_TIME
ARG PORT=6091

RUN RESOLVED_COMMIT="${COMMIT:-${SOURCE_COMMIT:-unknown}}" && \
    RESOLVED_BUILD_TIME="${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" && \
    echo "Building with COMMIT=${RESOLVED_COMMIT}, BUILD_TIME=${RESOLVED_BUILD_TIME}" && \
    CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s \
        -X 'github.com/zenfulcode/zencial/internal/infrastructure/buildinfo.Commit=${RESOLVED_COMMIT}' \
        -X 'github.com/zenfulcode/zencial/internal/infrastructure/buildinfo.BuildTime=${RESOLVED_BUILD_TIME}'" \
      -o /app/bin/api \
      ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s" \
      -o /app/bin/migrate \
      ./cmd/migrate

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata ffmpeg

WORKDIR /app

COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/bin/migrate /app/migrate
COPY --from=builder /app/migrations /app/migrations

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE ${PORT}

ENTRYPOINT ["/app/api"]
