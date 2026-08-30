#!/usr/bin/env bash
set -euo pipefail

voci_root="${1:-/private/SDF/Voci}"
loom_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
swift_source="$voci_root/platform/macos/VociMacApp.swift"
xaml_source="$voci_root/platform/windows-winui/VociWindows/MainWindow.xaml"

swift run --package-path "$loom_root" loom analyze "$swift_source" --root-view ContentView
swift run --package-path "$loom_root" loom generate "$swift_source" --root-view ContentView --theme-prefix Voci --output "$loom_root/Examples/Voci/ContentView.generated.xaml"
swift run --package-path "$loom_root" loom parity "$swift_source" --root-view ContentView --xaml "$xaml_source"
