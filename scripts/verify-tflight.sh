#!/bin/sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
APP="$ROOT/tflight/dist/TFlight.app"

SWIFT_MODULECACHE_PATH="${SWIFT_MODULECACHE_PATH:-/tmp/tflight-swift-cache}"
export SWIFT_MODULECACHE_PATH
mkdir -p "$SWIFT_MODULECACHE_PATH"

sh "$ROOT/scripts/package-tflight.sh"
test -x "$APP/Contents/MacOS/TFlight"
plutil -lint "$APP/Contents/Info.plist" >/dev/null
npm run validate:content
npm run build
printf '%s\n' "TFlight verification passed."
