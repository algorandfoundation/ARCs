#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: $0 <output-dir> [version]" >&2
  exit 1
fi

OUTPUT_DIR="$1"
VERSION="${2:-${GITHUB_REF_NAME:-dev}}"
VERSION="${VERSION#arckit/}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MODULE_ROOT="${REPO_ROOT}/arckit"
BUILD_ROOT="${OUTPUT_DIR}/build"
DIST_ROOT="${OUTPUT_DIR}/dist"

rm -rf "${BUILD_ROOT}" "${DIST_ROOT}"
mkdir -p "${BUILD_ROOT}" "${DIST_ROOT}"

platforms=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
  "windows amd64"
  "windows arm64"
)

for entry in "${platforms[@]}"; do
  read -r goos goarch <<<"${entry}"
  artifact="arckit_${VERSION}_${goos}_${goarch}"
  stage_dir="${BUILD_ROOT}/${artifact}"
  mkdir -p "${stage_dir}"

  binary_name="arckit"
  if [[ "${goos}" == "windows" ]]; then
    binary_name="arckit.exe"
  fi

  (
    cd "${MODULE_ROOT}"
    CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" \
      go build -trimpath -o "${stage_dir}/${binary_name}" ./cmd/arckit
  )

  if [[ "${goos}" == "windows" ]]; then
    (
      cd "${stage_dir}"
      zip -q "${DIST_ROOT}/${artifact}.zip" "${binary_name}"
    )
  else
    (
      cd "${stage_dir}"
      tar -czf "${DIST_ROOT}/${artifact}.tar.gz" "${binary_name}"
    )
  fi
done

(
  cd "${DIST_ROOT}"
  shasum -a 256 ./* > SHA256SUMS
)

echo "${DIST_ROOT}"
