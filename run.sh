#!/usr/bin/env bash
# run.sh — gitcli 빌드 스크립트
#
# Usage:
#   ./run.sh build [DIR]   Build + install to DIR (default: ./gitcli/)
#   ./run.sh test          Run tests
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO_DIR="$SCRIPT_DIR/gitcli"

case "${1:-build}" in
    build)
        echo "Building gitcli..."
        INSTALL_DIR="${2:-$GO_DIR}"
        mkdir -p "$INSTALL_DIR"
        (cd "$GO_DIR" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$INSTALL_DIR/gitcli" .)
        echo "✅ Installed: $INSTALL_DIR/gitcli ($(uname -m))"
        ls -lh "$INSTALL_DIR/gitcli"
        ;;
    test)
        echo "Running tests..."
        (cd "$GO_DIR" && go test -v -count=1 ./...)
        ;;
    -h|--help|help)
        echo "gitcli — Local git timeline CLI"
        echo ""
        echo "Usage:"
        echo "  ./run.sh build [DIR]   Build + install (default: ./gitcli/)"
        echo "  ./run.sh test          Run tests"
        ;;
    *)
        echo "Unknown: $1" >&2; exit 1 ;;
esac
