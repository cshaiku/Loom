#!/usr/bin/env bash
set -euo pipefail

voci_root="${1:-/private/SDF/Voci}"
loom_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
manifest="$loom_root/Examples/Voci/loom.json"
output_dir="$loom_root/Examples/Voci/Generated"

swift run --package-path "$loom_root" loom config:validate "$manifest" --project-root "$voci_root"
swift run --package-path "$loom_root" loom project:build "$manifest" --project-root "$voci_root" --output-dir "$output_dir"
