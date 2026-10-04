#!/bin/sh
# Builds Recuerdo's web version into dist/web (or $1): the WebAssembly,
# Go's wasm_exec.js and the page in web/. Serve the folder with any web
# server (WebAssembly needs http(s), not file://).
set -eu
root=$(cd "$(dirname "$0")/.." && pwd)
out=${1:-$root/dist/web}
mkdir -p "$out"
GOOS=js GOARCH=wasm go build -C "$root" -ldflags="-s -w" -o "$out/recuerdo.wasm" ./cmd/recuerdo-web
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$out/"
cp -r "$root"/web/. "$out/"
# the translations, and the list of languages for the page's menu
mkdir -p "$out/translations"
cp "$root"/data/translations/*.po "$out/translations/"
(cd "$root/data/translations" && ls *.po | sed 's/\.po$//; s/^recuerdo-//' | sort -u |
	awk 'BEGIN { printf "[" } { printf "%s\"%s\"", (NR > 1 ? "," : ""), $0 } END { print "]" }') > "$out/languages.json"
echo "web version in $out ($(du -h "$out/recuerdo.wasm" | cut -f1) of WebAssembly)"
