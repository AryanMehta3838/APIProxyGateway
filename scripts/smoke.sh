#!/usr/bin/env bash
# Local regression gate: build, vet, and test (no live servers required).
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

go build ./...
go vet ./...
go test ./...

echo "smoke: ok"
