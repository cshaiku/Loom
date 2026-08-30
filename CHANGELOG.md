# Changelog

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
