#!/bin/sh
# Builds Recuerdo.app around a recuerdo binary and bundles Qt into it:
#   packaging/macos/build-app.sh <recuerdo binary> <version> <outdir>
# Needs Homebrew's qt@5 (for macdeployqt).
set -eu
bin=$1
version=$2
out=$3
root=$(cd "$(dirname "$0")/../.." && pwd)
app="$out/Recuerdo.app"
rm -rf "$app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
cp "$bin" "$app/Contents/MacOS/recuerdo"
cp -r "$root/data" "$root/config" "$root/LICENSE" "$app/Contents/Resources/"

iconset=$(mktemp -d)/recuerdo.iconset
mkdir -p "$iconset"
for s in 16 32 128 256 512; do
	sips -z $s $s "$root/assets/recuerdo.png" --out "$iconset/icon_${s}x${s}.png" >/dev/null
	d=$((s * 2))
	[ $d -le 512 ] && sips -z $d $d "$root/assets/recuerdo.png" --out "$iconset/icon_${s}x${s}@2x.png" >/dev/null
done
iconutil -c icns "$iconset" -o "$app/Contents/Resources/recuerdo.icns"

cat > "$app/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key><string>Recuerdo</string>
	<key>CFBundleDisplayName</key><string>Recuerdo</string>
	<key>CFBundleIdentifier</key><string>eu.kiefte.Recuerdo</string>
	<key>CFBundleExecutable</key><string>recuerdo</string>
	<key>CFBundleIconFile</key><string>recuerdo</string>
	<key>CFBundlePackageType</key><string>APPL</string>
	<key>CFBundleShortVersionString</key><string>${version%%-*}</string>
	<key>CFBundleVersion</key><string>${version}</string>
	<key>LSMinimumSystemVersion</key><string>11.0</string>
	<key>NSHighResolutionCapable</key><true/>
</dict>
</plist>
PLIST

"$(brew --prefix qt@5)/bin/macdeployqt" "$app"
