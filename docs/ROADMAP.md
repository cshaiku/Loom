# Loom roadmap

## Completed foundation

- SwiftParser and SwiftSyntax source frontend.
- Platform-neutral, Codable layout tree.
- WinUI 3 XAML fragment emitter and parity diagnostics.
- Versioned project manifests and multi-component project builds.
- Vigil-style grouped command registry, manuals, JSON catalog, and access
  metadata.
- Voci project profile with six translated shell regions.
- Versioned semantic Patterns catalog covering every meaningful layout node.

## 0.4 — Recursive component graph

- Discover computed SwiftUI subviews without requiring every property in the
  manifest.
- Resolve custom `View` structs across multiple Swift files.
- Emit a component dependency graph with cycle and unresolved-reference
  diagnostics.
- Add project include/exclude rules so large applications can constrain source
  discovery deterministically.

Acceptance: Voci's root shell graph resolves its local computed subviews and
dialog `View` types across the macOS source set without duplicate work.

## 0.5 — Pattern-driven mapping registry and owned output

- Drive semantic color, control, event, and binding mappings from Patterns.
- Generate C++/WinRT event and view-model contract stubs.
- Introduce syntax-aware owned XAML regions or generated UserControls.
- Refuse to overwrite handwritten XAML nodes outside Loom-owned output.

Acceptance: rerunning Loom updates a generated Voci region idempotently while
leaving native Windows behavior and unrelated XAML byte-for-byte unchanged.

## 0.6 — Windows proof adapter

- Compile generated fragments on a Windows App SDK host.
- Capture deterministic reference-size renders and accessibility trees.
- Compare measured geometry, theme states, and automation identifiers.
- Produce machine-readable parity evidence suitable for CI.

Acceptance: a Windows-hosted check proves generated XAML compiles and reports
visual/accessibility drift against Voci's declared reference states.

## Later targets

- Additional XAML dialect policies where semantics differ from WinUI 3.
- Pluggable emitters for other declarative desktop frameworks.
- Language-server or editor integration driven by the stable JSON contracts.
