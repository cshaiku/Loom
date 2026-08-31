#!/usr/bin/env bash
set -euo pipefail

loom_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
sample_root="$loom_root/examples/sampleapp"
output_dir="$sample_root/generated"
xaml_path="${1:-$sample_root/mainwindow.xaml}"

if [[ ! -f "$xaml_path" ]]; then
  echo "loom sample app error: XAML source not found at $xaml_path" >&2
  echo "pass a XAML file path as the first argument, or run this script from a complete loom checkout." >&2
  exit 1
fi

mkdir -p "$output_dir"

go run "$loom_root/cmd/loom" inspect:xaml "$xaml_path" --format json --output "$output_dir/layout.json"
go run "$loom_root/cmd/loom" inspect:ascii "$xaml_path" --output "$output_dir/layout.txt"
go run "$loom_root/cmd/loom" accessibility:audit "$xaml_path" --format json --output "$output_dir/audit.json"
go run "$loom_root/cmd/loom" patterns:transfer "$xaml_path" --from winui3 --to macos --format json --output "$output_dir/transfer.json"
