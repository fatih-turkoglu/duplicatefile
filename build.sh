#!/usr/bin/env bash
# macOS release: ./build.sh [--universal|--arch arm64|--arch amd64] [--skip-tests]
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")"
source ./env.sh

if [[ "$(uname -s)" != Darwin ]]; then
  echo "build.sh requires macOS. Use build.ps1 on Windows." >&2
  exit 1
fi

target="darwin/$(uname -m)"
[[ "$target" != darwin/x86_64 ]] || target=darwin/amd64
skip_tests=false
while [[ $# -gt 0 ]]; do
  case "$1" in
    --universal) target=darwin/universal; shift ;;
    --arch)
      case "${2:-}" in arm64|amd64) target="darwin/$2" ;; *) echo "--arch requires arm64 or amd64" >&2; exit 1 ;; esac
      shift 2 ;;
    --skip-tests) skip_tests=true; shift ;;
    *) echo "Usage: ./build.sh [--universal|--arch arm64|--arch amd64] [--skip-tests]" >&2; exit 1 ;;
  esac
done
for tool in go node npm wails; do
  command -v "$tool" >/dev/null || { echo "Missing $tool; see README.md (macOS development)." >&2; exit 1; }
done
node -e 'const [major, minor] = process.versions.node.split(".").map(Number); if (major < 22 || (major === 22 && minor < 13)) { console.error("Node.js 22.13+ is required (24 LTS recommended)."); process.exit(1); }'
xcode-select -p >/dev/null
export CGO_ENABLED=1
export MACOSX_DEPLOYMENT_TARGET=12.0
export CGO_CFLAGS="${CGO_CFLAGS:--O2 -g} -mmacosx-version-min=12.0"
export CGO_LDFLAGS="${CGO_LDFLAGS:--O2 -g} -mmacosx-version-min=12.0 -framework UniformTypeIdentifiers"
version="$(node -p 'JSON.parse(require("node:fs").readFileSync("wails.json", "utf8")).info.productVersion')"

# main.go embeds frontend/dist, so build assets before Go checks on a clean clone.
(cd frontend && npm ci --no-audit --no-fund && npm run build)
if [[ "$skip_tests" == false ]]; then
  go vet ./...
  go test ./... -count=1
  (cd frontend && npm test)
fi
wails build -clean -s -platform "$target" -trimpath -ldflags "-X main.version=$version"
app="build/bin/DuplicateFile.app"
mkdir -p "$app/Contents/Resources/lang"
cp lang/*.lng "$app/Contents/Resources/lang/"
plutil -lint "$app/Contents/Info.plist"
# Adding resources invalidates the signature Wails creates, so sign the final bundle.
# An optional Developer ID can be supplied for distribution; otherwise sign locally.
codesign --force --deep --sign "${DUPLICATEFILE_SIGN_IDENTITY:--}" "$app"
codesign --verify --deep --strict "$app"
archive="build/bin/DuplicateFile-$version-${target#darwin/}.zip"
ditto -c -k --sequesterRsrc --keepParent "$app" "$archive"
echo "DuplicateFile $version: $app"
echo "Archive: $archive"
