#!/usr/bin/env bash
# 三平台打包脚本：Linux / macOS / Windows
set -euo pipefail

cd "$(dirname "$0")/.."

VERSION="${1:-v0.1.0}"
DIST="dist"
rm -rf "$DIST"
mkdir -p "$DIST"

PLATFORMS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
  "windows/arm64"
)

echo "=== Building KA Harness $VERSION ==="

for platform in "${PLATFORMS[@]}"; do
  os="${platform%/*}"
  arch="${platform#*/}"
  bin_name="ka-harness"
  archive_name="ka-harness_${VERSION}_${os}_${arch}"

  echo "  → $os/$arch"

  if [ "$os" = "windows" ]; then
    bin_name="ka-harness.exe"
  fi

  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -ldflags="-s -w -X main.version=$VERSION" \
    -o "$DIST/$bin_name" ./cmd/harness

  if [ "$os" = "windows" ]; then
    (cd "$DIST" && zip -q "${archive_name}.zip" "$bin_name")
    rm "$DIST/$bin_name"
  else
    tar -C "$DIST" -czf "$DIST/${archive_name}.tar.gz" "$bin_name"
    rm "$DIST/$bin_name"
  fi
done

echo ""
echo "=== Packages ready ==="
ls -lh "$DIST/"
