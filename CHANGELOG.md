# Changelog

## 0.24.0 - 2026-09-01

- Implemented `graph:components` for component and dependency discovery across
  SwiftUI, WinUI XAML, and Qt source trees, with text, JSON, and DOT output.
- Implemented `project:build` as an analysis-only manifest workflow that emits
  validation, source analysis, component graph, transfer, parity, and summary
  artifacts under a guarded output directory.
- Updated the sample app workflow to produce a component graph and full project
  build bundle.
- Added regression coverage for component graph discovery and project build
  artifact creation.
- Kept `generate:xaml`, `generate:swiftui`, and `generate:contracts` as the
  remaining v1.0 generator blockers.

## 0.23.0 - 2026-09-01

- Clarified the `v1.0.0` product goal as analyzer, generator, and translator.
- Added public contribution, security, conduct, support, testing, release,
  support-policy, deprecation, notes, and todo documentation.
- Added GitHub CI, Dependabot, issue templates, and pull request template.
- Added a Vigil-style repository policy configuration.
- Expanded the sample app with a manifest, visual profile, README, repeatable
  report workflow, and generated analysis/parity/transfer output targets.
- Changed Loom's repository license from MIT to 0BSD.
- Prepared public GitHub and `sltd.ca` release positioning for the pre-1.0
  analyzer and transfer-planning line.

## 0.22.0 - 2026-08-31

- Added initial `inspect:qt` support for Qt QML, Qt Designer `.ui`, and common
  Qt C++ layout/control constructs.
- Added Qt platform mappings to the `patterns` catalog for Linux/Qt transfer
  planning.
- Implemented `inspect:parity` for structural layout comparison across SwiftUI,
  WinUI XAML, and Qt sources.
- Added `inspect:visual-parity` infrastructure for profile-normalized visual
  regression metrics across SwiftUI, WinUI XAML, and Qt sources.
- Added `inspect:font` to extract intrinsic font material properties from
  supplied OpenType/TrueType fonts or installed family names.
- Extended `inspect:visual-parity` with source/target font material overrides.
- Added visual parity provenance, per-finding confidence, and strict visual
  profile validation.
- Extended visual parity source extraction for SwiftUI visual modifiers and XAML
  visual attributes/resource references.
- Added local XAML resource and implicit style resolution for visual parity
  provenance.
- Extended XAML visual parity trust to load local merged resource dictionaries
  and warn when referenced dictionaries cannot be resolved.
- Added explicit XAML style and simple BasedOn style-chain resolution for visual
  parity provenance.
- Added object-valued XAML visual resource/setter extraction, resource alias
  chain resolution, and unresolved explicit style diagnostics.
- Extended `inspect:source` auto-detection to `.qml`, `.ui`, and common Qt C++
  file extensions.
- Added neutral Qt sample input and sample outputs for Qt transfer/parity.
- Added regression coverage for Qt inspection, Qt diagnostics, Qt transfer, and
  SwiftUI ↔ Qt parity.

## 0.21.0 - 2026-08-30

- Added initial Go-native `inspect:swiftui` support for common SwiftUI layout
  and control constructs.
- Implemented `inspect:source` as an auto-detecting source inspector for
  SwiftUI and WinUI XAML inputs.
- Added macOS/SwiftUI → Windows/WinUI transfer planning through the existing
  OS-agnostic `patterns` vocabulary.
- Added SwiftUI delimiter diagnostics to `inspect:errors --kind swift`.
- Added regression coverage for SwiftUI inspection, auto-detection, parse
  diagnostics, and Mac → Windows transfer mappings.

## 0.20.0 - 2026-08-30

- Normalized public repository folders to lowercase: `patterns` and
  `examples/sampleapp`.
- Replaced private sample app references with neutral public sample app
  terminology and a checked-in `examples/sampleapp/mainwindow.xaml` fixture.
