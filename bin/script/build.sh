#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$ROOT/build"

# --- version info ---
VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")"
LDFLAGS="-X skills/bin/internal/version.Version=${VERSION#v} -X skills/bin/internal/version.Commit=${COMMIT} -X skills/bin/internal/version.Date=${DATE}"

echo "=== Building skills ==="
echo "  Version: $VERSION"
echo "  Commit:  $COMMIT"
echo "  Date:    $DATE"
echo

mkdir -p "$OUT"

echo "[clean] Removing old build artifacts..."
rm -rf "$OUT"
echo "OK"

echo "[fetch] go mod tidy..."
cd "$ROOT"
go mod tidy
echo "OK"

# Build each subcommand (each directory with a main.go).
for dir in "$ROOT"/*/; do
	name="$(basename "$dir")"
	if [ ! -f "$dir/main.go" ]; then
		continue
	fi

	echo "[build] $name (windows amd64)..."
	GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$OUT/${name}.exe" "./${name}/"
	echo "  OK - $OUT/${name}.exe"

	echo "[build] $name (linux amd64)..."
	GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$OUT/${name}" "./${name}/"
	echo "  OK - $OUT/${name}"
done

echo
echo "Done. Outputs:"
ls -lh "$OUT"
