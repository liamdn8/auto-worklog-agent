#!/usr/bin/env bash
set -euo pipefail

# Build fully static awagent binary with no external dependencies
ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
GO_CMD=${GO:-go}

if ! command -v "${GO_CMD}" >/dev/null 2>&1; then
  echo "go toolchain not found" >&2
  exit 1
fi

mkdir -p "${ROOT_DIR}/deploy"

echo "==> Building fully static awagent binary"
echo "    Platform: linux/amd64"
echo "    CGO: disabled (no libc dependency)"
echo "    Stripping: enabled (reduce size)"
echo ""

# Build fully static binary
# - CGO_ENABLED=0: Disable CGO to avoid GLIBC dependency
# - GOOS=linux: Target Linux
# - GOARCH=amd64: Target x86_64 architecture
# - -ldflags="-s -w": Strip debug info to reduce size
#   -s: Omit symbol table
#   -w: Omit DWARF debug info
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  "${GO_CMD}" build \
  -ldflags="-s -w" \
  -o "${ROOT_DIR}/deploy/awagent" \
  ./cmd/awagent

echo ""
echo "✓ Build complete: ${ROOT_DIR}/deploy/awagent"

# Show file size
if command -v du >/dev/null 2>&1; then
  SIZE=$(du -h "${ROOT_DIR}/deploy/awagent" | cut -f1)
  echo "  Size: ${SIZE}"
fi

# Verify it's statically linked
if command -v file >/dev/null 2>&1; then
  echo "  Info: $(file "${ROOT_DIR}/deploy/awagent")"
fi

# Check for dynamic dependencies (should be none)
if command -v ldd >/dev/null 2>&1; then
  echo ""
  echo "Checking for dynamic dependencies:"
  LDD_OUTPUT=$(ldd "${ROOT_DIR}/deploy/awagent" 2>&1 || true)
  if echo "$LDD_OUTPUT" | grep -q "not a dynamic executable"; then
    echo "  ✓ Fully static binary - no external dependencies"
  elif echo "$LDD_OUTPUT" | grep -q "statically linked"; then
    echo "  ✓ Fully static binary - no external dependencies"
  else
    echo "  ⚠ Warning: Binary has dynamic dependencies:"
    echo "$LDD_OUTPUT"
  fi
fi

echo ""
echo "Binary is ready to deploy on any Linux x86_64 system"