- Updated default pattern discovery so loom uses its own `patterns` catalog from
  the repository or from `/opt/homebrew/share/loom/patterns` after install.
- Updated `make build` to install both `/opt/homebrew/bin/loom` and loom's
  shared pattern catalog.
- Lowercased public-facing command help, diagnostics, and documentation labels
  where practical while preserving Go exported type names and stable JSON fields.

## 0.19.0 - 2026-08-30

- Added repository line-ending policy through `.gitattributes`.
- Added global `--line-ending lf|crlf|native` handling for deterministic LF,
  Windows CRLF, and host-native text output.
- Kept LF as the default for stable cross-platform reports and CI artifacts.
- Improved function-facing JSON output by preserving readable `<placeholder>`
  synopsis text and emitting stable empty arrays instead of `null`.
- Updated README and AI-agent docs with line-ending guidance.

## 0.18.0 - 2026-08-30

- Completed the Go port as the active runtime by removing SwiftPM manifest and
  Swift runtime source/test files.
- Added build-time installation behavior (`make build`) that installs
  `loom` into `/opt/homebrew/bin` for shared local use.
- Added robust unsupported-native-boundary warnings and transfer classifications
  for WinUI-only controls through the Go diagnostics pipeline.
- Added malformed/redundant layout risk diagnostics and audit surfaced fixes for
  both users and agents.
- Added catalog-reserved placeholders for generation/parity/project commands with
  explicit guarded unsupported-command output.
- Added `inspect:errors` + `audit` + `suggestions:os-errors` integration
  improvements for WinUI/SwiftUI boundary and design-risk guidance.

## 0.17.0 - 2026-08-30

- Started the Go migration so loom can ship as a single cross-platform CLI for
  macOS, Windows, and Linux.
- Added `cmd/loom` and `internal/loom` as the forward Go runtime.
- Ported pattern catalog operations, WinUI XAML ingestion, ASCII pattern
  rendering, accessibility/layout audit, pattern transfer, OS error
  suggestions, command listing/help, `status`, and `verify`.
- Preserved the Swift implementation as the 0.16.0 reference while SwiftUI
  parsing, generation, project workflows, owned-region guards, and parity
  commands are ported.
- Added Go regression tests for pattern validation, native WinUI component
  boundaries, structured suggested fixes, transfer classification, OS error
  matching, CLI JSON output, version output, and `--quiet` write behavior.

## 0.16.0 - 2026-08-30

- Added explicit `XAML.UNSUPPORTED_COMPONENT_BOUNDARY` warnings for native
  WinUI controls without loom semantic mappings.
- Preserved unsupported native WinUI controls as component nodes with
  `componentBoundary`, `unsupportedXamlElement`, and
  `requiresNativeImplementation` metadata.
- Surfaced native WinUI component boundaries in accessibility audit and pattern
  transfer reports.
- Added structured `suggested_fixes` to accessibility audit findings for user
  decisions and AI-agent implementation actions.
- Added a curated OS/framework error suggestion catalog for SwiftUI, WinUI,
  XAML, macOS, and Windows issues.
- Added `suggestions:os-errors` and automatic `suggested_fixes` enrichment for
  `inspect:errors` findings.
- Added `docs/os-error-suggestions.md`.
- Added tests for native WinUI component-boundary diagnostics, audit/transfer
  surfacing, OS error suggestions, and enriched error inspection findings.

## 0.15.0 - 2026-08-30

- Added `accessibility:audit` for SwiftUI and WinUI XAML sources.
- Added audit rules for missing accessible names, unlabeled images, unlabeled
  text inputs, placeholder-only input naming, undersized interactive targets,
  color-only semantic risks, unsupported/malformed nodes, empty containers,
  redundant wrappers, repeated nested layout wrappers, nested scroll regions,
  and geometry-dependent layout.
- Added text/JSON audit output plus `--output` and
  `--fail-on none|error|warning`.
