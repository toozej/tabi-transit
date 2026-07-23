# syntax=docker/dockerfile:1.7
# The base digests are reviewed with the version matrix. Build metadata is
# supplied by CI; no credentials or runtime configuration are baked in.
ARG GO_BUILDER_IMAGE=docker.io/library/golang@sha256:2a0ba12e116687098780d3ce700f9ce3cb340783779646aafbabed748fa6677c
ARG RUNTIME_IMAGE=gcr.io/distroless/static-debian12:nonroot@sha256:f5b485ea962d9bd1186b2f6b3a061191539b905b82ec395de78cbfae51f20e35

FROM ${GO_BUILDER_IMAGE} AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY db ./db
COPY internal ./internal
COPY services ./services
COPY docker ./docker

ARG TARGETOS=linux
ARG TARGETARCH
ARG VERSION=development
ARG REVISION=unknown

ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath \
      -ldflags="-s -w -X main.version=${VERSION} -X main.revision=${REVISION}" \
      -o /out/transit-api ./services/transit-api/cmd/transit-api && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath \
      -ldflags="-s -w" \
      -o /out/realtime-poller ./services/realtime-poller/cmd/realtime-poller && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath \
      -ldflags="-s -w" \
      -o /out/gtfs-importer ./services/gtfs-importer/cmd/gtfs-importer && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath \
      -ldflags="-s -w" \
      -o /out/tabi-healthcheck ./docker/healthcheck && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath \
      -ldflags="-s -w" \
      -o /out/tabi-migrate ./docker/migrate

FROM ${RUNTIME_IMAGE} AS runtime

ARG VERSION=development
ARG REVISION=unknown
ARG CREATED=unknown

LABEL org.opencontainers.image.title="tabi-backend" \
      org.opencontainers.image.description="Tabi transit backend runtime" \
      org.opencontainers.image.source="https://github.com/toozej/tabi-transit" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.created="${CREATED}"

WORKDIR /app
COPY --from=build --chown=nonroot:nonroot /out/ ./
COPY --chown=nonroot:nonroot db/migrations ./migrations

# Distroless supplies the numeric nonroot account. The image has no shell or
# package manager and works with Compose's read-only root filesystem plus /tmp.
USER nonroot:nonroot
CMD ["/app/transit-api"]
