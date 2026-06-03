#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/build"
APP="skills-cli"

echo "=== Building $APP ==="
echo

mkdir -p "$OUT"

echo "[clean] Removing old build artifacts..."
rm -f "$OUT/${APP}" "$OUT/${APP}.exe"
echo "OK"

echo "[1/3] Fetching dependencies..."
cd "$ROOT"
go mod tidy
echo "OK"

echo "[2/3] Building Linux amd64..."
GOOS=linux GOARCH=amd64 go build -o "$OUT/${APP}" "$ROOT/main.go"
echo "OK - $OUT/${APP}"

echo "[3/3] Building Windows amd64..."
GOOS=windows GOARCH=amd64 go build -o "$OUT/${APP}.exe" "$ROOT/main.go"
echo "OK - $OUT/${APP}.exe"

echo
echo "Done. Outputs:"
ls -lh "$OUT"
