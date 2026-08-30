# Changelog

## 0.12.0 — 2026-08-30

- Added `inspect:errors` to report Swift parser diagnostics, Loom extraction
  diagnostics, XAML parse failures, manifest validation issues, and Pattern
  validation/lint issues.
- Added `--kind swift|xaml|manifest|patterns`, `--format text|json`,
  `--output`, and `--fail-on none|error|warning` for error inspection.
- Added SwiftParserDiagnostics integration for syntax errors without shelling
  out to a compiler.
- Added `docs/AI_AGENTS.md` covering safe agent workflows, JSON output,
  write guards, self-healing limits, and translation boundaries.
- Added tests for Swift syntax error inspection, XAML parse errors, Pattern
  errors, fail modes, and command catalog coverage.

## 0.11.0 — 2026-08-30

- Added Vigil-inspired read-only diagnostics commands: `status`, `verify`,
  `checks:command-catalog`, `guards:summary`, and `self-heal:plan`.
- Added text and JSON output for diagnostics commands.
- Added command-catalog validation for duplicate commands, alias collisions,
  incomplete metadata, synopsis drift, and read/write access flag drift.
- Added tests for diagnostics readiness, guard summaries, self-healing plan
  visibility, and JSON report encoding.

## 0.10.0 — 2026-08-30

- Added `generate:contracts` to emit WinUI behavior, binding, resource,
  accessibility, visibility, collection, component, and unsupported-work
  contracts from SwiftUI-normalized IR.
- Added text and JSON contract reports with `--output` support.
- Added command catalog metadata and alias `contracts`.
- Added tests for contract extraction, JSON encoding, and command registry
  coverage.

- Replaced recursive `LoomNode` aggregation helpers with iterative traversals
  to avoid stack growth on deeply nested generated layout trees.
- Precompiled component-graph include/exclude glob expressions once per scan
  instead of rebuilding regexes for every candidate file.
- Reused manifest Swift source contents across multi-component validation and
  project builds to avoid repeated source-file reads.
- Added a deep-layout regression test for node counting and component-reference
  traversal.

## 0.9.0 — 2026-08-30

- Added global `--quiet` / `-q` and `--verbose` / `-v` runtime output controls.
- Added explicit `--init-region` self-healing for missing generated XAML host
  files used with `generate:xaml --replace-region`.
- Kept existing-file writes strict: Loom still refuses to alter XAML without
  explicit ownership markers.
- Added guard coverage for unsafe region ids, malformed marker order, malformed
  XAML, graph cycles, no-op region replacement, path-level region
  initialization, and operational pattern lint failures.
- Corrected command metadata so `--replace-region` is advertised only for
  `generate:xaml`.

## 0.8.0 — 2026-08-30

- Added safe Loom-owned XAML region replacement through
  `generate:xaml --replace-region ... --region-id ...`.
- Added strict marker validation for missing, duplicated, malformed, or
  out-of-order `LOOM-BEGIN` / `LOOM-END` comments.
- Preserved handwritten XAML outside the marked generated region.
- Added tests for successful replacement and refusal of unsafe marker states.

## 0.7.0 — 2026-08-30

- Added `generate:swiftui` to emit reviewable SwiftUI scaffolds from WinUI XAML.
- Added a SwiftUI emitter that maps XAML-normalized IR containers and controls
  into SwiftUI source.
- Preserved unsupported XAML controls as comments and `EmptyView()`
  placeholders.
- Added tests for supported XAML-to-SwiftUI generation and unsupported-control
  preservation.

## 0.6.0 — 2026-08-30

- Added `inspect:xaml` to parse WinUI XAML into Loom's platform-neutral
  `LoomNode` layout tree.
- Added XAML mappings for common WinUI controls including `Grid`, `StackPanel`,
  `TextBlock`, `Button`, `TextBox`, `Image`, `ScrollViewer`, and `ListView`.
- Preserved original XAML attributes as `xaml.*` node properties for future
  reverse-generation work.
- Added tests proving hand-written and Loom-generated XAML normalize into
  comparable IR.

## 0.5.0 — 2026-08-30

- Added an operational Pattern registry for looking up semantic mappings by
  `LoomNodeKind`.
- Added `patterns:lint` to enforce bidirectional mapping quality beyond basic
  schema validity.
- Added `generate:xaml --patterns-dir ... --pattern-comments` to trace emitted
  XAML nodes back to OS-agnostic Pattern ids and WinUI constructs.
- Kept default XAML output stable unless pattern comments are explicitly
  requested.
- Documented the direction-neutral mapping model needed for future
  Windows-to-macOS translation.

## 0.4.0 — 2026-08-30

- Added `graph:components` to discover reachable SwiftUI layout components from
  a Swift file or source directory.
- Resolved same-view computed subview references and cross-file custom `View`
  structs without manifest enumeration.
- Added text, JSON, and DOT graph output.
- Added cycle and unresolved custom-view diagnostics.
- Added include/exclude globs for deterministic source discovery in larger
  applications.

## 0.3.0 — 2026-08-30

- Added a normative `Patterns` catalog with one OS-agnostic semantic definition
  for every meaningful layout and control kind Loom recognizes.
- Added precise intent, structure, sizing, typed attribute, range, constraint,
  accessibility, and platform-mapping metadata.
- Added `patterns:list`, `patterns:show`, and `patterns:validate` commands.
- Added typed pattern APIs, cross-file identity and kind validation, and full
  semantic-kind coverage tests.

## 0.2.0 — 2026-08-30

- Added a Vigil-inspired command registry with colon-namespaced canonical
  commands, grouped categories, access metadata, aliases, JSON catalogs, and
  command manuals.
- Added `config:schema` and read-only `config:validate` manifest tooling.
- Added `project:build` to generate all declared analyses, XAML fragments,
  parity evidence, and a deterministic project summary in one run.
- Preserved `analyze`, `generate`, `parity`, and `project` as compatibility
  aliases.
- Upgraded the Voci integration to the manifest-driven project workflow.

## 0.1.0 — 2026-08-29

- Introduced the SwiftParser/SwiftSyntax frontend and SwiftUI view discovery.
- Added the platform-neutral layout model and text/JSON analysis reports.
- Added conservative WinUI 3 XAML generation with configurable theme prefixes.
- Added existing-XAML parity diagnostics.
- Added Voci analysis, generated-shell, and parity examples.
- Added Swift Testing coverage for extraction, generation, bindings, lifecycle
  modifiers, and error reporting.
