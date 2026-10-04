#!/bin/sh
# Builds Recuerdo-<version>-<arch>.AppImage with Qt 5 bundled.
#   packaging/linux/build-appimage.sh <recuerdo binary> <version> [outdir]
# Needs Qt 5 development files (qmake); QMAKE may point at qmake.
set -eu
bin=$1
version=$2
out=${3:-dist}
arch=$(uname -m)
root=$(cd "$(dirname "$0")/../.." && pwd)
tools=$(mktemp -d)
appdir=$(mktemp -d)/AppDir

mkdir -p "$appdir/usr/bin" "$appdir/usr/share/recuerdo"
install -m755 "$bin" "$appdir/usr/bin/recuerdo"
cp -r "$root/data" "$appdir/usr/share/recuerdo/"
install -Dm644 "$root/LICENSE" "$appdir/usr/share/licenses/recuerdo/LICENSE"

for t in linuxdeploy linuxdeploy-plugin-qt; do
	curl -fsSL -o "$tools/$t" "https://github.com/linuxdeploy/$t/releases/download/continuous/$t-$arch.AppImage"
	chmod +x "$tools/$t"
done

# linuxdeploy's bundled strip is too old for newer distributions' libraries
export PATH="$tools:$PATH" APPIMAGE_EXTRACT_AND_RUN=1 NO_STRIP=1
export QMAKE=${QMAKE:-$(command -v qmake-qt5 || command -v qmake)}
# X11 is bundled by default; add Wayland so it runs natively there too
export EXTRA_PLATFORM_PLUGINS="libqwayland-generic.so;libqwayland-egl.so"
export EXTRA_QT_PLUGINS="wayland-shell-integration;wayland-graphics-integration-client;wayland-decoration-client"
export LDAI_OUTPUT="Recuerdo-$version-$arch.AppImage"
mkdir -p "$out"
(cd "$out" && linuxdeploy --appdir "$appdir" \
	--executable "$appdir/usr/bin/recuerdo" \
	--desktop-file "$root/packaging/recuerdo.desktop" \
	--icon-file "$root/assets/recuerdo.png" \
	--plugin qt --output appimage)
