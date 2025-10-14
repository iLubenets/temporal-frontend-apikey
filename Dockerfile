# Global ARG for version (must be before any FROM to use in FROM statements)
ARG TEMPORAL_VERSION=1.28.1

# Build stage - compile custom temporal with API key auth
# Use native platform for builder (fast) and cross-compile for target
FROM --platform=$BUILDPLATFORM golang:1.25 AS builder

# Docker BuildKit provides these automatically for multiarch builds
ARG TARGETOS
ARG TARGETARCH

WORKDIR /

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build static binary (no CGO needed)
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -trimpath -o /out/temporal-server ./src/server

# Reuse official Temporal image
FROM temporalio/server:${TEMPORAL_VERSION} AS runtime

# Replace official binary with our custom one
COPY --from=builder /out/temporal-server /usr/local/bin/temporal-server

# Everything else (entrypoint, user, paths, etc.) inherited from official image
