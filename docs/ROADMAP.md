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
- Recursive component graph discovery for computed subviews and custom `View`
  structs across Swift source directories.
- Pattern registry, operational linting, and opt-in XAML mapping trace comments.
- WinUI XAML ingestion into the shared Loom IR for the future Windows-to-macOS
  path.
- SwiftUI scaffold generation from XAML-normalized IR.
- Safe Loom-owned XAML region replacement that refuses to alter handwritten
  content outside explicit markers.
- Reliability guards for malformed XAML, graph cycles, unsafe owned-region ids,
  explicit generated-region initialization, and automation-friendly CLI output.

## 0.4 — Recursive component graph

Status: shipped in 0.4.0.

- `graph:components` discovers computed SwiftUI subviews without requiring
  every property in the manifest.
- It resolves custom `View` structs across multiple Swift files.
- It emits text, JSON, and DOT component dependency graphs.
- It reports cycles and unresolved custom-view references.
- It supports include/exclude globs so large applications can constrain source
  discovery deterministically.

## 0.5 — Pattern-driven mapping registry

Status: initial mapping registry shipped in 0.5.0.

- `patterns:lint` enforces operational mapping quality beyond structural
  validity.
- `generate:xaml --patterns-dir Patterns --pattern-comments` traces emitted
  nodes back to OS-agnostic Pattern ids and WinUI constructs.
- Pattern mappings remain direction-neutral so future XAML ingestion can target
  the same IR used by SwiftUI extraction.

## 0.6 — XAML ingestion

Status: initial WinUI ingestion shipped in 0.6.0.

- `inspect:xaml` normalizes WinUI controls into `LoomNode` trees.
- Original XAML attributes are retained as `xaml.*` properties.
- Generated XAML can be ingested back into comparable IR for supported controls.

## 0.7 — SwiftUI scaffold generation

Status: initial XAML-to-SwiftUI generation shipped in 0.7.0.

- `generate:swiftui` emits reviewable SwiftUI scaffolds from WinUI XAML.
- Supported layout and control nodes map through Loom IR.
- Unsupported XAML remains explicit as comments and placeholders.

## 0.8 — Owned XAML regions

Status: owned-region replacement shipped in 0.8.0.

- `generate:xaml --replace-region <xaml-file> --region-id <id>` updates only a
  marked Loom region.
- Missing, duplicated, malformed, or out-of-order markers fail before writing.
- Existing handwritten XAML outside the region remains byte-preserved.

## 0.9 — Reliability guards

Status: shipped in 0.9.0.

- Added global `--quiet` and `--verbose` output controls.
- Added explicit `--init-region` creation for missing generated XAML host files.
- Preserved strict refusal for existing files without ownership markers.
- Added tests for malformed XAML, graph cycles, unsafe region ids, no-op region
  updates, path-level region initialization, and operational pattern lint gaps.

## 1.0 — Target contracts

- Drive semantic color, control, event, and binding mappings from Patterns.
- Generate C++/WinRT event and view-model contract stubs.

Acceptance: rerunning Loom updates a generated Voci region idempotently while
leaving native Windows behavior and unrelated XAML byte-for-byte unchanged.

## 1.1 — Windows proof adapter

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
