#!/bin/bash

set -e

PLATFORMS="linux:amd64 linux:arm64 windows:amd64 windows:arm64 darwin:arm64 darwin:amd64"

for PLATFORM in $PLATFORMS; do
    GOOS="${PLATFORM%%:*}"
    GOARCH="${PLATFORM#*:}"

    EXT=""
    if [ "$GOOS" = "windows" ]; then
        EXT=".exe"
    fi

    case "$GOARCH" in
        amd64) OUTPUT_ARCH="x86_64" ;;
        arm64) OUTPUT_ARCH="aarch64" ;;
        *)     OUTPUT_ARCH="$GOARCH" ;;
    esac

    OUTPUT="bin/pomodoro-cli-${GOOS}-${OUTPUT_ARCH}${EXT}"

    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$OUTPUT" .
done
