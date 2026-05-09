# syntax=docker/dockerfile:1.7

# =====================================================================
# Build stage
# =====================================================================
FROM golang:1.25-bookworm AS builder

WORKDIR /src

# Cache go modules across builds.
COPY go.mod go.sum ./
RUN --mount=type=cache,id=gomodcache,target=/go/pkg/mod \
    --mount=type=cache,id=gobuildcache,target=/root/.cache/go-build \
    go mod download

COPY . .

# CGO is enabled because the SQLite driver requires it. The image still
# works in production with DB_DRIVER=postgres (sqlite is just dead code in
# that path) and we publish a small distroless runtime image so the build
# overhead is acceptable. Strip symbols and disable DWARF to slim the
# binary; -trimpath removes local paths from the binary.
RUN --mount=type=cache,id=gomodcache,target=/go/pkg/mod \
    --mount=type=cache,id=gobuildcache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/lambdavault ./cmd/api

# =====================================================================
# Runtime stage
# =====================================================================
# Distroless cc image ships glibc + libstdc++ (needed by CGO sqlite) and
# runs as the non-root "nonroot" user (uid/gid 65532) by default.
FROM gcr.io/distroless/cc-debian12:nonroot

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /out/lambdavault /app/lambdavault

# Default mount point for the SQLite database when DB_DRIVER=sqlite. In
# production with Postgres this directory is unused.
USER nonroot:nonroot
ENV APP_PORT=8080 \
    DB_PATH=/app/data/lambdavault.db
EXPOSE 8080

ENTRYPOINT ["/app/lambdavault"]
