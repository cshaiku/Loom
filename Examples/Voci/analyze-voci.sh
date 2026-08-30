#!/usr/bin/env bash
set -euo pipefail

voci_root="${1:-/private/SDF/Voci}"
loom_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
output_dir="$loom_root/Examples/Voci/Generated"
xaml_path="${2:-$voci_root/platform/windows-winui/VociWindows/MainWindow.xaml}"

mkdir -p "$output_dir"

go run "$loom_root/cmd/loom" inspect:xaml "$xaml_path" --format json --output "$output_dir/layout.json"
go run "$loom_root/cmd/loom" inspect:ascii "$xaml_path" --output "$output_dir/layout.txt"
go run "$loom_root/cmd/loom" accessibility:audit "$xaml_path" --format json --output "$output_dir/audit.json"
go run "$loom_root/cmd/loom" patterns:transfer "$xaml_path" --from winui3 --to macos --format json --output "$output_dir/transfer.json"
