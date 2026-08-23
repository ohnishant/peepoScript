#!/usr/bin/env bash
# Builds everything the playground page needs into web/.
# Generated artifacts are gitignored; CI runs this before deploying.
set -euo pipefail
cd "$(dirname "$0")/.."

go run ./cmd/emotefetch -out web/emotes
GOOS=js GOARCH=wasm go build -o web/peepo.wasm ./cmd/wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/
cp assets/peepo.png web/

echo "web/ is ready - serve it statically, e.g.:"
echo "  cd web && python3 -m http.server 8080"
