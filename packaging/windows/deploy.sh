#!/bin/sh
# Collects recuerdo.exe with everything it needs to run into one folder
# (run in an MSYS2 UCRT64 shell):
#   packaging/windows/deploy.sh <recuerdo.exe> <outdir>
set -eu
exe=$1
out=$2
root=$(cd "$(dirname "$0")/../.." && pwd)
mkdir -p "$out"
cp "$exe" "$out/recuerdo.exe"
cp -r "$root/data" "$root/config" "$out/"
cp "$root/LICENSE" "$out/LICENSE.txt"
# Qt DLLs and plugins (platforms, styles, image formats)
windeployqt-qt5 --release --no-translations --no-system-d3d-compiler --no-opengl-sw "$out/recuerdo.exe" ||
	windeployqt --release --no-translations --no-system-d3d-compiler --no-opengl-sw "$out/recuerdo.exe"
# windeployqt leaves out the MinGW runtime and Qt's own dependencies
# (ICU, zlib, libpng, ...): copy every DLL from the MSYS2 prefix that the
# program or a plugin loads, until nothing new turns up.
prefix=$(cygpath -u "$MSYSTEM_PREFIX")
while :; do
	new=$(find "$out" -name '*.exe' -o -name '*.dll' | xargs ldd 2>/dev/null |
		awk -v p="$prefix/" 'index($3, p) == 1 { print $3 }' | sort -u |
		while read -r dll; do [ -e "$out/$(basename "$dll")" ] || echo "$dll"; done)
	[ -z "$new" ] && break
	echo "$new" | xargs -I{} cp {} "$out/"
done
