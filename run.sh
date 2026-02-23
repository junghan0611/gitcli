#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/gitcli"
CGO_ENABLED=0 go build -o gitcli .
cp gitcli ~/.local/bin/gitcli
echo "gitcli installed to ~/.local/bin/gitcli"
