#!/usr/bin/env bash
set -euo pipefail

# Simple helper to compile the awagent binary for local testing.
ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
GO_CMD=${GO:-go}

if ! command -v "${GO_CMD}" >/dev/null 2>&1; then
  echo "go toolchain not found" >&2
  exit 1
fi

mkdir -p "${ROOT_DIR}/bin"

echo "==> building awagent"
"${GO_CMD}" build -o "${ROOT_DIR}/bin/awagent" ./cmd/awagent

echo "Build complete: ${ROOT_DIR}/bin/awagent"
