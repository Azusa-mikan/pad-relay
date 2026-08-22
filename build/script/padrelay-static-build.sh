#!/bin/bash
set -euo pipefail

SOURCE_DIR=${SOURCE_DIR:-/src}
DIST_DIR=${DIST_DIR:-"${SOURCE_DIR}/build/dist"}
OUTPUT=${OUTPUT:-"${DIST_DIR}/pad-relay-client_linux_amd64_static"}

cd "${SOURCE_DIR}"
mkdir -p "${DIST_DIR}"

echo "Running Go tests..."
go test ./...

echo "Building ${OUTPUT}..."
PKG_CONFIG=/usr/local/bin/padrelay-static-pkg-config \
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
go build -trimpath -ldflags "-s -w" -o "${OUTPUT}" ./cmd/gamepad-client

if ldd "${OUTPUT}" 2>/dev/null | grep -q "libSDL2"; then
    echo "SDL2 was dynamically linked" >&2
    exit 1
fi

if [[ -n "${HOST_UID:-}" && -n "${HOST_GID:-}" ]]; then
    chown "${HOST_UID}:${HOST_GID}" "${OUTPUT}"
fi

echo "Built: ${OUTPUT}"
