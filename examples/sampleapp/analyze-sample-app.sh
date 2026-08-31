#!/usr/bin/env bash
set -euo pipefail

loom_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
sample_root="$loom_root/examples/sampleapp"
output_dir="$sample_root/generated"
xaml_path="${1:-$sample_root/mainwindow.xaml}"
swiftui_path="${2:-$sample_root/contentview.swift}"
qt_path="${3:-$sample_root/mainwindow.qml}"

if [[ ! -f "$xaml_path" ]]; then
  echo "loom sample app error: XAML source not found at $xaml_path" >&2
  echo "pass a XAML file path as the first argument, or run this script from a complete loom checkout." >&2
  exit 1
fi

if [[ ! -f "$swiftui_path" ]]; then
  echo "loom sample app error: SwiftUI source not found at $swiftui_path" >&2
  echo "pass a SwiftUI file path as the second argument, or run this script from a complete loom checkout." >&2
  exit 1
fi

if [[ ! -f "$qt_path" ]]; then
  echo "loom sample app error: Qt source not found at $qt_path" >&2
  echo "pass a Qt file path as the third argument, or run this script from a complete loom checkout." >&2
  exit 1
fi

mkdir -p "$output_dir"

go run "$loom_root/cmd/loom" inspect:xaml "$xaml_path" --format json --output "$output_dir/layout.json"
go run "$loom_root/cmd/loom" inspect:ascii "$xaml_path" --output "$output_dir/layout.txt"
go run "$loom_root/cmd/loom" accessibility:audit "$xaml_path" --format json --output "$output_dir/audit.json"
go run "$loom_root/cmd/loom" patterns:transfer "$xaml_path" --from winui3 --to macos --format json --output "$output_dir/transfer.json"
go run "$loom_root/cmd/loom" inspect:swiftui "$swiftui_path" --format json --output "$output_dir/swiftui-layout.json"
go run "$loom_root/cmd/loom" patterns:transfer "$swiftui_path" --from swiftui --to windows --format json --output "$output_dir/swiftui-to-windows-transfer.json"
go run "$loom_root/cmd/loom" inspect:qt "$qt_path" --format json --output "$output_dir/qt-layout.json"
go run "$loom_root/cmd/loom" patterns:transfer "$qt_path" --from qt --to windows --format json --output "$output_dir/qt-to-windows-transfer.json"
go run "$loom_root/cmd/loom" inspect:parity "$swiftui_path" --target "$qt_path" --from swiftui --to qt --format json --output "$output_dir/swiftui-qt-parity.json"
