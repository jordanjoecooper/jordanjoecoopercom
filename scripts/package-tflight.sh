#!/bin/sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
APP="$ROOT/tflight/dist/TFlight.app"

swift build --package-path "$ROOT/tflight" -c release
BUILD_DIR="$(swift build --package-path "$ROOT/tflight" -c release --show-bin-path)"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
cp "$BUILD_DIR/TFlight" "$APP/Contents/MacOS/TFlight"
cp "$ROOT/tflight/Info.plist" "$APP/Contents/Info.plist"
printf '%s\n' "Packaged $APP"
