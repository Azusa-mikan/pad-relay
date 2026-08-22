#!/bin/bash
set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
PROJECT_DIR=$(cd -- "${SCRIPT_DIR}/../.." && pwd)
DIST_DIR="${PROJECT_DIR}/build/dist"
IMAGE_NAME=${IMAGE_NAME:-padrelay-client-builder}
FORCE_BUILD=${FORCE_BUILD:-0}

mkdir -p "${DIST_DIR}"

if [[ "${FORCE_BUILD}" == "1" ]] || ! docker image inspect "${IMAGE_NAME}" >/dev/null 2>&1; then
    echo "Building Docker image ${IMAGE_NAME}..."
    docker build \
        --file "${SCRIPT_DIR}/Dockerfile" \
        --tag "${IMAGE_NAME}" \
        "${PROJECT_DIR}"
else
    echo "Using existing Docker image ${IMAGE_NAME} (set FORCE_BUILD=1 to rebuild)."
fi

echo "Building static client into ${DIST_DIR}..."
docker run --rm \
    --user 0:0 \
    --env HOST_UID="$(id -u)" \
    --env HOST_GID="$(id -g)" \
    --volume "${PROJECT_DIR}:/src" \
    "${IMAGE_NAME}"

echo "Build complete:"
ls -lh "${DIST_DIR}"
