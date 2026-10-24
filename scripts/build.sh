#!/usr/bin/env bash
# Builds the release binary with the player and admin console embedded:
#   pnpm install → lint and build both apps → copy dist → go test → go build
#
# Environment:
#   VERSION, COMMIT   injected into the binary (default: from git)
#   OUT               output path (default: release/sonora)
#   GOOS, GOARCH      honoured by go build for cross-compiling
set -euo pipefail
cd "$(dirname "$0")/.."

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)}"
OUT="${OUT:-release/sonora}"

echo "==> web apps"
pnpm install --frozen-lockfile
pnpm build
pnpm build:admin

echo "==> embed"
for app in player:frontend admin:admin; do
  dest="backend/web/${app%%:*}"
  find "$dest" -mindepth 1 ! -name .gitkeep -delete
  cp -R "${app##*:}/dist/." "$dest/"
done

echo "==> go test"
(cd backend && go test ./...)

echo "==> go build ${VERSION} (${COMMIT})"
mkdir -p "$(dirname "$OUT")"
(cd backend && CGO_ENABLED=0 go build -trimpath \
  -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
  -o "../${OUT}" ./cmd/sonora)

echo "==> ${OUT}"
