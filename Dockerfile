FROM --platform=$BUILDPLATFORM node:24-bookworm-slim AS frontend

ENV BUILD_TARGET=docker
WORKDIR /client

RUN npm install --global --force corepack@0.34.6 && corepack enable
COPY client/package.json client/yarn.lock client/.yarnrc.yml ./
RUN yarn install --immutable
COPY client/ ./
COPY internal/version/VERSION /internal/version/VERSION
ARG KOITO_VERSION=0.3.2-SUM-RC2
ENV VITE_KOITO_VERSION=$KOITO_VERSION
RUN yarn build && \
    find build/client -type f \( -name "*.js" -o -name "*.css" -o -name "*.html" -o -name "*.svg" \) -exec gzip -kf -9 {} \;

FROM golang:1.25-bookworm AS backend

ENV CGO_ENABLED=1
WORKDIR /app
RUN apt-get update && \
    apt-get install -y --no-install-recommends libvips-dev pkg-config && \
    rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY engine/ ./engine/
COPY imagecache/ ./imagecache/
COPY internal/ ./internal/
COPY queue/ ./queue/
COPY db/ ./db/
ARG KOITO_VERSION=0.3.2-SUM-RC2
RUN go build -trimpath -ldflags "-s -w -X main.Version=$KOITO_VERSION" -o /koito ./cmd/api

FROM debian:bookworm-slim AS final

WORKDIR /app
RUN apt-get update && \
    apt-get install -y --no-install-recommends libvips42 curl ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/*
COPY --from=backend /koito ./app
COPY --from=frontend /client/build/client ./client/build/client
COPY client/public/ ./client/public/
COPY assets/ ./assets/
ARG KOITO_VERSION=0.3.2-SUM-RC2
LABEL org.opencontainers.image.title="Koito SUM" \
      org.opencontainers.image.description="ListenBrainz-compatible scrobbling with simple public listening profiles" \
      org.opencontainers.image.source="https://github.com/ebrgr/Koito-SUM" \
      org.opencontainers.image.version=$KOITO_VERSION
ENV KOITO_CONFIG_DIR=/etc/koito \
    KOITO_BIND_ADDR=0.0.0.0 \
    KOITO_LISTEN_PORT=4110
EXPOSE 4110
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
    CMD curl --fail --silent http://127.0.0.1:4110/apis/web/v1/health || exit 1
ENTRYPOINT ["./app"]