- Extended optional pattern accessibility metadata with keyboard behavior,
  states, required accessibility properties, and minimum target size.
- Updated README, command docs, architecture notes, pattern docs, AI agent
  guidance, roadmap, and VERSION.
- Added tests for accessibility/layout/design audit findings and JSON report
  round trips.

## 0.14.0 - 2026-08-30

- Added `patterns:transfer` to classify interface layout transfer readiness
  between SwiftUI and WinUI.
- Added transfer dispositions for direct transfer, policy-dependent transfer,
  native-contract-dependent transfer, lossy transfer, and unsupported nodes.
- Added `inspect:ascii` for plaintext ASCII pattern layout trees.
- Added optional pattern variants to the pattern model and schema for adaptive,
  dense, compact, or accessibility layout policies.
- Updated README, command docs, architecture notes, pattern docs, AI agent
  guidance, roadmap, and VERSION.
- Added tests for transfer classification, ASCII rendering, and JSON report
  round trips.

## 0.13.0 - 2026-08-30

- Added `patterns:export` to derive integration-oriented JSON from the
  canonical pattern catalog.
- Added export formats for loom-native JSON, DTCG-style token objects, Open
  UI-style component metadata, ARIA accessibility summaries, and Style
  Dictionary-compatible token packages.
- Added `--format`, `--directory`, and `--output` handling for pattern exports.
- Added tests for external export shapes and command catalog visibility.

## 0.12.0 - 2026-08-30

- Added `inspect:errors` to report Swift parser diagnostics, loom extraction
  diagnostics, XAML parse failures, manifest validation issues, and pattern
  validation/lint issues.
- Added `--kind swift|xaml|manifest|patterns`, `--format text|json`,
  `--output`, and `--fail-on none|error|warning` for error inspection.
- Added SwiftParserDiagnostics integration for syntax errors without shelling
  out to a compiler.
- Added `docs/ai-agents.md` covering safe agent workflows, JSON output,
  write guards, self-healing limits, and translation boundaries.
- Added tests for Swift syntax error inspection, XAML parse errors, pattern
  errors, fail modes, and command catalog coverage.

## 0.11.0 - 2026-08-30

- Added Vigil-inspired read-only diagnostics commands: `status`, `verify`,
  `checks:command-catalog`, `guards:summary`, and `self-heal:plan`.
- Added text and JSON output for diagnostics commands.
- Added command-catalog validation for duplicate commands, alias collisions,
  incomplete metadata, synopsis drift, and read/write access flag drift.
- Added tests for diagnostics readiness, guard summaries, self-healing plan
  visibility, and JSON report encoding.

## 0.10.0 - 2026-08-30

- Added `generate:contracts` to emit WinUI behavior, binding, resource,
  accessibility, visibility, collection, component, and unsupported-work
  contracts from SwiftUI-normalized IR.
- Added text and JSON contract reports with `--output` support.
- Added command catalog metadata and alias `contracts`.
- Added tests for contract extraction, JSON encoding, and command registry
  coverage.

- Replaced recursive `loom node` aggregation helpers with iterative traversals
  to avoid stack growth on deeply nested generated layout trees.
- Precompiled component-graph include/exclude glob expressions once per scan
  instead of rebuilding regexes for every candidate file.
- Reused manifest Swift source contents across multi-component validation and
  project builds to avoid repeated source-file reads.
- Added a deep-layout regression test for node counting and component-reference
  traversal.

## 0.9.0 - 2026-08-30

- Added global `--quiet` / `-q` and `--verbose` / `-v` runtime output controls.
- Added explicit `--init-region` self-healing for missing generated XAML host
  files used with `generate:xaml --replace-region`.
- Kept existing-file writes strict: loom still refuses to alter XAML without
  explicit ownership markers.
- Added guard coverage for unsafe region ids, malformed marker order, malformed
  XAML, graph cycles, no-op region replacement, path-level region
  initialization, and operational pattern lint failures.
