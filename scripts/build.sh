#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SRC_DIR="$ROOT_DIR/src"

TARGET_OS="${TARGET_OS:-$(go env GOOS)}"
TARGET_ARCH="${TARGET_ARCH:-$(go env GOARCH)}"
VERSION="${VERSION:-$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null || echo dev)}"

OUTPUT_DIR="${OUTPUT_DIR:-$ROOT_DIR}"
OUTPUT_NAME="${OUTPUT_NAME:-gitmirror}"

if [[ "$TARGET_OS" == "windows" ]]; then
	OUTPUT_FILE="$OUTPUT_DIR/${OUTPUT_NAME}.exe"
else
	OUTPUT_FILE="$OUTPUT_DIR/$OUTPUT_NAME"
fi

echo "Building $OUTPUT_NAME for $TARGET_OS/$TARGET_ARCH (version: $VERSION)..."

(
	cd "$SRC_DIR"
	CGO_ENABLED=0 GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" \
		go build -ldflags "-X main.version=$VERSION -X main.buildOS=$TARGET_OS -X main.buildArch=$TARGET_ARCH -X main.buildTime=$(date +%Y-%m-%dT%H:%M:%S)" -o "$OUTPUT_FILE" ./cli
)

echo "Build completed successfully: $OUTPUT_FILE"
