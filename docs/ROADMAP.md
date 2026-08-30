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
- WinUI target contract generation for native behavior, bindings, resources,
  accessibility, and component wiring that layout emission cannot own.
- Vigil-inspired diagnostics commands for readiness, verification, command
  catalog checks, write guards, and self-healing visibility.
- Error-focused inspection for Swift syntax, Loom analysis, XAML parse,
  manifest, and Pattern issues.
- AI-agent operating guidance for safe sequencing, JSON output, write guards,
  and translation boundaries.
- Pattern export for Loom-native JSON, DTCG-style tokens, Open UI-style
  component metadata, ARIA summaries, and Style Dictionary-compatible token
  packages.
- Pattern transfer planning for SwiftUI-to-WinUI and WinUI-to-SwiftUI layout
  movement.
- ASCII Pattern rendering for plaintext layout sketches in logs, reviews, and
  agent workflows.
- Optional Pattern variants for adaptive layout policy metadata.
- Accessibility/layout auditing for missing names, unlabeled media/inputs,
  undersized controls, redundant wrappers, malformed nodes, nested scroll
  regions, and weak layout-transfer design.
- Optional richer Pattern accessibility metadata for keyboard behavior, states,
  required properties, and minimum target size.
- Unsupported native WinUI controls are preserved as explicit component
  boundaries and surfaced in audit/transfer reports.
- Audit findings include structured suggested fixes for user decisions and AI
  agent implementation actions.

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

## 0.10 — Target contracts

Status: shipped in 0.10.0.

- Added `generate:contracts` for the native WinUI implementation surface beside
  generated layout.
- Reports actions, bindings, lifecycle hooks, visibility rules, collection
  templates, accessibility metadata, theme resources, component boundaries, and
  unsupported expressions.
- Supports text and JSON output plus `--output` for automation.

## 0.11 — Vigil-style diagnostics

Status: shipped in 0.11.0.

- Added `status` for local Loom readiness and Pattern health.
- Added `verify` for read-only command catalog and Pattern validation/lint
  checks.
- Added `checks:command-catalog` to audit command metadata, aliases, synopsis,
  and access flags.
- Added `guards:summary` to show commands capable of writing and the flags that
  authorize writes.
- Added `self-heal:plan` to document explicit self-healing actions and their
  guardrails.

## 0.12 — Error inspection and AI agent guide

Status: shipped in 0.12.0.

- Added `inspect:errors` for Swift syntax diagnostics, Loom extraction
  diagnostics, XAML parse failures, manifest validation, and Pattern
  validation/lint issues.
- Added `--kind swift|xaml|manifest|patterns`, `--format text|json`,
  `--output`, and `--fail-on none|error|warning`.
- Added `docs/AI_AGENTS.md` with safe command sequencing, JSON-first
  automation guidance, write-guard expectations, and translation boundaries.

## 0.13 — Pattern export interop

Status: shipped in 0.13.0.

- Added `patterns:export` for integration-oriented Pattern catalog exports.
- Supports `loom`, `dtcg`, `open-ui`, `aria`, and `style-dictionary` formats.
- Keeps `.pattern.json` as the canonical source while exposing derived shapes
  for design-token pipelines, component inventories, accessibility review, and
  design-system tooling.

## 0.14 — Pattern transfer and ASCII Patterns

Status: shipped in 0.14.0.

- Added `patterns:transfer` to classify each normalized layout element for
  OS-to-OS interface transfer.
- Added transfer dispositions: `direct`, `needs-policy`,
  `needs-native-contract`, `lossy`, and `unsupported`.
- Added `inspect:ascii` to render SwiftUI or WinUI layouts as plaintext ASCII
  Pattern trees.
- Added optional Pattern variants in the schema/model for compact, dense,
  accessibility, and adaptive layout policy metadata.

## 0.15 — Accessibility and layout audit

Status: shipped in 0.15.0.

- Added `accessibility:audit` for SwiftUI and WinUI XAML inputs.
- Reports missing accessible names, unlabeled images and inputs, small
  interactive targets, color-only semantic risks, unsupported/malformed nodes,
  empty or redundant containers, nested scroll regions, and geometry-dependent
  layouts.
- Supports text/JSON output, `--output`, and `--fail-on none|error|warning`.
- Extended optional Pattern accessibility metadata for keyboard behavior,
  states, required properties, and minimum target size.

## 0.16 — Native boundary warnings and suggested fixes

Status: shipped in 0.16.0.

- Added explicit `XAML.UNSUPPORTED_COMPONENT_BOUNDARY` warnings for native
  WinUI controls without Loom semantic mappings.
- Preserved those controls as component nodes with native-boundary metadata.
- Surfaced native WinUI component boundaries in accessibility audit and Pattern
  transfer reports.
- Added structured `suggested_fixes` to accessibility audit findings for user
  decisions and AI agent implementation actions.

## 1.0 — Pattern-driven emitters

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
