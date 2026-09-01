# Loom Roadmap

Loom's `v1.0.0` target is a complete analyzer, generator, and translator for
moving UI layout intent between SwiftUI, WinUI XAML, and Qt.

The current `0.24.0` release is the analyzer, transfer-planning,
component-graph, and analysis-only project build foundation. Main now adds
reviewable generator scaffolds, target contracts, and project-build generator
artifacts toward `v1.0.0`.

## Completed Foundation

### 0.17 - Go Runtime Foundation

- Forward Go CLI introduced under `cmd/loom`.
- Command catalog, status/verification diagnostics, pattern operations, and
  command guard outputs.
- `inspect:xaml`, `inspect:ascii`, `inspect:errors`, `accessibility:audit`,
  `patterns:*`, OS error suggestions, and read/write output guards.

### 0.18 - Analyzer Baseline

- Go runtime became the active implementation.
- WinUI XAML analysis and diagnostics moved into Go ownership.
- Unsupported native WinUI boundaries became explicit diagnostics instead of
  silent flattening.
- Placeholder generator/project commands were retained for catalog continuity.

### 0.19 - Deterministic Output

- Repository line-ending policy added.
- `--line-ending lf|crlf|native` added for deterministic text artifacts.
- JSON output normalized for function and automation callers.

### 0.20 - Public Repository Shape

- Repository folders normalized for public source.
- Neutral sample app fixtures added.
- Default pattern discovery aligned with repository and installed locations.

### 0.21 - SwiftUI Analysis

- SwiftUI layout/control inspection added.
- `inspect:source` auto-detection added for SwiftUI and WinUI XAML.
- SwiftUI-to-Windows transfer planning added.

### 0.22 - Qt, Parity, And Visual Material

- Qt QML, Qt Designer UI, and common Qt C++ layout inspection added.
- SwiftUI, WinUI XAML, and Qt structural parity added.
- Visual parity infrastructure added for typography, spacing, controls,
  provenance, and profile-normalized comparisons.
- Font material inspection added for OpenType, TrueType, TrueType Collection,
  WOFF, and installed family names.

### 0.24 - Component Graph And Project Build

- `graph:components` added for source-tree component and dependency discovery.
- `project:build` added for manifest-directed analysis bundles.
- Sample workflow now emits component graph and project build artifacts.

### Main - Generator Scaffolds And Contracts

- `generate:xaml` emits reviewable WinUI XAML fragments and supports guarded
  owned-region replacement.
- `generate:swiftui` emits reviewable SwiftUI scaffolds from normalized WinUI,
  Qt, or SwiftUI source.
- `generate:contracts` emits behavior, state, collection, component-boundary,
  and policy handoff reports.
- `project:build` now includes generated XAML, generated SwiftUI, and target
  contract artifacts.

## v1.0.0 Release Blockers

- Freeze public JSON schemas for analyzer, generator, translator, audit, parity,
  and diagnostics output.
- Add cross-platform CI and release artifacts for macOS, Linux, and Windows.
- Add release evidence for generated artifact reviewability, overwrite guards,
  malformed input, large input, and platform path behavior.

## Near-Term Priorities

- Strengthen generator data structures without changing analyzer output.
- Deepen Qt parsing beyond conservative QML/UI/C++ layout heuristics.
- Improve malformed layout diagnostics for nested scroll and geometry-dependent
  constructs.
- Extend suggestions to cover additional WinUI, XAML, SwiftUI, Qt, macOS,
  Linux, and Windows parity failure classes.
- Add explicit malformed-input and memory/guard regression tests.