- Corrected command metadata so `--replace-region` is advertised only for
  `generate:xaml`.

## 0.8.0 - 2026-08-30

- Added safe loom-owned XAML region replacement through
  `generate:xaml --replace-region ... --region-id ...`.
- Added strict marker validation for missing, duplicated, malformed, or
  out-of-order `LOOM-BEGIN` / `LOOM-END` comments.
- Preserved handwritten XAML outside the marked generated region.
- Added tests for successful replacement and refusal of unsafe marker states.

## 0.7.0 - 2026-08-30

- Added `generate:swiftui` to emit reviewable SwiftUI scaffolds from WinUI XAML.
- Added a SwiftUI emitter that maps XAML-normalized IR containers and controls
  into SwiftUI source.
- Preserved unsupported XAML controls as comments and `EmptyView()`
  placeholders.
- Added tests for supported XAML-to-SwiftUI generation and unsupported-control
  preservation.

## 0.6.0 - 2026-08-30

- Added `inspect:xaml` to parse WinUI XAML into loom's platform-neutral
  `loom node` layout tree.
- Added XAML mappings for common WinUI controls including `Grid`, `StackPanel`,
  `TextBlock`, `Button`, `TextBox`, `Image`, `ScrollViewer`, and `ListView`.
- Preserved original XAML attributes as `xaml.*` node properties for future
  reverse-generation work.
- Added tests proving hand-written and loom-generated XAML normalize into
  comparable IR.

## 0.5.0 - 2026-08-30

- Added an operational pattern registry for looking up semantic mappings by
  `loom node kind`.
- Added `patterns:lint` to enforce bidirectional mapping quality beyond basic
  schema validity.
- Added `generate:xaml --patterns-dir ... --pattern-comments` to trace emitted
  XAML nodes back to OS-agnostic pattern ids and WinUI constructs.
- Kept default XAML output stable unless pattern comments are explicitly
  requested.
- Documented the direction-neutral mapping model needed for future
  Windows-to-macOS translation.

## 0.4.0 - 2026-08-30

- Added `graph:components` to discover reachable SwiftUI layout components from
  a Swift file or source directory.
- Resolved same-view computed subview references and cross-file custom `View`
  structs without manifest enumeration.
- Added text, JSON, and DOT graph output.
- Added cycle and unresolved custom-view diagnostics.
- Added include/exclude globs for deterministic source discovery in larger
  applications.

## 0.3.0 - 2026-08-30

- Added a normative `patterns` catalog with one OS-agnostic semantic definition
  for every meaningful layout and control kind loom recognizes.
- Added precise intent, structure, sizing, typed attribute, range, constraint,
  accessibility, and platform-mapping metadata.
- Added `patterns:list`, `patterns:show`, and `patterns:validate` commands.
- Added typed pattern APIs, cross-file identity and kind validation, and full
  semantic-kind coverage tests.

## 0.2.0 - 2026-08-30

- Added a Vigil-inspired command registry with colon-namespaced canonical
  commands, grouped categories, access metadata, aliases, JSON catalogs, and
  command manuals.
- Added `config:schema` and read-only `config:validate` manifest tooling.
- Added `project:build` to generate all declared analyses, XAML fragments,
  parity evidence, and a deterministic project summary in one run.
- Preserved `analyze`, `generate`, `parity`, and `project` as compatibility
  aliases.
- Upgraded the sample app integration to the manifest-driven project workflow.

## 0.1.0 - 2026-08-29

- Introduced the SwiftParser/SwiftSyntax frontend and SwiftUI view discovery.
- Added the platform-neutral layout model and text/JSON analysis reports.
- Added conservative WinUI 3 XAML generation with configurable theme prefixes.
- Added existing-XAML parity diagnostics.
- Added sample app analysis, generated-shell, and parity examples.
- Added Swift Testing coverage for extraction, generation, bindings, lifecycle
  modifiers, and error reporting.
