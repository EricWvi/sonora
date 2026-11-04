# syntax=docker/dockerfile:1

# Both build stages run on the build host; Go cross-compiles for the target.
# The steps follow scripts/build.sh.
FROM --platform=$BUILDPLATFORM node:24-alpine AS web
RUN corepack enable
WORKDIR /src
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY frontend/package.json frontend/
COPY admin/package.json admin/
RUN pnpm install --frozen-lockfile
COPY frontend/ frontend/
COPY admin/ admin/
ARG VERSION=dev
RUN pnpm lint && pnpm lint:admin && VITE_APP_VERSION=$VERSION pnpm build && pnpm build:admin
# Development fixtures must never reach a release build.
RUN ! grep -rl -e "listener@sonora.test" -e "fixtures/baseline" -e "fixture:audio" frontend/dist admin/dist

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# .dockerignore keeps any local web/ build out; only the fresh one is embedded.
COPY --from=web /src/frontend/dist ./web/player
COPY --from=web /src/admin/dist ./web/admin
RUN go test ./...
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
      -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
      -o /out/sonora ./cmd/sonora \
 && mkdir -p /out/data /out/media

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/sonora /app/sonora
COPY --from=build --chown=nonroot:nonroot /out/data /data
COPY --from=build --chown=nonroot:nonroot /out/media /media
VOLUME ["/data", "/media"]
EXPOSE 14560
ENTRYPOINT ["/app/sonora", "-config", "/app/config.yaml"]
