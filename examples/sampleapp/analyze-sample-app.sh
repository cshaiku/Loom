#!/usr/bin/env bash
set -euo pipefail

loom_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
sample_root="$loom_root/examples/sampleapp"
output_dir="$sample_root/generated"
xaml_path="${1:-$sample_root/mainwindow.xaml}"
swiftui_path="${2:-$sample_root/contentview.swift}"
qt_path="${3:-$sample_root/mainwindow.qml}"
overwrite_flag=""
allow_overwrite=false

if [[ "${1:-}" == "--overwrite" ]]; then
  overwrite_flag="--overwrite"
  allow_overwrite=true
  xaml_path="$sample_root/mainwindow.xaml"
  swiftui_path="$sample_root/contentview.swift"
  qt_path="$sample_root/mainwindow.qml"
elif [[ "${4:-}" == "--overwrite" ]]; then
  overwrite_flag="--overwrite"
  allow_overwrite=true
fi

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
loom_bin="$output_dir/loom-sample-cli"
if [[ "${OS:-}" == "Windows_NT" ]]; then
  loom_bin="$output_dir/loom-sample-cli.exe"
fi
go build -o "$loom_bin" "$loom_root/cmd/loom"

write_stdout_report() {
  local output="$1"
  shift
  if [[ -e "$output" && "$allow_overwrite" != true ]]; then
    echo "loom sample app error: refusing to overwrite existing output $output without --overwrite" >&2
    exit 1
  fi
  local temp
  temp="$(mktemp "$output_dir/.report.XXXXXX")"
  "$@" > "$temp"
  mv "$temp" "$output"
}

write_findings_report() {
  set +e
  "$@"
  local status=$?
  set -e
  if [[ "$status" -ne 0 ]]; then
    echo "loom sample app note: wrote report with findings: $*" >&2
  fi
}

write_stdout_report "$output_dir/manifest-validation.json" "$loom_bin" --quiet config:validate "$sample_root/loom.json" --project-root "$sample_root" --format json
"$loom_bin" --quiet inspect:xaml "$xaml_path" --format json --output "$output_dir/winui-layout.json" $overwrite_flag
"$loom_bin" --quiet inspect:ascii "$xaml_path" --output "$output_dir/winui-layout.txt" $overwrite_flag
"$loom_bin" --quiet accessibility:audit "$xaml_path" --format json --output "$output_dir/winui-audit.json" $overwrite_flag
write_findings_report "$loom_bin" --quiet patterns:transfer "$xaml_path" --from winui3 --to macos --format json --output "$output_dir/winui-to-macos-transfer.json" $overwrite_flag
"$loom_bin" --quiet inspect:swiftui "$swiftui_path" --format json --output "$output_dir/swiftui-layout.json" $overwrite_flag
write_findings_report "$loom_bin" --quiet patterns:transfer "$swiftui_path" --from swiftui --to windows --format json --output "$output_dir/swiftui-to-windows-transfer.json" $overwrite_flag
"$loom_bin" --quiet inspect:qt "$qt_path" --format json --output "$output_dir/qt-layout.json" $overwrite_flag
write_findings_report "$loom_bin" --quiet patterns:transfer "$qt_path" --from qt --to windows --format json --output "$output_dir/qt-to-windows-transfer.json" $overwrite_flag
write_findings_report "$loom_bin" --quiet inspect:parity "$swiftui_path" --target "$xaml_path" --from swiftui --to winui3 --format json --output "$output_dir/swiftui-winui-parity.json" $overwrite_flag
write_findings_report "$loom_bin" --quiet inspect:parity "$swiftui_path" --target "$qt_path" --from swiftui --to qt --format json --output "$output_dir/swiftui-qt-parity.json" $overwrite_flag
write_findings_report "$loom_bin" --quiet inspect:visual-parity "$swiftui_path" --target "$xaml_path" --from swiftui --to winui3 --profile "$sample_root/visual-profile.json" --format json --output "$output_dir/swiftui-winui-visual-parity.json" $overwrite_flag

echo "loom sample reports written to $output_dir"
