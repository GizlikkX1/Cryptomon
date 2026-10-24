#!/usr/bin/env sh
# Cross-compiles Cryptomon for Linux, macOS and Windows (amd64 + arm64).
# Pure Go, no CGO: runs from any of these systems with Go 1.25+ installed.
set -eu
cd "$(dirname "$0")/.."

VERSION="${VERSION:-1.0.0}"
OUT="${OUT:-dist}"
LDFLAGS="-s -w -X cryptomon/internal/app.Version=${VERSION}"
TARGETS="${TARGETS:-linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64}"

mkdir -p "$OUT"
for target in $TARGETS; do
  os="${target%/*}"
  arch="${target#*/}"
  ext=""
  [ "$os" = "windows" ] && ext=".exe"
  dir="$OUT/cryptomon-${VERSION}-${os}-${arch}"
  mkdir -p "$dir"
  echo "building $os/$arch"
  for cmd in cryptomon currency-service gateway; do
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags "$LDFLAGS" -o "$dir/${cmd}${ext}" "./cmd/$cmd"
  done
  cp README.md "$dir/"
done
echo "done: $OUT/"
